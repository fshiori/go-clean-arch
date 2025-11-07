package repository

import (
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql" // MySQL driver
	"github.com/jmoiron/sqlx"
)

// DBConfig holds database configuration
type DBConfig struct {
	Driver   string // mysql
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// NewDatabase creates a new database connection using sqlx for MySQL
func NewDatabase(config DBConfig) (*sqlx.DB, error) {
	if config.Driver != "mysql" {
		return nil, fmt.Errorf("unsupported database driver: %s (only mysql is supported)", config.Driver)
	}

	// MySQL DSN format: user:password@tcp(host:port)/dbname?parseTime=true&charset=utf8mb4
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.User, config.Password, config.Host, config.Port, config.DBName,
	)

	db, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxIdleConns(10)
	db.SetMaxOpenConns(100)
	db.SetConnMaxLifetime(time.Hour)

	return db, nil
}
