package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"factual-docs/internal/server"
)

type Closer interface {
	Close() error
}

func gracefulShutdown(appServer *http.Server, cleanup func() error, done chan bool) {
	// Create context that listens for the interrupt signal from the OS.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Listen for the interrupt signal.
	<-ctx.Done()

	log.Println("shutting down gracefully, press Ctrl+C again to force")

	// Cancel the context, stop watching for termination signals.
	// Now if the user presses Ctrl+C again (or someone sends SIGINT, SIGTERM signal),
	// that will go straight to the OS and kill the process immediately,
	// bypassing the graceful shutdown.
	stop()

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// This is a blocking call.
	// Shutdown will wait indefinitely for connections to return to idle,
	// but in this case up to 5 seconds.
	if err := appServer.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown with error: %v", err)
	}

	log.Println("Closing Database and Redis connections...")
	if err := cleanup(); err != nil {
		log.Printf("Error during cleanup: %v", err)
	}

	log.Println("Server exiting...")

	// Notify the main goroutine that the shutdown is complete
	done <- true
}

func main() {
	server, cleanup := server.NewServer()

	// Create a done channel to signal when the shutdown is complete
	done := make(chan bool, 1)

	// Listen for SIGINT SIGTERM in a separate goroutine
	// Gracefully shut down the server. If so return ErrServerClosed.
	go gracefulShutdown(server, cleanup, done)

	fmt.Printf("Server running on http://%s\n", server.Addr)
	err := server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Fatalf("http server error: %v", err)
	}

	// Wait for the graceful shutdown to complete
	<-done
	log.Println("Graceful shutdown complete.")
}
