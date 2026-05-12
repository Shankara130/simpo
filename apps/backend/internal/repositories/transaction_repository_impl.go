package repositories

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/models"
)

// transactionRepository implements TransactionRepository interface
// AC2: GORM-based concrete implementation
type transactionRepository struct {
	db *gorm.DB
}

// NewTransactionRepository creates a new transaction repository
// AC5: Factory function for dependency injection
func NewTransactionRepository(db interface{}) TransactionRepository {
	gormDB, ok := db.(*gorm.DB)
	if !ok {
		panic("db must be *gorm.DB")
	}
	return &transactionRepository{db: gormDB}
}

// Create inserts a new transaction into the database
// AC3: Error handling with wrapping
func (r *transactionRepository) Create(ctx context.Context, transaction *models.Transaction) error {
	if transaction == nil {
		return fmt.Errorf("transaction cannot be nil")
	}
	err := r.db.WithContext(ctx).Create(transaction).Error
	if err != nil {
		return fmt.Errorf("failed to create transaction: %w", err)
	}
	return nil
}

// GetByID retrieves a transaction by its ID
// AC3: Distinguish between "not found" and other errors
// P-011: Add zero ID validation
func (r *transactionRepository) GetByID(ctx context.Context, id uint) (*models.Transaction, error) {
	if id == 0 {
		return nil, ErrInvalidInput
	}
	var transaction models.Transaction
	err := r.db.WithContext(ctx).Preload("TransactionItems").First(&transaction, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}
	return &transaction, nil
}

// GetByTransactionNumber retrieves a transaction by its transaction number
// AC3: Descriptive error for business logic consumption
func (r *transactionRepository) GetByTransactionNumber(ctx context.Context, transactionNumber string) (*models.Transaction, error) {
	var transaction models.Transaction
	err := r.db.WithContext(ctx).Preload("TransactionItems").Where("transaction_number = ?", transactionNumber).First(&transaction).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get transaction by number: %w", err)
	}
	return &transaction, nil
}

// Update modifies an existing transaction in the database
// AC3: Error wrapping for context
func (r *transactionRepository) Update(ctx context.Context, transaction *models.Transaction) error {
	if transaction == nil {
		return fmt.Errorf("transaction cannot be nil")
	}
	err := r.db.WithContext(ctx).Save(transaction).Error
	if err != nil {
		return fmt.Errorf("failed to update transaction: %w", err)
	}
	return nil
}

// Delete removes a transaction from the database (soft delete)
// AC3: Descriptive error message
// P-004: Check RowsAffected to detect non-existent records
func (r *transactionRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&models.Transaction{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete transaction: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// List retrieves transactions with optional filtering and pagination
// AC4: Complex query support with filtering by branch, cashier, date range, payment method, status
// P-001, P-003, P-005, P-006, P-007: Security and validation fixes
func (r *transactionRepository) List(ctx context.Context, filter *TransactionFilter) ([]*models.Transaction, int64, error) {
	// P-005: Check context cancellation
	select {
	case <-ctx.Done():
		return nil, 0, fmt.Errorf("context cancelled: %w", ctx.Err())
	default:
	}

	// P-007: Handle nil filter
	if filter == nil {
		filter = &TransactionFilter{}
	}

	var transactions []*models.Transaction
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Transaction{})

	// Apply filters
	if filter.BranchID != nil {
		query = query.Where("branch_id = ?", *filter.BranchID)
	}
	if filter.CashierID != nil {
		query = query.Where("cashier_id = ?", *filter.CashierID)
	}
	if filter.StartDate != nil {
		query = query.Where("created_at >= ?", *filter.StartDate)
	}
	if filter.EndDate != nil {
		query = query.Where("created_at <= ?", *filter.EndDate)
	}
	if filter.PaymentMethod != "" {
		query = query.Where("payment_method = ?", filter.PaymentMethod)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.TransactionNumber != "" {
		// Sanitize search query
		search := strings.ReplaceAll(filter.TransactionNumber, "%", "")
		search = strings.ReplaceAll(search, "_", "")
		if search != "" {
			query = query.Where("transaction_number LIKE ?", "%"+search+"%")
		}
	}
	if filter.CustomerName != "" {
		// Sanitize search query
		search := strings.ReplaceAll(filter.CustomerName, "%", "")
		search = strings.ReplaceAll(search, "_", "")
		if search != "" {
			query = query.Where("customer_name LIKE ?", "%"+search+"%")
		}
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count transactions: %w", err)
	}

	// P-003, P-006: Apply pagination with bounds checking
	page := filter.Page
	limit := filter.Limit
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20 // default
	}
	if limit > 1000 {
		limit = 1000 // maximum to prevent DoS
	}
	// P-006: Check for integer overflow in offset calculation
	if page > 1000000 {
		return nil, 0, fmt.Errorf("page number exceeds maximum allowed")
	}
	offset := (page - 1) * limit
	query = query.Offset(offset).Limit(limit)

	// P-001: Whitelist validation for sort fields
	allowedSortFields := map[string]bool{
		"id": true, "transaction_number": true, "created_at": true, "updated_at": true,
		"total": true, "status": true, "payment_method": true,
	}
	sortBy := "created_at"
	if filter.SortBy != "" {
		if allowedSortFields[filter.SortBy] {
			sortBy = filter.SortBy
		}
	}

	allowedSortOrders := map[string]bool{
		"ASC": true, "DESC": true, "asc": true, "desc": true,
	}
	sortOrder := "DESC"
	if filter.SortOrder != "" {
		normalized := strings.ToUpper(filter.SortOrder)
		if allowedSortOrders[normalized] {
			sortOrder = normalized
		}
	}
	query = query.Order(sortBy + " " + sortOrder)

	// Execute query with preload
	if err := query.Preload("TransactionItems").Find(&transactions).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list transactions: %w", err)
	}

	return transactions, total, nil
}

// GetDailySummary retrieves daily transaction summary for a branch
// P-009: Use UTC consistently for date boundaries
func (r *transactionRepository) GetDailySummary(ctx context.Context, branchID uint, date time.Time) (*TransactionSummary, error) {
	var summary TransactionSummary

	// P-009: Use UTC for consistent date boundaries
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	endOfDay := startOfDay.Add(24 * time.Hour)

	type summaryResult struct {
		Count          int64   `json:"count"`
		TotalAmount    float64 `json:"total_amount"`
		SubtotalAmount float64 `json:"subtotal_amount"`
		TaxAmount      float64 `json:"tax_amount"`
		DiscountAmount float64 `json:"discount_amount"`
	}

	var result summaryResult
	err := r.db.WithContext(ctx).Model(&models.Transaction{}).
		Where("branch_id = ? AND created_at >= ? AND created_at < ? AND status = ?",
			branchID, startOfDay, endOfDay, models.StatusCompleted).
		Select("COUNT(*) as count, COALESCE(SUM(total), 0) as total_amount, "+
			"COALESCE(SUM(subtotal), 0) as subtotal_amount, "+
			"COALESCE(SUM(tax), 0) as tax_amount, "+
			"COALESCE(SUM(discount), 0) as discount_amount").
		Scan(&result).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get daily summary: %w", err)
	}

	summary.TotalTransactions = result.Count
	summary.TotalAmount = fmt.Sprintf("%.2f", result.TotalAmount)
	summary.SubtotalAmount = fmt.Sprintf("%.2f", result.SubtotalAmount)
	summary.TaxAmount = fmt.Sprintf("%.2f", result.TaxAmount)
	summary.DiscountAmount = fmt.Sprintf("%.2f", result.DiscountAmount)

	return &summary, nil
}

// GetMonthlySummary retrieves monthly transaction summary for a branch
// P-009: Use UTC consistently for date boundaries
func (r *transactionRepository) GetMonthlySummary(ctx context.Context, branchID uint, year int, month time.Month) (*TransactionSummary, error) {
	var summary TransactionSummary

	// P-009: Use UTC for consistent date boundaries
	startOfMonth := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	endOfMonth := startOfMonth.AddDate(0, 1, 0).Add(-time.Second)

	type summaryResult struct {
		Count          int64   `json:"count"`
		TotalAmount    float64 `json:"total_amount"`
		SubtotalAmount float64 `json:"subtotal_amount"`
		TaxAmount      float64 `json:"tax_amount"`
		DiscountAmount float64 `json:"discount_amount"`
	}

	var result summaryResult
	err := r.db.WithContext(ctx).Model(&models.Transaction{}).
		Where("branch_id = ? AND created_at >= ? AND created_at <= ? AND status = ?",
			branchID, startOfMonth, endOfMonth, models.StatusCompleted).
		Select("COUNT(*) as count, COALESCE(SUM(total), 0) as total_amount, "+
			"COALESCE(SUM(subtotal), 0) as subtotal_amount, "+
			"COALESCE(SUM(tax), 0) as tax_amount, "+
			"COALESCE(SUM(discount), 0) as discount_amount").
		Scan(&result).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get monthly summary: %w", err)
	}

	summary.TotalTransactions = result.Count
	summary.TotalAmount = fmt.Sprintf("%.2f", result.TotalAmount)
	summary.SubtotalAmount = fmt.Sprintf("%.2f", result.SubtotalAmount)
	summary.TaxAmount = fmt.Sprintf("%.2f", result.TaxAmount)
	summary.DiscountAmount = fmt.Sprintf("%.2f", result.DiscountAmount)

	return &summary, nil
}

// CreateWithItems creates a transaction with its items in a single transaction
// P-010: Validate items slice is not empty
// P-014: Wrap errors with descriptive messages
func (r *transactionRepository) CreateWithItems(ctx context.Context, transaction *models.Transaction, items []*models.TransactionItem) error {
	// P-010: Validate items slice
	if len(items) == 0 {
		return fmt.Errorf("transaction must have at least one item")
	}
	if transaction == nil {
		return fmt.Errorf("transaction cannot be nil")
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Create transaction
		// P-014: Wrap error with descriptive message
		if err := tx.Create(transaction).Error; err != nil {
			return fmt.Errorf("failed to create transaction: %w", err)
		}

		// Create items with transaction ID
		for _, item := range items {
			item.TransactionID = transaction.ID
			// P-014: Wrap error with descriptive message
			if err := tx.Create(item).Error; err != nil {
				return fmt.Errorf("failed to create transaction item: %w", err)
			}
		}

		return nil
	})
}
