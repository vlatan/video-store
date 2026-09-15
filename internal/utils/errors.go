package utils

import (
	"context"
	"errors"
	"net/http"
)

type contextKey string

const errorContextKey contextKey = "requestError"

// HttpError provides shorter handling of http error
func HttpError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

// IsContextErr checks if a given error is context error
func IsContextErr(err error) bool {
	return errors.Is(err, context.Canceled) ||
		errors.Is(err, context.DeadlineExceeded)
}

// AttachError stores error in the context
func AttachErrorToCtx(r *http.Request, err error) *http.Request {
	if err == nil {
		return r
	}
	ctx := context.WithValue(r.Context(), errorContextKey, err)
	return r.WithContext(ctx)
}

// GetError retrieves error from the context
func GetErrorFromCtx(r *http.Request) error {
	if err, ok := r.Context().Value(errorContextKey).(error); ok {
		return err
	}
	return nil
}
