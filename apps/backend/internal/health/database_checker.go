package health

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type DatabaseChecker struct {
	db *gorm.DB
}

func NewDatabaseChecker(db *gorm.DB) *DatabaseChecker {
	return &DatabaseChecker{db: db}
}

func (d *DatabaseChecker) Name() string {
	return "database"
}

func (d *DatabaseChecker) Check(ctx context.Context) CheckResult {
	start := time.Now()

	sqlDB, err := d.db.DB()
	if err != nil {
		return CheckResult{
			Status:  CheckFail,
			Message: "Failed to get database instance",
		}
	}

	// Use the context timeout from handler (AC1: 400ms max for total response)
	// Only add timeout if parent context has no deadline
	pingCtx := ctx
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		pingCtx, cancel = context.WithTimeout(ctx, 200*time.Millisecond)
		defer cancel()
	}

	if err := sqlDB.PingContext(pingCtx); err != nil {
		return CheckResult{
			Status:  CheckFail,
			Message: "Database connection failed",
		}
	}

	var result int
	if err := d.db.WithContext(pingCtx).Raw("SELECT 1").Scan(&result).Error; err != nil {
		return CheckResult{
			Status:  CheckFail,
			Message: "Database query failed",
		}
	}

	duration := time.Since(start)
	status := CheckPass
	message := "Database connection healthy"

	// AC1: Health check should respond within 500ms total
	// Individual check taking >300ms is a concern
	if duration > 300*time.Millisecond {
		status = CheckWarn
		message = "Database response time degraded"
	}

	return CheckResult{
		Status:       status,
		Message:      message,
		ResponseTime: fmt.Sprintf("%dms", duration.Milliseconds()),
	}
}
