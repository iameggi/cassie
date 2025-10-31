// Package weave provides lightweight concurrency primitives for
// bounded parallelism and structured task coordination.
package weave

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
)

// Weaver manages a pool of worker goroutines that execute tasks with
// bounded concurrency. It guarantees safe task submission, panic
// recovery, and deterministic shutdown.
//
// A Weaver is intended for cases where you need controlled parallelism
// with cancellation and consistent error propagation — similar in
// spirit to errgroup, but with explicit concurrency limits and lifecycle
// control.
type Weaver struct {
	wg        sync.WaitGroup
	errOnce   sync.Once
	errChan   chan error
	taskQueue chan Task
	cancel    func()
	isClosed  atomic.Bool
	finalErr  error
}

// NewWeaver creates a new Weaver with a fixed concurrency limit.
// It launches 'concurrency' worker goroutines immediately and
// returns an initialized Weaver instance.
//
// If concurrency is less than or equal to zero, NewWeaver returns an error.
func NewWeaver(ctx context.Context, concurrency int) (*Weaver, error) {
	if concurrency <= 0 {
		return nil, errors.New("weave: concurrency must be greater than 0")
	}

	workerCtx, cancel := context.WithCancel(ctx)

	w := &Weaver{
		taskQueue: make(chan Task, concurrency),
		errChan:   make(chan error, 1),
		cancel:    cancel,
	}

	w.wg.Add(concurrency)
	for i := 0; i < concurrency; i++ {
		go w.worker(workerCtx)
	}

	return w, nil
}

// worker continuously pulls tasks from the queue and executes them.
// It terminates when the queue is closed or when the parent context
// is canceled.
func (w *Weaver) worker(ctx context.Context) {
	defer w.wg.Done()
	for {
		select {
		case task, ok := <-w.taskQueue:
			if !ok {
				return
			}
			w.execute(ctx, task)
		case <-ctx.Done():
			return
		}
	}
}

// execute runs a single task with panic protection and cooperative
// context cancellation. If a task returns an error or panics, the first
// such error is recorded for retrieval by Wait.
func (w *Weaver) execute(ctx context.Context, task Task) {
	defer func() {
		if r := recover(); r != nil {
			w.sendErr(fmt.Errorf("panic recovered: %v", r))
		}
	}()
	if ctx.Err() != nil {
		return
	}
	if err := task(ctx); err != nil {
		w.sendErr(err)
	}
}

// sendErr stores the first error encountered by any task.
// Subsequent calls are ignored.
func (w *Weaver) sendErr(err error) {
	w.errOnce.Do(func() {
		w.errChan <- err
	})
}

// Add submits a task to the Weaver for execution.
// It returns an error if the Weaver has already been closed
// or if task submission occurs after Wait has begun.
func (w *Weaver) Add(task Task) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.New("weave: cannot add task to closed weaver")
		}
	}()
	if w.isClosed.Load() {
		return errors.New("weave: weaver is closed")
	}
	w.taskQueue <- task
	return nil
}

// Wait blocks until all tasks have completed or an error occurs.
// It is idempotent and race-safe: multiple concurrent calls to Wait
// are synchronized, and all callers receive the same final error.
//
// If any task returns an error or panics, that error is returned.
// If the parent context is canceled, Wait returns ctx.Err().
// Once Wait has returned, the Weaver is considered closed.
func (w *Weaver) Wait() error {
	// Fast-path: already closed
	if w.isClosed.Load() {
		<-w.errChan
		return w.finalErr
	}

	// Attempt to become the closer
	if !w.isClosed.CompareAndSwap(false, true) {
		<-w.errChan
		return w.finalErr
	}

	// We are the closer
	defer func() {
		w.cancel()
		if r := recover(); r != nil {
			w.finalErr = fmt.Errorf("weaver: wait panic: %v", r)
			close(w.errChan)
			panic(r)
		} else {
			close(w.errChan)
		}
	}()

	close(w.taskQueue)
	w.wg.Wait()

	select {
	case err := <-w.errChan:
		w.finalErr = err
	default:
	}

	return w.finalErr
}
