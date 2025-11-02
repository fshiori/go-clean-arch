package app

import (
	"go-clean-arch/internal/delivery/job"
	"go-clean-arch/internal/interface/gateway"
	"go-clean-arch/internal/interface/repository"
	"go-clean-arch/internal/usecase"
	"log"

	"gorm.io/gorm"
)

// CronScheduler represents the cron job scheduler application
type CronScheduler struct {
	db *gorm.DB
}

// NewCronScheduler creates a new CronScheduler instance
func NewCronScheduler(db *gorm.DB) *CronScheduler {
	return &CronScheduler{
		db: db,
	}
}

// Start starts the cron scheduler
func (c *CronScheduler) Start() error {
	log.Println("Starting cron scheduler...")

	// Initialize repositories
	userRepo := repository.NewUserRepository(c.db)
	orderRepo := repository.NewOrderRepository(c.db)

	// Initialize gateways
	paymentGateway := gateway.NewStripeGateway("your-stripe-api-key")

	// Initialize use cases
	userInteractor := usecase.NewUserInteractor(userRepo)
	orderInteractor := usecase.NewOrderInteractor(orderRepo, userRepo, paymentGateway)

	// Initialize jobs
	dailyReportJob := job.NewDailyReportJob(userInteractor, orderInteractor)

	// Create scheduler
	scheduler := job.NewScheduler(dailyReportJob)

	// Start scheduler
	log.Println("Cron scheduler is ready")
	return scheduler.Start()
}
