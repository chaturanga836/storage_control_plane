// internal/shutdown/shutdown.go
package shutdown

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// HandleShutdown waits for termination signals and cancels the context.
func HandleShutdown(cancelFunc context.CancelFunc) {
	sig := make(chan os.Signal, 1)

	// Listen for SIGINT (Ctrl+C) and SIGTERM (kill)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	// Block until a signal is received
	<-sig

	log.Println("📴 Shutting down gracefully...")
	cancelFunc()

	// Optional: Give goroutines time to finish
	time.Sleep(1 * time.Second)
}
