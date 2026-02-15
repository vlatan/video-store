package server

import (
	"fmt"
	"log"
	"net/http"
)

// Run runs the app by making the HTTP server listen and serve
func (s *Server) Run() error {

	// Create a notification channel to receive a signal
	// from when a shutdown is complete
	done := make(chan struct{})

	// Listen for SIGINT SIGTERM in a separate goroutine
	// Gracefully shut down the server there if needed.
	go s.Shutdown(done)

	fmt.Printf("Server running on: http://%s\n", s.HttpServer.Addr)
	if s.Domain != "" {
		fmt.Printf("Website available at: https://%s\n", s.Domain)
	}

	// If the HTTP server was shut down, meaning
	// s.HttpServer.Shutdown(ctx) method was called,
	// ListenAndServe will return ErrServerClosed.
	err := s.HttpServer.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		return err
	}

	<-done // Wait for the graceful shutdown to complete
	log.Println("Graceful shutdown complete.")

	return nil
}
