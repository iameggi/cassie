package weave

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

//
// ────────────────────────────────────────────────
//   TESTS FOR SAIL()
// ────────────────────────────────────────────────
//

// TestSail_Success verifies that all tasks run successfully in parallel.
func TestSail_Success(t *testing.T) {
	var counter int32
	task := func(ctx context.Context) error {
		atomic.AddInt32(&counter, 1)
		return nil
	}

	err := Sail(context.Background(), task, task, task)
	assert.NoError(t, err)
	assert.Equal(t, int32(3), counter)
}

// TestSail_Error ensures that Sail returns the first error produced by any task.
func TestSail_Error(t *testing.T) {
	expectedErr := errors.New("task failed")

	taskOK := func(ctx context.Context) error { return nil }
	taskFail := func(ctx context.Context) error { return expectedErr }

	err := Sail(context.Background(), taskOK, taskFail, taskOK)
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
}

// TestSail_Panic verifies that panics within tasks are safely recovered.
func TestSail_Panic(t *testing.T) {
	taskOK := func(ctx context.Context) error { return nil }
	taskPanic := func(ctx context.Context) error {
		panic("something went wrong")
	}

	err := Sail(context.Background(), taskOK, taskPanic, taskOK)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "panic recovered: something went wrong")
}

// TestSail_ContextCancel ensures Sail respects external context cancellation.
func TestSail_ContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	task := func(ctx context.Context) error {
		<-ctx.Done()
		return nil
	}
	cancel()

	err := Sail(ctx, task, task)
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}

//
// ────────────────────────────────────────────────
//   TESTS FOR WEAVER
// ────────────────────────────────────────────────
//

// TestWeaver_New_InvalidConcurrency ensures NewWeaver fails when concurrency ≤ 0.
func TestWeaver_New_InvalidConcurrency(t *testing.T) {
	weaver, err := NewWeaver(context.Background(), 0)
	assert.Error(t, err)
	assert.Nil(t, weaver)
	assert.Contains(t, err.Error(), "concurrency must be greater than 0")

	weaver, err = NewWeaver(context.Background(), -1)
	assert.Error(t, err)
	assert.Nil(t, weaver)
}

// TestWeaver_Success verifies that all tasks execute successfully.
func TestWeaver_Success(t *testing.T) {
	weaver, err := NewWeaver(context.Background(), 4)
	assert.NoError(t, err)

	var counter int32
	for i := 0; i < 100; i++ {
		err := weaver.Add(func(ctx context.Context) error {
			atomic.AddInt32(&counter, 1)
			return nil
		})
		assert.NoError(t, err)
	}

	err = weaver.Wait()
	assert.NoError(t, err)
	assert.Equal(t, int32(100), counter)
}

// TestWeaver_Error ensures that Weaver propagates the first encountered error.
func TestWeaver_Error(t *testing.T) {
	weaver, err := NewWeaver(context.Background(), 4)
	assert.NoError(t, err)
	expectedErr := errors.New("task failed")

	for i := 0; i < 10; i++ {
		assert.NoError(t, weaver.Add(func(ctx context.Context) error { return nil }))
	}
	assert.NoError(t, weaver.Add(func(ctx context.Context) error { return expectedErr }))

	err = weaver.Wait()
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
}

// TestWeaver_Panic ensures that task panics are safely recovered.
func TestWeaver_Panic(t *testing.T) {
	weaver, err := NewWeaver(context.Background(), 4)
	assert.NoError(t, err)

	panicTask := func(ctx context.Context) error {
		panic("worker panic")
	}

	assert.NoError(t, weaver.Add(func(ctx context.Context) error { return nil }))
	assert.NoError(t, weaver.Add(panicTask))
	assert.NoError(t, weaver.Add(func(ctx context.Context) error { return nil }))

	err = weaver.Wait()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "panic recovered: worker panic")
}

// TestWeaver_ConcurrencyLimit ensures that concurrent workers do not exceed the configured limit.
func TestWeaver_ConcurrencyLimit(t *testing.T) {
	weaver, err := NewWeaver(context.Background(), 2)
	assert.NoError(t, err)

	var activeWorkers, maxActive int32
	block := make(chan struct{})

	for i := 0; i < 4; i++ {
		assert.NoError(t, weaver.Add(func(ctx context.Context) error {
			current := atomic.AddInt32(&activeWorkers, 1)
			if current > atomic.LoadInt32(&maxActive) {
				atomic.StoreInt32(&maxActive, current)
			}
			<-block
			atomic.AddInt32(&activeWorkers, -1)
			return nil
		}))
	}

	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, int32(2), atomic.LoadInt32(&maxActive))

	close(block)
	err = weaver.Wait()
	assert.NoError(t, err)
}

// TestWeaver_ContextCancel ensures that Weaver stops executing tasks after context cancellation.
func TestWeaver_ContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	weaver, err := NewWeaver(ctx, 4)
	assert.NoError(t, err)

	var started int32
	task := func(ctx context.Context) error {
		atomic.AddInt32(&started, 1)
		<-ctx.Done()
		return nil
	}

	go func() {
		for i := 0; i < 10; i++ {
			if err := weaver.Add(task); err != nil {
				return
			}
		}
		weaver.Wait()
	}()

	time.Sleep(50 * time.Millisecond)
	cancel()

	assert.Less(t, atomic.LoadInt32(&started), int32(10))
	assert.GreaterOrEqual(t, atomic.LoadInt32(&started), int32(4))
}

// TestWeaver_Add_After_Wait verifies that no tasks can be added after Wait() is called.
func TestWeaver_Add_After_Wait(t *testing.T) {
	weaver, err := NewWeaver(context.Background(), 1)
	assert.NoError(t, err)

	err = weaver.Wait()
	assert.NoError(t, err)

	err = weaver.Add(func(ctx context.Context) error { return nil })
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "weaver is closed")
}

// TestWeaver_Wait_Idempotent ensures that Wait() can be safely called multiple times.
func TestWeaver_Wait_Idempotent(t *testing.T) {
	weaver, err := NewWeaver(context.Background(), 2)
	assert.NoError(t, err)

	expectedErr := errors.New("the error")
	assert.NoError(t, weaver.Add(func(ctx context.Context) error { return expectedErr }))

	err1 := weaver.Wait()
	err2 := weaver.Wait()

	assert.Error(t, err1)
	assert.Equal(t, expectedErr, err1)
	assert.Error(t, err2)
	assert.Equal(t, expectedErr, err2)
}
