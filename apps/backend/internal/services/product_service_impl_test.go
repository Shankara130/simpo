package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/models"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/repositories"
)

// MockProductRepository is a mock implementation of ProductRepository
type MockProductRepository struct {
	mock.Mock
}

func (m *MockProductRepository) Create(ctx context.Context, product *models.Product) error {
	args := m.Called(ctx, product)
	return args.Error(0)
}

func (m *MockProductRepository) GetByID(ctx context.Context, id uint) (*models.Product, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Product), args.Error(1)
}

func (m *MockProductRepository) GetBySKU(ctx context.Context, branchID uint, sku string) (*models.Product, error) {
	args := m.Called(ctx, branchID, sku)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Product), args.Error(1)
}

func (m *MockProductRepository) Update(ctx context.Context, product *models.Product) error {
	args := m.Called(ctx, product)
	return args.Error(0)
}

func (m *MockProductRepository) UpdateStock(ctx context.Context, id uint, quantity int64) error {
	args := m.Called(ctx, id, quantity)
	return args.Error(0)
}

func (m *MockProductRepository) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockProductRepository) List(ctx context.Context, filter *repositories.ProductFilter) ([]*models.Product, int64, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*models.Product), args.Get(1).(int64), args.Error(2)
}

func (m *MockProductRepository) GetLowStockProducts(ctx context.Context, branchID uint) ([]*models.Product, error) {
	args := m.Called(ctx, branchID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Product), args.Error(1)
}

func (m *MockProductRepository) GetExpiredProducts(ctx context.Context, branchID uint) ([]*models.Product, error) {
	args := m.Called(ctx, branchID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Product), args.Error(1)
}

// Test NewProductService with nil dependencies
func TestNewProductService_PanicOnNilDependencies(t *testing.T) {
	assert.Panics(t, func() {
			NewProductService(nil, &MockAuditService{}, nil, nil, nil)
	}, "Should panic when productRepo is nil")

	assert.Panics(t, func() {
		NewProductService(&MockProductRepository{}, nil, nil, nil, nil)
	}, "Should panic when auditService is nil")
}

// Test CreateProduct
func TestProductService_CreateProduct_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockProductRepository)
	mockAudit := new(MockAuditService)
	service := NewProductService(mockRepo, mockAudit, nil, nil, nil)

	product := &models.Product{
		SKU:         "TEST-001",
		Name:        "Test Product",
		StockQty:    100,
		Price:       "50000.00",
		BranchID:    1,
		Description: "Test description",
	}

	// Mock expectations
	mockRepo.On("GetBySKU", mock.Anything, uint(1), "TEST-001").Return(nil, errors.New("not found"))
	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.Product")).Return(nil)

	// Act
	err := service.CreateProduct(context.Background(), product)

	// Assert
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestProductService_CreateProduct_EmptySKU(t *testing.T) {
	// Arrange
	mockRepo := new(MockProductRepository)
	mockAudit := new(MockAuditService)
	service := NewProductService(mockRepo, mockAudit, nil, nil, nil)

	product := &models.Product{
		SKU:      "",
		Name:     "Test Product",
		Price:    "50000.00",
		BranchID: 1,
	}

	// Act
	err := service.CreateProduct(context.Background(), product)

	// Assert
	assert.Error(t, err)
	var invErr *InvalidInputError
	assert.True(t, errors.As(err, &invErr))
	assert.Equal(t, "sku", invErr.Field)
}

func TestProductService_CreateProduct_DuplicateSKU(t *testing.T) {
	// Arrange
	mockRepo := new(MockProductRepository)
	mockAudit := new(MockAuditService)
	service := NewProductService(mockRepo, mockAudit, nil, nil, nil)

	product := &models.Product{
		SKU:      "TEST-001",
		Name:     "Test Product",
		Price:    "50000.00",
		BranchID: 1,
	}

	existingProduct := &models.Product{ID: 1, SKU: "TEST-001"}

	// Mock expectations
	mockRepo.On("GetBySKU", mock.Anything, uint(1), "TEST-001").Return(existingProduct, nil)

	// Act
	err := service.CreateProduct(context.Background(), product)

	// Assert
	assert.Error(t, err)
	var dupErr *DuplicateSKUError
	assert.True(t, errors.As(err, &dupErr))
	assert.Equal(t, "TEST-001", dupErr.SKU)
}

func TestProductService_CreateProduct_ContextCanceled(t *testing.T) {
	// Arrange
	mockRepo := new(MockProductRepository)
	mockAudit := new(MockAuditService)
	service := NewProductService(mockRepo, mockAudit, nil, nil, nil)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	product := &models.Product{
		SKU:      "TEST-001",
		Name:     "Test Product",
		Price:    "50000.00",
		BranchID: 1,
	}

	// Act
	err := service.CreateProduct(ctx, product)

	// Assert
	assert.Error(t, err)
}

// Test UpdateProduct
func TestProductService_UpdateProduct_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockProductRepository)
	mockAudit := new(MockAuditService)
	service := NewProductService(mockRepo, mockAudit, nil, nil, nil)

	existing := &models.Product{
		ID:        1,
		SKU:       "TEST-001",
		Name:      "Old Name",
		StockQty:  50,
		Price:     "50000.00",
		BranchID:  1,
		CreatedAt: time.Now(),
	}

	updated := &models.Product{
		Name:     "New Name",
		StockQty: 100,
		Price:    "60000.00",
	}

	// Mock expectations
	mockRepo.On("GetByID", mock.Anything, uint(1)).Return(existing, nil)
	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*models.Product")).Return(nil)

	// Act
	err := service.UpdateProduct(context.Background(), 1, updated)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "TEST-001", updated.SKU) // SKU preserved
	mockRepo.AssertExpectations(t)
}

func TestProductService_UpdateProduct_CannotChangeSKU(t *testing.T) {
	// Arrange
	mockRepo := new(MockProductRepository)
	mockAudit := new(MockAuditService)
	service := NewProductService(mockRepo, mockAudit, nil, nil, nil)

	existing := &models.Product{
		ID:        1,
		SKU:       "TEST-001",
		Name:      "Test Product",
		StockQty:  50,
		Price:     "50000.00",
		BranchID:  1,
		CreatedAt: time.Now(),
	}

	updated := &models.Product{
		SKU:  "NEW-SKU", // Attempting to change SKU
		Name: "Updated Name",
	}

	// Mock expectations
	mockRepo.On("GetByID", mock.Anything, uint(1)).Return(existing, nil)

	// Act
	err := service.UpdateProduct(context.Background(), 1, updated)

	// Assert
	assert.Error(t, err)
	var invErr *InvalidInputError
	assert.True(t, errors.As(err, &invErr))
	assert.Equal(t, "sku", invErr.Field)
}

// Test CheckAvailability
func TestProductService_CheckAvailability_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockProductRepository)
	mockAudit := new(MockAuditService)
	service := NewProductService(mockRepo, mockAudit, nil, nil, nil)

	product := &models.Product{
		ID:       1,
		Name:     "Test Product",
		StockQty: 50,
	}

	// Mock expectations
	mockRepo.On("GetByID", mock.Anything, uint(1)).Return(product, nil)

	// Act
	available, err := service.CheckAvailability(context.Background(), 1, 30)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, int64(30), available)
}

func TestProductService_CheckAvailability_InsufficientStock(t *testing.T) {
	// Arrange
	mockRepo := new(MockProductRepository)
	mockAudit := new(MockAuditService)
	service := NewProductService(mockRepo, mockAudit, nil, nil, nil)

	product := &models.Product{
		ID:       1,
		Name:     "Test Product",
		StockQty: 20,
	}

	// Mock expectations
	mockRepo.On("GetByID", mock.Anything, uint(1)).Return(product, nil)

	// Act
	available, err := service.CheckAvailability(context.Background(), 1, 50)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, int64(20), available) // Returns available stock
}

func TestProductService_CheckAvailability_ExpiredProduct(t *testing.T) {
	// Arrange
	mockRepo := new(MockProductRepository)
	mockAudit := new(MockAuditService)
	service := NewProductService(mockRepo, mockAudit, nil, nil, nil)

	past := time.Now().Add(-24 * time.Hour)
	product := &models.Product{
		ID:         1,
		Name:       "Expired Product",
		StockQty:   50,
		ExpiryDate: &past,
	}

	// Mock expectations
	mockRepo.On("GetByID", mock.Anything, uint(1)).Return(product, nil)

	// Act
	_, err := service.CheckAvailability(context.Background(), 1, 30)

	// Assert
	assert.Error(t, err)
	var expErr *ProductExpiredError
	assert.True(t, errors.As(err, &expErr))
}

// Test GetLowStockProducts
func TestProductService_GetLowStockProducts_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockProductRepository)
	mockAudit := new(MockAuditService)
	service := NewProductService(mockRepo, mockAudit, nil, nil, nil)

	products := []*models.Product{
		{ID: 1, Name: "Product 1", StockQty: 5, ReorderThreshold: 10},
		{ID: 2, Name: "Product 2", StockQty: 8, ReorderThreshold: 10},
	}

	// Mock expectations
	mockRepo.On("GetLowStockProducts", mock.Anything, uint(1)).Return(products, nil)

	// Act
	result, err := service.GetLowStockProducts(context.Background(), 1)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, result, 2)
	mockRepo.AssertExpectations(t)
}

// Test ListProducts
func TestProductService_ListProducts_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockProductRepository)
	mockAudit := new(MockAuditService)
	service := NewProductService(mockRepo, mockAudit, nil, nil, nil)

	products := []*models.Product{
		{ID: 1, Name: "Product 1"},
		{ID: 2, Name: "Product 2"},
	}

	filter := &ProductFilter{
		BranchID: uintPtr(1),
		Page:     1,
		Limit:    20,
	}

	// Mock expectations
	mockRepo.On("List", mock.Anything, mock.AnythingOfType("*repositories.ProductFilter")).Return(products, int64(2), nil)

	// Act
	result, total, err := service.ListProducts(context.Background(), filter)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, int64(2), total)
	mockRepo.AssertExpectations(t)
}

func TestProductService_ListProducts_SanitizesWildcardCharacters(t *testing.T) {
	// Arrange
	mockRepo := new(MockProductRepository)
	mockAudit := new(MockAuditService)
	service := NewProductService(mockRepo, mockAudit, nil, nil, nil)

	filter := &ProductFilter{
		SearchQuery: "test%_wildcards",
		Page:        1,
		Limit:       20,
	}

	// Mock expectations - expect sanitized input
	mockRepo.On("List", mock.Anything, mock.MatchedBy(func(f *repositories.ProductFilter) bool {
		return f.SearchQuery == "testwildcards" // Wildcards removed
	})).Return([]*models.Product{}, int64(0), nil)

	// Act
	_, _, err := service.ListProducts(context.Background(), filter)

	// Assert
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestProductService_ListProducts_DefaultPagination(t *testing.T) {
	// Arrange
	mockRepo := new(MockProductRepository)
	mockAudit := new(MockAuditService)
	service := NewProductService(mockRepo, mockAudit, nil, nil, nil)

	// Mock expectations - expect default pagination
	mockRepo.On("List", mock.Anything, mock.MatchedBy(func(f *repositories.ProductFilter) bool {
		return f.Page == 1 && f.Limit == 20
	})).Return([]*models.Product{}, int64(0), nil)

	// Act
	_, _, err := service.ListProducts(context.Background(), nil)

	// Assert
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestProductService_ListProducts_MaxPaginationLimit(t *testing.T) {
	// Arrange
	mockRepo := new(MockProductRepository)
	mockAudit := new(MockAuditService)
	service := NewProductService(mockRepo, mockAudit, nil, nil, nil)

	filter := &ProductFilter{
		Limit: 5000, // Exceeds max
		Page:  1,
	}

	// Mock expectations - expect capped limit
	mockRepo.On("List", mock.Anything, mock.MatchedBy(func(f *repositories.ProductFilter) bool {
		return f.Limit == 1000 // Max limit applied
	})).Return([]*models.Product{}, int64(0), nil)

	// Act
	_, _, err := service.ListProducts(context.Background(), filter)

	// Assert
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

// Helper function
func uintPtr(v uint) *uint {
	return &v
}

// Test ManualAdjustStock
func TestProductService_ManualAdjustStock_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockProductRepository)
	mockAudit := new(MockAuditService)
	service := NewProductService(mockRepo, mockAudit, nil, nil, nil)

	product := &models.Product{
		ID:               1,
		SKU:              "TEST-001",
		Name:             "Test Product",
		StockQty:         50,
		Price:            "50000.00",
		BranchID:         1,
		ReorderThreshold: 10,
	}

	req := &StockAdjustmentRequest{
		ProductID:   1,
		BranchID:    1,
		NewStockQty: 75,
		Reason:      "DeliveryReceipt",
	}

	// Mock expectations
	mockRepo.On("GetByID", mock.Anything, uint(1)).Return(product, nil)
	mockRepo.On("UpdateStock", mock.Anything, uint(1), int64(25)).Return(nil)

	// Act
	result, err := service.ManualAdjustStock(context.Background(), req, 1, "admin")

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, uint(1), result.ProductID)
	assert.Equal(t, "TEST-001", result.SKU)
	assert.Equal(t, int64(50), result.OldStockQty)
	assert.Equal(t, int64(75), result.NewStockQty)
	assert.Equal(t, int64(25), result.Change)
	assert.Equal(t, "DeliveryReceipt", result.Reason)
	assert.Equal(t, "admin", result.AdjustedBy)
	mockRepo.AssertExpectations(t)
}

func TestProductService_ManualAdjustStock_NegativeStockRejected(t *testing.T) {
	// Arrange
	mockRepo := new(MockProductRepository)
	mockAudit := new(MockAuditService)
	service := NewProductService(mockRepo, mockAudit, nil, nil, nil)

	req := &StockAdjustmentRequest{
		ProductID:   1,
		BranchID:    1,
		NewStockQty: -5, // Negative stock
		Reason:      "Damage",
	}

	// Act
	result, err := service.ManualAdjustStock(context.Background(), req, 1, "admin")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	var invErr *InvalidInputError
	assert.True(t, errors.As(err, &invErr))
	assert.Equal(t, "new_stock_qty", invErr.Field)
}

func TestProductService_ManualAdjustStock_MissingReasonRejected(t *testing.T) {
	// Arrange
	mockRepo := new(MockProductRepository)
	mockAudit := new(MockAuditService)
	service := NewProductService(mockRepo, mockAudit, nil, nil, nil)

	req := &StockAdjustmentRequest{
		ProductID:   1,
		BranchID:    1,
		NewStockQty: 50,
		Reason:      "", // Empty reason
	}

	// Act
	result, err := service.ManualAdjustStock(context.Background(), req, 1, "admin")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	var invErr *InvalidInputError
	assert.True(t, errors.As(err, &invErr))
	assert.Equal(t, "reason", invErr.Field)
}

func TestProductService_ManualAdjustStock_ProductNotFound(t *testing.T) {
	// Arrange
	mockRepo := new(MockProductRepository)
	mockAudit := new(MockAuditService)
	service := NewProductService(mockRepo, mockAudit, nil, nil, nil)

	req := &StockAdjustmentRequest{
		ProductID:   999,
		BranchID:    1,
		NewStockQty: 50,
		Reason:      "Damage",
	}

	// Mock expectations
	mockRepo.On("GetByID", mock.Anything, uint(999)).Return(nil, errors.New("not found"))

	// Act
	result, err := service.ManualAdjustStock(context.Background(), req, 1, "admin")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	var prodErr *ProductNotFoundError
	assert.True(t, errors.As(err, &prodErr))
	assert.Equal(t, uint(999), prodErr.ProductID)
}

func TestProductService_ManualAdjustStock_BranchMismatch(t *testing.T) {
	// Arrange
	mockRepo := new(MockProductRepository)
	mockAudit := new(MockAuditService)
	service := NewProductService(mockRepo, mockAudit, nil, nil, nil)

	product := &models.Product{
		ID:       1,
		SKU:      "TEST-001",
		Name:     "Test Product",
		StockQty: 50,
		BranchID: 2, // Product belongs to branch 2
	}

	req := &StockAdjustmentRequest{
		ProductID:   1,
		BranchID:    1, // Trying to adjust for branch 1
		NewStockQty: 75,
		Reason:      "Damage",
	}

	// Mock expectations
	mockRepo.On("GetByID", mock.Anything, uint(1)).Return(product, nil)

	// Act
	result, err := service.ManualAdjustStock(context.Background(), req, 1, "admin")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	var invErr *InvalidInputError
	assert.True(t, errors.As(err, &invErr))
	assert.Equal(t, "branch_id", invErr.Field)
}

func TestProductService_ManualAdjustStock_InvalidReason(t *testing.T) {
	// Arrange
	mockRepo := new(MockProductRepository)
	mockAudit := new(MockAuditService)
	service := NewProductService(mockRepo, mockAudit, nil, nil, nil)

	product := &models.Product{
		ID:       1,
		SKU:      "TEST-001",
		Name:     "Test Product",
		StockQty: 50,
		BranchID: 1,
	}

	req := &StockAdjustmentRequest{
		ProductID:   1,
		BranchID:    1,
		NewStockQty: 75,
		Reason:      "InvalidReason", // Not in allowed list
	}

	// Mock expectations
	mockRepo.On("GetByID", mock.Anything, uint(1)).Return(product, nil)

	// Act
	result, err := service.ManualAdjustStock(context.Background(), req, 1, "admin")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	var invErr *InvalidInputError
	assert.True(t, errors.As(err, &invErr))
	assert.Equal(t, "reason", invErr.Field)
}

func TestProductService_ManualAdjustStock_OtherReasonWithoutNotes(t *testing.T) {
	// Arrange
	mockRepo := new(MockProductRepository)
	mockAudit := new(MockAuditService)
	service := NewProductService(mockRepo, mockAudit, nil, nil, nil)

	product := &models.Product{
		ID:       1,
		SKU:      "TEST-001",
		Name:     "Test Product",
		StockQty: 50,
		BranchID: 1,
	}

	req := &StockAdjustmentRequest{
		ProductID:   1,
		BranchID:    1,
		NewStockQty: 75,
		Reason:      "Other",
		ReasonNotes: "", // Missing notes for "Other" reason
	}

	// Mock expectations
	mockRepo.On("GetByID", mock.Anything, uint(1)).Return(product, nil)

	// Act
	result, err := service.ManualAdjustStock(context.Background(), req, 1, "admin")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	var invErr *InvalidInputError
	assert.True(t, errors.As(err, &invErr))
	assert.Equal(t, "reason_notes", invErr.Field)
}

func TestProductService_ManualAdjustStock_ValidReasons(t *testing.T) {
	// Arrange
	product := &models.Product{
		ID:       1,
		SKU:      "TEST-001",
		Name:     "Test Product",
		StockQty: 50,
		BranchID: 1,
	}

	validReasons := []string{"Damage", "Expiration", "DeliveryReceipt", "PhysicalCount", "TheftLoss", "Other"}

	for _, reason := range validReasons {
		t.Run(reason, func(t *testing.T) {
			// Reset mock for each iteration
			mockRepo := new(MockProductRepository)
			mockAudit := new(MockAuditService)
			service := NewProductService(mockRepo, mockAudit, nil, nil, nil)

			req := &StockAdjustmentRequest{
				ProductID:   1,
				BranchID:    1,
				NewStockQty: 75,
				Reason:      reason,
			}

			if reason == "Other" {
				req.ReasonNotes = "Additional details"
			}

			// Mock expectations
			mockRepo.On("GetByID", mock.Anything, uint(1)).Return(product, nil)
			mockRepo.On("UpdateStock", mock.Anything, uint(1), int64(25)).Return(nil)

			// Act
			result, err := service.ManualAdjustStock(context.Background(), req, 1, "admin")

			// Assert
			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, reason, result.Reason)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestProductService_ManualAdjustStock_LowStockTriggered(t *testing.T) {
	// Arrange
	mockRepo := new(MockProductRepository)
	mockAudit := new(MockAuditService)
	service := NewProductService(mockRepo, mockAudit, nil, nil, nil)

	product := &models.Product{
		ID:               1,
		SKU:              "TEST-001",
		Name:             "Test Product",
		StockQty:         50,
		Price:            "50000.00",
		BranchID:         1,
		ReorderThreshold: 10, // Threshold is 10
	}

	req := &StockAdjustmentRequest{
		ProductID:   1,
		BranchID:    1,
		NewStockQty: 5, // Below threshold
		Reason:      "Damage",
	}

	// Mock expectations
	mockRepo.On("GetByID", mock.Anything, uint(1)).Return(product, nil)
	mockRepo.On("UpdateStock", mock.Anything, uint(1), int64(-45)).Return(nil)

	// Act
	result, err := service.ManualAdjustStock(context.Background(), req, 1, "admin")

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int64(5), result.NewStockQty)
	mockRepo.AssertExpectations(t)
}

func TestProductService_ManualAdjustStock_ContextCanceled(t *testing.T) {
	// Arrange
	mockRepo := new(MockProductRepository)
	mockAudit := new(MockAuditService)
	service := NewProductService(mockRepo, mockAudit, nil, nil, nil)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	req := &StockAdjustmentRequest{
		ProductID:   1,
		BranchID:    1,
		NewStockQty: 75,
		Reason:      "DeliveryReceipt",
	}

	// Act
	result, err := service.ManualAdjustStock(ctx, req, 1, "admin")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestProductService_ManualAdjustStock_AllowsZeroStock(t *testing.T) {
	// Arrange
	mockRepo := new(MockProductRepository)
	mockAudit := new(MockAuditService)
	service := NewProductService(mockRepo, mockAudit, nil, nil, nil)

	product := &models.Product{
		ID:       1,
		SKU:      "TEST-001",
		Name:     "Test Product",
		StockQty: 50,
		BranchID: 1,
	}

	req := &StockAdjustmentRequest{
		ProductID:   1,
		BranchID:    1,
		NewStockQty: 0, // Zero stock is allowed
		Reason:      "Expiration",
	}

	// Mock expectations
	mockRepo.On("GetByID", mock.Anything, uint(1)).Return(product, nil)
	mockRepo.On("UpdateStock", mock.Anything, uint(1), int64(-50)).Return(nil)

	// Act
	result, err := service.ManualAdjustStock(context.Background(), req, 1, "admin")

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int64(0), result.NewStockQty)
	mockRepo.AssertExpectations(t)
}
