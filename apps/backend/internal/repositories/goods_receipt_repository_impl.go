package repositories

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/models"
)

// goodsReceiptRepository implements GoodsReceiptRepository interface
// Story 10.3: GORM-based concrete implementation
type goodsReceiptRepository struct {
	db *gorm.DB
}

// NewGoodsReceiptRepository creates a new goods receipt repository
// Story 10.3: Factory function for dependency injection
func NewGoodsReceiptRepository(db interface{}) GoodsReceiptRepository {
	gormDB, ok := db.(*gorm.DB)
	if !ok {
		panic("db must be *gorm.DB")
	}
	return &goodsReceiptRepository{db: gormDB}
}

// Create inserts a new goods receipt into the database
// Story 10.3, AC1: Error handling with wrapping
func (r *goodsReceiptRepository) Create(ctx context.Context, receipt *models.GoodsReceipt) error {
	if receipt == nil {
		return fmt.Errorf("goods receipt cannot be nil")
	}
	if receipt.PurchaseInvoiceID == 0 {
		return fmt.Errorf("purchase invoice ID is required")
	}
	if receipt.ReceivedBy == 0 {
		return fmt.Errorf("received by user ID is required")
	}
	if receipt.BranchID == 0 {
		return fmt.Errorf("branch ID is required")
	}

	// Validate received date is not in future
	// Code review fix: LOW-002 - Use UTC for consistent date validation (PATCH-018 from Story 10.2)
	if receipt.ReceivedDate.After(time.Now().UTC()) {
		return fmt.Errorf("received date cannot be in the future")
	}

	// Set default received date to current date if not set
	if receipt.ReceivedDate.IsZero() {
		receipt.ReceivedDate = time.Now().UTC()
	}

	// Create goods receipt
	err := r.db.WithContext(ctx).Create(receipt).Error
	if err != nil {
		return fmt.Errorf("failed to create goods receipt: %w", err)
	}

	return nil
}

// GetByID retrieves a goods receipt by its ID with eager loaded relationships
// Story 10.3: Get goods receipt with purchase invoice details
func (r *goodsReceiptRepository) GetByID(ctx context.Context, id uint) (*models.GoodsReceipt, error) {
	if id == 0 {
		return nil, fmt.Errorf("ID is required")
	}

	var receipt models.GoodsReceipt
	err := r.db.WithContext(ctx).
		Preload("PurchaseInvoice").
		Preload("PurchaseInvoice.Supplier").
		Preload("PurchaseInvoice.Items").
		Preload("PurchaseInvoice.Items.Product").
		Preload("Branch").
		First(&receipt, id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("goods receipt not found")
		}
		return nil, fmt.Errorf("failed to get goods receipt: %w", err)
	}

	return &receipt, nil
}

// GetByInvoiceID retrieves a goods receipt by its purchase invoice ID
// Story 10.3: Get goods receipt for a specific invoice (one-to-one relationship)
func (r *goodsReceiptRepository) GetByInvoiceID(ctx context.Context, invoiceID uint) (*models.GoodsReceipt, error) {
	if invoiceID == 0 {
		return nil, fmt.Errorf("invoice ID is required")
	}

	var receipt models.GoodsReceipt
	err := r.db.WithContext(ctx).
		Preload("PurchaseInvoice").
		Preload("PurchaseInvoice.Supplier").
		Preload("PurchaseInvoice.Items").
		Preload("Branch").
		Where("purchase_invoice_id = ?", invoiceID).
		First(&receipt).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("goods receipt not found for invoice")
		}
		return nil, fmt.Errorf("failed to get goods receipt by invoice ID: %w", err)
	}

	return &receipt, nil
}

// List retrieves goods receipts with optional filtering and pagination
// Story 10.3: List goods receipts with filters (branch, date range, pagination)
func (r *goodsReceiptRepository) List(ctx context.Context, filter *GoodsReceiptFilter) ([]*models.GoodsReceipt, int64, error) {
	query := r.db.WithContext(ctx).Model(&models.GoodsReceipt{})

	// Apply filters
	if filter != nil {
		// Filter by branch
		if filter.BranchID != nil {
			query = query.Where("branch_id = ?", *filter.BranchID)
		}

		// Filter by received by user
		if filter.ReceivedBy != nil {
			query = query.Where("received_by = ?", *filter.ReceivedBy)
		}

		// Filter by received date range
		if filter.StartDate != nil {
			startDate, err := time.Parse("2006-01-02", *filter.StartDate)
			if err != nil {
				return nil, 0, fmt.Errorf("invalid start date format: %w", err)
			}
			query = query.Where("received_date >= ?", startDate)
		}

		if filter.EndDate != nil {
			endDate, err := time.Parse("2006-01-02", *filter.EndDate)
			if err != nil {
				return nil, 0, fmt.Errorf("invalid end date format: %w", err)
			}
			query = query.Where("received_date <= ?", endDate)
		}
	}

	// Get total count
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count goods receipts: %w", err)
	}

	// Apply sorting
	if filter != nil && filter.SortBy != "" {
		sortOrder := "desc"
		if filter.SortOrder == "asc" {
			sortOrder = "asc"
		}

		// Whitelist allowed sort columns
		switch filter.SortBy {
		case "received_date":
			if sortOrder == "asc" {
				query = query.Order("received_date asc")
			} else {
				query = query.Order("received_date desc")
			}
		case "created_at":
			if sortOrder == "asc" {
				query = query.Order("created_at asc")
			} else {
				query = query.Order("created_at desc")
			}
		default:
			// Default sort by received_date desc
			query = query.Order("received_date desc")
		}
	} else {
		// Default sort by received_date desc
		query = query.Order("received_date desc")
	}

	// Apply pagination
	if filter != nil && filter.Limit > 0 {
		page := 1
		if filter.Page > 0 {
			page = filter.Page
		}

		offset := (page - 1) * filter.Limit
		query = query.Offset(offset).Limit(filter.Limit)
	}

	// Execute query with eager loading
	var receipts []*models.GoodsReceipt
	err := query.
		Preload("PurchaseInvoice").
		Preload("PurchaseInvoice.Supplier").
		Preload("Branch").
		Find(&receipts).Error

	if err != nil {
		return nil, 0, fmt.Errorf("failed to list goods receipts: %w", err)
	}

	return receipts, total, nil
}
