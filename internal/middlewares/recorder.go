package middlewares

import (
	"bytes"
	"log"
	"net/http"
)

// A custom http.ResponseWriter that captures the status code
// and the body of the response. This allows the middleware to inspect the
// response from the next handler before writing it to the client.
type responseRecorder struct {
	http.ResponseWriter
	body   *bytes.Buffer
	status int
}

// Creates a new responseRecorder
func NewResponseRecorder(w http.ResponseWriter) *responseRecorder {
	return &responseRecorder{
		ResponseWriter: w,
		body:           new(bytes.Buffer),
		status:         http.StatusOK, // Default to 200 OK
	}
}

// Captures the response status code
func (r *responseRecorder) WriteHeader(statusCode int) {
	r.status = statusCode
}

// Captures the response body.
func (r *responseRecorder) Write(b []byte) (int, error) {
	return r.body.Write(b)
}

// Sends the captured response (or a modified one) to the client.
func (r *responseRecorder) flush() {
	r.ResponseWriter.WriteHeader(r.status)
	if r.body.Len() > 0 {
		_, err := r.ResponseWriter.Write(r.body.Bytes())
		if err != nil {
			// Too late for recovery here, just log the error
			log.Printf("Error writing response body: %v", err)
		}
	}
}
