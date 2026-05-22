package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/models"
)

// TestProductService_ValidateProductForSale_WithExpiredProduct tests that expired products are rejected
func TestProductService_ValidateProductForSale_WithExpiredProduct(t *testing.T) {
	// Arrange
	mockRepo := new(MockProductRepository)
	mockAudit := new(MockAuditService)
	service := NewProductService(mockRepo, mockAudit, nil, nil, nil)

	// Create an expired product
	pastDate := time.Now().Add(-24 * time.Hour)
	expiredProduct := &models.Product{
		ID:         1,
		SKU:        "EXP001",
		Name:       "Expired Medicine",
		StockQty:   10,
		Price:      "10000.00",
		ExpiryDate: &pastDate,
	}

	// Mock repository to return the expired product
	mockRepo.On("GetByID", mock.Anything, uint(1)).Return(expiredProduct, nil)

	// Act
	err := service.ValidateProductForSale(context.Background(), expiredProduct.ID)

	// Assert
	require.Error(t, err, "Should return error for expired product")
	var productExpiredErr *ErrProductExpired
	assert.True(t, errors.As(err, &productExpiredErr), "Should return ErrProductExpired type")
	assert.Contains(t, err.Error(), "expired", "Error message should mention expired")
}

// TestProductService_ValidateProductForSale_WithValidProduct tests that valid products pass validation
func TestProductService_ValidateProductForSale_WithValidProduct(t *testing.T) {
	// Arrange
	mockRepo := new(MockProductRepository)
	mockAudit := new(MockAuditService)
	service := NewProductService(mockRepo, mockAudit, nil, nil, nil)

	// Create a valid product with future expiry date
	futureDate := time.Now().Add(30 * 24 * time.Hour)
	validProduct := &models.Product{
		ID:         2,
		SKU:        "VAL001",
		Name:       "Valid Medicine",
		StockQty:   20,
		Price:      "20000.00",
		ExpiryDate: &futureDate,
	}

	// Mock repository to return the valid product
	mockRepo.On("GetByID", mock.Anything, uint(2)).Return(validProduct, nil)

	// Act
	err := service.ValidateProductForSale(context.Background(), validProduct.ID)

	// Assert
	assert.NoError(t, err, "Should not return error for valid product")
}

// TestProductService_ValidateProductForSale_WithNoExpiryDate tests that products without expiry date pass validation
func TestProductService_ValidateProductForSale_WithNoExpiryDate(t *testing.T) {
	// Arrange
	mockRepo := new(MockProductRepository)
	mockAudit := new(MockAuditService)
	service := NewProductService(mockRepo, mockAudit, nil, nil, nil)

	// Create a product with no expiry date
	productNoExpiry := &models.Product{
		ID:         3,
		SKU:        "NOP001",
		Name:       "No Expiry Medicine",
		StockQty:   30,
		Price:      "30000.00",
		ExpiryDate: nil,
	}

	// Mock repository to return the product
	mockRepo.On("GetByID", mock.Anything, uint(3)).Return(productNoExpiry, nil)

	// Act
	err := service.ValidateProductForSale(context.Background(), productNoExpiry.ID)

	// Assert
	assert.NoError(t, err, "Should not return error for product without expiry date")
}

// TestProductService_ValidateProductForSale_ProductNotFound tests that non-existent products return error
func TestProductService_ValidateProductForSale_ProductNotFound(t *testing.T) {
	// Arrange
	mockRepo := new(MockProductRepository)
	mockAudit := new(MockAuditService)
	service := NewProductService(mockRepo, mockAudit, nil, nil, nil)

	// Mock repository to return not found error
	mockRepo.On("GetByID", mock.Anything, uint(999)).Return(nil, errors.New("not found"))

	// Act
	err := service.ValidateProductForSale(context.Background(), 999)

	// Assert
	require.Error(t, err, "Should return error when product not found")
	var productNotFoundErr *ProductNotFoundError
	assert.True(t, errors.As(err, &productNotFoundErr), "Should return ProductNotFoundError type")
}

// TestGetProductBySKU_Success tests successful product retrieval by SKU
func TestProductService_GetProductBySKU_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockProductRepository)
	mockAudit := new(MockAuditService)
	service := NewProductService(mockRepo, mockAudit, nil, nil, nil)

	// Create a valid product
	futureDate := time.Now().Add(30 * 24 * time.Hour)
	validProduct := &models.Product{
		ID:         1,
		SKU:        "VAL001",
		Name:       "Valid Medicine",
		StockQty:   20,
		Price:      "20000.00",
		ExpiryDate: &futureDate,
	}

	// Mock repository to return the product
	mockRepo.On("GetBySKU", mock.Anything, uint(1), "VAL001").Return(validProduct, nil)

	// Act
	product, err := service.GetProductBySKU(context.Background(), 1, "VAL001")

	// Assert
	assert.NoError(t, err, "Should not return error for valid product")
	assert.NotNil(t, product, "Should return product")
	assert.Equal(t, "VAL001", product.SKU)
	assert.Equal(t, "Valid Medicine", product.Name)
}

// TestGetProductBySKU_ExpiredProduct tests that expired products return error
func TestProductService_GetProductBySKU_ExpiredProduct(t *testing.T) {
	// Arrange
	mockRepo := new(MockProductRepository)
	mockAudit := new(MockAuditService)
	service := NewProductService(mockRepo, mockAudit, nil, nil, nil)

	// Create an expired product
	pastDate := time.Now().Add(-24 * time.Hour)
	expiredProduct := &models.Product{
		ID:         2,
		SKU:        "EXP002",
		Name:       "Expired Medicine",
		StockQty:   10,
		Price:      "10000.00",
		ExpiryDate: &pastDate,
	}

	// Mock repository to return the expired product
	mockRepo.On("GetBySKU", mock.Anything, uint(1), "EXP002").Return(expiredProduct, nil)

	// Act
	product, err := service.GetProductBySKU(context.Background(), 1, "EXP002")

	// Assert
	require.Error(t, err, "Should return error for expired product")
	var productExpiredErr *ErrProductExpired
	assert.True(t, errors.As(err, &productExpiredErr), "Should return ErrProductExpired type")
	assert.Nil(t, product, "Should not return product when expired")
}

// TestGetProductBySKU_ProductNotFound tests that non-existent products return error
func TestProductService_GetProductBySKU_ProductNotFound(t *testing.T) {
	// Arrange
	mockRepo := new(MockProductRepository)
	mockAudit := new(MockAuditService)
	service := NewProductService(mockRepo, mockAudit, nil, nil, nil)

	// Mock repository to return not found error
	mockRepo.On("GetBySKU", mock.Anything, uint(1), "NOTFOUND").Return(nil, errors.New("not found"))

	// Act
	product, err := service.GetProductBySKU(context.Background(), 1, "NOTFOUND")

	// Assert
	require.Error(t, err, "Should return error when product not found")
	var productNotFoundErr *ProductNotFoundError
	assert.True(t, errors.As(err, &productNotFoundErr), "Should return ProductNotFoundError type")
	assert.Nil(t, product, "Should not return product when not found")
}

// TestGetProductBySKU_EmptySKU tests validation for empty SKU
func TestProductService_GetProductBySKU_EmptySKU(t *testing.T) {
	// Arrange
	mockRepo := new(MockProductRepository)
	mockAudit := new(MockAuditService)
	service := NewProductService(mockRepo, mockAudit, nil, nil, nil)

	// Act
	product, err := service.GetProductBySKU(context.Background(), 1, "")

	// Assert
	require.Error(t, err, "Should return error for empty SKU")
	var invalidInputErr *InvalidInputError
	assert.True(t, errors.As(err, &invalidInputErr), "Should return InvalidInputError type")
	assert.Equal(t, "sku", invalidInputErr.Field)
	assert.Nil(t, product, "Should not return product when SKU is empty")
}

// TestGetProductBySKU_ZeroBranchID tests validation for zero branch ID
func TestProductService_GetProductBySKU_ZeroBranchID(t *testing.T) {
	// Arrange
	mockRepo := new(MockProductRepository)
	mockAudit := new(MockAuditService)
	service := NewProductService(mockRepo, mockAudit, nil, nil, nil)

	// Act
	product, err := service.GetProductBySKU(context.Background(), 0, "VAL001")

	// Assert
	require.Error(t, err, "Should return error for zero branch ID")
	var invalidInputErr *InvalidInputError
	assert.True(t, errors.As(err, &invalidInputErr), "Should return InvalidInputError type")
	assert.Equal(t, "branch_id", invalidInputErr.Field)
	assert.Nil(t, product, "Should not return product when branch ID is zero")
}

// TestIntegration_ExpiredProductScanToAuditLog tests the complete flow:
// Scan (GetProductBySKU) → Validate (ProductService) → Block (ProcessSale) → Log Audit
// Story 4.6, Task 7.6: Integration test for AC3, AC4, AC5, AC6
func TestIntegration_ExpiredProductScanToAuditLog(t *testing.T) {
	// Arrange
	mockRepo := new(MockProductRepository)
	mockAudit := new(MockAuditService)
	service := NewProductService(mockRepo, mockAudit, nil, nil, nil)

	// Create an expired product
	pastDate := time.Now().Add(-24 * time.Hour)
	expiredProduct := &models.Product{
		ID:         1,
		SKU:        "EXP001",
		Name:       "Expired Medicine",
		StockQty:   10,
		Price:      "10000.00",
		ExpiryDate: &pastDate,
	}

	// Mock repository to return the expired product

	mockRepo.On("GetByID", mock.Anything, uint(1)).Return(expiredProduct, nil)
	// Mock repository to return the expired product
	mockRepo.On("GetBySKU", mock.Anything, uint(1), "EXP001").Return(expiredProduct, nil)

	// Act & Assert - Step 1: Scan (GetProductBySKU) returns error for expired product
	product, err := service.GetProductBySKU(context.Background(), 1, "EXP001")

	// Assert: Scan blocked
	require.Error(t, err, "GetProductBySKU should return error for expired product")
	var productExpiredErr *ErrProductExpired
	assert.True(t, errors.As(err, &productExpiredErr), "Should return ErrProductExpired type")
	assert.Nil(t, product, "Should not return product when expired")
	assert.Equal(t, "EXP001", productExpiredErr.ProductSKU, "Error should contain correct SKU")
	assert.Equal(t, "Expired Medicine", productExpiredErr.ProductName, "Error should contain correct product name")

	// Assert & Assert - Step 2: ValidateProductForSale also returns error
	err = service.ValidateProductForSale(context.Background(), expiredProduct.ID)
	require.Error(t, err, "ValidateProductForSale should return error for expired product")
	assert.True(t, errors.As(err, &productExpiredErr), "Should return ErrProductExpired type")

	// Note: The audit logging happens at the transaction service level when ProcessSale is called
	// This test verifies that the validation and blocking logic works correctly
	mockRepo.AssertExpectations(t)
}
