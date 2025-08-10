package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"highperf-api/internal/config"
	"highperf-api/internal/logger"

	_ "github.com/lib/pq" // PostgreSQL driver
)

type DB struct {
	*sql.DB
	logger *logger.Logger
}

// Connect establishes a database connection with the given configuration
func Connect(cfg config.DatabaseConfig, log *logger.Logger) (*DB, error) {
	db, err := sql.Open(cfg.Driver, formatDSN(cfg))
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Info("database connection established",
		"driver", cfg.Driver,
		"host", cfg.Host,
		"port", cfg.Port,
		"database", cfg.Name,
	)

	return &DB{
		DB:     db,
		logger: log,
	}, nil
}

// Close closes the database connection
func (db *DB) Close() error {
	db.logger.Info("closing database connection")
	return db.DB.Close()
}

// Health checks the database connection health
func (db *DB) Health(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}

	return nil
}

// WithTx executes a function within a database transaction
func (db *DB) WithTx(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				db.logger.Error("failed to rollback transaction after panic",
					"panic", p,
					"rollback_error", rollbackErr.Error(),
				)
			}
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			db.logger.Error("failed to rollback transaction",
				"error", err.Error(),
				"rollback_error", rollbackErr.Error(),
			)
			return fmt.Errorf("transaction failed (rollback error: %v): %w", rollbackErr, err)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// formatDSN formats the database connection string
func formatDSN(cfg config.DatabaseConfig) string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Password,
		cfg.Name,
		cfg.SSLMode,
	)
}