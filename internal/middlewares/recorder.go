package middlewares

import (
	"bytes"
	"net/http"
)

// responseRecorder is a custom http.ResponseWriter that captures the status code
// and the body of the response. This allows the middleware to inspect the
// response from the next handler before writing it to the client.
type responseRecorder struct {
	http.ResponseWriter
	body        *bytes.Buffer
	status      int
	wroteHeader bool
}

// NewResponseRecorder creates a new responseRecorder.
func NewResponseRecorder(w http.ResponseWriter) *responseRecorder {
	return &responseRecorder{
		ResponseWriter: w,
		body:           new(bytes.Buffer),
		status:         http.StatusOK, // Default to 200 OK
	}
}

// WriteHeader captures the response status code if error
func (r *responseRecorder) WriteHeader(statusCode int) {
	if r.wroteHeader {
		return
	}

	r.status = statusCode
	r.wroteHeader = true

	// Stream successful headers immediately
	if r.status < http.StatusBadRequest {
		r.ResponseWriter.WriteHeader(statusCode)
	}
}

// Write captures the response body.
func (r *responseRecorder) Write(b []byte) (int, error) {
	if !r.wroteHeader {
		r.WriteHeader(http.StatusOK) // Implicit 200 OK
	}

	// Only buffer the body if it's an error
	if r.status >= http.StatusBadRequest {
		return r.body.Write(b)
	}

	// Stream successful bodies directly to the client
	return r.ResponseWriter.Write(b)
}

// flush sends the captured response (or a modified one) to the client.
func (r *responseRecorder) flush() {
	if r.status < http.StatusBadRequest {
		return
	}

	// Successful requests were already streamed.
	// Only write headers and body for errors.
	r.ResponseWriter.WriteHeader(r.status)
	if r.body.Len() > 0 {
		r.ResponseWriter.Write(r.body.Bytes())
	}
}
