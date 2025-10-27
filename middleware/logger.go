package middleware

import (
	"net/http"
	"time"

	"github.com/rs/zerolog"
)

// responseWriterInterceptor is a custom wrapper around http.ResponseWriter.
// It intercepts and records the status code written by downstream handlers.
type responseWriterInterceptor struct {
	http.ResponseWriter
	statusCode int
}

// newResponseWriterInterceptor creates a new response writer interceptor.
// Defaults to status code 200 (OK), since handlers that never call WriteHeader
// implicitly send a 200 OK response.
func newResponseWriterInterceptor(w http.ResponseWriter) *responseWriterInterceptor {
	return &responseWriterInterceptor{w, http.StatusOK}
}

// WriteHeader captures the response status code before delegating
// the actual header writing to the underlying ResponseWriter.
func (rwi *responseWriterInterceptor) WriteHeader(code int) {
	rwi.statusCode = code
	rwi.ResponseWriter.WriteHeader(code)
}

// Logger returns an HTTP middleware that provides structured access logging.
//
// It leverages zerolog for high-performance, zero-allocation JSON logging.
// Each request log entry includes method, path, HTTP status code, and latency.
//
// Example:
//
//	r.Use(middleware.Logger(log))
//	// Logs: {"level":"info","method":"GET","path":"/api","status":200,"latency_ms":1.23,"message":"Request processed"}
func Logger(logger zerolog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Wrap the original ResponseWriter with our interceptor
			interceptor := newResponseWriterInterceptor(w)

			// Execute the next handler with the wrapped writer
			next.ServeHTTP(interceptor, r)

			// Measure request latency
			latency := time.Since(start)

			// Log structured request metadata
			logger.Info().
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Int("status", interceptor.statusCode).
				Dur("latency_ms", latency).
				Msg("Request processed")
		})
	}
}
