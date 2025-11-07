package cmd

import (
	"fmt"

	"go-clean-arch/internal/app"

	"github.com/jmoiron/sqlx"
	"github.com/spf13/cobra"
)

// apiCmd represents the api command
var apiCmd = &cobra.Command{
	Use:   "api",
	Short: "Start the HTTP API server",
	Long: `Start the HTTP API server to handle web requests.

The API server provides RESTful endpoints for user management and other operations.
It uses Gin as the HTTP framework and includes request logging and trace ID tracking.

Example:
  app api
  app api --config configs/config.toml`,
	RunE: runAPI,
}

func init() {
	rootCmd.AddCommand(apiCmd)
}

func runAPI(cmd *cobra.Command, args []string) error {
	// Get database connection
	dbInterface, err := getDB()
	if err != nil {
		return err
	}

	// Type assert to concrete type
	db, ok := dbInterface.(*sqlx.DB)
	if !ok {
		return fmt.Errorf("invalid database type")
	}

	// Get configuration
	cfg := getConfig()

	// Create and start API server
	server := app.NewAPIServer(db, cfg, cfg.Server.Port)
	return server.Start()
}
