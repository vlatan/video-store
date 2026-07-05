package middlewares

import (
	"net/http"
)

type statusTracker struct {
	http.ResponseWriter
	status int
}

// NewStatusTracker creates a new statusTracker.
func NewStatusTracker(w http.ResponseWriter) *statusTracker {
	return &statusTracker{
		ResponseWriter: w,
		status:         http.StatusOK, // Default to 200 OK
	}
}

func (st *statusTracker) WriteHeader(statusCode int) {
	st.status = statusCode
	st.ResponseWriter.WriteHeader(statusCode)
}

func (st *statusTracker) Write(b []byte) (int, error) {
	if st.status == 0 {
		st.status = http.StatusOK
	}
	return st.ResponseWriter.Write(b)
}
