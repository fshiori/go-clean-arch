package main

import (
	"flag"
	"go-clean-arch/internal/app"
	"go-clean-arch/internal/adapter/repository"
	"go-clean-arch/pkg/config"
	"go-clean-arch/pkg/logger"
	"log"
	"os"
)

func main() {
	// Define command-line flags
	mode := flag.String("mode", "api", "The mode to run the application in (api, worker, cron)")
	configPath := flag.String("config", "configs/config.toml", "Path to configuration file")
	flag.Parse()

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger with config
	logger.Init(logger.Config{
		Level:  cfg.Logger.Level,
		Format: cfg.Logger.Format,
	})
	logger.Info("Starting application", "mode", *mode)

	// Initialize database connection (shared by all modes)
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
		logger.Error("Failed to connect to database", "error", err)
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Run auto-migration
	if cfg.Database.AutoMigrate {
		logger.Info("Running database migrations...")
		if err := repository.AutoMigrate(db); err != nil {
			logger.Error("Failed to run migrations", "error", err)
			log.Fatalf("Failed to run migrations: %v", err)
		}
		logger.Info("Database migrations completed")
	}

	// Start the application based on mode
	switch *mode {
	case "api":
		logger.Info("Starting API server", "port", cfg.Server.Port)
		apiServer := app.NewAPIServer(db, cfg.Server.Port)
		if err := apiServer.Start(); err != nil {
			logger.Error("Failed to start API server", "error", err)
			log.Fatalf("Failed to start API server: %v", err)
		}

	case "worker":
		logger.Info("Starting worker")
		worker := app.NewWorker(db)
		if err := worker.Start(); err != nil {
			logger.Error("Failed to start worker", "error", err)
			log.Fatalf("Failed to start worker: %v", err)
		}

	case "cron":
		logger.Info("Starting cron scheduler")
		cronScheduler := app.NewCronScheduler(db)
		if err := cronScheduler.Start(); err != nil {
			logger.Error("Failed to start cron scheduler", "error", err)
			log.Fatalf("Failed to start cron scheduler: %v", err)
		}

	default:
		logger.Error("Unknown mode", "mode", *mode)
		log.Fatalf("Unknown mode: %s. Valid modes are: api, worker, cron", *mode)
		os.Exit(1)
	}
}
