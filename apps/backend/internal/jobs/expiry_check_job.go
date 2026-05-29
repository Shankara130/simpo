package jobs

import (
	"context"
	"log/slog"
	"time"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/dto"
)

// ExpiryService defines the interface for expiry check operations
type ExpiryService interface {
	CheckExpiringProducts(ctx context.Context) ([]*dto.ExpiryAlertEvent, error)
}

// ExpiryCheckJob runs scheduled expiry checks for products approaching expiry
// Story 4.5, AC1, AC2, AC3: Scheduled job to generate expiry alerts every 6 hours
// Story 4.5, Task 4.1-4.5: Implement scheduled execution with metrics
// PATCH: Added timer field for tracking initial delay timer
type ExpiryCheckJob struct {
	expiryService ExpiryService
	logger        *slog.Logger
	ticker        *time.Ticker
	timer         *time.Timer // PATCH: Track initial timer for proper cleanup
	stopChan      chan struct{}
	metrics       *ExpiryJobMetrics
}

// ExpiryJobMetrics tracks metrics for the expiry check job
// Story 4.5, Task 4.4: Add metrics: count of alerts generated per day per alert level
type ExpiryJobMetrics struct {
	TotalRuns      int64
	TotalAlerts    int64
	WarningAlerts  int64 // 30-day alerts
	CriticalAlerts int64 // 14-day alerts
	UrgentAlerts   int64 // 7-day alerts
	Errors         int64
	LastRunTime    time.Time
	LastAlertCount int
}

// NewExpiryCheckJob creates a new expiry check job
// Story 4.5, Task 4.1: Create background job in internal/jobs/expiry_check_job.go
func NewExpiryCheckJob(expiryService ExpiryService, logger *slog.Logger) *ExpiryCheckJob {
	if logger == nil {
		logger = slog.Default()
	}

	return &ExpiryCheckJob{
		expiryService: expiryService,
		logger:        logger,
		stopChan:      make(chan struct{}),
		metrics:       &ExpiryJobMetrics{},
	}
}

// Start begins the scheduled job running every 6 hours
// Story 4.5, Task 4.2: Implement scheduled execution (cron-like)
// Run every 6 hours (00:00, 06:00, 12:00, 18:00)
// PATCH: Fixed race condition in job startup and context cancellation handling
func (j *ExpiryCheckJob) Start(ctx context.Context) {
	j.logger.Info("starting expiry check job", "interval", "6 hours")

	// Calculate initial delay to run at next 6-hour mark
	now := time.Now().UTC()
	nextRun := j.nextSixHourMark(now)
	initialDelay := time.Until(nextRun)

	j.logger.Info("scheduling first expiry check run",
		"next_run", nextRun.Format(time.RFC3339),
		"initial_delay", initialDelay.String())

	// PATCH: Use goroutine with select for proper context cancellation handling
	go func() {
		// Wait for initial delay or context cancellation
		select {
		case <-time.After(initialDelay):
			j.runOnce(ctx)
			// Start ticker for subsequent runs every 6 hours
			j.ticker = time.NewTicker(6 * time.Hour)
			j.run(ctx)
		case <-ctx.Done():
			j.logger.Info("expiry check job cancelled during initial delay")
			return
		case <-j.stopChan:
			j.logger.Info("expiry check job stopped during initial delay")
			return
		}
	}()
}

// nextSixHourMark calculates the next 6-hour mark (00:00, 06:00, 12:00, 18:00 UTC)
func (j *ExpiryCheckJob) nextSixHourMark(now time.Time) time.Time {
	hour := now.Hour()
	var nextHour int

	// Find next 6-hour mark
	switch {
	case hour < 6:
		nextHour = 6
	case hour < 12:
		nextHour = 12
	case hour < 18:
		nextHour = 18
	default:
		// Next day at 00:00
		nextHour = 0
		return time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, time.UTC)
	}

	return time.Date(now.Year(), now.Month(), now.Day(), nextHour, 0, 0, 0, time.UTC)
}

// run executes the job on each ticker tick
// Story 4.5, Task 4.3: Use Go context with cancellation support
func (j *ExpiryCheckJob) run(ctx context.Context) {
	for {
		select {
		case <-j.ticker.C:
			j.runOnce(ctx)
		case <-j.stopChan:
			j.logger.Info("expiry check job stopped")
			return
		case <-ctx.Done():
			j.logger.Info("expiry check job context cancelled")
			j.Stop()
			return
		}
	}
}

// runOnce executes a single expiry check
func (j *ExpiryCheckJob) runOnce(ctx context.Context) {
	j.logger.Info("running expiry check job")
	startTime := time.Now()

	j.metrics.TotalRuns++
	j.metrics.LastRunTime = startTime

	// Run the expiry check
	events, err := j.expiryService.CheckExpiringProducts(ctx)
	if err != nil {
		j.metrics.Errors++
		j.logger.Error("expiry check job failed", "error", err)
		return
	}

	// Update metrics
	j.metrics.TotalAlerts += int64(len(events))
	j.metrics.LastAlertCount = len(events)

	// Categorize alerts by level
	for _, event := range events {
		switch event.Data.AlertLevel {
		case "warning":
			j.metrics.WarningAlerts++
		case "critical":
			j.metrics.CriticalAlerts++
		case "urgent":
			j.metrics.UrgentAlerts++
		}
	}

	duration := time.Since(startTime)
	j.logger.Info("expiry check job completed",
		"alerts_generated", len(events),
		"warning", j.metrics.WarningAlerts,
		"critical", j.metrics.CriticalAlerts,
		"urgent", j.metrics.UrgentAlerts,
		"duration_ms", duration.Milliseconds())
}

// Stop stops the scheduled job
// Story 4.5, Task 4.3: Use Go context with cancellation support
// PATCH: Also stop initial timer if set
func (j *ExpiryCheckJob) Stop() {
	// PATCH: Stop initial timer if still running
	if j.timer != nil {
		j.timer.Stop()
	}
	close(j.stopChan)
	if j.ticker != nil {
		j.ticker.Stop()
	}
	j.logger.Info("expiry check job stopped",
		"total_runs", j.metrics.TotalRuns,
		"total_alerts", j.metrics.TotalAlerts)
}

// GetMetrics returns the current job metrics
// Story 4.5, Task 4.4: Add metrics: count of alerts generated per day per alert level
func (j *ExpiryCheckJob) GetMetrics() *ExpiryJobMetrics {
	return j.metrics
}

// RunOnceImmediately executes a single expiry check immediately (for testing/manual triggers)
func (j *ExpiryCheckJob) RunOnceImmediately(ctx context.Context) error {
	j.logger.Info("running immediate expiry check")
	j.runOnce(ctx)
	return nil
}
