package app

import (
	"fmt"
	"go-clean-arch/internal/delivery/http"
	"go-clean-arch/internal/delivery/http/handler"
	"go-clean-arch/internal/interface/gateway"
	"go-clean-arch/internal/interface/repository"
	"go-clean-arch/internal/usecase"
	"log"

	"gorm.io/gorm"
)

// APIServer represents the API server application
type APIServer struct {
	db     *gorm.DB
	port   int
}

// NewAPIServer creates a new API server instance
func NewAPIServer(db *gorm.DB, port int) *APIServer {
	return &APIServer{
		db:   db,
		port: port,
	}
}

// Start starts the API server
func (s *APIServer) Start() error {
	log.Println("Starting API server...")

	// Initialize repositories
	userRepo := repository.NewUserRepository(s.db)
	orderRepo := repository.NewOrderRepository(s.db)

	// Initialize gateways
	paymentGateway := gateway.NewStripeGateway("your-stripe-api-key")

	// Initialize use cases (interactors)
	userInteractor := usecase.NewUserInteractor(userRepo)
	orderInteractor := usecase.NewOrderInteractor(orderRepo, userRepo, paymentGateway)

	// Initialize HTTP handlers
	userHandler := handler.NewUserHandler(userInteractor)

	// Setup router
	router := http.Router(userHandler)

	// Start server
	addr := fmt.Sprintf(":%d", s.port)
	log.Printf("API server listening on %s", addr)

	if err := router.Run(addr); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}
