package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/vlatan/video-store/internal/server"
)

func main() {

	// Create new server
	server, domain, cleanup := server.NewServer()

	// Create a notification channel to receive a signal
	// from when a  shutdown is complete
	done := make(chan struct{})

	// Listen for SIGINT SIGTERM in a separate goroutine
	// Gracefully shut down the server.
	// If so ListenAndServe will return ErrServerClosed.
	go gracefulShutdown(server, cleanup, done)

	fmt.Printf("Server running on: http://%s\n", server.Addr)
	if domain != "" {
		fmt.Printf("Website available at: https://%s\n", domain)
	}

	err := server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Fatalf("http server error: %v", err)
	}

	<-done // Wait for the graceful shutdown to complete
	log.Println("Graceful shutdown complete.")
}

// gracefulShutdown listens for SIGINT and SIGTERM signals,
// shuts down the server, performs cleanup and informs the
// main goroutine
func gracefulShutdown(server *http.Server, cleanup func() error, done chan<- struct{}) {
	// Create context that listens for the interrupt signal from the OS.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// This is a blocking call.
	// If context is done an interruption signal was received.
	// This channel is closed by the sender and the program proceeds.
	<-ctx.Done()

	log.Println("shutting down gracefully, press Ctrl+C again to force")

	// Cancel the context, stop watching for termination signals.
	// Now if the user presses Ctrl+C again (or someone sends SIGINT, SIGTERM signal),
	// that will go straight to the OS and kill the process immediately,
	// bypassing the graceful shutdown.
	stop()

	// This context will give the server 5 seconds
	// to finish the requests it is currently handling.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// This is a blocking call.
	// Shutdown will wait indefinitely for connections to return to idle,
	// but in this case up to 5 seconds.
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown with error: %v", err)
	}

	// Perform cleanup. Close the DB pool and Redis connections.
	log.Println("Closing Database and Redis connections...")
	if err := cleanup(); err != nil {
		log.Printf("Error during cleanup: %v", err)
	}

	log.Println("Server exiting...")

	// Notify the main goroutine that the shutdown is complete
	done <- struct{}{}
}
