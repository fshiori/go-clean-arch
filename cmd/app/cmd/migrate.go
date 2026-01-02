package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"go-clean-arch/pkg/logger"

	"github.com/spf13/cobra"
)

// migrateCmd represents the migrate command using Atlas
var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Database migration management using Atlas (Twelve-Factor compliant admin process)",
	Long: `Run database migrations as one-off admin processes using Atlas.

Atlas is a modern database schema management tool that provides:
- Versioned migrations with automatic tracking
- Schema validation and safety checks
- Support for PostgreSQL, MySQL, SQLite, and more
- Declarative schema definitions

This command follows Twelve-Factor App methodology (Factor XII - Admin Processes):
- Runs as one-off process using the same codebase
- Uses the same environment configuration as main app
- Executes in the same environment as the application

Available subcommands:
  up     - Apply all pending migrations
  down   - Rollback the last migration
  status - Show current migration status
  create - Create a new migration file
  validate - Validate migration files

Examples:
  app migrate up
  app migrate down
  app migrate status
  app migrate create "add_users_table"
  app migrate validate

Environment Variables:
  DATABASE_URL - Database connection string (required)
    Format: postgres://user:pass@host:port/dbname?sslmode=disable
           mysql://user:pass@tcp(host:port)/dbname
           sqlite://file.db
  ATLAS_ENV    - Atlas environment to use (default: local)`,
}

var migrateUpCmd = &cobra.Command{
	Use:   "up",
	Short: "Apply all pending migrations using Atlas",
	Long: `Apply all pending database migrations using Atlas.

This will execute all SQL migration files in the migrations/ directory
that haven't been applied yet.

Atlas will:
- Validate migration files before applying
- Track applied migrations in the database
- Run migrations in a transaction (when supported)
- Provide detailed output of changes`,
	RunE: runMigrateUp,
}

var migrateDownCmd = &cobra.Command{
	Use:   "down [steps]",
	Short: "Rollback migrations using Atlas",
	Long: `Rollback one or more migrations using Atlas.

By default, rolls back 1 migration. You can specify the number of steps:
  app migrate down 2

WARNING: This operation cannot be undone. Make sure you have backups.`,
	RunE: runMigrateDown,
}

var migrateStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show migration status using Atlas",
	Long:  `Display which migrations have been applied and which are pending using Atlas.`,
	RunE:  runMigrateStatus,
}

var migrateCreateCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "Create a new migration file using Atlas",
	Long: `Create a new SQL migration file with the given name using Atlas.

Example:
  app migrate create "add_orders_table"

This will create:
  migrations/YYYYMMDDHHMMSS_add_orders_table.sql

Atlas uses timestamps instead of sequential numbers for better
conflict resolution in team environments.`,
	Args: cobra.ExactArgs(1),
	RunE: runMigrateCreate,
}

var migrateValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate migration files using Atlas",
	Long: `Validate migration files for correctness using Atlas.

This will check:
- SQL syntax errors
- Migration file naming and ordering
- Schema consistency`,
	RunE: runMigrateValidate,
}

func init() {
	rootCmd.AddCommand(migrateCmd)
	migrateCmd.AddCommand(migrateUpCmd)
	migrateCmd.AddCommand(migrateDownCmd)
	migrateCmd.AddCommand(migrateStatusCmd)
	migrateCmd.AddCommand(migrateCreateCmd)
	migrateCmd.AddCommand(migrateValidateCmd)
}

func runMigrateUp(_ *cobra.Command, _ []string) error {
	logger.Info("Running database migrations using Atlas...")

	// Check if Atlas is installed
	if err := checkAtlasInstalled(); err != nil {
		return err
	}

	// Get DATABASE_URL from environment or config
	dbURL, err := getDatabaseURL()
	if err != nil {
		return err
	}

	// Get Atlas environment (default: local)
	atlasEnv := getAtlasEnv()

	logger.Info("Applying migrations", "environment", atlasEnv)

	// Run Atlas migrate apply
	cmd := exec.Command("atlas", "migrate", "apply",
		"--env", atlasEnv,
		"--url", dbURL,
	)

	// Set environment variables
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("DATABASE_URL=%s", dbURL),
	)

	// Capture output
	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Error("Failed to apply migrations", "error", err, "output", string(output))
		return fmt.Errorf("failed to apply migrations: %w\n%s", err, string(output))
	}

	// Print output
	fmt.Println(string(output))
	logger.Info("Migrations applied successfully")

	return nil
}

func runMigrateDown(cmd *cobra.Command, args []string) error {
	logger.Info("Rolling back migrations using Atlas...")

	// Check if Atlas is installed
	if err := checkAtlasInstalled(); err != nil {
		return err
	}

	// Get DATABASE_URL from environment or config
	dbURL, err := getDatabaseURL()
	if err != nil {
		return err
	}

	// Get Atlas environment (default: local)
	atlasEnv := getAtlasEnv()

	// Determine number of steps to rollback (default: 1)
	steps := "1"
	if len(args) > 0 {
		steps = args[0]
	}

	logger.Info("Rolling back migrations", "environment", atlasEnv, "steps", steps)
	logger.Warn("WARNING: This operation cannot be undone. Make sure you have backups.")

	// Run Atlas migrate down
	atlasCmd := exec.Command("atlas", "migrate", "down",
		"--env", atlasEnv,
		"--url", dbURL,
		steps,
	)

	// Set environment variables
	atlasCmd.Env = append(os.Environ(),
		fmt.Sprintf("DATABASE_URL=%s", dbURL),
	)

	// Capture output
	output, execErr := atlasCmd.CombinedOutput()
	if execErr != nil {
		logger.Error("Failed to rollback migrations", "error", execErr, "output", string(output))
		return fmt.Errorf("failed to rollback migrations: %w\n%s", execErr, string(output))
	}

	// Print output
	fmt.Println(string(output))
	logger.Info("Migrations rolled back successfully")

	return nil
}

func runMigrateStatus(_ *cobra.Command, _ []string) error {
	logger.Info("Checking migration status using Atlas...")

	// Check if Atlas is installed
	if err := checkAtlasInstalled(); err != nil {
		return err
	}

	// Get DATABASE_URL from environment or config
	dbURL, err := getDatabaseURL()
	if err != nil {
		return err
	}

	// Get Atlas environment (default: local)
	atlasEnv := getAtlasEnv()

	logger.Info("Checking status", "environment", atlasEnv)

	// Run Atlas migrate status
	cmd := exec.Command("atlas", "migrate", "status",
		"--env", atlasEnv,
		"--url", dbURL,
	)

	// Set environment variables
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("DATABASE_URL=%s", dbURL),
	)

	// Capture output
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Atlas might return non-zero exit code if there are pending migrations
		// But we still want to show the status
		fmt.Println(string(output))
		return nil
	}

	// Print output
	fmt.Println(string(output))

	return nil
}

func runMigrateCreate(_ *cobra.Command, args []string) error {
	migrationName := args[0]

	logger.Info("Creating new migration file using Atlas...", "name", migrationName)

	// Check if Atlas is installed
	if err := checkAtlasInstalled(); err != nil {
		return err
	}

	// Get Atlas environment (default: local)
	atlasEnv := getAtlasEnv()

	// Run Atlas migrate new
	cmd := exec.Command("atlas", "migrate", "new",
		"--env", atlasEnv,
		migrationName,
	)

	// Capture output
	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Error("Failed to create migration file", "error", err, "output", string(output))
		return fmt.Errorf("failed to create migration file: %w\n%s", err, string(output))
	}

	// Print output
	fmt.Println(string(output))
	logger.Info("Migration file created successfully")
	fmt.Println("\nEdit the file to add your SQL migration, then run:")
	fmt.Println("  app migrate up")

	return nil
}

func runMigrateValidate(_ *cobra.Command, _ []string) error {
	logger.Info("Validating migration files using Atlas...")

	// Check if Atlas is installed
	if err := checkAtlasInstalled(); err != nil {
		return err
	}

	// Get DATABASE_URL from environment or config
	dbURL, err := getDatabaseURL()
	if err != nil {
		return err
	}

	// Get Atlas environment (default: local)
	atlasEnv := getAtlasEnv()

	logger.Info("Validating migrations", "environment", atlasEnv)

	// Run Atlas migrate validate
	cmd := exec.Command("atlas", "migrate", "validate",
		"--env", atlasEnv,
		"--url", dbURL,
	)

	// Set environment variables
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("DATABASE_URL=%s", dbURL),
	)

	// Capture output
	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Error("Migration validation failed", "error", err, "output", string(output))
		return fmt.Errorf("migration validation failed: %w\n%s", err, string(output))
	}

	// Print output
	fmt.Println(string(output))
	logger.Info("Migration validation successful")

	return nil
}

// Helper functions

// checkAtlasInstalled checks if Atlas CLI is installed on the system
func checkAtlasInstalled() error {
	cmd := exec.Command("atlas", "version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf(`Atlas CLI is not installed. Please install it first:

macOS:
  brew install ariga/tap/atlas

Linux:
  curl -sSf https://atlasgo.sh | sh

Windows:
  Download from https://release.ariga.io/atlas/atlas-windows-amd64-latest.exe

Or use Docker:
  docker pull arigaio/atlas

For more information, visit: https://atlasgo.io/getting-started/

Error: %w
Output: %s`, err, string(output))
	}

	// Log Atlas version
	logger.Info("Atlas CLI found", "version", strings.TrimSpace(string(output)))
	return nil
}

// getDatabaseURL gets the database connection URL from environment or config
func getDatabaseURL() (string, error) {
	// Try DATABASE_URL environment variable first (Twelve-Factor compliant)
	if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
		return dbURL, nil
	}

	// Try APP_DATABASE_URL as an alternative
	if dbURL := os.Getenv("APP_DATABASE_URL"); dbURL != "" {
		return dbURL, nil
	}

	// Build from individual config values
	cfg := getConfig()
	if cfg == nil {
		return "", fmt.Errorf("configuration not loaded")
	}

	// Build connection string based on driver
	var dbURL string
	switch cfg.Database.Driver {
	case "postgres":
		dbURL = fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
			cfg.Database.User,
			cfg.Database.Password,
			cfg.Database.Host,
			cfg.Database.Port,
			cfg.Database.DBName,
			cfg.Database.SSLMode,
		)
	case "mysql":
		dbURL = fmt.Sprintf("mysql://%s:%s@tcp(%s:%d)/%s",
			cfg.Database.User,
			cfg.Database.Password,
			cfg.Database.Host,
			cfg.Database.Port,
			cfg.Database.DBName,
		)
	case "sqlite":
		dbURL = fmt.Sprintf("sqlite://%s", cfg.Database.DBName)
	default:
		return "", fmt.Errorf("unsupported database driver: %s", cfg.Database.Driver)
	}

	return dbURL, nil
}

// getAtlasEnv returns the Atlas environment to use (default: local)
func getAtlasEnv() string {
	if env := os.Getenv("ATLAS_ENV"); env != "" {
		return env
	}
	return "local"
}
