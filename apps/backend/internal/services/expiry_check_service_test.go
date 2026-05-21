package services

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/dto"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/models"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/repositories"
)

// MockProductRepository is a mock for ProductRepository
type MockProductRepositoryForExpiry struct {
	mock.Mock
}

func (m *MockProductRepositoryForExpiry) Create(ctx context.Context, product *models.Product) error {
	args := m.Called(ctx, product)
	return args.Error(0)
}

func (m *MockProductRepositoryForExpiry) GetByID(ctx context.Context, id uint) (*models.Product, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Product), args.Error(1)
}

func (m *MockProductRepositoryForExpiry) GetBySKU(ctx context.Context, branchID uint, sku string) (*models.Product, error) {
	args := m.Called(ctx, branchID, sku)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Product), args.Error(1)
}

func (m *MockProductRepositoryForExpiry) Update(ctx context.Context, product *models.Product) error {
	args := m.Called(ctx, product)
	return args.Error(0)
}

func (m *MockProductRepositoryForExpiry) UpdateStock(ctx context.Context, id uint, quantity int64) error {
	args := m.Called(ctx, id, quantity)
	return args.Error(0)
}

func (m *MockProductRepositoryForExpiry) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockProductRepositoryForExpiry) List(ctx context.Context, filter *repositories.ProductFilter) ([]*models.Product, int64, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*models.Product), args.Get(1).(int64), args.Error(2)
}

func (m *MockProductRepositoryForExpiry) GetLowStockProducts(ctx context.Context, branchID uint) ([]*models.Product, error) {
	args := m.Called(ctx, branchID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Product), args.Error(1)
}

func (m *MockProductRepositoryForExpiry) GetExpiredProducts(ctx context.Context, branchID uint) ([]*models.Product, error) {
	args := m.Called(ctx, branchID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Product), args.Error(1)
}

func (m *MockProductRepositoryForExpiry) GetExpiringProducts(ctx context.Context, branchID uint, startDate, endDate time.Time) ([]*models.Product, error) {
	args := m.Called(ctx, branchID, startDate, endDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Product), args.Error(1)
}

// MockAlertServiceForExpiry is a mock for AlertService
type MockAlertServiceForExpiry struct {
	mock.Mock
}

func (m *MockAlertServiceForExpiry) CheckLowStockAlerts(ctx context.Context, branchID uint) ([]*LowStockAlert, error) {
	args := m.Called(ctx, branchID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*LowStockAlert), args.Error(1)
}

func (m *MockAlertServiceForExpiry) CheckExpiryAlerts(ctx context.Context, branchID uint, daysThreshold int) ([]*ExpiryAlert, error) {
	args := m.Called(ctx, branchID, daysThreshold)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*ExpiryAlert), args.Error(1)
}

func (m *MockAlertServiceForExpiry) SendNotification(ctx context.Context, alert interface{}) error {
	args := m.Called(ctx, alert)
	return args.Error(0)
}

func (m *MockAlertServiceForExpiry) PublishLowStockAlert(ctx context.Context, event *dto.LowStockNotificationEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockAlertServiceForExpiry) PublishExpiryAlert(ctx context.Context, event *dto.ExpiryAlertEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockAlertServiceForExpiry) ClearLowStockState(ctx context.Context, productID uint, branchID uint) error {
	args := m.Called(ctx, productID, branchID)
	return args.Error(0)
}

// TestCheckExpiringProducts_30DayExpiry tests 30-day expiry detection
// Story 4.5, Subtask 6.2: Test 30-day expiry detection
func TestCheckExpiringProducts_30DayExpiry(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := new(MockProductRepositoryForExpiry)
	mockAlert := new(MockAlertServiceForExpiry)
	mockRedis := redis.NewClient(&redis.Options{Addr: ":6379"})

	now := time.Now().UTC()
	expiryDate := now.AddDate(0, 0, 30)

	products := []*models.Product{
		{
			ID:         1,
			SKU:        "SKU-001",
			Name:       "Test Product 1",
			ExpiryDate: &expiryDate,
			BranchID:   1,
			Branch:     &models.Branch{ID: 1, Name: "Test Branch"},
		},
	}

	mockRepo.On("GetExpiringProducts", ctx, uint(0), mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).
		Return(products, nil)
	mockAlert.On("PublishExpiryAlert", ctx, mock.AnythingOfType("*dto.ExpiryAlertEvent")).Return(nil)

	service := NewExpiryCheckService(mockRepo, mockAlert, mockRedis, slog.Default())

	// Act
	events, err := service.CheckExpiringProducts(ctx)

	// Assert
	require.NoError(t, err, "Should not return error")
	assert.Len(t, events, 1, "Should generate 1 event")
	if len(events) > 0 {
		assert.Equal(t, "warning", events[0].Data.AlertLevel, "Alert level should be warning for 30 days")
		assert.GreaterOrEqual(t, events[0].Data.DaysRemaining, 29, "Days remaining should be ~30")
		assert.LessOrEqual(t, events[0].Data.DaysRemaining, 30, "Days remaining should be ~30")
	}
	mockRepo.AssertExpectations(t)
	mockAlert.AssertExpectations(t)
}

// TestCheckExpiringProducts_14DayExpiry tests 14-day expiry detection
// Story 4.5, Subtask 6.3: Test 14-day expiry detection
func TestCheckExpiringProducts_14DayExpiry(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := new(MockProductRepositoryForExpiry)
	mockAlert := new(MockAlertServiceForExpiry)
	mockRedis := redis.NewClient(&redis.Options{Addr: ":6379"})

	now := time.Now().UTC()
	expiryDate := now.AddDate(0, 0, 14)

	products := []*models.Product{
		{
			ID:         2,
			SKU:        "SKU-002",
			Name:       "Test Product 2",
			ExpiryDate: &expiryDate,
			BranchID:   1,
			Branch:     &models.Branch{ID: 1, Name: "Test Branch"},
		},
	}

	mockRepo.On("GetExpiringProducts", ctx, uint(0), mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).
		Return(products, nil)
	mockAlert.On("PublishExpiryAlert", ctx, mock.AnythingOfType("*dto.ExpiryAlertEvent")).Return(nil)

	service := NewExpiryCheckService(mockRepo, mockAlert, mockRedis, slog.Default())

	// Act
	events, err := service.CheckExpiringProducts(ctx)

	// Assert
	require.NoError(t, err, "Should not return error")
	assert.Len(t, events, 1, "Should generate 1 event")
	if len(events) > 0 {
		assert.Equal(t, "critical", events[0].Data.AlertLevel, "Alert level should be critical for 14 days")
		assert.GreaterOrEqual(t, events[0].Data.DaysRemaining, 13, "Days remaining should be ~14")
		assert.LessOrEqual(t, events[0].Data.DaysRemaining, 14, "Days remaining should be ~14")
	}
}

// TestCheckExpiringProducts_7DayExpiry tests 7-day expiry detection
// Story 4.5, Subtask 6.4: Test 7-day expiry detection
func TestCheckExpiringProducts_7DayExpiry(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := new(MockProductRepositoryForExpiry)
	mockAlert := new(MockAlertServiceForExpiry)
	mockRedis := redis.NewClient(&redis.Options{Addr: ":6379"})

	now := time.Now().UTC()
	expiryDate := now.AddDate(0, 0, 7)

	products := []*models.Product{
		{
			ID:         3,
			SKU:        "SKU-003",
			Name:       "Test Product 3",
			ExpiryDate: &expiryDate,
			BranchID:   1,
			Branch:     &models.Branch{ID: 1, Name: "Test Branch"},
		},
	}

	mockRepo.On("GetExpiringProducts", ctx, uint(0), mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).
		Return(products, nil)
	mockAlert.On("PublishExpiryAlert", ctx, mock.AnythingOfType("*dto.ExpiryAlertEvent")).Return(nil)

	service := NewExpiryCheckService(mockRepo, mockAlert, mockRedis, slog.Default())

	// Act
	events, err := service.CheckExpiringProducts(ctx)

	// Assert
	require.NoError(t, err, "Should not return error")
	assert.Len(t, events, 1, "Should generate 1 event")
	if len(events) > 0 {
		assert.Equal(t, "urgent", events[0].Data.AlertLevel, "Alert level should be urgent for 7 days")
		assert.GreaterOrEqual(t, events[0].Data.DaysRemaining, 6, "Days remaining should be ~7")
		assert.LessOrEqual(t, events[0].Data.DaysRemaining, 7, "Days remaining should be ~7")
	}
}

// TestCheckExpiringProducts_AlertLevelCategorization tests alert level categorization
// Story 4.5, Subtask 6.5: Test alert level categorization
func TestCheckExpiringProducts_AlertLevelCategorization(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := new(MockProductRepositoryForExpiry)
	mockAlert := new(MockAlertServiceForExpiry)
	mockRedis := redis.NewClient(&redis.Options{Addr: ":6379"})

	now := time.Now().UTC()
	day30 := now.AddDate(0, 0, 30)
	day14 := now.AddDate(0, 0, 14)
	day7 := now.AddDate(0, 0, 7)

	products := []*models.Product{
		{ID: 1, SKU: "SKU-30D", Name: "30 Day Product", ExpiryDate: &day30, BranchID: 1, Branch: &models.Branch{ID: 1, Name: "Branch 1"}},
		{ID: 2, SKU: "SKU-14D", Name: "14 Day Product", ExpiryDate: &day14, BranchID: 1, Branch: &models.Branch{ID: 1, Name: "Branch 1"}},
		{ID: 3, SKU: "SKU-7D", Name: "7 Day Product", ExpiryDate: &day7, BranchID: 1, Branch: &models.Branch{ID: 1, Name: "Branch 1"}},
	}

	mockRepo.On("GetExpiringProducts", ctx, uint(0), mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).
		Return(products, nil)
	mockAlert.On("PublishExpiryAlert", ctx, mock.AnythingOfType("*dto.ExpiryAlertEvent")).Return(nil).Times(3)

	service := NewExpiryCheckService(mockRepo, mockAlert, mockRedis, slog.Default())

	// Act
	events, err := service.CheckExpiringProducts(ctx)

	// Assert
	require.NoError(t, err)
	assert.Len(t, events, 3, "Should generate 3 events")

	// Check categorization
	alertLevels := make(map[string]int)
	for _, event := range events {
		alertLevels[event.Data.AlertLevel]++
	}

	assert.Equal(t, 1, alertLevels["warning"], "Should have 1 warning (30 days)")
	assert.Equal(t, 1, alertLevels["critical"], "Should have 1 critical (14 days)")
	assert.Equal(t, 1, alertLevels["urgent"], "Should have 1 urgent (7 days)")
}

// TestCheckExpiringProducts_RepositoryError tests error handling when repository fails
func TestCheckExpiringProducts_RepositoryError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := new(MockProductRepositoryForExpiry)
	mockAlert := new(MockAlertServiceForExpiry)
	mockRedis := redis.NewClient(&redis.Options{Addr: ":6379"})

	mockRepo.On("GetExpiringProducts", ctx, uint(0), mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).
		Return(nil, errors.New("database connection failed"))

	service := NewExpiryCheckService(mockRepo, mockAlert, mockRedis, slog.Default())

	// Act
	events, err := service.CheckExpiringProducts(ctx)

	// Assert
	assert.Error(t, err, "Should return error when repository fails")
	assert.Nil(t, events, "Should return nil events on error")
	assert.Contains(t, err.Error(), "failed to get expiring products", "Error should mention the operation that failed")
}

// TestCheckExpiringProducts_NilExpiryDate tests handling of products with nil expiry dates
func TestCheckExpiringProducts_NilExpiryDate(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := new(MockProductRepositoryForExpiry)
	mockAlert := new(MockAlertServiceForExpiry)
	mockRedis := redis.NewClient(&redis.Options{Addr: ":6379"})

	products := []*models.Product{
		{ID: 1, SKU: "SKU-001", Name: "Product without expiry", ExpiryDate: nil, BranchID: 1},
	}

	mockRepo.On("GetExpiringProducts", ctx, uint(0), mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).
		Return(products, nil)

	service := NewExpiryCheckService(mockRepo, mockAlert, mockRedis, slog.Default())

	// Act
	events, err := service.CheckExpiringProducts(ctx)

	// Assert
	require.NoError(t, err)
	assert.Empty(t, events, "Should skip products with nil expiry dates")
}

// TestCheckExpiringProducts_EventStructure tests event payload structure
// Story 4.5, Subtask 6.8: Test event payload structure matches specification
func TestCheckExpiringProducts_EventStructure(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := new(MockProductRepositoryForExpiry)
	mockAlert := new(MockAlertServiceForExpiry)
	mockRedis := redis.NewClient(&redis.Options{Addr: ":6379"})

	now := time.Now().UTC()
	expiryDate := now.AddDate(0, 0, 14)

	products := []*models.Product{
		{
			ID:         123,
			SKU:        "SKU-TEST-123",
			Name:       "Test Medicine",
			ExpiryDate: &expiryDate,
			BranchID:   5,
			Branch:     &models.Branch{ID: 5, Name: "Jakarta Branch"},
		},
	}

	mockRepo.On("GetExpiringProducts", ctx, uint(0), mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).
		Return(products, nil)
	mockAlert.On("PublishExpiryAlert", ctx, mock.MatchedBy(func(event *dto.ExpiryAlertEvent) bool {
		// Validate event structure
		return event.EventType == "product.expiry" &&
			event.Data.ProductID == 123 &&
			event.Data.SKU == "SKU-TEST-123" &&
			event.Data.ProductName == "Test Medicine" &&
			event.Data.AlertLevel == "critical" &&
			event.Data.BranchID == 5 &&
			event.Data.BranchName == "Jakarta Branch"
	})).Return(nil)

	service := NewExpiryCheckService(mockRepo, mockAlert, mockRedis, slog.Default())

	// Act
	events, err := service.CheckExpiringProducts(ctx)

	// Assert
	require.NoError(t, err)
	assert.Len(t, events, 1)

	event := events[0]
	assert.NotEmpty(t, event.EventID, "EventID should be set")
	assert.Equal(t, "product.expiry", event.EventType, "EventType should be product.expiry")
	assert.NotEmpty(t, event.Timestamp, "Timestamp should be set")
	assert.Equal(t, uint(123), event.Data.ProductID, "ProductID should match")
	assert.Equal(t, "SKU-TEST-123", event.Data.SKU, "SKU should match")
	assert.Equal(t, "Test Medicine", event.Data.ProductName, "ProductName should match")
	assert.NotEmpty(t, event.Data.ExpiryDate, "ExpiryDate should be set")
	assert.GreaterOrEqual(t, event.Data.DaysRemaining, 13, "DaysRemaining should be ~14")
	assert.LessOrEqual(t, event.Data.DaysRemaining, 14, "DaysRemaining should be ~14")
	assert.Equal(t, "critical", event.Data.AlertLevel, "AlertLevel should be critical")
	assert.Equal(t, uint(5), event.Data.BranchID, "BranchID should match")
	assert.Equal(t, "Jakarta Branch", event.Data.BranchName, "BranchName should match")
}

// TestCheckExpiringProducts_PublishErrorContinues tests that processing continues if one publish fails
func TestCheckExpiringProducts_PublishErrorContinues(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := new(MockProductRepositoryForExpiry)
	mockAlert := new(MockAlertServiceForExpiry)
	mockRedis := redis.NewClient(&redis.Options{Addr: ":6379"})

	now := time.Now().UTC()
	day30 := now.AddDate(0, 0, 30)
	day14 := now.AddDate(0, 0, 14)

	products := []*models.Product{
		{ID: 1, SKU: "SKU-001", Name: "Product 1", ExpiryDate: &day30, BranchID: 1, Branch: &models.Branch{ID: 1, Name: "Branch 1"}},
		{ID: 2, SKU: "SKU-002", Name: "Product 2", ExpiryDate: &day14, BranchID: 1, Branch: &models.Branch{ID: 1, Name: "Branch 1"}},
	}

	mockRepo.On("GetExpiringProducts", ctx, uint(0), mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).
		Return(products, nil)
	// First call fails, second succeeds
	mockAlert.On("PublishExpiryAlert", ctx, mock.MatchedBy(func(e *dto.ExpiryAlertEvent) bool { return e.Data.ProductID == 1 })).
		Return(errors.New("Redis connection failed"))
	mockAlert.On("PublishExpiryAlert", ctx, mock.MatchedBy(func(e *dto.ExpiryAlertEvent) bool { return e.Data.ProductID == 2 })).
		Return(nil)

	service := NewExpiryCheckService(mockRepo, mockAlert, mockRedis, slog.Default())

	// Act
	events, err := service.CheckExpiringProducts(ctx)

	// Assert
	require.NoError(t, err, "Should not return error even if one publish fails")
	assert.Len(t, events, 1, "Should only return successfully published events")
	assert.Equal(t, uint(2), events[0].Data.ProductID, "Should return the second product")
}

// TestClearExpiryAlertState tests clearing alert state
// Story 4.5, Subtask 3.5: Add ClearExpiryAlertState method
func TestClearExpiryAlertState(t *testing.T) {
	// This test would require a real Redis instance or more sophisticated mocking
	// For now, we just verify the method exists and handles nil Redis gracefully
	ctx := context.Background()
	mockRepo := new(MockProductRepositoryForExpiry)
	mockAlert := new(MockAlertServiceForExpiry)

	service := NewExpiryCheckService(mockRepo, mockAlert, nil, nil)

	// Act - should not panic with nil Redis
	err := service.ClearExpiryAlertState(ctx, 123, 1)

	// Assert
	assert.NoError(t, err, "Should handle nil Redis gracefully")
}
