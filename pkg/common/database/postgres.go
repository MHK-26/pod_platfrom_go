// pkg/common/database/postgres.go
package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // PostgreSQL driver
	"github.com/MHK-26/pod_platfrom_go/pkg/common/config"
)

// NewPostgresDB creates a new PostgreSQL connection
func NewPostgresDB(cfg *config.DBConfig) (*sqlx.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, err
	}

	// Set connection pool settings
	db.SetMaxOpenConns(cfg.MaxConns)
	db.SetMaxIdleConns(cfg.MaxIdle)
	db.SetConnMaxLifetime(cfg.Timeout)

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

// MigrateDatabase runs database migrations
func MigrateDatabase(db *sqlx.DB) error {
	// In a real application, you would use a migration tool like golang-migrate
	// For simplicity, we'll just print a message here
	fmt.Println("Database migrations would be run here.")
	return nil
}

// CloseDB closes the database connection
func CloseDB(db *sqlx.DB) error {
	return db.Close()
}

// WithTransaction executes a function within a transaction
func WithTransaction(db *sqlx.DB, fn func(*sqlx.Tx) error) error {
	tx, err := db.Beginx()
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p) // Re-throw panic after rollback
		}
	}()

	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

// PostgresHealthCheck checks if the database is healthy
func PostgresHealthCheck(db *sqlx.DB) error {
	// Set a short timeout to detect when the database is unreachable
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	return db.PingContext(ctx)
}