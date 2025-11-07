// Package app provides application setup and database initialization.
package app

import (
	"fmt"

	"go-clean-arch/internal/adapter/repository"
	"go-clean-arch/pkg/config"
	"go-clean-arch/pkg/logger"

	"github.com/jmoiron/sqlx"
)

// InitializeDatabase initializes the database connection with configuration.
// It handles connection setup. Note: sqlx doesn't have auto-migration,
// you should use migration tools like golang-migrate or sql files.
//
// This function encapsulates database setup logic to keep it out of the CMD layer,
// maintaining clean architecture principles by avoiding infrastructure dependencies
// in outer layers.
func InitializeDatabase(cfg *config.Config) (*sqlx.DB, error) {
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

	// Note: AutoMigrate is removed. Use SQL migration files instead.
	// See migrations/ directory for migration files.
	if cfg.Database.AutoMigrate {
		logger.Warn("AutoMigrate is enabled but not supported with sqlx. Please use migration tools like golang-migrate.")
	}

	return db, nil
}
