package app

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/vlatan/video-store/internal/utils/color"
)

// Run runs the app by making the HTTP server listen and serve
func (a *App) Run() error {

	// Create a notification channel to receive a signal
	// from when a shutdown is complete
	done := make(chan struct{})

	// Listen for SIGINT SIGTERM in a separate goroutine
	// Gracefully shut down the server there if needed.
	go a.Shutdown(done)

	color.Yellow(fmt.Sprintf("Server running on: http://%s", a.server.Addr))
	if a.domain != "" {
		color.Yellow(fmt.Sprintf("Website available at: https://%s", a.domain))
	}

	// If the HTTP server was shut down, meaning
	// s.Server.Shutdown(ctx) method was called,
	// ListenAndServe will return ErrServerClosed.
	err := a.server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	<-done // Wait for the graceful shutdown to complete

	return nil
}
