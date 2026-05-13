package services

import (
	"context"
	"time"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/models"
)

// TransactionService defines the interface for transaction business operations
// AC1: Service interface for transaction domain with clear business method signatures
type TransactionService interface {
	// CreateTransaction creates a new transaction with transaction number generation
	CreateTransaction(ctx context.Context, transaction *models.Transaction) error

	// ProcessSale processes a sale with transactional operations
	// Validates all products exist and have sufficient stock
	// Performs atomic operations: deduct stock, create transaction, create items
	// All-or-nothing semantics with rollback on any error
	ProcessSale(ctx context.Context, sale *SaleRequest, cashierID uint, branchID uint) (*models.Transaction, error)

	// CalculateTotal calculates the total amount for a sale
	// Returns sum of (quantity * unit_price) for all items
	CalculateTotal(items []*SaleItem) (string, error)

	// GenerateReceiptData generates receipt structure for printing
	GenerateReceiptData(ctx context.Context, transactionID uint) (*ReceiptData, error)

	// GetTransactionByID retrieves a transaction by ID with relationships
	GetTransactionByID(ctx context.Context, id uint) (*models.Transaction, error)

	// ListTransactions retrieves transactions with filtering and pagination
	// Filters: date range, cashier, payment method
	ListTransactions(ctx context.Context, filter *TransactionFilter) ([]*models.Transaction, int64, error)
}

// SaleRequest represents a sale request
type SaleRequest struct {
	Items         []*SaleItem
	PaymentMethod string
	CustomerName  string
	TaxAmount     string
	DiscountAmount string
}

// SaleItem represents a single item in a sale
type SaleItem struct {
	ProductID uint
	Quantity  int64
	UnitPrice string
}

// ReceiptData represents structured receipt data for printing
type ReceiptData struct {
	TransactionNumber string
	TransactionDate   time.Time
	CashierName       string
	BranchName        string
	CustomerName      string
	Items             []*ReceiptItem
	Subtotal          string
	TaxAmount         string
	DiscountAmount    string
	Total             string
	PaymentMethod     string
}

// ReceiptItem represents a single item in receipt
type ReceiptItem struct {
	ProductName string
	SKU         string
	Quantity    int64
	UnitPrice   string
	Total       string
}

// TransactionFilter defines filtering options for transaction listing
type TransactionFilter struct {
	BranchID          *uint
	CashierID         *uint
	StartDate         *time.Time
	EndDate           *time.Time
	PaymentMethod     string
	Status            string
	TransactionNumber string
	CustomerName      string
	Page              int
	Limit             int
	SortBy            string
	SortOrder         string
}
