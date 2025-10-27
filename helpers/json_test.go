package helpers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSendJSON(t *testing.T) {
	type testData struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	data := testData{ID: 1, Name: "Cassie"}

	rr := httptest.NewRecorder()

	err := SendJSON(rr, http.StatusCreated, data)

	assert.NoError(t, err, "SendJSON should not fail")

	assert.Equal(t, http.StatusCreated, rr.Code, "Status code should be 201 Created")

	expectedHeader := "application/json; charset=utf-8"
	assert.Equal(t, expectedHeader, rr.Header().Get("Content-Type"), "Incorrect Content-Type header")

	var responseData testData
	err = json.Unmarshal(rr.Body.Bytes(), &responseData)

	assert.NoError(t, err, "Response body should be valid JSON")
	assert.Equal(t, data, responseData, "Response JSON body does not match input data")
}

func TestSendError(t *testing.T) {
	rr := httptest.NewRecorder()

	SendError(rr, http.StatusNotFound, "User not found")

	assert.Equal(t, http.StatusNotFound, rr.Code, "Status code should be 404 Not Found")

	expectedHeader := "application/json; charset=utf-8"
	assert.Equal(t, expectedHeader, rr.Header().Get("Content-Type"), "Incorrect Content-Type header")

	expectedJSON := `{"error":"User not found"}`
	assert.JSONEq(t, expectedJSON, rr.Body.String(), "Error JSON body does not match expected value")
}
