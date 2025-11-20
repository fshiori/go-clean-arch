package cmd

import (
	"fmt"

	"go-clean-arch/internal/app"

	"github.com/jmoiron/sqlx"
	"github.com/spf13/cobra"
)

// workerCmd represents the worker command
var workerCmd = &cobra.Command{
	Use:   "worker",
	Short: "Start the background worker",
	Long: `Start the background worker to process messages from the queue.

The worker consumes messages from RabbitMQ and processes order-related events
such as shipping, completion, and cancellation.

Example:
  app worker
  app worker --config configs/config.toml`,
	RunE: runWorker,
}

func init() {
	rootCmd.AddCommand(workerCmd)
}

func runWorker(cmd *cobra.Command, args []string) error {
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

	// Create and start worker
	worker := app.NewWorker(db, cfg)
	return worker.Start()
}
