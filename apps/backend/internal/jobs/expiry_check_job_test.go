package jobs

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/dto"
)

// MockExpiryCheckService is a mock for ExpiryCheckService
type MockExpiryCheckService struct {
	mock.Mock
}

func (m *MockExpiryCheckService) CheckExpiringProducts(ctx context.Context) ([]*dto.ExpiryAlertEvent, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*dto.ExpiryAlertEvent), args.Error(1)
}

func (m *MockExpiryCheckService) PublishExpiryAlert(ctx context.Context, event *dto.ExpiryAlertEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockExpiryCheckService) ClearExpiryAlertState(ctx context.Context, productID uint, branchID uint) error {
	args := m.Called(ctx, productID, branchID)
	return args.Error(0)
}

// TestNewExpiryCheckJob tests job creation
// Story 4.5, Task 4.1: Create background job
func TestNewExpiryCheckJob(t *testing.T) {
	mockService := new(MockExpiryCheckService)
	logger := slog.Default()

	job := NewExpiryCheckJob(mockService, logger)

	assert.NotNil(t, job, "Job should be created")
	assert.NotNil(t, job.logger, "Logger should be set")
	assert.NotNil(t, job.stopChan, "Stop channel should be created")
	assert.NotNil(t, job.metrics, "Metrics should be initialized")
	assert.Equal(t, int64(0), job.metrics.TotalRuns, "Initial total runs should be 0")
}

// TestExpiryCheckJob_RunOnceImmediately tests immediate execution
// Story 4.5, Task 4.2: Implement scheduled execution
func TestExpiryCheckJob_RunOnceImmediately(t *testing.T) {
	mockService := new(MockExpiryCheckService)
	logger := slog.Default()
	ctx := context.Background()

	now := time.Now().UTC()
	day30 := now.AddDate(0, 0, 30)

	events := []*dto.ExpiryAlertEvent{
		{
			EventID:   "evt-1",
			EventType: "product.expiry",
			Timestamp: now.Format(time.RFC3339),
			Data: dto.ProductExpiryData{
				ProductID:     1,
				SKU:           "SKU-001",
				ProductName:   "Test Product",
				ExpiryDate:    day30.Format(time.RFC3339),
				DaysRemaining: 30,
				AlertLevel:    "warning",
				BranchID:      1,
				BranchName:    "Test Branch",
			},
		},
	}

	mockService.On("CheckExpiringProducts", ctx).Return(events, nil)

	job := NewExpiryCheckJob(mockService, logger)

	// Act
	err := job.RunOnceImmediately(ctx)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, int64(1), job.metrics.TotalRuns, "Total runs should be 1")
	assert.Equal(t, int64(1), job.metrics.TotalAlerts, "Total alerts should be 1")
	assert.Equal(t, int64(1), job.metrics.WarningAlerts, "Warning alerts should be 1")
	assert.Equal(t, 1, job.metrics.LastAlertCount, "Last alert count should be 1")
	assert.False(t, job.metrics.LastRunTime.IsZero(), "Last run time should be set")

	mockService.AssertExpectations(t)
}

// TestExpiryCheckJob_RunOnceImmediately_Error tests error handling
func TestExpiryCheckJob_RunOnceImmediately_Error(t *testing.T) {
	mockService := new(MockExpiryCheckService)
	logger := slog.Default()
	ctx := context.Background()

	mockService.On("CheckExpiringProducts", ctx).Return(nil, assert.AnError)

	job := NewExpiryCheckJob(mockService, logger)

	// Act
	err := job.RunOnceImmediately(ctx)

	// Assert
	require.NoError(t, err, "RunOnceImmediately should not return error even if check fails")
	assert.Equal(t, int64(1), job.metrics.TotalRuns, "Total runs should be 1")
	assert.Equal(t, int64(1), job.metrics.Errors, "Errors should be 1")
	assert.Equal(t, int64(0), job.metrics.TotalAlerts, "Total alerts should be 0")

	mockService.AssertExpectations(t)
}

// TestExpiryCheckJob_MultiLevelAlerts tests metrics for different alert levels
// Story 4.5, Task 4.4: Add metrics: count of alerts generated per day per alert level
func TestExpiryCheckJob_MultiLevelAlerts(t *testing.T) {
	mockService := new(MockExpiryCheckService)
	logger := slog.Default()
	ctx := context.Background()

	events := []*dto.ExpiryAlertEvent{
		{Data: dto.ProductExpiryData{AlertLevel: "warning", ProductID: 1}},
		{Data: dto.ProductExpiryData{AlertLevel: "critical", ProductID: 2}},
		{Data: dto.ProductExpiryData{AlertLevel: "urgent", ProductID: 3}},
		{Data: dto.ProductExpiryData{AlertLevel: "warning", ProductID: 4}},
		{Data: dto.ProductExpiryData{AlertLevel: "urgent", ProductID: 5}},
	}

	mockService.On("CheckExpiringProducts", ctx).Return(events, nil)

	job := NewExpiryCheckJob(mockService, logger)
	job.RunOnceImmediately(ctx)

	// Assert
	assert.Equal(t, int64(5), job.metrics.TotalAlerts, "Total alerts should be 5")
	assert.Equal(t, int64(2), job.metrics.WarningAlerts, "Warning alerts should be 2")
	assert.Equal(t, int64(1), job.metrics.CriticalAlerts, "Critical alerts should be 1")
	assert.Equal(t, int64(2), job.metrics.UrgentAlerts, "Urgent alerts should be 2")
}

// TestExpiryCheckJob_GetMetrics tests metrics retrieval
func TestExpiryCheckJob_GetMetrics(t *testing.T) {
	mockService := new(MockExpiryCheckService)
	logger := slog.Default()
	ctx := context.Background()

	job := NewExpiryCheckJob(mockService, logger)

	// Get initial metrics
	metrics := job.GetMetrics()
	assert.NotNil(t, metrics, "Metrics should not be nil")
	assert.Equal(t, int64(0), metrics.TotalRuns, "Initial total runs should be 0")

	// Run once and check metrics again
	events := []*dto.ExpiryAlertEvent{
		{Data: dto.ProductExpiryData{AlertLevel: "urgent", ProductID: 1}},
	}
	mockService.On("CheckExpiringProducts", ctx).Return(events, nil)
	job.RunOnceImmediately(ctx)

	metrics = job.GetMetrics()
	assert.Equal(t, int64(1), metrics.TotalRuns, "Total runs should be 1")
	assert.Equal(t, int64(1), metrics.UrgentAlerts, "Urgent alerts should be 1")
}

// TestExpiryCheckJob_Stop tests job stopping
// Story 4.5, Task 4.3: Use Go context with cancellation support
func TestExpiryCheckJob_Stop(t *testing.T) {
	mockService := new(MockExpiryCheckService)
	logger := slog.Default()
	ctx, cancel := context.WithCancel(context.Background())

	job := NewExpiryCheckJob(mockService, logger)

	// Start the job (but don't wait for ticker)
	go job.Start(ctx)

	// Stop immediately
	job.Stop()
	cancel()

	// Should not panic
	assert.True(t, true, "Stop should not panic")
}

// TestExpiryCheckJob_NextSixHourMark tests the 6-hour mark calculation
func TestExpiryCheckJob_NextSixHourMark(t *testing.T) {
	job := NewExpiryCheckJob(nil, slog.Default())

	tests := []struct {
		name     string
		hour     int
		expected int
	}{
		{"Before 6 AM", 0, 6},
		{"Before 6 AM", 5, 6},
		{"Before noon", 6, 12},
		{"Before noon", 11, 12},
		{"Before 6 PM", 12, 18},
		{"Before 6 PM", 17, 18},
		{"After 6 PM", 18, 0}, // Next day
		{"Midnight", 23, 0},   // Next day
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			now := time.Date(2026, 5, 20, tt.hour, 30, 0, 0, time.UTC)
			next := job.nextSixHourMark(now)

			if tt.expected == 0 {
				// Should be next day at midnight
				assert.Equal(t, 21, next.Day(), "Should be next day")
				assert.Equal(t, 0, next.Hour(), "Should be midnight")
			} else {
				assert.Equal(t, 20, next.Day(), "Should be same day")
				assert.Equal(t, tt.expected, next.Hour(), "Hour should match")
			}
		})
	}
}

// TestExpiryJobMetrics tests the metrics structure
func TestExpiryJobMetrics(t *testing.T) {
	metrics := &ExpiryJobMetrics{}

	assert.NotNil(t, metrics, "Metrics should be created")
	assert.Equal(t, int64(0), metrics.TotalRuns, "Initial total runs should be 0")
	assert.Equal(t, int64(0), metrics.TotalAlerts, "Initial total alerts should be 0")
	assert.Equal(t, int64(0), metrics.WarningAlerts, "Initial warning alerts should be 0")
	assert.Equal(t, int64(0), metrics.CriticalAlerts, "Initial critical alerts should be 0")
	assert.Equal(t, int64(0), metrics.UrgentAlerts, "Initial urgent alerts should be 0")
	assert.Equal(t, int64(0), metrics.Errors, "Initial errors should be 0")
}
