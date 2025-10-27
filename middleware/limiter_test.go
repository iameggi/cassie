package middleware

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLimiter(t *testing.T) {
	const maxConcurrency = 2

	limiter := NewLimiter(maxConcurrency)

	handlerRunning := make(chan struct{}, maxConcurrency+1)
	handlerFinish := make(chan struct{})

	mockHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerRunning <- struct{}{}
		<-handlerFinish
	})

	handlerToTest := limiter.Wrap(mockHandler)
	const totalRequests = maxConcurrency + 1
	var wg sync.WaitGroup
	wg.Add(totalRequests)

	for i := 0; i < totalRequests; i++ {
		go func() {
			defer wg.Done()
			req := httptest.NewRequest("GET", "/", nil)
			rr := httptest.NewRecorder()
			handlerToTest.ServeHTTP(rr, req)
		}()
	}

	time.Sleep(50 * time.Millisecond)

	runningCount := len(handlerRunning)
	assert.Equal(t, maxConcurrency, runningCount, "ONLY %d handlers should be running concurrently (the rest must be blocked)", maxConcurrency)

	handlerFinish <- struct{}{}
	time.Sleep(50 * time.Millisecond)

	runningCount = len(handlerRunning)
	assert.Equal(t, totalRequests, runningCount, "Handler #%d should now be allowed to start", totalRequests)

	for i := 0; i < maxConcurrency; i++ {
		handlerFinish <- struct{}{}
	}

	wg.Wait()
}
