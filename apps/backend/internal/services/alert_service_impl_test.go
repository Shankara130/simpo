package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/models"
)

// Test NewAlertService with nil dependencies
func TestNewAlertService_PanicOnNilDependencies(t *testing.T) {
	mockProdRepo := new(MockProductRepository)
	mockAudit := new(MockAuditService)

	assert.Panics(t, func() {
		NewAlertService(nil, mockAudit)
	}, "Should panic when productRepo is nil")

	assert.Panics(t, func() {
		NewAlertService(mockProdRepo, nil)
	}, "Should panic when auditService is nil")
}

// Test CheckLowStockAlerts
func TestAlertService_CheckLowStockAlerts_Success(t *testing.T) {
	// Arrange
	mockProdRepo := new(MockProductRepository)
	mockAudit := new(MockAuditService)
	service := NewAlertService(mockProdRepo, mockAudit)

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
	service := NewAlertService(mockProdRepo, mockAudit)

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
	service := NewAlertService(mockProdRepo, mockAudit)

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
	service := NewAlertService(mockProdRepo, mockAudit)

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
	service := NewAlertService(mockProdRepo, mockAudit)

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
	service := NewAlertService(mockProdRepo, mockAudit)

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
	service := NewAlertService(mockProdRepo, mockAudit)

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
	service := NewAlertService(mockProdRepo, mockAudit)

	// Act
	err := service.SendNotification(context.Background(), &LowStockAlert{})

	// Assert
	assert.Error(t, err)
	var invErr *InvalidInputError
	assert.True(t, errors.As(err, &invErr))
	assert.Equal(t, "notification", invErr.Field)
}
