package app

import (
	"go-clean-arch/pkg/config"
	"go-clean-arch/pkg/logger"

	"gorm.io/gorm"
)

// CronScheduler represents the cron job scheduler application
type CronScheduler struct {
	db  *gorm.DB
	cfg *config.Config
}

// NewCronScheduler creates a new CronScheduler instance
func NewCronScheduler(db *gorm.DB, cfg *config.Config) *CronScheduler {
	return &CronScheduler{
		db:  db,
		cfg: cfg,
	}
}

// Start starts the cron scheduler
func (c *CronScheduler) Start() error {
	logger.Info("Starting cron scheduler...")

	// Wire automatically injects all dependencies
	scheduler := InitializeCronScheduler(c.db, c.cfg)

	// Start scheduler
	logger.Info("Cron scheduler is ready")
	return scheduler.Start()
}
