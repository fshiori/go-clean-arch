package app

import (
	"go-clean-arch/internal/adapter/gateway"
	"go-clean-arch/internal/adapter/repository"
	"go-clean-arch/internal/delivery/consumer"
	"go-clean-arch/internal/usecase"
	"log"

	"gorm.io/gorm"
)

// Worker represents the background worker application
type Worker struct {
	db *gorm.DB
}

// NewWorker creates a new Worker instance
func NewWorker(db *gorm.DB) *Worker {
	return &Worker{
		db: db,
	}
}

// Start starts the worker
func (w *Worker) Start() error {
	log.Println("Starting worker...")

	// Initialize repositories
	userRepo := repository.NewUserRepository(w.db)
	orderRepo := repository.NewOrderRepository(w.db)

	// Initialize gateways
	paymentGateway := gateway.NewStripeGateway("your-stripe-api-key")

	// Initialize use cases
	orderInteractor := usecase.NewOrderInteractor(orderRepo, userRepo, paymentGateway)

	// Initialize consumers
	orderConsumer := consumer.NewOrderConsumer(orderInteractor)

	// Start consuming messages
	log.Println("Worker is ready to consume messages")
	return orderConsumer.Start()
}
