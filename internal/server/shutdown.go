package server

import (
	"context"
	"log"
	"os/signal"
	"syscall"
	"time"
)

// Shutdown listens for SIGINT and SIGTERM signals,
// gracefully shuts down the HTTP server,
// performs cleanup and informs the main goroutine when done.
func (s *Server) Shutdown(done chan<- struct{}) {
	// Create context that listens for the interrupt signal from the OS.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// This is a blocking call.
	// If context is done an interruption signal was received.
	// This channel is closed by the sender and the program proceeds.
	<-ctx.Done()

	log.Println("Shutting down gracefully, press Ctrl+C again to force...")

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
	// Shutdown will wait for connections to return to idle,
	// but in this case up to 5 seconds.
	if err := s.HttpServer.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown with error: %v", err)
	}

	// Perform cleanup. Close the DB pool and Redis connections.
	log.Println("Closing Database and Redis connections...")
	if err := s.cleanup(); err != nil {
		log.Printf("Error during cleanup: %v", err)
	}

	log.Println("Server exiting...")

	// Notify the main goroutine that the shutdown is complete
	done <- struct{}{}
}
