// Package app provides application setup and database initialization.
package app

import (
	"fmt"

	"go-clean-arch/internal/adapter/repository"
	"go-clean-arch/pkg/config"
	"go-clean-arch/pkg/logger"

	"gorm.io/gorm"
)

// InitializeDatabase initializes the database connection with configuration.
// It handles connection setup and optional auto-migration.
//
// This function encapsulates database setup logic to keep it out of the CMD layer,
// maintaining clean architecture principles by avoiding infrastructure dependencies
// in outer layers.
func InitializeDatabase(cfg *config.Config) (*gorm.DB, error) {
	dbConfig := repository.DBConfig{
		Driver:   cfg.Database.Driver,
		Host:     cfg.Database.Host,
		Port:     cfg.Database.Port,
		User:     cfg.Database.User,
		Password: cfg.Database.Password,
		DBName:   cfg.Database.DBName,
		SSLMode:  cfg.Database.SSLMode,
	}

	db, err := repository.NewDatabase(dbConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	logger.Info("Database connected",
		"driver", cfg.Database.Driver,
		"host", cfg.Database.Host,
		"database", cfg.Database.DBName,
	)

	// Run auto-migration if enabled
	if cfg.Database.AutoMigrate {
		logger.Info("Running database migrations...")
		if err := repository.AutoMigrate(db); err != nil {
			return nil, fmt.Errorf("failed to run migrations: %w", err)
		}
		logger.Info("Database migrations completed")
	}

	return db, nil
}
