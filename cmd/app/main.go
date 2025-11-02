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
	configPath := flag.String("config", "configs/config.yaml", "Path to configuration file")
	flag.Parse()

	// Initialize logger
	logger.Init()
	log.Printf("Starting application in %s mode", *mode)

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

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
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Run auto-migration
	if cfg.Database.AutoMigrate {
		log.Println("Running database migrations...")
		if err := repository.AutoMigrate(db); err != nil {
			log.Fatalf("Failed to run migrations: %v", err)
		}
	}

	// Start the application based on mode
	switch *mode {
	case "api":
		apiServer := app.NewAPIServer(db, cfg.Server.Port)
		if err := apiServer.Start(); err != nil {
			log.Fatalf("Failed to start API server: %v", err)
		}

	case "worker":
		worker := app.NewWorker(db)
		if err := worker.Start(); err != nil {
			log.Fatalf("Failed to start worker: %v", err)
		}

	case "cron":
		cronScheduler := app.NewCronScheduler(db)
		if err := cronScheduler.Start(); err != nil {
			log.Fatalf("Failed to start cron scheduler: %v", err)
		}

	default:
		log.Fatalf("Unknown mode: %s. Valid modes are: api, worker, cron", *mode)
		os.Exit(1)
	}
}
