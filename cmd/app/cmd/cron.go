package cmd

import (
	"fmt"

	"go-clean-arch/internal/app"

	"github.com/jmoiron/sqlx"
	"github.com/spf13/cobra"
)

// cronCmd represents the cron command
var cronCmd = &cobra.Command{
	Use:   "cron",
	Short: "Start the cron job scheduler",
	Long: `Start the cron job scheduler to run scheduled tasks.

The cron scheduler runs periodic jobs such as daily reports and cleanup tasks.
Jobs are configured using cron syntax in the configuration file.

Example:
  app cron
  app cron --config configs/config.toml`,
	RunE: runCron,
}

func init() {
	rootCmd.AddCommand(cronCmd)
}

func runCron(_ *cobra.Command, _ []string) error {
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

	// Create and start cron scheduler
	scheduler := app.NewCronScheduler(db, cfg)
	return scheduler.Start()
}
