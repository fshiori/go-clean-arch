// Package app provides application setup and execution for different modes.
// It uses Wire for dependency injection to manage the component lifecycle.
package app

import (
	"context"
	"errors"
	"fmt"
	"go-clean-arch/pkg/config"
	"go-clean-arch/pkg/logger"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jmoiron/sqlx"
)

// APIServer represents the API server application
type APIServer struct {
	db   *sqlx.DB
	cfg  *config.Config
	port int
}

// NewAPIServer creates a new API server instance
func NewAPIServer(db *sqlx.DB, cfg *config.Config, port int) *APIServer {
	return &APIServer{
		db:   db,
		cfg:  cfg,
		port: port,
	}
}

// Start starts the API server with graceful shutdown support.
// This implementation follows the official Gin documentation:
// https://gin-gonic.com/docs/examples/graceful-restart-or-stop/
// https://github.com/gin-gonic/examples/tree/master/graceful-shutdown
func (s *APIServer) Start() error {
	logger.Info("Starting API server...")

	// Wire automatically injects all dependencies
	router := InitializeAPIRouter(s.db, s.cfg)

	// Create HTTP server with the Gin router
	addr := fmt.Sprintf(":%d", s.port)
	srv := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	// Channel to listen for errors from server
	serverErrors := make(chan error, 1)

	// Start server in a goroutine (non-blocking)
	go func() {
		logger.Info("API server listening", "address", addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrors <- err
		}
	}()

	// Create context that listens for interrupt signals (SIGINT, SIGTERM)
	// This uses signal.NotifyContext which is the recommended approach since Go 1.16
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	// Block until we receive a signal or server error
	select {
	case err := <-serverErrors:
		logger.Error("Server error", "error", err)
		return fmt.Errorf("server error: %w", err)

	case <-ctx.Done():
		// Restore default signal handling
		stop()

		logger.Info("Shutdown signal received, starting graceful shutdown...")

		// Create a deadline for graceful shutdown (30 seconds)
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Attempt graceful shutdown
		if err := srv.Shutdown(shutdownCtx); err != nil {
			logger.Error("Server forced to shutdown", "error", err)
			return fmt.Errorf("server forced to shutdown: %w", err)
		}

		// Close database connection
		if err := s.db.Close(); err != nil {
			logger.Error("Error closing database connection", "error", err)
		}

		logger.Info("API server stopped gracefully")
		return nil
	}
}
