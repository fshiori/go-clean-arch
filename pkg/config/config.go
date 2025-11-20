// Package config provides application configuration management using Viper.
// It supports multiple configuration formats (TOML, YAML, JSON) and automatic
// environment variable overrides.
//
// Configuration files are OPTIONAL. The application can run with only environment
// variables and defaults, supporting fully Twelve-Factor App compliant deployments.
//
// Environment variables can override any configuration value using the APP_ prefix.
// For example:
//   - APP_DATABASE_PASSWORD overrides database.password
//   - APP_LOGGER_LEVEL overrides logger.level
//   - APP_SERVER_PORT overrides server.port
//
// Usage with config file (optional):
//   cfg, err := config.Load("configs/config.toml")
//
// Usage with environment variables only:
//   cfg, err := config.Load("")  // Empty path = env vars + defaults only
//
// Example environment-only deployment:
//   export APP_SERVER_PORT=8080
//   export APP_DATABASE_HOST=postgres.example.com
//   export APP_DATABASE_USER=myuser
//   export APP_DATABASE_PASSWORD=secret
//   ./app api
package config

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/viper"
)

// Config represents the application configuration
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	RabbitMQ RabbitMQConfig `mapstructure:"rabbitmq"`
	Cron     CronConfig     `mapstructure:"cron"`
	Stripe   StripeConfig   `mapstructure:"stripe"`
	Logger   LoggerConfig   `mapstructure:"logger"`
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Host string `mapstructure:"host"`
	Mode string `mapstructure:"mode"` // debug, release, test
}

// DatabaseConfig holds database-related configuration
type DatabaseConfig struct {
	Driver      string `mapstructure:"driver"`       // postgres, mysql, sqlite
	Host        string `mapstructure:"host"`
	Port        int    `mapstructure:"port"`
	User        string `mapstructure:"user"`
	Password    string `mapstructure:"password"`
	DBName      string `mapstructure:"dbname"`
	SSLMode     string `mapstructure:"sslmode"`
	AutoMigrate bool   `mapstructure:"auto_migrate"`
}

// RabbitMQConfig holds RabbitMQ configuration
type RabbitMQConfig struct {
	URL       string `mapstructure:"url"`
	QueueName string `mapstructure:"queue_name"`
}

// CronConfig holds cron job configuration
type CronConfig struct {
	DailyReportSchedule string `mapstructure:"daily_report_schedule"`
}

// StripeConfig holds Stripe payment gateway configuration
type StripeConfig struct {
	APIKey    string `mapstructure:"api_key"`
	SecretKey string `mapstructure:"secret_key"`
}

// LoggerConfig holds logger configuration
type LoggerConfig struct {
	Level  string `mapstructure:"level"`  // debug, info, warn, error
	Format string `mapstructure:"format"` // json, text
}

// Load loads configuration from a file using Viper.
// The config file is OPTIONAL. If path is empty or file doesn't exist,
// configuration will be loaded from environment variables and defaults only.
//
// Supported formats: toml, yaml, json, etc.
//
// Twelve-Factor App Compliance:
// This implementation follows Factor III (Config) by supporting environment-only
// configuration. Config files are provided for local development convenience only.
func Load(path string) (*Config, error) {
	v := viper.New()

	// Set default values first (before file and env vars)
	setDefaults(v)

	// Enable automatic environment variable override
	// Environment variables should be prefixed with APP_
	// Example: APP_SERVER_PORT, APP_DATABASE_PASSWORD
	v.SetEnvPrefix("APP")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Try to read config file (optional - non-fatal if missing)
	if path != "" {
		v.SetConfigFile(path)

		if err := v.ReadInConfig(); err != nil {
			// Check if file exists
			if _, statErr := os.Stat(path); os.IsNotExist(statErr) {
				// File doesn't exist - this is OK, use env vars and defaults
				log.Printf("[Config] Config file not found at %s, using environment variables and defaults", path)
			} else {
				// File exists but couldn't be read - this might be a problem
				log.Printf("[Config] Warning: Config file exists but couldn't be read: %v. Using environment variables and defaults", err)
			}
		} else {
			log.Printf("[Config] Loaded configuration from file: %s", path)
		}
	} else {
		log.Printf("[Config] No config file specified, using environment variables and defaults (Twelve-Factor compliant)")
	}

	// Unmarshal config (from defaults, file if present, and env vars)
	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate config
	if err := validate(&config); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &config, nil
}

// setDefaults sets default configuration values
func setDefaults(v *viper.Viper) {
	// Server defaults
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.mode", "debug")

	// Database defaults
	v.SetDefault("database.driver", "postgres")
	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", 5432)
	v.SetDefault("database.sslmode", "disable")
	// Auto-migrate disabled by default for production safety
	// Use explicit migration commands instead (Factor XII - Admin Processes)
	v.SetDefault("database.auto_migrate", false)

	// RabbitMQ defaults
	v.SetDefault("rabbitmq.url", "amqp://guest:guest@localhost:5672/")
	v.SetDefault("rabbitmq.queue_name", "order_queue")

	// Cron defaults
	v.SetDefault("cron.daily_report_schedule", "0 0 * * *")

	// Logger defaults
	v.SetDefault("logger.level", "info")
	v.SetDefault("logger.format", "json")
}

// validate validates the configuration
func validate(config *Config) error {
	// Validate server config
	if config.Server.Port < 1 || config.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", config.Server.Port)
	}

	// Validate database config
	if config.Database.Driver == "" {
		return fmt.Errorf("database driver is required")
	}

	// Validate logger config
	validLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
	if !validLevels[config.Logger.Level] {
		return fmt.Errorf("invalid logger level: %s (must be debug, info, warn, or error)", config.Logger.Level)
	}

	validFormats := map[string]bool{"json": true, "text": true}
	if !validFormats[config.Logger.Format] {
		return fmt.Errorf("invalid logger format: %s (must be json or text)", config.Logger.Format)
	}

	return nil
}
