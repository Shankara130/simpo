package services

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/models"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/repositories"
)

// MockTransactionRepository is a mock implementation of TransactionRepository
type MockTransactionRepository struct {
	mock.Mock
}

func (m *MockTransactionRepository) Create(ctx context.Context, transaction *models.Transaction) error {
	args := m.Called(ctx, transaction)
	return args.Error(0)
}

func (m *MockTransactionRepository) GetByID(ctx context.Context, id uint) (*models.Transaction, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Transaction), args.Error(1)
}

func (m *MockTransactionRepository) GetByTransactionNumber(ctx context.Context, transactionNumber string) (*models.Transaction, error) {
	args := m.Called(ctx, transactionNumber)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Transaction), args.Error(1)
}

// CRITICAL-003: GetByIdempotencyKey mock implementation
func (m *MockTransactionRepository) GetByIdempotencyKey(ctx context.Context, key string) (*models.Transaction, error) {
	args := m.Called(ctx, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Transaction), args.Error(1)
}

func (m *MockTransactionRepository) Update(ctx context.Context, transaction *models.Transaction) error {
	args := m.Called(ctx, transaction)
	return args.Error(0)
}

func (m *MockTransactionRepository) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockTransactionRepository) List(ctx context.Context, filter *repositories.TransactionFilter) ([]*models.Transaction, int64, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*models.Transaction), args.Get(1).(int64), args.Error(2)
}

func (m *MockTransactionRepository) GetDailySummary(ctx context.Context, branchID uint, date time.Time) (*repositories.TransactionSummary, error) {
	args := m.Called(ctx, branchID, date)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repositories.TransactionSummary), args.Error(1)
}

func (m *MockTransactionRepository) GetMonthlySummary(ctx context.Context, branchID uint, year int, month time.Month) (*repositories.TransactionSummary, error) {
	args := m.Called(ctx, branchID, year, month)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repositories.TransactionSummary), args.Error(1)
}

func (m *MockTransactionRepository) CreateWithItems(ctx context.Context, transaction *models.Transaction, items []*models.TransactionItem) error {
	args := m.Called(ctx, transaction, items)
	return args.Error(0)
}

func (m *MockTransactionRepository) ProcessSaleWithStockUpdate(ctx context.Context, transaction *models.Transaction, items []*models.TransactionItem, stockUpdates map[uint]int64) error {
	args := m.Called(ctx, transaction, items, stockUpdates)
	return args.Error(0)
}

func (m *MockTransactionRepository) GetNextTransactionNumber(ctx context.Context, branchID uint, dateStr string) (int, error) {
	args := m.Called(ctx, branchID, dateStr)
	return args.Int(0), args.Error(1)
}

// MockTransactionItemRepository is a mock implementation of TransactionItemRepository
type MockTransactionItemRepository struct {
	mock.Mock
}

func (m *MockTransactionItemRepository) Create(ctx context.Context, item *models.TransactionItem) error {
	args := m.Called(ctx, item)
	return args.Error(0)
}

func (m *MockTransactionItemRepository) GetByID(ctx context.Context, id uint) (*models.TransactionItem, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.TransactionItem), args.Error(1)
}

func (m *MockTransactionItemRepository) GetByTransactionID(ctx context.Context, transactionID uint) ([]*models.TransactionItem, error) {
	args := m.Called(ctx, transactionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.TransactionItem), args.Error(1)
}

func (m *MockTransactionItemRepository) Update(ctx context.Context, item *models.TransactionItem) error {
	args := m.Called(ctx, item)
	return args.Error(0)
}

func (m *MockTransactionItemRepository) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockTransactionItemRepository) List(ctx context.Context, filter *repositories.TransactionItemFilter) ([]*models.TransactionItem, int64, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*models.TransactionItem), args.Get(1).(int64), args.Error(2)
}

func (m *MockTransactionItemRepository) CreateBatch(ctx context.Context, items []*models.TransactionItem) error {
	args := m.Called(ctx, items)
	return args.Error(0)
}

// Test NewTransactionService with nil dependencies
func TestNewTransactionService_PanicOnNilDependencies(t *testing.T) {
	mockItemRepo := new(MockTransactionItemRepository)
	mockProdRepo := new(MockProductRepository)
	mockAudit := new(MockAuditService)
	mockTxnRepo := new(MockTransactionRepository)

	assert.Panics(t, func() {
		NewTransactionService(nil, mockItemRepo, mockProdRepo, mockAudit, nil)
	}, "Should panic when transactionRepo is nil")

	assert.Panics(t, func() {
		NewTransactionService(mockTxnRepo, nil, mockProdRepo, mockAudit, nil)
	}, "Should panic when transactionItemRepo is nil")

	assert.Panics(t, func() {
		NewTransactionService(mockTxnRepo, mockItemRepo, nil, mockAudit, nil)
	}, "Should panic when productRepo is nil")

	assert.Panics(t, func() {
		NewTransactionService(mockTxnRepo, mockItemRepo, mockProdRepo, nil, nil)
	}, "Should panic when auditService is nil")
}

// Test CreateTransaction
func TestTransactionService_CreateTransaction_Success(t *testing.T) {
	// Arrange
	mockTxnRepo := new(MockTransactionRepository)
	mockItemRepo := new(MockTransactionItemRepository)
	mockProdRepo := new(MockProductRepository)
	mockAudit := new(MockAuditService)
	service := NewTransactionService(mockTxnRepo, mockItemRepo, mockProdRepo, mockAudit, nil)

	transaction := &models.Transaction{
		CashierID:     1,
		BranchID:      1,
		Total:         "100.00",
		PaymentMethod: "CASH",
	}

	// Mock expectations
	mockTxnRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.Transaction")).Return(nil)

	// Act
	err := service.CreateTransaction(context.Background(), transaction)

	// Assert
	assert.NoError(t, err)
	assert.NotEmpty(t, transaction.TransactionNumber)
	mockTxnRepo.AssertExpectations(t)
}

// Test ProcessSale
func TestTransactionService_ProcessSale_Success(t *testing.T) {
	// Arrange
	mockTxnRepo := new(MockTransactionRepository)
	mockItemRepo := new(MockTransactionItemRepository)
	mockProdRepo := new(MockProductRepository)
	mockAudit := new(MockAuditService)
	service := NewTransactionService(mockTxnRepo, mockItemRepo, mockProdRepo, mockAudit, nil)

	sale := &SaleRequest{
		Items: []*SaleItem{
			{ProductID: 1, Quantity: 2, UnitPrice: "50.00"},
		},
		PaymentMethod: "CASH",
		CustomerName:  "",
	}

	product := &models.Product{
		ID:       1,
		Name:     "Test Product",
		StockQty: 100,
	}

	// Mock expectations
	mockProdRepo.On("GetByID", mock.Anything, uint(1)).Return(product, nil)
	mockTxnRepo.On("ProcessSaleWithStockUpdate", mock.Anything, mock.Anything, mock.Anything, map[uint]int64{1: -2}).Return(nil)

	// Act
	result, err := service.ProcessSale(context.Background(), sale, 1, 1)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.TransactionNumber)
	mockProdRepo.AssertExpectations(t)
	mockTxnRepo.AssertExpectations(t)
}

func TestTransactionService_ProcessSale_EmptyItems(t *testing.T) {
	// Arrange
	mockTxnRepo := new(MockTransactionRepository)
	mockItemRepo := new(MockTransactionItemRepository)
	mockProdRepo := new(MockProductRepository)
	mockAudit := new(MockAuditService)
	service := NewTransactionService(mockTxnRepo, mockItemRepo, mockProdRepo, mockAudit, nil)

	sale := &SaleRequest{
		Items:         []*SaleItem{},
		PaymentMethod: "CASH",
	}

	// Act
	_, err := service.ProcessSale(context.Background(), sale, 1, 1)

	// Assert
	assert.Error(t, err)
	var invErr *InvalidInputError
	assert.True(t, errors.As(err, &invErr))
	assert.Equal(t, "items", invErr.Field)
}

func TestTransactionService_ProcessSale_InsufficientStock(t *testing.T) {
	// Arrange
	mockTxnRepo := new(MockTransactionRepository)
	mockItemRepo := new(MockTransactionItemRepository)
	mockProdRepo := new(MockProductRepository)
	mockAudit := new(MockAuditService)
	service := NewTransactionService(mockTxnRepo, mockItemRepo, mockProdRepo, mockAudit, nil)

	sale := &SaleRequest{
		Items: []*SaleItem{
			{ProductID: 1, Quantity: 50, UnitPrice: "50.00"},
		},
		PaymentMethod: "CASH",
	}

	product := &models.Product{
		ID:       1,
		Name:     "Test Product",
		StockQty: 10, // Less than requested
	}

	// Mock expectations - ProcessSaleWithStockUpdate will handle stock validation
	mockProdRepo.On("GetByID", mock.Anything, uint(1)).Return(product, nil)
	// ProcessSaleWithStockUpdate returns error for insufficient stock
	stockErr := fmt.Errorf("insufficient stock for product Test Product (ID: 1). Available: 10, Requested: 50")
	mockTxnRepo.On("ProcessSaleWithStockUpdate", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(stockErr)

	// Act
	_, err := service.ProcessSale(context.Background(), sale, 1, 1)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient stock")
}

// Test CalculateTotal
func TestTransactionService_CalculateTotal_EmptyItems(t *testing.T) {
	// Arrange
	mockTxnRepo := new(MockTransactionRepository)
	mockItemRepo := new(MockTransactionItemRepository)
	mockProdRepo := new(MockProductRepository)
	mockAudit := new(MockAuditService)
	service := NewTransactionService(mockTxnRepo, mockItemRepo, mockProdRepo, mockAudit, nil)

	items := []*SaleItem{}

	// Act
	total, err := service.CalculateTotal(items)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "0.00", total)
}

// Test GetTransactionByID
func TestTransactionService_GetTransactionByID_ZeroID(t *testing.T) {
	// Arrange
	mockTxnRepo := new(MockTransactionRepository)
	mockItemRepo := new(MockTransactionItemRepository)
	mockProdRepo := new(MockProductRepository)
	mockAudit := new(MockAuditService)
	service := NewTransactionService(mockTxnRepo, mockItemRepo, mockProdRepo, mockAudit, nil)

	// Act
	_, err := service.GetTransactionByID(context.Background(), 0)

	// Assert
	assert.Error(t, err)
	var invErr *InvalidInputError
	assert.True(t, errors.As(err, &invErr))
	assert.Equal(t, "id", invErr.Field)
}

// Test ListTransactions
func TestTransactionService_ListTransactions_DefaultPagination(t *testing.T) {
	// Arrange
	mockTxnRepo := new(MockTransactionRepository)
	mockItemRepo := new(MockTransactionItemRepository)
	mockProdRepo := new(MockProductRepository)
	mockAudit := new(MockAuditService)
	service := NewTransactionService(mockTxnRepo, mockItemRepo, mockProdRepo, mockAudit, nil)

	transactions := []*models.Transaction{
		{ID: 1, TransactionNumber: "TRX-001"},
	}

	// Mock expectations - expect default pagination
	mockTxnRepo.On("List", mock.Anything, mock.MatchedBy(func(f *repositories.TransactionFilter) bool {
		return f.Page == 1 && f.Limit == 20
	})).Return(transactions, int64(1), nil)

	// Act
	_, _, err := service.ListTransactions(context.Background(), nil)

	// Assert
	assert.NoError(t, err)
	mockTxnRepo.AssertExpectations(t)
}

func TestTransactionService_ListTransactions_MaxPaginationLimit(t *testing.T) {
	// Arrange
	mockTxnRepo := new(MockTransactionRepository)
	mockItemRepo := new(MockTransactionItemRepository)
	mockProdRepo := new(MockProductRepository)
	mockAudit := new(MockAuditService)
	service := NewTransactionService(mockTxnRepo, mockItemRepo, mockProdRepo, mockAudit, nil)

	filter := &TransactionFilter{
		Limit: 5000, // Exceeds max
		Page:  1,
	}

	// Mock expectations - expect capped limit
	mockTxnRepo.On("List", mock.Anything, mock.MatchedBy(func(f *repositories.TransactionFilter) bool {
		return f.Limit == 1000 // Max limit applied
	})).Return([]*models.Transaction{}, int64(0), nil)

	// Act
	_, _, err := service.ListTransactions(context.Background(), filter)

	// Assert
	assert.NoError(t, err)
	mockTxnRepo.AssertExpectations(t)
}

func TestTransactionService_ProcessSale_ContextCanceled(t *testing.T) {
	// Arrange
	mockTxnRepo := new(MockTransactionRepository)
	mockItemRepo := new(MockTransactionItemRepository)
	mockProdRepo := new(MockProductRepository)
	mockAudit := new(MockAuditService)
	service := NewTransactionService(mockTxnRepo, mockItemRepo, mockProdRepo, mockAudit, nil)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	sale := &SaleRequest{
		Items: []*SaleItem{
			{ProductID: 1, Quantity: 2, UnitPrice: "50.00"},
		},
		PaymentMethod: "CASH",
	}

	// Act
	_, err := service.ProcessSale(ctx, sale, 1, 1)

	// Assert
	assert.Error(t, err)
}

// Story 3.7: Transaction History Service Tests

// TestTransactionService_ListTransactions_WithStatusFilter tests filtering by status
func TestTransactionService_ListTransactions_WithStatusFilter(t *testing.T) {
	// Arrange
	mockTxnRepo := new(MockTransactionRepository)
	mockItemRepo := new(MockTransactionItemRepository)
	mockProdRepo := new(MockProductRepository)
	mockAudit := new(MockAuditService)
	service := NewTransactionService(mockTxnRepo, mockItemRepo, mockProdRepo, mockAudit, nil)

	status := "COMPLETED"
	branchID := uint(1)
	filter := &TransactionFilter{
		BranchID: &branchID,
		Status:   status,
		Page:     1,
		Limit:    20,
	}

	// Mock expectations - expect status filter to be passed through
	mockTxnRepo.On("List", mock.Anything, mock.MatchedBy(func(f *repositories.TransactionFilter) bool {
		return f.Status == status && *f.BranchID == branchID
	})).Return([]*models.Transaction{}, int64(0), nil)

	// Act
	_, _, err := service.ListTransactions(context.Background(), filter)

	// Assert
	assert.NoError(t, err)
	mockTxnRepo.AssertExpectations(t)
}

// TestTransactionService_ListTransactions_WithDateRangeFilter tests filtering by date range
func TestTransactionService_ListTransactions_WithDateRangeFilter(t *testing.T) {
	// Arrange
	mockTxnRepo := new(MockTransactionRepository)
	mockItemRepo := new(MockTransactionItemRepository)
	mockProdRepo := new(MockProductRepository)
	mockAudit := new(MockAuditService)
	service := NewTransactionService(mockTxnRepo, mockItemRepo, mockProdRepo, mockAudit, nil)

	startDate := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 5, 15, 0, 0, 0, 0, time.UTC)
	branchID := uint(1)

	filter := &TransactionFilter{
		BranchID:  &branchID,
		StartDate: &startDate,
		EndDate:   &endDate,
		Page:      1,
		Limit:     20,
	}

	// Mock expectations - expect date range filter to be passed through
	mockTxnRepo.On("List", mock.Anything, mock.MatchedBy(func(f *repositories.TransactionFilter) bool {
		return f.StartDate != nil && f.EndDate != nil
	})).Return([]*models.Transaction{}, int64(0), nil)

	// Act
	_, _, err := service.ListTransactions(context.Background(), filter)

	// Assert
	assert.NoError(t, err)
	mockTxnRepo.AssertExpectations(t)
}

// TestTransactionService_ListTransactions_DateRangeExceedsMax tests validation of date range
func TestTransactionService_ListTransactions_DateRangeExceedsMax(t *testing.T) {
	// Arrange
	mockTxnRepo := new(MockTransactionRepository)
	mockItemRepo := new(MockTransactionItemRepository)
	mockProdRepo := new(MockProductRepository)
	mockAudit := new(MockAuditService)
	service := NewTransactionService(mockTxnRepo, mockItemRepo, mockProdRepo, mockAudit, nil)

	// Date range exceeding 1 year
	startDate := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2027, 6, 1, 0, 0, 0, 0, time.UTC)
	branchID := uint(1)

	filter := &TransactionFilter{
		BranchID:  &branchID,
		StartDate: &startDate,
		EndDate:   &endDate,
		Page:      1,
		Limit:     20,
	}

	// Act
	_, _, err := service.ListTransactions(context.Background(), filter)

	// Assert
	assert.Error(t, err)
	var invErr *InvalidInputError
	assert.True(t, errors.As(err, &invErr))
	assert.Equal(t, "date_range", invErr.Field)
}

// TestTransactionService_ListTransactions_WithSortOrder tests descending sort order
func TestTransactionService_ListTransactions_WithSortOrder(t *testing.T) {
	// Arrange
	mockTxnRepo := new(MockTransactionRepository)
	mockItemRepo := new(MockTransactionItemRepository)
	mockProdRepo := new(MockProductRepository)
	mockAudit := new(MockAuditService)
	service := NewTransactionService(mockTxnRepo, mockItemRepo, mockProdRepo, mockAudit, nil)

	branchID := uint(1)
	filter := &TransactionFilter{
		BranchID:  &branchID,
		SortBy:    "created_at",
		SortOrder: "desc", // Newest first
		Page:      1,
		Limit:     20,
	}

	// Mock expectations - expect sort order to be passed through
	mockTxnRepo.On("List", mock.Anything, mock.MatchedBy(func(f *repositories.TransactionFilter) bool {
		return f.SortBy == "created_at" && f.SortOrder == "desc"
	})).Return([]*models.Transaction{}, int64(0), nil)

	// Act
	_, _, err := service.ListTransactions(context.Background(), filter)

	// Assert
	assert.NoError(t, err)
	mockTxnRepo.AssertExpectations(t)
}

// TestTransactionService_ListTransactions_InvalidSortField tests validation of sort field
func TestTransactionService_ListTransactions_InvalidSortField(t *testing.T) {
	// Arrange
	mockTxnRepo := new(MockTransactionRepository)
	mockItemRepo := new(MockTransactionItemRepository)
	mockProdRepo := new(MockProductRepository)
	mockAudit := new(MockAuditService)
	service := NewTransactionService(mockTxnRepo, mockItemRepo, mockProdRepo, mockAudit, nil)

	branchID := uint(1)
	filter := &TransactionFilter{
		BranchID:  &branchID,
		SortBy:    "invalid_field", // Invalid sort field
		SortOrder: "desc",
		Page:      1,
		Limit:     20,
	}

	// Act
	_, _, err := service.ListTransactions(context.Background(), filter)

	// Assert
	assert.Error(t, err)
	var invErr *InvalidInputError
	assert.True(t, errors.As(err, &invErr))
	assert.Equal(t, "sort_by", invErr.Field)
}

// TestTransactionService_ListTransactions_RBACBranchFilter tests RBAC via branch filtering
func TestTransactionService_ListTransactions_RBACBranchFilter(t *testing.T) {
	// Arrange
	mockTxnRepo := new(MockTransactionRepository)
	mockItemRepo := new(MockTransactionItemRepository)
	mockProdRepo := new(MockProductRepository)
	mockAudit := new(MockAuditService)
	service := NewTransactionService(mockTxnRepo, mockItemRepo, mockProdRepo, mockAudit, nil)

	cashierBranchID := uint(1)
	filter := &TransactionFilter{
		BranchID: &cashierBranchID, // Only transactions from cashier's branch
		Page:     1,
		Limit:    20,
	}

	// Mock expectations - verify branch ID filter is passed
	mockTxnRepo.On("List", mock.Anything, mock.MatchedBy(func(f *repositories.TransactionFilter) bool {
		return f.BranchID != nil && *f.BranchID == cashierBranchID
	})).Return([]*models.Transaction{}, int64(0), nil)

	// Act
	_, _, err := service.ListTransactions(context.Background(), filter)

	// Assert
	assert.NoError(t, err)
	mockTxnRepo.AssertExpectations(t)
}

// TestTransactionService_GetTransactionByID_Success tests successful retrieval
func TestTransactionService_GetTransactionByID_Success(t *testing.T) {
	// Arrange
	mockTxnRepo := new(MockTransactionRepository)
	mockItemRepo := new(MockTransactionItemRepository)
	mockProdRepo := new(MockProductRepository)
	mockAudit := new(MockAuditService)
	service := NewTransactionService(mockTxnRepo, mockItemRepo, mockProdRepo, mockAudit, nil)

	expectedTransaction := &models.Transaction{
		ID:                1,
		TransactionNumber: "TRX-20260515-0001",
		CashierID:         100,
		BranchID:          1,
		Total:             "150000.00",
		Status:            models.StatusCompleted,
	}

	mockTxnRepo.On("GetByID", mock.Anything, uint(1)).Return(expectedTransaction, nil)

	// Act
	result, err := service.GetTransactionByID(context.Background(), 1)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedTransaction.ID, result.ID)
	assert.Equal(t, expectedTransaction.TransactionNumber, result.TransactionNumber)
	mockTxnRepo.AssertExpectations(t)
}

// TestTransactionService_GetTransactionByID_NotFound tests not found error
func TestTransactionService_GetTransactionByID_NotFound(t *testing.T) {
	// Arrange
	mockTxnRepo := new(MockTransactionRepository)
	mockItemRepo := new(MockTransactionItemRepository)
	mockProdRepo := new(MockProductRepository)
	mockAudit := new(MockAuditService)
	service := NewTransactionService(mockTxnRepo, mockItemRepo, mockProdRepo, mockAudit, nil)

	mockTxnRepo.On("GetByID", mock.Anything, uint(999)).Return(nil, repositories.ErrNotFound)

	// Act
	_, err := service.GetTransactionByID(context.Background(), 999)

	// Assert
	assert.Error(t, err)
	var svcErr *ServiceError
	assert.True(t, errors.As(err, &svcErr))
}
