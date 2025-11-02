package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	RabbitMQ RabbitMQConfig `yaml:"rabbitmq"`
	Cron     CronConfig     `yaml:"cron"`
	Stripe   StripeConfig   `yaml:"stripe"`
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Port int    `yaml:"port"`
	Host string `yaml:"host"`
	Mode string `yaml:"mode"` // debug, release, test
}

// DatabaseConfig holds database-related configuration
type DatabaseConfig struct {
	Driver      string `yaml:"driver"`       // postgres, mysql, sqlite
	Host        string `yaml:"host"`
	Port        int    `yaml:"port"`
	User        string `yaml:"user"`
	Password    string `yaml:"password"`
	DBName      string `yaml:"dbname"`
	SSLMode     string `yaml:"sslmode"`
	AutoMigrate bool   `yaml:"auto_migrate"`
}

// RabbitMQConfig holds RabbitMQ configuration
type RabbitMQConfig struct {
	URL       string `yaml:"url"`
	QueueName string `yaml:"queue_name"`
}

// CronConfig holds cron job configuration
type CronConfig struct {
	DailyReportSchedule string `yaml:"daily_report_schedule"`
}

// StripeConfig holds Stripe payment gateway configuration
type StripeConfig struct {
	APIKey    string `yaml:"api_key"`
	SecretKey string `yaml:"secret_key"`
}

// Load loads configuration from a YAML file
func Load(path string) (*Config, error) {
	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Override with environment variables if present
	overrideWithEnv(&config)

	return &config, nil
}

// overrideWithEnv overrides config values with environment variables
func overrideWithEnv(config *Config) {
	if port := os.Getenv("SERVER_PORT"); port != "" {
		// In production, parse string to int
		config.Server.Port = 8080
	}

	if dbHost := os.Getenv("DB_HOST"); dbHost != "" {
		config.Database.Host = dbHost
	}

	if dbPassword := os.Getenv("DB_PASSWORD"); dbPassword != "" {
		config.Database.Password = dbPassword
	}

	if stripeKey := os.Getenv("STRIPE_API_KEY"); stripeKey != "" {
		config.Stripe.APIKey = stripeKey
	}

	// Add more environment variable overrides as needed
}
