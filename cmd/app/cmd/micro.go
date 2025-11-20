package cmd

import (
	"fmt"

	"go-clean-arch/internal/app"

	"github.com/jmoiron/sqlx"
	"github.com/spf13/cobra"
)

var (
	microServiceName string
	microVersion     string
	microAddress     string
)

// microCmd represents the micro command
var microCmd = &cobra.Command{
	Use:   "micro",
	Short: "Start the go-micro microservice server",
	Long: `Start the go-micro microservice server to handle RPC requests.

The microservice provides RPC endpoints for user management operations.
It uses go-micro framework for service discovery and RPC handling.

Example:
  app micro
  app micro --config configs/config.toml
  app micro --service-name user.service --version v1.0.0 --address :8081`,
	RunE: runMicro,
}

func init() {
	rootCmd.AddCommand(microCmd)

	// Microservice-specific flags
	microCmd.Flags().StringVar(&microServiceName, "service-name", "go.micro.service.user", "service name for registration")
	microCmd.Flags().StringVar(&microVersion, "version", "latest", "service version")
	microCmd.Flags().StringVar(&microAddress, "address", ":8081", "microservice address")
}

func runMicro(cmd *cobra.Command, args []string) error {
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

	// Create and start microservice server
	server := app.NewMicroServer(db, cfg, microServiceName, microVersion, microAddress)
	return server.Start()
}
