package app

import (
	"context"
	"go-clean-arch/pkg/config"
	"go-clean-arch/pkg/logger"
	"os"
	"os/signal"
	"syscall"

	"github.com/jmoiron/sqlx"
)

// Worker represents the background worker application
type Worker struct {
	db  *sqlx.DB
	cfg *config.Config
}

// NewWorker creates a new Worker instance
func NewWorker(db *sqlx.DB, cfg *config.Config) *Worker {
	return &Worker{
		db:  db,
		cfg: cfg,
	}
}

// Start starts the worker with graceful shutdown support
func (w *Worker) Start() error {
	logger.Info("Starting worker...")

	// Wire automatically injects all dependencies
	orderConsumer := InitializeWorker(w.db, w.cfg)

	// Create context that listens for interrupt signals (SIGINT, SIGTERM)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	// Channel to listen for consumer errors
	consumerErrors := make(chan error, 1)

	// Start consumer in a goroutine (non-blocking)
	go func() {
		logger.Info("Worker is ready to consume messages")
		consumerErrors <- orderConsumer.Start(ctx)
	}()

	// Block until we receive a signal or consumer error
	select {
	case err := <-consumerErrors:
		if err != nil {
			logger.Error("Consumer error", "error", err)
			return err
		}
		logger.Info("Consumer stopped")

	case <-ctx.Done():
		// Restore default signal handling
		stop()

		logger.Info("Shutdown signal received, stopping worker...")

		// The consumer will stop when it detects context cancellation
		// Wait for consumer to finish
		if err := <-consumerErrors; err != nil {
			logger.Error("Error during consumer shutdown", "error", err)
		}

		// Close database connection
		if err := w.db.Close(); err != nil {
			logger.Error("Error closing database connection", "error", err)
		}

		logger.Info("Worker stopped gracefully")
	}

	return nil
}
