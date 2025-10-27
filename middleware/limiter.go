package middleware

import (
	"net/http"
)

// Limiter is an HTTP middleware that limits the number of concurrent requests
// being processed at any given time.
type Limiter struct {
	// semaphore acts as a concurrency control mechanism.
	// Each slot represents one active request being processed.
	semaphore chan struct{}
}

// NewLimiter creates a new Limiter instance with the specified maximum concurrency.
//
// Panics if maxConcurrency is less than or equal to zero.
func NewLimiter(maxConcurrency int) *Limiter {
	if maxConcurrency <= 0 {
		panic("middleware.NewLimiter: maxConcurrency must be greater than 0")
	}

	return &Limiter{
		// The buffered channel acts as a semaphore with a capacity equal
		// to the allowed number of concurrent requests.
		semaphore: make(chan struct{}, maxConcurrency),
	}
}

// Wrap returns a new http.Handler that enforces the concurrency limit.
//
// When all slots are full, new requests will block until a slot is released.
func (l *Limiter) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Acquire a slot — this will block if the semaphore is full.
		l.semaphore <- struct{}{}

		// Ensure the slot is released even if the handler panics.
		defer func() {
			<-l.semaphore
		}()

		// Continue to the next handler in the chain.
		next.ServeHTTP(w, r)
	})
}
