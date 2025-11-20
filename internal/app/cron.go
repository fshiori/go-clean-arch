package app

import (
	"context"
	"go-clean-arch/pkg/config"
	"go-clean-arch/pkg/logger"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jmoiron/sqlx"
)

// CronScheduler represents the cron job scheduler application
type CronScheduler struct {
	db  *sqlx.DB
	cfg *config.Config
}

// NewCronScheduler creates a new CronScheduler instance
func NewCronScheduler(db *sqlx.DB, cfg *config.Config) *CronScheduler {
	return &CronScheduler{
		db:  db,
		cfg: cfg,
	}
}

// Start starts the cron scheduler with graceful shutdown support
func (c *CronScheduler) Start() error {
	logger.Info("Starting cron scheduler...")

	// Wire automatically injects all dependencies
	scheduler := InitializeCronScheduler(c.db, c.cfg)

	// Start scheduler
	if err := scheduler.Start(); err != nil {
		logger.Error("Failed to start scheduler", "error", err)
		return err
	}

	logger.Info("Cron scheduler is ready")

	// Create context that listens for interrupt signals (SIGINT, SIGTERM)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	// Block until we receive a shutdown signal
	<-ctx.Done()

	// Restore default signal handling
	stop()

	logger.Info("Shutdown signal received, stopping cron scheduler...")

	// Stop scheduler and get context for waiting on running jobs
	// The cron.Stop() method returns a context that will be done when all jobs complete
	jobsCtx := scheduler.Stop()

	// Wait for running jobs to complete with a timeout
	select {
	case <-jobsCtx.Done():
		logger.Info("All cron jobs completed successfully")
	case <-time.After(30 * time.Second):
		logger.Warn("Timeout waiting for cron jobs to complete, forcing shutdown")
	}

	// Close database connection
	if err := c.db.Close(); err != nil {
		logger.Error("Error closing database connection", "error", err)
	}

	logger.Info("Cron scheduler stopped gracefully")
	return nil
}
