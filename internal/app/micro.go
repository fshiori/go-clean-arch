// Package app provides application setup and execution for different modes.
package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"go-clean-arch/internal/delivery/micro/handler"
	"go-clean-arch/pkg/config"
	"go-clean-arch/pkg/logger"

	"github.com/jmoiron/sqlx"
	"go-micro.dev/v5"
	"go-micro.dev/v5/server"
)

// MicroServer represents the microservice server application
type MicroServer struct {
	db          *sqlx.DB
	cfg         *config.Config
	serviceName string
	version     string
	address     string
}

// NewMicroServer creates a new microservice server instance
func NewMicroServer(db *sqlx.DB, cfg *config.Config, serviceName, version, address string) *MicroServer {
	return &MicroServer{
		db:          db,
		cfg:         cfg,
		serviceName: serviceName,
		version:     version,
		address:     address,
	}
}

// Start starts the microservice server with graceful shutdown support
func (m *MicroServer) Start() error {
	logger.Info("Starting go-micro microservice...",
		"service", m.serviceName,
		"version", m.version,
		"address", m.address,
	)

	// Wire automatically injects all dependencies
	userService := InitializeMicroService(m.db, m.cfg)

	// Create a new micro service
	srv := micro.NewService(
		micro.Name(m.serviceName),
		micro.Version(m.version),
		micro.Address(m.address),
	)

	// Initialize service
	srv.Init()

	// Register handler with the service
	// The handler is registered with the service name "UserServiceSimple"
	if err := server.RegisterHandler(srv.Server(), userService); err != nil {
		return fmt.Errorf("failed to register handler: %w", err)
	}

	// Create context that listens for interrupt signals
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	// Channel to listen for server errors
	serverErrors := make(chan error, 1)

	// Start server in a goroutine (non-blocking)
	go func() {
		logger.Info("Go-micro microservice is ready to accept requests",
			"transport", "http",
			"address", m.address,
		)
		if err := srv.Run(); err != nil {
			serverErrors <- err
		}
	}()

	// Block until we receive a signal or server error
	select {
	case err := <-serverErrors:
		logger.Error("Server error", "error", err)
		return fmt.Errorf("server error: %w", err)

	case <-ctx.Done():
		// Restore default signal handling
		stop()

		logger.Info("Shutdown signal received, starting graceful shutdown...")

		// Close database connection
		if err := m.db.Close(); err != nil {
			logger.Error("Error closing database connection", "error", err)
		}

		logger.Info("Go-micro microservice stopped gracefully")
		return nil
	}
}
