// Package app provides application setup and execution for different modes.
package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-clean-arch/internal/delivery/micro/handler"
	"go-clean-arch/pkg/config"
	"go-clean-arch/pkg/logger"

	"github.com/jmoiron/sqlx"
)

// MicroServer represents the microservice server application
// This is a lightweight JSON-RPC style microservice implementation
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
	logger.Info("Starting microservice...",
		"service", m.serviceName,
		"version", m.version,
		"address", m.address,
	)

	// Wire automatically injects all dependencies
	userService := InitializeMicroService(m.db, m.cfg)

	// Create HTTP mux for RPC-style endpoints
	mux := http.NewServeMux()

	// Register service info endpoint
	mux.HandleFunc("/info", func(w http.ResponseWriter, r *http.Request) {
		info := map[string]string{
			"service": m.serviceName,
			"version": m.version,
			"status":  "running",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(info)
	})

	// Register RPC-style endpoints for user service
	mux.HandleFunc("/rpc/user/create", m.wrapHandler(userService.CreateUser))
	mux.HandleFunc("/rpc/user/get", m.wrapHandler(userService.GetUser))
	mux.HandleFunc("/rpc/user/list", m.wrapHandler(userService.ListUsers))
	mux.HandleFunc("/rpc/user/update-password", m.wrapHandler(userService.UpdatePassword))
	mux.HandleFunc("/rpc/user/delete", m.wrapHandler(userService.DeleteUser))

	// Create HTTP server
	srv := &http.Server{
		Addr:    m.address,
		Handler: mux,
	}

	// Channel to listen for errors from server
	serverErrors := make(chan error, 1)

	// Start server in a goroutine (non-blocking)
	go func() {
		logger.Info("Microservice is ready to accept requests")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrors <- err
		}
	}()

	// Create context that listens for interrupt signals
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
		if err := m.db.Close(); err != nil {
			logger.Error("Error closing database connection", "error", err)
		}

		logger.Info("Microservice stopped gracefully")
		return nil
	}
}

// wrapHandler wraps an RPC handler function to handle HTTP requests
func (m *MicroServer) wrapHandler(fn interface{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		// Handle different function signatures
		switch f := fn.(type) {
		case func(context.Context, *handler.CreateUserReq, *handler.CreateUserRsp) error:
			var req handler.CreateUserReq
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			var rsp handler.CreateUserRsp
			if err := f(r.Context(), &req, &rsp); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			json.NewEncoder(w).Encode(rsp)

		case func(context.Context, *handler.GetUserReq, *handler.GetUserRsp) error:
			var req handler.GetUserReq
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			var rsp handler.GetUserRsp
			if err := f(r.Context(), &req, &rsp); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			json.NewEncoder(w).Encode(rsp)

		case func(context.Context, *handler.ListUsersReq, *handler.ListUsersRsp) error:
			var req handler.ListUsersReq
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			var rsp handler.ListUsersRsp
			if err := f(r.Context(), &req, &rsp); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			json.NewEncoder(w).Encode(rsp)

		case func(context.Context, *handler.UpdatePasswordReq, *handler.UpdatePasswordRsp) error:
			var req handler.UpdatePasswordReq
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			var rsp handler.UpdatePasswordRsp
			if err := f(r.Context(), &req, &rsp); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			json.NewEncoder(w).Encode(rsp)

		case func(context.Context, *handler.DeleteUserReq, *handler.DeleteUserRsp) error:
			var req handler.DeleteUserReq
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			var rsp handler.DeleteUserRsp
			if err := f(r.Context(), &req, &rsp); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			json.NewEncoder(w).Encode(rsp)

		default:
			http.Error(w, "Unknown handler type", http.StatusInternalServerError)
		}
	}
}
