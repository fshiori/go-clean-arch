// Package config provides application configuration management using Viper.
// It supports multiple configuration formats (TOML, YAML, JSON) and automatic
// environment variable overrides.
//
// Configuration files should be placed in the configs/ directory.
// The default configuration file is configs/config.toml.
//
// Environment variables can override any configuration value using the APP_ prefix.
// For example:
//   - APP_DATABASE_PASSWORD overrides database.password
//   - APP_LOGGER_LEVEL overrides logger.level
//   - APP_SERVER_PORT overrides server.port
//
// Usage:
//   cfg, err := config.Load("configs/config.toml")
//   if err != nil {
//       log.Fatal(err)
//   }
//   fmt.Println(cfg.Server.Port)
package config

import (
	"fmt"
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

// Load loads configuration from a file using Viper
// Supported formats: toml, yaml, json, etc.
func Load(path string) (*Config, error) {
	v := viper.New()

	// Set config file path
	v.SetConfigFile(path)

	// Enable automatic environment variable override
	// Environment variables should be prefixed with APP_
	// Example: APP_SERVER_PORT, APP_DATABASE_PASSWORD
	v.SetEnvPrefix("APP")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Set default values
	setDefaults(v)

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Unmarshal config
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
	v.SetDefault("database.auto_migrate", true)

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
