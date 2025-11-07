// Package app provides application setup and execution for different modes.
// It uses Wire for dependency injection to manage the component lifecycle.
package app

import (
	"fmt"
	"go-clean-arch/pkg/config"
	"go-clean-arch/pkg/logger"

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

// Start starts the API server
func (s *APIServer) Start() error {
	logger.Info("Starting API server...")

	// Wire automatically injects all dependencies
	router := InitializeAPIRouter(s.db, s.cfg)

	// Start server
	addr := fmt.Sprintf(":%d", s.port)
	logger.Info("API server listening", "address", addr)

	if err := router.Run(addr); err != nil {
		logger.Error("Failed to start server", "error", err)
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}
