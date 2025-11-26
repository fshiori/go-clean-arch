package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"go-clean-arch/pkg/logger"

	"github.com/jmoiron/sqlx"
	"github.com/spf13/cobra"
)

// migrateCmd represents the migrate command
var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Database migration management (Twelve-Factor compliant admin process)",
	Long: `Run database migrations as one-off admin processes.

This command follows Twelve-Factor App methodology (Factor XII - Admin Processes):
- Runs as one-off process using the same codebase
- Uses the same environment configuration as main app
- Executes in the same environment as the application

Available subcommands:
  up     - Apply all pending migrations
  down   - Rollback the last migration
  status - Show current migration status
  create - Create a new migration file

Examples:
  app migrate up
  app migrate down
  app migrate status
  app migrate create "add_users_table"`,
}

var migrateUpCmd = &cobra.Command{
	Use:   "up",
	Short: "Apply all pending migrations",
	Long: `Apply all pending database migrations.

This will execute all SQL migration files in the migrations/ directory
that haven't been applied yet.`,
	RunE: runMigrateUp,
}

var migrateDownCmd = &cobra.Command{
	Use:   "down",
	Short: "Rollback the last migration",
	Long: `Rollback the most recently applied migration.

WARNING: This operation cannot be undone. Make sure you have backups.`,
	RunE: runMigrateDown,
}

var migrateStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show migration status",
	Long:  `Display which migrations have been applied and which are pending.`,
	RunE:  runMigrateStatus,
}

var migrateCreateCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "Create a new migration file",
	Long: `Create a new SQL migration file with the given name.

Example:
  app migrate create "add_orders_table"

This will create:
  migrations/003_add_orders_table.sql`,
	Args: cobra.ExactArgs(1),
	RunE: runMigrateCreate,
}

func init() {
	rootCmd.AddCommand(migrateCmd)
	migrateCmd.AddCommand(migrateUpCmd)
	migrateCmd.AddCommand(migrateDownCmd)
	migrateCmd.AddCommand(migrateStatusCmd)
	migrateCmd.AddCommand(migrateCreateCmd)
}

func runMigrateUp(_ *cobra.Command, _ []string) error {
	logger.Info("Running database migrations...")

	// Get database connection
	dbInterface, err := getDB()
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	db, ok := dbInterface.(*sqlx.DB)
	if !ok {
		return fmt.Errorf("invalid database type")
	}
	defer func() {
		if err := db.Close(); err != nil {
			logger.Error("Failed to close database connection", "error", err)
		}
	}()

	// Create migrations table if it doesn't exist
	if err := createMigrationsTable(db); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get list of applied migrations
	appliedMigrations, err := getAppliedMigrations(db)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Get list of migration files
	migrationFiles, err := getMigrationFiles()
	if err != nil {
		return fmt.Errorf("failed to read migration files: %w", err)
	}

	// Apply pending migrations
	appliedCount := 0
	for _, file := range migrationFiles {
		if _, applied := appliedMigrations[file]; !applied {
			logger.Info("Applying migration", "file", file)
			if err := applyMigration(db, file); err != nil {
				return fmt.Errorf("failed to apply migration %s: %w", file, err)
			}
			appliedCount++
		}
	}

	if appliedCount == 0 {
		logger.Info("No pending migrations")
	} else {
		logger.Info("Migrations applied successfully", "count", appliedCount)
	}

	return nil
}

func runMigrateDown(_ *cobra.Command, _ []string) error {
	logger.Info("Rolling back last migration...")

	// Get database connection
	dbInterface, err := getDB()
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	db, ok := dbInterface.(*sqlx.DB)
	if !ok {
		return fmt.Errorf("invalid database type")
	}
	defer func() {
		if err := db.Close(); err != nil {
			logger.Error("Failed to close database connection", "error", err)
		}
	}()

	// Get last applied migration
	var lastMigration string
	err = db.Get(&lastMigration, "SELECT filename FROM schema_migrations ORDER BY applied_at DESC LIMIT 1")
	if err != nil {
		return fmt.Errorf("no migrations to rollback")
	}

	logger.Info("Rolling back migration", "file", lastMigration)

	// Delete from migrations table
	_, err = db.Exec("DELETE FROM schema_migrations WHERE filename = $1", lastMigration)
	if err != nil {
		return fmt.Errorf("failed to rollback migration: %w", err)
	}

	logger.Info("Migration rolled back successfully")
	logger.Warn("Note: SQL changes were NOT reverted. You may need to manually undo schema changes.")

	return nil
}

func runMigrateStatus(_ *cobra.Command, _ []string) error {
	// Get database connection
	dbInterface, err := getDB()
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	db, ok := dbInterface.(*sqlx.DB)
	if !ok {
		return fmt.Errorf("invalid database type")
	}
	defer func() {
		if err := db.Close(); err != nil {
			logger.Error("Failed to close database connection", "error", err)
		}
	}()

	// Create migrations table if it doesn't exist
	if err := createMigrationsTable(db); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get applied migrations
	appliedMigrations, err := getAppliedMigrations(db)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Get migration files
	migrationFiles, err := getMigrationFiles()
	if err != nil {
		return fmt.Errorf("failed to read migration files: %w", err)
	}

	fmt.Println("\nMigration Status:")
	fmt.Println("================")

	if len(migrationFiles) == 0 {
		fmt.Println("No migration files found")
		return nil
	}

	pendingCount := 0
	for _, file := range migrationFiles {
		if _, applied := appliedMigrations[file]; applied {
			fmt.Printf("✅ %s (applied)\n", file)
		} else {
			fmt.Printf("⏳ %s (pending)\n", file)
			pendingCount++
		}
	}

	fmt.Printf("\nTotal: %d migrations (%d applied, %d pending)\n",
		len(migrationFiles),
		len(appliedMigrations),
		pendingCount)

	return nil
}

func runMigrateCreate(_ *cobra.Command, args []string) error {
	migrationName := args[0]

	// Get next migration number
	files, err := getMigrationFiles()
	if err != nil {
		return fmt.Errorf("failed to read migration files: %w", err)
	}

	nextNumber := len(files) + 1

	// Create migration filename
	filename := fmt.Sprintf("%03d_%s.sql", nextNumber, migrationName)
	filepath := filepath.Join("migrations", filename)

	// Create migration file
	content := fmt.Sprintf(`-- Migration: %s
-- Created: %s
--
-- Add your SQL migration here
-- Example:
-- CREATE TABLE example (
--     id SERIAL PRIMARY KEY,
--     name VARCHAR(255) NOT NULL,
--     created_at TIMESTAMP DEFAULT NOW()
-- );

`, migrationName, "now")

	if err := os.WriteFile(filepath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to create migration file: %w", err)
	}

	logger.Info("Migration file created", "file", filepath)
	fmt.Printf("✅ Created: %s\n", filepath)
	fmt.Println("\nEdit this file to add your SQL migration, then run:")
	fmt.Println("  app migrate up")

	return nil
}

// Helper functions

func createMigrationsTable(db *sqlx.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			id SERIAL PRIMARY KEY,
			filename VARCHAR(255) NOT NULL UNIQUE,
			applied_at TIMESTAMP DEFAULT NOW()
		)
	`
	_, err := db.Exec(query)
	return err
}

func getAppliedMigrations(db *sqlx.DB) (map[string]bool, error) {
	var migrations []string
	err := db.Select(&migrations, "SELECT filename FROM schema_migrations ORDER BY applied_at")
	if err != nil {
		return nil, err
	}

	result := make(map[string]bool)
	for _, m := range migrations {
		result[m] = true
	}
	return result, nil
}

func getMigrationFiles() ([]string, error) {
	files, err := os.ReadDir("migrations")
	if err != nil {
		return nil, err
	}

	var migrations []string
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".sql" && file.Name() != "README.md" {
			migrations = append(migrations, file.Name())
		}
	}

	return migrations, nil
}

func applyMigration(db *sqlx.DB, filename string) error {
	// Read migration file
	content, err := os.ReadFile(filepath.Join("migrations", filename))
	if err != nil {
		return fmt.Errorf("failed to read migration file: %w", err)
	}

	// Begin transaction
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Execute migration
	if _, err := tx.Exec(string(content)); err != nil {
		return fmt.Errorf("failed to execute migration: %w", err)
	}

	// Record migration
	if _, err := tx.Exec("INSERT INTO schema_migrations (filename) VALUES ($1)", filename); err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
