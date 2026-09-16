package utils

import (
	"context"
	"errors"
	"net/http"

	"github.com/vlatan/video-store/internal/ctxerrors"
)

// HttpError provides shorter handling of http error.
// Adds an error to context if any.
func HttpError(w http.ResponseWriter, r *http.Request, status int, err error) {
	ctxerrors.Add(r.Context(), err)
	http.Error(w, http.StatusText(status), status)
}

// IsContextErr checks if a given error is context error
func IsContextErr(err error) bool {
	return errors.Is(err, context.Canceled) ||
		errors.Is(err, context.DeadlineExceeded)
}
