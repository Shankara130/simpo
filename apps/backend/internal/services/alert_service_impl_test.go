package services

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/dto"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/models"
)

// Test NewAlertService with nil dependencies
func TestNewAlertService_PanicOnNilDependencies(t *testing.T) {
	mockProdRepo := new(MockProductRepository)
	mockAudit := new(MockAuditService)

	assert.Panics(t, func() {
		NewAlertService(nil, mockAudit, nil)
	}, "Should panic when productRepo is nil")

	assert.Panics(t, func() {
		NewAlertService(mockProdRepo, nil, nil)
	}, "Should panic when auditService is nil")
}

// Test CheckLowStockAlerts
func TestAlertService_CheckLowStockAlerts_Success(t *testing.T) {
	// Arrange
	mockProdRepo := new(MockProductRepository)
	mockAudit := new(MockAuditService)
	service := NewAlertService(mockProdRepo, mockAudit, nil)

	products := []*models.Product{
		{
			ID:               1,
			Name:             "Product 1",
			SKU:              "SKU-001",
			StockQty:         5,
			ReorderThreshold: 10,
			BranchID:         1,
		},
		{
			ID:               2,
			Name:             "Product 2",
			SKU:              "SKU-002",
			StockQty:         0,  // Out of stock - HIGH severity
			ReorderThreshold: 10,
			BranchID:         1,
		},
	}

	// Mock expectations
	mockProdRepo.On("GetLowStockProducts", mock.Anything, uint(1)).Return(products, nil)

	// Act
	alerts, err := service.CheckLowStockAlerts(context.Background(), 1)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, alerts, 2)

	// Check first alert (LOW severity - stock >= 50% of threshold)
	assert.Equal(t, "LOW", alerts[0].Severity)
	assert.Equal(t, uint(1), alerts[0].ProductID)

	// Check second alert (HIGH severity - out of stock)
	assert.Equal(t, "HIGH", alerts[1].Severity)
	assert.Equal(t, uint(2), alerts[1].ProductID)

	mockProdRepo.AssertExpectations(t)
}

func TestAlertService_CheckLowStockAlerts_ZeroBranchID(t *testing.T) {
	// Arrange
	mockProdRepo := new(MockProductRepository)
	mockAudit := new(MockAuditService)
	service := NewAlertService(mockProdRepo, mockAudit, nil)

	// Act
	_, err := service.CheckLowStockAlerts(context.Background(), 0)

	// Assert
	assert.Error(t, err)
	var invErr *InvalidInputError
	assert.True(t, errors.As(err, &invErr))
	assert.Equal(t, "branch_id", invErr.Field)
}

func TestAlertService_CheckLowStockAlerts_ContextCanceled(t *testing.T) {
	// Arrange
	mockProdRepo := new(MockProductRepository)
	mockAudit := new(MockAuditService)
	service := NewAlertService(mockProdRepo, mockAudit, nil)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Act
	_, err := service.CheckLowStockAlerts(ctx, 1)

	// Assert
	assert.Error(t, err)
}

// Test CheckExpiryAlerts
func TestAlertService_CheckExpiryAlerts_Success(t *testing.T) {
	// Arrange
	mockProdRepo := new(MockProductRepository)
	mockAudit := new(MockAuditService)
	service := NewAlertService(mockProdRepo, mockAudit, nil)

	// Set up expiry dates
	fiveDays := time.Now().AddDate(0, 0, 5)  // CRITICAL
	tenDays := time.Now().AddDate(0, 0, 10)  // WARNING
	twentyDays := time.Now().AddDate(0, 0, 20) // INFO

	products := []*models.Product{
		{
			ID:         1,
			Name:       "Expiring Soon",
			SKU:        "SKU-001",
			ExpiryDate: &fiveDays,
			BranchID:   1,
		},
		{
			ID:         2,
			Name:       "Expiring Later",
			SKU:        "SKU-002",
			ExpiryDate: &tenDays,
			BranchID:   1,
		},
		{
			ID:         3,
			Name:       "Expiring Much Later",
			SKU:        "SKU-003",
			ExpiryDate: &twentyDays,
			BranchID:   1,
		},
	}

	// Mock expectations
	mockProdRepo.On("GetExpiredProducts", mock.Anything, uint(1)).Return(products, nil)

	// Act
	alerts, err := service.CheckExpiryAlerts(context.Background(), 1, 30)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, alerts, 3)

	// Check severities
	assert.Equal(t, "CRITICAL", alerts[0].Severity)
	assert.Equal(t, "WARNING", alerts[1].Severity)
	assert.Equal(t, "INFO", alerts[2].Severity)

	mockProdRepo.AssertExpectations(t)
}

func TestAlertService_CheckExpiryAlerts_ZeroBranchID(t *testing.T) {
	// Arrange
	mockProdRepo := new(MockProductRepository)
	mockAudit := new(MockAuditService)
	service := NewAlertService(mockProdRepo, mockAudit, nil)

	// Act
	_, err := service.CheckExpiryAlerts(context.Background(), 0, 30)

	// Assert
	assert.Error(t, err)
	var invErr *InvalidInputError
	assert.True(t, errors.As(err, &invErr))
	assert.Equal(t, "branch_id", invErr.Field)
}

func TestAlertService_CheckExpiryAlerts_InvalidDaysThreshold(t *testing.T) {
	// Arrange
	mockProdRepo := new(MockProductRepository)
	mockAudit := new(MockAuditService)
	service := NewAlertService(mockProdRepo, mockAudit, nil)

	// Act
	_, err := service.CheckExpiryAlerts(context.Background(), 1, 0)

	// Assert
	assert.Error(t, err)
	var invErr *InvalidInputError
	assert.True(t, errors.As(err, &invErr))
	assert.Equal(t, "days_threshold", invErr.Field)
}

func TestAlertService_CheckExpiryAlerts_OnlyWithinThreshold(t *testing.T) {
	// Arrange
	mockProdRepo := new(MockProductRepository)
	mockAudit := new(MockAuditService)
	service := NewAlertService(mockProdRepo, mockAudit, nil)

	// Set up expiry dates
	fiveDays := time.Now().AddDate(0, 0, 5)   // Within threshold
	fortyDays := time.Now().AddDate(0, 0, 40) // Outside threshold

	products := []*models.Product{
		{
			ID:         1,
			Name:       "Expiring Soon",
			SKU:        "SKU-001",
			ExpiryDate: &fiveDays,
			BranchID:   1,
		},
		{
			ID:         2,
			Name:       "Expiring Later",
			SKU:        "SKU-002",
			ExpiryDate: &fortyDays,
			BranchID:   1,
		},
	}

	// Mock expectations
	mockProdRepo.On("GetExpiredProducts", mock.Anything, uint(1)).Return(products, nil)

	// Act
	alerts, err := service.CheckExpiryAlerts(context.Background(), 1, 30)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, alerts, 1) // Only the 5-day product
	assert.Equal(t, "CRITICAL", alerts[0].Severity)

	mockProdRepo.AssertExpectations(t)
}

// Test SendNotification
func TestAlertService_SendNotification_NotImplemented(t *testing.T) {
	// Arrange
	mockProdRepo := new(MockProductRepository)
	mockAudit := new(MockAuditService)
	service := NewAlertService(mockProdRepo, mockAudit, nil)

	// Act
	err := service.SendNotification(context.Background(), &LowStockAlert{})

	// Assert
	assert.Error(t, err)
	var invErr *InvalidInputError
	assert.True(t, errors.As(err, &invErr))
	assert.Equal(t, "notification", invErr.Field)
}

// Story 4.4, Task 6: Comprehensive Low Stock Testing

// Test PublishLowStockAlert with mock Redis
func TestAlertService_PublishLowStockAlert_Success(t *testing.T) {
	// Arrange - Setup miniredis
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	redisClient := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	mockProdRepo := new(MockProductRepository)
	mockAudit := new(MockAuditService)
	service := NewAlertService(mockProdRepo, mockAudit, redisClient)

	event := &dto.LowStockNotificationEvent{
		EventID:   uuid.New().String(),
		EventType: "stock.low",
		Timestamp: time.Now().Format(time.RFC3339),
		Data: dto.ProductLowStockData{
			ProductID:         1,
			SKU:               "TEST-001",
			ProductName:       "Test Product",
			CurrentStock:      5,
			ReorderThreshold:  10,
			SuggestedOrderQty: 15,
			BranchID:          1,
			BranchName:        "Main Branch",
		},
	}

	// Act
	err = service.PublishLowStockAlert(context.Background(), event)

	// Assert
	assert.NoError(t, err)

	// Verify Redis Set was created for debounce tracking
	ctx := context.Background()
	key := "low_stock:1:1"
	exists, err := redisClient.Exists(ctx, key).Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(1), exists)

	// Verify TTL is set to 24 hours
	ttl, err := redisClient.TTL(ctx, key).Result()
	assert.NoError(t, err)
	assert.True(t, ttl > 23*time.Hour && ttl <= 24*time.Hour, "TTL should be approximately 24 hours")
}

// Test PublishLowStockAlert with debounce logic (duplicate notification suppression)
func TestAlertService_PublishLowStockAlert_DebounceLogic(t *testing.T) {
	// Arrange - Setup miniredis
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	redisClient := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	mockProdRepo := new(MockProductRepository)
	mockAudit := new(MockAuditService)
	service := NewAlertService(mockProdRepo, mockAudit, redisClient)

	event := &dto.LowStockNotificationEvent{
		EventID:   uuid.New().String(),
		EventType: "stock.low",
		Timestamp: time.Now().Format(time.RFC3339),
		Data: dto.ProductLowStockData{
			ProductID:         1,
			SKU:               "TEST-001",
			ProductName:       "Test Product",
			CurrentStock:      5,
			ReorderThreshold:  10,
			SuggestedOrderQty: 15,
			BranchID:          1,
			BranchName:        "Main Branch",
		},
	}

	ctx := context.Background()

	// Act - First call should succeed
	err = service.PublishLowStockAlert(ctx, event)
	assert.NoError(t, err)

	// Act - Second call with same product/branch should be suppressed (debounce)
	err = service.PublishLowStockAlert(ctx, event)
	assert.NoError(t, err)

	// Assert - Verify only one Redis Set entry exists
	key := "low_stock:1:1"
	exists, err := redisClient.Exists(ctx, key).Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(1), exists)
}

// Test PublishLowStockAlert event payload structure
func TestAlertService_PublishLowStockAlert_EventPayloadStructure(t *testing.T) {
	// Arrange
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	redisClient := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	mockProdRepo := new(MockProductRepository)
	mockAudit := new(MockAuditService)
	service := NewAlertService(mockProdRepo, mockAudit, redisClient)

	event := &dto.LowStockNotificationEvent{
		EventID:   uuid.New().String(),
		EventType: "stock.low",
		Timestamp: time.Now().Format(time.RFC3339),
		Data: dto.ProductLowStockData{
			ProductID:         1,
			SKU:               "TEST-001",
			ProductName:       "Test Product",
			CurrentStock:      5,
			ReorderThreshold:  10,
			SuggestedOrderQty: 15,
			BranchID:          1,
			BranchName:        "Main Branch",
		},
	}

	// Act
	err = service.PublishLowStockAlert(context.Background(), event)

	// Assert
	assert.NoError(t, err)

	// Verify event structure matches specification
	assert.Equal(t, "stock.low", event.EventType)
	assert.NotEmpty(t, event.EventID)
	assert.NotEmpty(t, event.Timestamp)

	// Verify Data structure
	assert.Equal(t, uint(1), event.Data.ProductID)
	assert.Equal(t, "TEST-001", event.Data.SKU)
	assert.Equal(t, "Test Product", event.Data.ProductName)
	assert.Equal(t, 5, event.Data.CurrentStock)
	assert.Equal(t, 10, event.Data.ReorderThreshold)
	assert.Equal(t, 15, event.Data.SuggestedOrderQty)
	assert.Equal(t, uint(1), event.Data.BranchID)
	assert.Equal(t, "Main Branch", event.Data.BranchName)

	// Verify event can be marshaled to JSON
	jsonData, err := json.Marshal(event)
	assert.NoError(t, err)
	assert.NotEmpty(t, jsonData)

	// Verify JSON structure
	var unmarshaled map[string]interface{}
	err = json.Unmarshal(jsonData, &unmarshaled)
	assert.NoError(t, err)
	assert.Equal(t, "stock.low", unmarshaled["eventType"])
	assert.NotEmpty(t, unmarshaled["eventId"])
	assert.NotEmpty(t, unmarshaled["timestamp"])
	assert.NotEmpty(t, unmarshaled["data"])
}

// Test PublishLowStockAlert with Redis unavailable (graceful degradation)
func TestAlertService_PublishLowStockAlert_RedisUnavailable(t *testing.T) {
	// Arrange - Use invalid Redis address
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:9999", // Invalid address
	})

	mockProdRepo := new(MockProductRepository)
	mockAudit := new(MockAuditService)
	service := NewAlertService(mockProdRepo, mockAudit, redisClient)

	event := &dto.LowStockNotificationEvent{
		EventID:   uuid.New().String(),
		EventType: "stock.low",
		Timestamp: time.Now().Format(time.RFC3339),
		Data: dto.ProductLowStockData{
			ProductID:         1,
			SKU:               "TEST-001",
			ProductName:       "Test Product",
			CurrentStock:      5,
			ReorderThreshold:  10,
			SuggestedOrderQty: 15,
			BranchID:          1,
			BranchName:        "Main Branch",
		},
	}

	// Act
	err := service.PublishLowStockAlert(context.Background(), event)

	// Assert - Graceful degradation: should return nil (no error) even though Redis is unavailable
	assert.NoError(t, err)
}

// Test ClearLowStockState
func TestAlertService_ClearLowStockState_Success(t *testing.T) {
	// Arrange - Setup miniredis
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	redisClient := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	mockProdRepo := new(MockProductRepository)
	mockAudit := new(MockAuditService)
	service := NewAlertService(mockProdRepo, mockAudit, redisClient)

	ctx := context.Background()

	// Set up low stock state
	key := "low_stock:1:1"
	err = redisClient.Set(ctx, key, "1", 24*time.Hour).Err()
	require.NoError(t, err)

	// Verify state exists
	exists, err := redisClient.Exists(ctx, key).Result()
	require.NoError(t, err)
	require.Equal(t, int64(1), exists)

	// Act
	err = service.ClearLowStockState(ctx, 1, 1)

	// Assert
	assert.NoError(t, err)

	// Verify state was cleared
	exists, err = redisClient.Exists(ctx, key).Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(0), exists)
}

// Test ClearLowStockState when not in low stock state
func TestAlertService_ClearLowStockState_NotInLowStock(t *testing.T) {
	// Arrange - Setup miniredis
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	redisClient := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	mockProdRepo := new(MockProductRepository)
	mockAudit := new(MockAuditService)
	service := NewAlertService(mockProdRepo, mockAudit, redisClient)

	ctx := context.Background()

	// Act - Clear state that doesn't exist
	err = service.ClearLowStockState(ctx, 1, 1)

	// Assert - Should not error
	assert.NoError(t, err)
}

// Test PublishLowStockAlert with context cancellation
func TestAlertService_PublishLowStockAlert_ContextCanceled(t *testing.T) {
	// Arrange
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	redisClient := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	mockProdRepo := new(MockProductRepository)
	mockAudit := new(MockAuditService)
	service := NewAlertService(mockProdRepo, mockAudit, redisClient)

	event := &dto.LowStockNotificationEvent{
		EventID:   uuid.New().String(),
		EventType: "stock.low",
		Timestamp: time.Now().Format(time.RFC3339),
		Data: dto.ProductLowStockData{
			ProductID:         1,
			SKU:               "TEST-001",
			ProductName:       "Test Product",
			CurrentStock:      5,
			ReorderThreshold:  10,
			SuggestedOrderQty: 15,
			BranchID:          1,
			BranchName:        "Main Branch",
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Act
	err = service.PublishLowStockAlert(ctx, event)

	// Assert
	assert.Error(t, err)
}

// Test ClearLowStockState with context cancellation
func TestAlertService_ClearLowStockState_ContextCanceled(t *testing.T) {
	// Arrange
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	redisClient := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	mockProdRepo := new(MockProductRepository)
	mockAudit := new(MockAuditService)
	service := NewAlertService(mockProdRepo, mockAudit, redisClient)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Act
	err = service.ClearLowStockState(ctx, 1, 1)

	// Assert
	assert.Error(t, err)
}
