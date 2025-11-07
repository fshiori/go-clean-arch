package app

import (
	"go-clean-arch/pkg/config"
	"go-clean-arch/pkg/logger"

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

// Start starts the worker
func (w *Worker) Start() error {
	logger.Info("Starting worker...")

	// Wire automatically injects all dependencies
	orderConsumer := InitializeWorker(w.db, w.cfg)

	// Start consuming messages
	logger.Info("Worker is ready to consume messages")
	return orderConsumer.Start()
}
