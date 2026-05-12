package db

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"gorm.io/gorm"
)

// Ping checks if database connection is alive
// Story 2.4: Database health check for /health/db endpoint
// Returns error if database is unreachable, nil if healthy
func Ping(db *gorm.DB) error {
	return PingContext(context.Background(), db)
}

// PingContext checks if database connection is alive with context timeout
// Story 2.4: Database health check with timeout support
func PingContext(ctx context.Context, db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("database connection is nil")
	}

	sqlDB, err := db.DB()
	if err != nil {
		slog.Error("Failed to get sql.DB from gorm DB", "error", err)
		return fmt.Errorf("failed to get database instance: %w", err)
	}

	// Use context with timeout for health check
	pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(pingCtx); err != nil {
		slog.Error("Database ping failed", "error", err)
		return fmt.Errorf("database ping failed: %w", err)
	}

	slog.Debug("Database health check passed")
	return nil
}
