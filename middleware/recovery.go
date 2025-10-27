package middleware

import (
	"log"
	"net/http"
	"runtime/debug"
)

// Recovery returns an HTTP middleware that recovers from panics
// in downstream handlers and logs the error details.
//
// This middleware prevents the entire server from crashing due to
// unexpected panics. When a panic occurs, it logs the error and
// full stack trace using the provided *log.Logger, then returns a
// safe 500 Internal Server Error response to the client.
//
// Example:
//
//	logger := log.New(os.Stderr, "", log.LstdFlags)
//	r.Use(middleware.Recovery(logger))
//
// Output example:
//
//	PANIC: runtime error: index out of range
//	goroutine 18 [running]:
//	...stack trace...
func Recovery(logger *log.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			// Defer a panic recovery function
			defer func() {
				if err := recover(); err != nil {
					// Log the panic message and full stack trace
					logger.Printf("PANIC: %v\n\n%s", err, debug.Stack())

					// Send a generic 500 response to the client.
					// Safe to call even if headers were partially written.
					http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				}
			}()

			// Continue to the next handler
			next.ServeHTTP(w, r)
		})
	}
}
