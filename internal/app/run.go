package app

import (
	"fmt"
	"log"
	"net/http"
)

// Run runs the app by making the HTTP server listen and serve
func (a *App) Run() error {

	// Create a notification channel to receive a signal
	// from when a shutdown is complete
	done := make(chan struct{})

	// Listen for SIGINT SIGTERM in a separate goroutine
	// Gracefully shut down the server there if needed.
	go a.Shutdown(done)

	fmt.Printf("Server running on: http://%s\n", a.server.Addr)
	if a.domain != "" {
		fmt.Printf("Website available at: https://%s\n", a.domain)
	}

	// If the HTTP server was shut down, meaning
	// s.Server.Shutdown(ctx) method was called,
	// ListenAndServe will return ErrServerClosed.
	err := a.server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		return err
	}

	<-done // Wait for the graceful shutdown to complete
	log.Println("Graceful shutdown complete.")

	return nil
}
