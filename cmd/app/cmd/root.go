// Package cmd provides Cobra command definitions for the application CLI.
// This package focuses on CLI concerns only and delegates application logic
// to the internal/app package, maintaining clean architecture principles.
package cmd

import (
	"fmt"
	"os"

	"go-clean-arch/internal/app"
	"go-clean-arch/pkg/config"
	"go-clean-arch/pkg/logger"

	"github.com/spf13/cobra"
)

var (
	cfgFile string
	cfg     *config.Config
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "app",
	Short: "A clean architecture Go application",
	Long: `A clean architecture Go application with multiple execution modes.

This application can run in three different modes:
- api: HTTP API server
- worker: Background message queue worker
- cron: Scheduled job runner

Use subcommands to run the application in different modes.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	// Config file is OPTIONAL - supports Twelve-Factor App (Factor III)
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file path (optional, defaults to env vars)")
}

// initConfig reads in config file and initializes logger.
// Config files are optional - application can run with only environment variables.
// This follows Twelve-Factor App methodology (Factor III - Config).
func initConfig() {
	var err error

	// If no config file specified, check if default exists
	if cfgFile == "" {
		// Try default path if it exists
		if _, err := os.Stat("configs/config.toml"); err == nil {
			cfgFile = "configs/config.toml"
		}
		// If default doesn't exist, that's OK - use env vars only
	}

	// Load configuration (file is optional, env vars will be used)
	cfg, err = config.Load(cfgFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	logger.Init(logger.Config{
		Level:  cfg.Logger.Level,
		Format: cfg.Logger.Format,
	})
}

// getDB initializes and returns a database connection.
// Returns interface{} to avoid direct gorm dependency in CMD layer,
// maintaining clean architecture by delegating to app layer.
func getDB() (interface{}, error) {
	return app.InitializeDatabase(cfg)
}

// getConfig returns the loaded configuration
func getConfig() *config.Config {
	return cfg
}
