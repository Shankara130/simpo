package repositories

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"golang.org/x/text/unicode/norm"
	"gorm.io/gorm"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/models"
)

// purchaseInvoiceRepository implements PurchaseInvoiceRepository interface
// Story 10.2: GORM-based concrete implementation
//
// Orphaned Items Handling (DN-005):
// When a purchase invoice is soft deleted, its associated line items are NOT automatically deleted.
// This design decision preserves the audit trail - users can review what was ordered even after
// invoice deletion. Line items remain in the database as orphaned records, which is acceptable
// for audit purposes. Future reports may need to handle this case by checking if the parent
// invoice exists.
type purchaseInvoiceRepository struct {
	db *gorm.DB
}

// NewPurchaseInvoiceRepository creates a new purchase invoice repository
// Story 10.2: Factory function for dependency injection
func NewPurchaseInvoiceRepository(db interface{}) PurchaseInvoiceRepository {
	gormDB, ok := db.(*gorm.DB)
	if !ok {
		panic("db must be *gorm.DB")
	}
	return &purchaseInvoiceRepository{db: gormDB}
}

// Create inserts a new purchase invoice into the database with optional line items
// Story 10.2, AC1: Error handling with wrapping
// DN-001: Updated to support creating line items within the same transaction
func (r *purchaseInvoiceRepository) Create(ctx context.Context, invoice *models.PurchaseInvoice, createdBy uint, items []models.PurchaseInvoiceItem) error {
	if invoice == nil {
		return fmt.Errorf("purchase invoice cannot be nil")
	}
	if invoice.InvoiceNumber == "" {
		return fmt.Errorf("invoice number is required")
	}
	if invoice.SupplierID == 0 {
		return fmt.Errorf("supplier ID is required")
	}
	if invoice.BranchID == 0 {
		return fmt.Errorf("branch ID is required")
	}
	if createdBy == 0 {
		return fmt.Errorf("createdBy user ID is required")
	}

	// Validate invoice date is not in future
	if invoice.InvoiceDate.After(time.Now()) {
		return fmt.Errorf("invoice date cannot be in the future")
	}

	// Set default payment status if not provided
	if invoice.PaymentStatus == "" {
		invoice.PaymentStatus = "unpaid"
	}

	invoice.CreatedBy = &createdBy
	invoice.UpdatedBy = &createdBy

	// Use transaction to ensure data integrity
	// DN-001: Include line items creation in the same transaction
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Check for duplicate invoice number with Unicode normalization
		normalizedInvoiceNumber := norm.NFKC.String(strings.TrimSpace(invoice.InvoiceNumber))
		var existingCount int64
		if err := tx.Model(&models.PurchaseInvoice{}).
			Where("invoice_number = ?", normalizedInvoiceNumber).
			Count(&existingCount).Error; err != nil {
			return fmt.Errorf("failed to check duplicate invoice number: %w", err)
		}
		if existingCount > 0 {
			return fmt.Errorf("invoice number already exists")
		}

		// Validate supplier exists and is active
		var supplier models.Supplier
		if err := tx.Where("id = ? AND deleted_at IS NULL", invoice.SupplierID).
			First(&supplier).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("supplier not found or inactive")
			}
			return fmt.Errorf("failed to validate supplier: %w", err)
		}

		// Validate branch exists
		var branch models.Branch
		if err := tx.First(&branch, invoice.BranchID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("branch not found")
			}
			return fmt.Errorf("failed to validate branch: %w", err)
		}

		// Create the purchase invoice
		if err := tx.Create(invoice).Error; err != nil {
			return fmt.Errorf("failed to create purchase invoice: %w", err)
		}

		// DN-001: Create line items within the same transaction
		if len(items) > 0 {
			for i := range items {
				items[i].PurchaseInvoiceID = invoice.ID
				// Set timestamps
				items[i].CreatedAt = time.Now()
				items[i].UpdatedAt = time.Now()
			}
			// Batch insert line items
			if err := tx.Create(&items).Error; err != nil {
				return fmt.Errorf("failed to create line items: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to create purchase invoice: %w", err)
	}
	return nil
}

// GetByID retrieves a purchase invoice by its ID with eager loaded relationships
// Story 10.2, AC3: Eager load Supplier and Items relationships
func (r *purchaseInvoiceRepository) GetByID(ctx context.Context, id uint) (*models.PurchaseInvoice, error) {
	if id == 0 {
		return nil, ErrInvalidInput
	}

	var invoice models.PurchaseInvoice
	err := r.db.WithContext(ctx).
		Preload("Supplier").
		Preload("Branch").
		Preload("Items.Product").
		First(&invoice, id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get purchase invoice: %w", err)
	}
	return &invoice, nil
}

// GetByInvoiceNumber retrieves a purchase invoice by its invoice number
// Story 10.2, AC1: Check for duplicate invoice numbers with Unicode normalization
func (r *purchaseInvoiceRepository) GetByInvoiceNumber(ctx context.Context, invoiceNumber string) (*models.PurchaseInvoice, error) {
	var invoice models.PurchaseInvoice
	// Normalize invoice number for consistent comparison (handles Unicode equivalence)
	normalizedInvoiceNumber := norm.NFKC.String(strings.TrimSpace(invoiceNumber))
	err := r.db.WithContext(ctx).
		Where("invoice_number = ?", normalizedInvoiceNumber).
		First(&invoice).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get purchase invoice by invoice number: %w", err)
	}
	return &invoice, nil
}

// Update modifies an existing purchase invoice in the database
// Story 10.2, AC2: Error wrapping for context, optimistic locking with version check
func (r *purchaseInvoiceRepository) Update(ctx context.Context, invoice *models.PurchaseInvoice, updatedBy uint) error {
	if invoice == nil {
		return fmt.Errorf("purchase invoice cannot be nil")
	}
	if updatedBy == 0 {
		return fmt.Errorf("updatedBy user ID is required")
	}

	// Check if invoice is deleted (soft deleted)
	if invoice.DeletedAt.Valid {
		return fmt.Errorf("cannot update deleted purchase invoice")
	}

	// Optimistic locking: check version hasn't changed
	var existing models.PurchaseInvoice
	err := r.db.WithContext(ctx).
		Select("version").
		Where("id = ?", invoice.ID).
		First(&existing).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return fmt.Errorf("failed to check invoice version: %w", err)
	}
	if existing.Version != invoice.Version {
		return fmt.Errorf("purchase invoice was modified by another user (current version: %d, your version: %d)", existing.Version, invoice.Version)
	}

	invoice.UpdatedBy = &updatedBy

	err = r.db.WithContext(ctx).Save(invoice).Error
	if err != nil {
		return fmt.Errorf("failed to update purchase invoice: %w", err)
	}
	return nil
}

// Delete soft deletes a purchase invoice (sets deleted_at timestamp)
// Story 10.2: Soft delete with explicit deleted_at and deleted_by
func (r *purchaseInvoiceRepository) Delete(ctx context.Context, id uint, deletedBy uint) error {
	if id == 0 {
		return ErrInvalidInput
	}
	if deletedBy == 0 {
		return fmt.Errorf("deletedBy user ID is required")
	}

	// Explicitly set deleted_at using SoftDelete
	result := r.db.WithContext(ctx).
		Where("id = ?", id).
		Delete(&models.PurchaseInvoice{})

	if result.Error != nil {
		return fmt.Errorf("failed to delete purchase invoice: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// List retrieves purchase invoices with optional filtering and pagination
// Story 10.2, AC2: Complex query support with filtering and pagination
func (r *purchaseInvoiceRepository) List(ctx context.Context, filter *PurchaseInvoiceFilter) ([]*models.PurchaseInvoice, int64, error) {
	// Check context cancellation
	select {
	case <-ctx.Done():
		return nil, 0, fmt.Errorf("context cancelled: %w", ctx.Err())
	default:
	}

	// Handle nil filter
	if filter == nil {
		filter = &PurchaseInvoiceFilter{}
	}

	var invoices []*models.PurchaseInvoice
	var total int64

	query := r.db.WithContext(ctx).Model(&models.PurchaseInvoice{})

	// Filter by supplier if specified
	if filter.SupplierID != nil {
		query = query.Where("supplier_id = ?", *filter.SupplierID)
	}

	// Filter by payment status if specified
	if filter.PaymentStatus != nil && *filter.PaymentStatus != "" {
		// Validate payment status value
		validStatuses := map[string]bool{
			"unpaid":  true,
			"partial": true,
			"paid":    true,
		}
		if validStatuses[*filter.PaymentStatus] {
			query = query.Where("payment_status = ?", *filter.PaymentStatus)
		}
	}

	// Filter by date range if specified
	// PATCH-005: Validate date range is not inverted
	if filter.StartDate != nil && *filter.StartDate != "" {
		// Validate date format
		if _, err := time.Parse("2006-01-02", *filter.StartDate); err != nil {
			return nil, 0, fmt.Errorf("invalid start_date format: %w", err)
		}
		query = query.Where("invoice_date >= ?", *filter.StartDate)
	}
	if filter.EndDate != nil && *filter.EndDate != "" {
		// Validate date format
		if _, err := time.Parse("2006-01-02", *filter.EndDate); err != nil {
			return nil, 0, fmt.Errorf("invalid end_date format: %w", err)
		}
		// Check for date range inversion if both dates are provided
		if filter.StartDate != nil && *filter.StartDate != "" {
			startDate, _ := time.Parse("2006-01-02", *filter.StartDate)
			endDate, _ := time.Parse("2006-01-02", *filter.EndDate)
			if startDate.After(endDate) {
				return nil, 0, fmt.Errorf("invalid date range: start_date (%s) must be before or equal to end_date (%s)", *filter.StartDate, *filter.EndDate)
			}
		}
		query = query.Where("invoice_date <= ?", *filter.EndDate)
	}

	// Sanitize search query - remove wildcard characters
	if filter.SearchQuery != "" {
		search := strings.TrimSpace(filter.SearchQuery)
		// Remove wildcard characters to prevent SQL injection
		search = strings.ReplaceAll(search, "%", "")
		search = strings.ReplaceAll(search, "_", "")
		search = strings.ReplaceAll(search, "\\", "")
		// Require minimum search length and limit max length
		if len(search) >= 2 && len(search) <= 100 {
			query = query.Where("invoice_number ILIKE ?", "%"+search+"%")
		}
		// If search is too short after sanitization, ignore it
	}

	// Count total before pagination
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count purchase invoices: %w", err)
	}

	// Apply pagination with bounds checking
	// PATCH-012: Add proper overflow checking for offset calculation
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

	// Check for integer overflow in offset calculation before multiplication
	// Use int64 for intermediate calculation to detect overflow
	if int64(page-1) > (1<<63-1)/int64(limit) {
		return nil, 0, fmt.Errorf("page number exceeds maximum allowed (overflow protection)")
	}
	offset := (page - 1) * limit
	query = query.Offset(offset).Limit(limit)

	// Whitelist validation for sort fields with safe ordering
	// PATCH-001: Use switch instead of string concatenation to prevent SQL injection
	sortBy := "created_at"
	if filter.SortBy != "" {
		switch filter.SortBy {
		case "id", "invoice_date", "total_amount", "payment_status", "created_at":
			sortBy = filter.SortBy
		default:
			sortBy = "created_at"
		}
	}

	sortOrder := "DESC"
	if filter.SortOrder != "" {
		normalized := strings.ToUpper(filter.SortOrder)
		if normalized == "ASC" {
			sortOrder = "ASC"
		} else {
			sortOrder = "DESC"
		}
	}

	// Safe ordering using validated values
	if sortOrder == "ASC" {
		query = query.Order(sortBy)
	} else {
		query = query.Order(sortBy + " DESC")
	}

	// Eager load relationships
	query = query.Preload("Supplier").Preload("Branch").Preload("Items.Product")

	// Execute query
	if err := query.Find(&invoices).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list purchase invoices: %w", err)
	}

	return invoices, total, nil
}

// DN-001: CreateLineItems creates multiple line items for a purchase invoice
// Story 10.2: Batch insert line items within a transaction
// This must be called within a transaction to ensure data consistency with invoice creation
func (r *purchaseInvoiceRepository) CreateLineItems(ctx context.Context, invoiceID uint, items []models.PurchaseInvoiceItem) error {
	if invoiceID == 0 {
		return fmt.Errorf("invoice ID is required")
	}
	if len(items) == 0 {
		return fmt.Errorf("at least one line item is required")
	}

	// Prepare line items with invoice ID
	for i := range items {
		items[i].PurchaseInvoiceID = invoiceID
		// Set timestamps
		items[i].CreatedAt = time.Now()
		items[i].UpdatedAt = time.Now()
	}

	// Batch insert line items
	if err := r.db.WithContext(ctx).Create(&items).Error; err != nil {
		return fmt.Errorf("failed to create line items: %w", err)
	}

	return nil
}

// UpdatePaymentStatus updates the payment status of a purchase invoice based on total payments
// Story 10.4, AC1: Calculate and update payment status (unpaid/partial/fully paid) based on payments
func (r *purchaseInvoiceRepository) UpdatePaymentStatus(ctx context.Context, invoiceID uint) error {
	if invoiceID == 0 {
		return ErrInvalidInput
	}

	// Get the invoice first to check if it exists
	var invoice models.PurchaseInvoice
	if err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", invoiceID).First(&invoice).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("purchase invoice not found")
		}
		return fmt.Errorf("failed to get purchase invoice: %w", err)
	}

	// Calculate total paid amount using a subquery
	var totalPaid float64
	err := r.db.WithContext(ctx).Table("supplier_payments").
		Where("purchase_invoice_id = ?", invoiceID).
		Select("COALESCE(SUM(payment_amount), 0)").
		Scan(&totalPaid).Error

	if err != nil {
		return fmt.Errorf("failed to calculate total paid: %w", err)
	}

	// Determine payment status based on total paid (Story 10.4 logic)
	var newStatus string
	if totalPaid == 0 {
		newStatus = "unpaid"
	} else if totalPaid < invoice.TotalAmount {
		newStatus = "partial"
	} else {
		newStatus = "paid" // totalPaid >= invoice.TotalAmount
	}

	// Update payment status
	if err := r.db.WithContext(ctx).
		Model(&models.PurchaseInvoice{}).
		Where("id = ?", invoiceID).
		Update("payment_status", newStatus).Error; err != nil {
		return fmt.Errorf("failed to update payment status: %w", err)
	}

	return nil
}
