package bucket

import "sync"

// Pool is a type-safe wrapper around sync.Pool.
// It ensures objects are properly reset before being reused,
// preventing state leakage and reducing garbage collector pressure.
type Pool[T any] struct {
	pool  *sync.Pool
	reset func(*T) // Reset function called before returning an object to the pool.
}

// New creates a new type-safe Pool for the given type T.
//
// The newFunc parameter constructs a new instance when the pool is empty.
// The resetFunc parameter is required and is automatically called on every
// object before it is put back into the pool.
//
// Panics if resetFunc is nil.
func New[T any](newFunc func() *T, resetFunc func(*T)) *Pool[T] {
	if resetFunc == nil {
		panic("bucket.New: resetFunc must not be nil")
	}

	return &Pool[T]{
		pool: &sync.Pool{
			New: func() any {
				return newFunc()
			},
		},
		reset: resetFunc,
	}
}

// --- Pattern 1: Manual Get/Put ---

// Get retrieves an object from the pool.
// The caller is responsible for returning it to the pool via Put().
// Typically used with `defer p.Put(obj)` for safety.
func (p *Pool[T]) Get() *T {
	return p.pool.Get().(*T)
}

// Put returns the given object to the pool after calling its reset function.
// Nil objects are ignored.
func (p *Pool[T]) Put(obj *T) {
	if obj == nil {
		return
	}
	p.reset(obj)
	p.pool.Put(obj)
}

// --- Pattern 2: Automatic Callback (Safe) ---

// With retrieves an object from the pool, passes it to the given function f,
// and automatically returns it to the pool after f completes.
// This pattern eliminates the risk of forgetting to call Put().
func (p *Pool[T]) With(f func(obj *T)) {
	obj := p.Get()
	defer p.Put(obj)
	f(obj)
}

// WithErr behaves like With, but supports functions that return an error.
// The pooled object is always returned, even if f returns an error.
func (p *Pool[T]) WithErr(f func(obj *T) error) error {
	obj := p.Get()
	defer p.Put(obj)
	return f(obj)
}
