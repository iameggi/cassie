package middleware

import (
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRecovery_NoPanic(t *testing.T) {
	logger := log.New(io.Discard, "", 0)

	recoveryMiddleware := Recovery(logger)

	mockHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Success"))
	})

	handlerToTest := recoveryMiddleware(mockHandler)

	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()

	handlerToTest.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code, "Status code should be 200 OK")
	assert.Equal(t, "Success", rr.Body.String(), "Response body should be 'Success'")
}

func TestRecovery_WithPanic(t *testing.T) {
	logger := log.New(io.Discard, "", 0)

	recoveryMiddleware := Recovery(logger)

	mockHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("Something went horribly wrong!")
	})

	handlerToTest := recoveryMiddleware(mockHandler)

	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()

	assert.NotPanics(t, func() {
		handlerToTest.ServeHTTP(rr, req)
	}, "Middleware should recover from panic instead of propagating it")

	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Status code should be 500 Internal Server Error")
	assert.Equal(t, http.StatusText(http.StatusInternalServerError)+"\n", rr.Body.String(), "Response body should match the default 500 error text")
}
