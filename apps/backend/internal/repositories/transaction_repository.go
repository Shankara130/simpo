package repositories

import (
	"context"
	"time"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/models"
)

// TransactionRepository defines the interface for transaction data operations
// AC1: Repository interface with CRUD methods for Transaction entity
type TransactionRepository interface {
	// Create inserts a new transaction into the database
	Create(ctx context.Context, transaction *models.Transaction) error

	// GetByID retrieves a transaction by its ID
	// Returns ErrNotFound if transaction doesn't exist
	GetByID(ctx context.Context, id uint) (*models.Transaction, error)

	// GetByTransactionNumber retrieves a transaction by its transaction number
	// Returns ErrNotFound if transaction doesn't exist
	GetByTransactionNumber(ctx context.Context, transactionNumber string) (*models.Transaction, error)

	// CRITICAL-003: GetByIdempotencyKey retrieves a transaction by its idempotency key
	// Used to implement idempotent transaction creation
	GetByIdempotencyKey(ctx context.Context, key string) (*models.Transaction, error)

	// GetNextTransactionNumber gets the next sequential transaction number for a branch and date
	// Returns the next sequential number (e.g., 1 for first transaction of the day)
	GetNextTransactionNumber(ctx context.Context, branchID uint, dateStr string) (int, error)

	// Update modifies an existing transaction in the database
	Update(ctx context.Context, transaction *models.Transaction) error

	// Delete removes a transaction from the database (soft delete)
	Delete(ctx context.Context, id uint) error

	// List retrieves transactions with optional filtering and pagination
	// Returns slice of transactions, total count, and error
	List(ctx context.Context, filter *TransactionFilter) ([]*models.Transaction, int64, error)

	// GetDailySummary retrieves daily transaction summary for a branch
	GetDailySummary(ctx context.Context, branchID uint, date time.Time) (*TransactionSummary, error)

	// GetMonthlySummary retrieves monthly transaction summary for a branch
	GetMonthlySummary(ctx context.Context, branchID uint, year int, month time.Month) (*TransactionSummary, error)

	// CreateWithItems creates a transaction with its items in a single transaction
	CreateWithItems(ctx context.Context, transaction *models.Transaction, items []*models.TransactionItem) error

	// ProcessSaleWithStockUpdate processes a sale with atomic operations
	// Wraps stock updates and transaction creation in a single database transaction
	// Uses SELECT FOR UPDATE to prevent race conditions on stock
	ProcessSaleWithStockUpdate(ctx context.Context, transaction *models.Transaction, items []*models.TransactionItem, stockUpdates map[uint]int64) error
}

// TransactionFilter defines filtering options for transaction listing
// AC4: Complex query support with filtering by branch, cashier, date range, payment method, status
type TransactionFilter struct {
	BranchID         *uint     // Filter by branch
	CashierID        *uint     // Filter by cashier
	StartDate         *time.Time // Filter for transactions on or after this date
	EndDate           *time.Time // Filter for transactions on or before this date
	PaymentMethod    string    // Filter by payment method (CASH, TRANSFER, E-WALLET, etc.)
	Status           string    // Filter by status (COMPLETED, CANCELLED, etc.)
	TransactionNumber string   // Filter by transaction number (partial match)
	CustomerName     string    // Filter by customer name (partial match)
	Page              int       // Page number (1-indexed)
	Limit             int       // Items per page
	SortBy            string    // Field to sort by
	SortOrder         string    // "asc" or "desc"
}

// TransactionSummary represents aggregated transaction data
type TransactionSummary struct {
	TotalTransactions int64    `json:"total_transactions"`
	TotalAmount        string   `json:"total_amount"`     // Decimal string for precision
	SubtotalAmount     string   `json:"subtotal_amount"`   // Before tax and discount
	TaxAmount          string   `json:"tax_amount"`        // Total tax collected
	DiscountAmount     string   `json:"discount_amount"`   // Total discount given
	PaymentMethods     []PaymentMethodSummary `json:"payment_methods"`
}

// PaymentMethodSummary represents transaction count and total per payment method
type PaymentMethodSummary struct {
	PaymentMethod string `json:"payment_method"`
	Count         int64  `json:"count"`
	TotalAmount   string `json:"total_amount"`
}
