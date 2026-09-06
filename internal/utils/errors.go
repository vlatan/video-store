package utils

import (
	"context"
	"errors"
	"net/http"
)

// HttpError provides shorter handling of http error
func HttpError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

// IsContextErr checks if a given error is context error
func IsContextErr(err error) bool {
	return errors.Is(err, context.Canceled) ||
		errors.Is(err, context.DeadlineExceeded)
}
