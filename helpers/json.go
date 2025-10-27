package helpers

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/iameggi/cassie/bucket"
)

// defaultErrorLogger is used only if SendError itself fails.
// It writes to stderr with a consistent prefix.
var defaultErrorLogger = log.New(os.Stderr, "CASSIE HELPER ERROR: ", log.LstdFlags)

// SendJSON writes a high-performance JSON response using Cassie's pooled buffers.
//
// This helper automatically sets the Content-Type header and encodes the given data
// into a pooled *bytes.Buffer to minimize memory allocations and GC overhead.
//
// Returns an error if JSON encoding or writing to the client fails.
func SendJSON(w http.ResponseWriter, statusCode int, data interface{}) error {
	return bucket.WithByteBufferErr(func(buf *bytes.Buffer) error {
		// Encode JSON directly into the pooled buffer.
		if err := json.NewEncoder(buf).Encode(data); err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return err
		}

		// Write headers and response body.
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(statusCode)

		if _, err := w.Write(buf.Bytes()); err != nil {
			// Handle client write errors (e.g., broken pipe).
			return err
		}
		return nil
	})
}

// SendError is a convenience helper for sending structured JSON error responses.
//
// It wraps SendJSON to ensure consistent error formatting across your application.
// SendError does not return an error itself — if the response write fails,
// the failure is logged using defaultErrorLogger.
func SendError(w http.ResponseWriter, statusCode int, message string) {
	type errorResponse struct {
		Error string `json:"error"`
	}

	if err := SendJSON(w, statusCode, errorResponse{Error: message}); err != nil {
		defaultErrorLogger.Printf("failed to send SendError response: %v", err)
	}
}
