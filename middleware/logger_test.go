package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestLogger(t *testing.T) {
	logOutput := &bytes.Buffer{}

	logger := zerolog.New(logOutput)

	loggerMiddleware := Logger(logger)

	mockHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte("OK"))
	})

	handlerToTest := loggerMiddleware(mockHandler)

	req := httptest.NewRequest("GET", "/testpath", nil)
	rr := httptest.NewRecorder()

	handlerToTest.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusAccepted, rr.Code, "Status code should be 202")
	assert.Equal(t, "OK", rr.Body.String(), "Response body should be 'OK'")

	logString := logOutput.String()

	assert.True(t, strings.Contains(logString, `"level":"info"`), "Log level should be 'info'")
	assert.True(t, strings.Contains(logString, `"method":"GET"`), "Log should contain HTTP method")
	assert.True(t, strings.Contains(logString, `"path":"/testpath"`), "Log should contain request path")
	assert.True(t, strings.Contains(logString, `"status":202`), "Log should contain status code 202")
	assert.True(t, strings.Contains(logString, `"latency_ms"`), "Log should contain latency field")

}
