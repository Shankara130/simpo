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
		NewProductService(nil, &MockAuditService{}, nil, nil)
	}, "Should panic when productRepo is nil")

	assert.Panics(t, func() {
		NewProductService(&MockProductRepository{}, nil, nil, nil)
	}, "Should panic when auditService is nil")
}

// Test CreateProduct
func TestProductService_CreateProduct_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockProductRepository)
	mockAudit := new(MockAuditService)
	service := NewProductService(mockRepo, mockAudit, nil, nil)

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
	service := NewProductService(mockRepo, mockAudit, nil, nil)

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
	service := NewProductService(mockRepo, mockAudit, nil, nil)

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
	service := NewProductService(mockRepo, mockAudit, nil, nil)

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
	service := NewProductService(mockRepo, mockAudit, nil, nil)

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
	service := NewProductService(mockRepo, mockAudit, nil, nil)

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
	service := NewProductService(mockRepo, mockAudit, nil, nil)

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
	service := NewProductService(mockRepo, mockAudit, nil, nil)

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
	service := NewProductService(mockRepo, mockAudit, nil, nil)

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
	service := NewProductService(mockRepo, mockAudit, nil, nil)

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
	service := NewProductService(mockRepo, mockAudit, nil, nil)

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
	service := NewProductService(mockRepo, mockAudit, nil, nil)

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
	service := NewProductService(mockRepo, mockAudit, nil, nil)

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
	service := NewProductService(mockRepo, mockAudit, nil, nil)

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
