// Package weave provides lightweight concurrency primitives for structured task orchestration.
// It focuses on predictable cancellation, panic recovery, and deterministic error propagation.
package weave

import (
	"context"
	"fmt"
	"sync"
)

// Task represents a unit of concurrent work that accepts a context
// for cooperative cancellation. A non-nil error returned by a Task
// will cause the orchestration to terminate early.
type Task func(ctx context.Context) error

// Sail runs a set of tasks concurrently, respecting context cancellation
// and ensuring deterministic error propagation.
//
// Sail guarantees the following:
//   - Each task is executed in its own goroutine.
//   - If any task returns a non-nil error or panics, Sail returns that error immediately.
//   - If the provided context is canceled, Sail stops scheduling new tasks
//     and returns ctx.Err().
//   - All panics are safely recovered and returned as formatted errors.
//
// The function blocks until all tasks have completed, an error occurs, or the context is canceled.
func Sail(ctx context.Context, tasks ...Task) error {
	var wg sync.WaitGroup
	wg.Add(len(tasks))

	errChan := make(chan error, 1)
	var once sync.Once

	sendErr := func(err error) {
		once.Do(func() {
			errChan <- err
		})
	}

	for _, task := range tasks {
		// Skip task if context is already canceled.
		if ctx.Err() != nil {
			wg.Done()
			continue
		}

		go func(t Task) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					sendErr(fmt.Errorf("panic recovered: %v", r))
				}
			}()

			if err := t(ctx); err != nil {
				sendErr(err)
			}
		}(task)
	}

	// Close errChan once all tasks have completed.
	go func() {
		wg.Wait()
		close(errChan)
	}()

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}
