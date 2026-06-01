package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"golang.org/x/text/unicode/norm"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/models"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/repositories"
)

// purchaseInvoiceServiceImpl implements PurchaseInvoiceService interface
// Story 10.2: Service layer with business logic and audit logging
// Story 10.5: Added supplierProductCatalogService for catalog price integration
// Story 10.7: Added SupplierAuditService for supplier transaction audit trail
type purchaseInvoiceServiceImpl struct {
	purchaseInvoiceRepo        repositories.PurchaseInvoiceRepository
	supplierRepo               repositories.SupplierRepository
	productRepo                repositories.ProductRepository
	auditService               AuditService
	supplierProductCatalogService SupplierProductCatalogService
	supplierAuditService       SupplierAuditService
}

// NewPurchaseInvoiceService creates a new purchase invoice service
// Story 10.2: Factory function with dependency injection
// Story 10.5: Added supplierProductCatalogService parameter for catalog price integration
// Story 10.7: Added SupplierAuditService parameter for audit trail integration
func NewPurchaseInvoiceService(
	purchaseInvoiceRepo repositories.PurchaseInvoiceRepository,
	supplierRepo repositories.SupplierRepository,
	productRepo repositories.ProductRepository,
	auditService AuditService,
	supplierProductCatalogService SupplierProductCatalogService,
	supplierAuditService SupplierAuditService,
) PurchaseInvoiceService {
	return &purchaseInvoiceServiceImpl{
		purchaseInvoiceRepo:           purchaseInvoiceRepo,
		supplierRepo:                  supplierRepo,
		productRepo:                   productRepo,
		auditService:                   auditService,
		supplierProductCatalogService:  supplierProductCatalogService,
		supplierAuditService:           supplierAuditService,
	}
}

// CreatePurchaseInvoice creates a new purchase invoice with validation and audit logging
// Story 10.2, AC1: Validates invoice data, calculates total, logs creation
func (s *purchaseInvoiceServiceImpl) CreatePurchaseInvoice(ctx context.Context, invoice *models.PurchaseInvoice, items []CreatePurchaseInvoiceItemRequest, createdBy uint, ipAddress string) (*models.PurchaseInvoice, error) {
	// Validate inputs
	if invoice == nil {
		return nil, fmt.Errorf("purchase invoice cannot be nil")
	}
	// PATCH-015: Add reasonable limit for items array to prevent memory issues
	if len(items) == 0 {
		return nil, fmt.Errorf("at least one line item is required")
	}
	if len(items) > 500 {
		return nil, fmt.Errorf("too many line items: maximum 500 items per invoice")
	}
	if createdBy == 0 {
		return nil, fmt.Errorf("createdBy user ID is required")
	}

	// PATCH-017: Standardize Unicode normalization - apply to all string inputs
	// Normalize invoice number
	invoice.InvoiceNumber = norm.NFKC.String(strings.TrimSpace(invoice.InvoiceNumber))
	if invoice.InvoiceNumber == "" {
		return nil, fmt.Errorf("invoice number is required")
	}
	// PATCH-017: Also normalize notes to prevent encoding issues
	if invoice.Notes != "" {
		invoice.Notes = norm.NFKC.String(strings.TrimSpace(invoice.Notes))
	}
	// PATCH-017: Normalize document URL to prevent encoding issues
	if invoice.DocumentURL != "" {
		invoice.DocumentURL = norm.NFKC.String(strings.TrimSpace(invoice.DocumentURL))
	}

	// Validate invoice date with proper timezone handling
	if invoice.InvoiceDate.IsZero() {
		return nil, fmt.Errorf("invoice date is required")
	}
	// PATCH-018: Use UTC for date comparisons to avoid timezone issues
	now := time.Now().UTC()
	// Convert invoice date to UTC for comparison (assumes input is local date)
	invoiceDateUTC := time.Date(invoice.InvoiceDate.Year(), invoice.InvoiceDate.Month(), invoice.InvoiceDate.Day(), 0, 0, 0, 0, time.UTC)
	if invoiceDateUTC.After(now) {
		return nil, fmt.Errorf("invoice date cannot be in the future")
	}
	// Store as UTC for consistency
	invoice.InvoiceDate = invoiceDateUTC

	// Validate supplier exists and is active
	supplier, err := s.supplierRepo.GetByID(ctx, invoice.SupplierID)
	if err != nil {
		return nil, fmt.Errorf("supplier not found: %w", err)
	}
	if supplier.DeletedAt.Valid {
		return nil, fmt.Errorf("supplier is inactive")
	}

	// PATCH-011: Validate Document URL format if provided
	if invoice.DocumentURL != "" {
		// Check for safe URL protocols only
		lowerURL := strings.ToLower(strings.TrimSpace(invoice.DocumentURL))
		if !strings.HasPrefix(lowerURL, "http://") && !strings.HasPrefix(lowerURL, "https://") && !strings.HasPrefix(lowerURL, "/") {
			return nil, fmt.Errorf("document URL must use http, https protocol or be a relative path")
		}
		// Additional check for dangerous protocols
		if strings.HasPrefix(lowerURL, "javascript:") || strings.HasPrefix(lowerURL, "data:") || strings.HasPrefix(lowerURL, "file:") {
			return nil, fmt.Errorf("document URL protocol not allowed")
		}
	}

	// PATCH-006: Validate BranchID
	if invoice.BranchID == 0 {
		return nil, fmt.Errorf("branch ID is required")
	}

	// DN-007: Aggregate duplicate products - if the same product appears multiple times,
	// combine quantities and use the first unit cost (assumes same product at same cost)
	aggregatedItems := make([]CreatePurchaseInvoiceItemRequest, 0)
	productMap := make(map[uint]CreatePurchaseInvoiceItemRequest) // ProductID -> Item
	for _, item := range items {
		if existing, found := productMap[item.ProductID]; found {
			// Aggregate quantities
			existing.Quantity += item.Quantity
			productMap[item.ProductID] = existing
		} else {
			// First occurrence, add to map
			productMap[item.ProductID] = item
		}
	}
	// Convert map back to slice
	for _, item := range productMap {
		aggregatedItems = append(aggregatedItems, item)
	}
	// Replace items with aggregated items
	items = aggregatedItems

	// Validate line items and calculate total
	// PATCH-004: Add overflow protection and reasonable limits
	totalAmount := 0.0
	for i, item := range items {
		if item.Quantity <= 0 {
			return nil, fmt.Errorf("line item %d: quantity must be greater than 0", i+1)
		}
		// PATCH-004: Add reasonable quantity limit to prevent overflow
		if item.Quantity > 1000000 {
			return nil, fmt.Errorf("line item %d: quantity exceeds maximum allowed", i+1)
		}
		if item.UnitCost < 0 {
			return nil, fmt.Errorf("line item %d: unit cost cannot be negative", i+1)
		}
		// PATCH-004: Add reasonable unit cost limit to prevent overflow
		if item.UnitCost > 1000000000 {
			return nil, fmt.Errorf("line item %d: unit cost exceeds maximum allowed", i+1)
		}

		// Validate product exists
		product, err := s.productRepo.GetByID(ctx, item.ProductID)
		if err != nil {
			return nil, fmt.Errorf("line item %d: product not found: %w", i+1, err)
		}
		if product.DeletedAt.Valid {
			return nil, fmt.Errorf("line item %d: product is inactive", i+1)
		}

		// Calculate and validate subtotal with overflow check
		subtotal := float64(item.Quantity) * item.UnitCost
		// PATCH-004: Check for overflow
		if subtotal < 0 || subtotal > 1e15 { // Sanity check for overflow
			return nil, fmt.Errorf("line item %d: subtotal calculation overflow", i+1)
		}
		totalAmount += subtotal
		// PATCH-004: Check for total overflow
		if totalAmount < 0 || totalAmount > 1e15 {
			return nil, fmt.Errorf("line item %d: total amount overflow", i+1)
		}
	}

	// PATCH-009: Validate total amount is non-negative
	if totalAmount < 0 {
		return nil, fmt.Errorf("total amount cannot be negative: %.2f", totalAmount)
	}

	// Set total amount and default payment status
	invoice.TotalAmount = totalAmount
	invoice.PaymentStatus = "unpaid"

	// PATCH-002: Check for duplicate invoice number
	// Note: This check has a race condition gap. The unique constraint in database will catch duplicates.
	// PATCH-013: Generic error to prevent information leakage
	existing, err := s.purchaseInvoiceRepo.GetByInvoiceNumber(ctx, invoice.InvoiceNumber)
	if err == nil && existing != nil {
		return nil, &DuplicateInvoiceError{InvoiceNumber: "<REDACTED>"}
	}

	// DN-001: Convert service items to model items for repository
	modelItems := make([]models.PurchaseInvoiceItem, len(items))
	for i, item := range items {
		modelItems[i] = models.PurchaseInvoiceItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			UnitCost:  item.UnitCost,
			Subtotal:  float64(item.Quantity) * item.UnitCost,
		}
	}

	// Create invoice with line items in same transaction
	err = s.purchaseInvoiceRepo.Create(ctx, invoice, createdBy, modelItems)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			return nil, fmt.Errorf("invoice number '%s' already exists", invoice.InvoiceNumber)
		}
		return nil, fmt.Errorf("failed to create purchase invoice: %w", err)
	}

	// Log to audit trail
	entry := AuditLogEntry{
		UserID:    &createdBy,
		Username:  fmt.Sprintf("user_%d", createdBy),
		Action:    "purchase_invoice.created",
		IPAddress: ipAddress,
		Outcome:   "success",
		Reason:    fmt.Sprintf("Created purchase invoice: %s (ID: %d) for supplier %s (ID: %d), total: %.2f",
			invoice.InvoiceNumber, invoice.ID, supplier.Name, supplier.ID, totalAmount),
	}
	// PATCH-014: Add proper audit failure logging with context
	if err := s.auditService.LogLoginAttempt(ctx, entry); err != nil {
		// Log to stderr as fallback for audit failures - this ensures critical audit trail gaps are visible
		// In production, this should go to a monitoring system
		fmt.Printf("WARN: Audit logging failed for purchase_invoice.created (invoice ID: %d): %v\n", invoice.ID, err)
		// Note: Operation continues despite audit failure to maintain availability
		// Consider implementing retry logic or dead letter queue for audit events in production
	}

	// Story 10.7: Log to supplier audit trail for Badan POM compliance
	if s.supplierAuditService != nil {
		auditLog := &models.SupplierAuditTrail{
			TransactionType:     "purchase_invoice",
			EntityType:         "purchase_invoice",
			EntityID:           invoice.ID,
			UserID:             createdBy,
			UserRole:           "Admin",
			ActionType:         "CREATE",
			ActionDescription:   fmt.Sprintf("Merekam faktur pembelian: %s untuk supplier %s", invoice.InvoiceNumber, supplier.Name),
			Reason:             "",
			TransactionAmount:   &totalAmount,
			AffectedItemsCount: len(items),
			IPAddress:          ipAddress,
			BranchID:           invoice.BranchID,
		}
		if err := s.supplierAuditService.LogSupplierOperation(ctx, auditLog); err != nil {
			fmt.Printf("WARN: Supplier audit logging failed for purchase_invoice.created (invoice ID: %d): %v\n", invoice.ID, err)
		}
	}

	return invoice, nil
}

// GetPurchaseInvoiceByID retrieves a purchase invoice by ID
// Story 10.2, AC3: Returns invoice details with line items and supplier
func (s *purchaseInvoiceServiceImpl) GetPurchaseInvoiceByID(ctx context.Context, id uint) (*models.PurchaseInvoice, error) {
	if id == 0 {
		return nil, fmt.Errorf("invalid purchase invoice ID")
	}

	invoice, err := s.purchaseInvoiceRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get purchase invoice: %w", err)
	}

	return invoice, nil
}

// ListPurchaseInvoices retrieves purchase invoices with filtering and pagination
// Story 10.2, AC2: Supports filtering by supplier, date range, payment status
func (s *purchaseInvoiceServiceImpl) ListPurchaseInvoices(ctx context.Context, filter *PurchaseInvoiceListFilter) ([]*models.PurchaseInvoice, int64, error) {
	// Set defaults if filter is nil
	if filter == nil {
		filter = &PurchaseInvoiceListFilter{
			Page:  1,
			Limit: 20,
		}
	}

	// Convert to repository filter
	repoFilter := &repositories.PurchaseInvoiceFilter{
		SupplierID:    filter.SupplierID,
		StartDate:     filter.StartDate,
		EndDate:       filter.EndDate,
		PaymentStatus: filter.PaymentStatus,
		SearchQuery:   filter.SearchQuery,
		Page:          filter.Page,
		Limit:         filter.Limit,
		SortBy:        filter.SortBy,
		SortOrder:     filter.SortOrder,
	}

	invoices, total, err := s.purchaseInvoiceRepo.List(ctx, repoFilter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list purchase invoices: %w", err)
	}

	return invoices, total, nil
}

// UpdatePurchaseInvoice updates an existing purchase invoice with validation and audit logging
// Story 10.2: Validates changes, recalculates total, logs update
func (s *purchaseInvoiceServiceImpl) UpdatePurchaseInvoice(ctx context.Context, id uint, updates *UpdatePurchaseInvoiceRequest, updatedBy uint, ipAddress string) (*models.PurchaseInvoice, error) {
	// Validate inputs
	if id == 0 {
		return nil, fmt.Errorf("invalid purchase invoice ID")
	}
	if updates == nil {
		return nil, fmt.Errorf("updates cannot be nil")
	}
	if strings.TrimSpace(updates.Reason) == "" {
		return nil, fmt.Errorf("reason is required for update")
	}
	if updatedBy == 0 {
		return nil, fmt.Errorf("updatedBy user ID is required")
	}

	// PATCH-006: Validate SupplierID is not zero
	if updates.SupplierID == 0 {
		return nil, fmt.Errorf("supplier ID is required")
	}

	// PATCH-007: Validate invoice number after normalization
	updates.InvoiceNumber = norm.NFKC.String(strings.TrimSpace(updates.InvoiceNumber))
	if updates.InvoiceNumber == "" {
		return nil, fmt.Errorf("invoice number cannot be empty or whitespace only")
	}
	// PATCH-007: Additional check for minimum length after normalization
	if len(updates.InvoiceNumber) < 3 {
		return nil, fmt.Errorf("invoice number must be at least 3 characters")
	}

	// Validate invoice date
	invoiceDate, err := time.Parse(time.RFC3339, updates.InvoiceDate)
	if err != nil {
		return nil, fmt.Errorf("invalid invoice date format: %w", err)
	}
	if invoiceDate.After(time.Now()) {
		return nil, fmt.Errorf("invoice date cannot be in the future")
	}

	// Validate supplier exists and is active
	supplier, err := s.supplierRepo.GetByID(ctx, updates.SupplierID)
	if err != nil {
		return nil, fmt.Errorf("supplier not found: %w", err)
	}
	if supplier.DeletedAt.Valid {
		return nil, fmt.Errorf("supplier is inactive")
	}

	// Validate line items
	if len(updates.Items) == 0 {
		return nil, fmt.Errorf("at least one line item is required")
	}

	// DN-007: Aggregate duplicate products in updates
	aggregatedItems := make([]CreatePurchaseInvoiceItemRequest, 0)
	productMap := make(map[uint]CreatePurchaseInvoiceItemRequest) // ProductID -> Item
	for _, item := range updates.Items {
		if existing, found := productMap[item.ProductID]; found {
			// Aggregate quantities
			existing.Quantity += item.Quantity
			productMap[item.ProductID] = existing
		} else {
			// First occurrence, add to map
			productMap[item.ProductID] = item
		}
	}
	// Convert map back to slice
	for _, item := range productMap {
		aggregatedItems = append(aggregatedItems, item)
	}
	// Replace updates.Items with aggregated items
	updates.Items = aggregatedItems

	// Calculate new total
	totalAmount := 0.0
	for i, item := range updates.Items {
		if item.Quantity <= 0 {
			return nil, fmt.Errorf("line item %d: quantity must be greater than 0", i+1)
		}
		if item.UnitCost < 0 {
			return nil, fmt.Errorf("line item %d: unit cost cannot be negative", i+1)
		}

		// Validate product exists
		product, err := s.productRepo.GetByID(ctx, item.ProductID)
		if err != nil {
			return nil, fmt.Errorf("line item %d: product not found: %w", i+1, err)
		}
		if product.DeletedAt.Valid {
			return nil, fmt.Errorf("line item %d: product is inactive", i+1)
		}

		totalAmount += float64(item.Quantity) * item.UnitCost
	}

	// Get existing invoice
	existing, err := s.purchaseInvoiceRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get purchase invoice: %w", err)
	}

	// Check for duplicate invoice number if number is being changed
	if updates.InvoiceNumber != existing.InvoiceNumber {
		duplicate, err := s.purchaseInvoiceRepo.GetByInvoiceNumber(ctx, updates.InvoiceNumber)
		if err == nil && duplicate != nil && duplicate.ID != id {
			return nil, fmt.Errorf("invoice number '%s' already exists", updates.InvoiceNumber)
		}
	}

	// Apply updates
	existing.InvoiceNumber = updates.InvoiceNumber
	existing.InvoiceDate = invoiceDate
	existing.SupplierID = updates.SupplierID
	existing.TotalAmount = totalAmount
	existing.Notes = updates.Notes
	existing.DocumentURL = updates.DocumentURL

	// Update invoice
	err = s.purchaseInvoiceRepo.Update(ctx, existing, updatedBy)
	if err != nil {
		return nil, fmt.Errorf("failed to update purchase invoice: %w", err)
	}

	// Log to audit trail
	entry := AuditLogEntry{
		UserID:    &updatedBy,
		Username:  fmt.Sprintf("user_%d", updatedBy),
		Action:    "purchase_invoice.updated",
		IPAddress: ipAddress,
		Outcome:   "success",
		Reason:    fmt.Sprintf("Updated purchase invoice: %s (ID: %d). Reason: %s", existing.InvoiceNumber, id, updates.Reason),
	}
	if err := s.auditService.LogLoginAttempt(ctx, entry); err != nil {
		// Log error but don't fail the operation
	}

	// Story 10.7: Log to supplier audit trail for Badan POM compliance
	if s.supplierAuditService != nil {
		auditLog := &models.SupplierAuditTrail{
			TransactionType:     "purchase_invoice",
			EntityType:         "purchase_invoice",
			EntityID:           existing.ID,
			UserID:             updatedBy,
			UserRole:           "Admin",
			ActionType:         "UPDATE",
			ActionDescription:   fmt.Sprintf("Memperbarui faktur pembelian: %s", existing.InvoiceNumber),
			Reason:             updates.Reason,
			TransactionAmount:   &existing.TotalAmount,
			AffectedItemsCount: len(updates.Items),
			IPAddress:          ipAddress,
			BranchID:           existing.BranchID,
		}
		if err := s.supplierAuditService.LogSupplierOperation(ctx, auditLog); err != nil {
			fmt.Printf("WARN: Supplier audit logging failed for purchase_invoice.updated (invoice ID: %d): %v\n", existing.ID, err)
		}
	}

	return existing, nil
}

// DeletePurchaseInvoice deletes (soft deletes) a purchase invoice with audit logging
// Story 10.2: Soft deletes invoice and logs deletion
func (s *purchaseInvoiceServiceImpl) DeletePurchaseInvoice(ctx context.Context, id uint, deletedBy uint, ipAddress string) error {
	// Validate inputs
	if id == 0 {
		return fmt.Errorf("invalid purchase invoice ID")
	}
	if deletedBy == 0 {
		return fmt.Errorf("deletedBy user ID is required")
	}

	// Get invoice for audit logging
	invoice, err := s.purchaseInvoiceRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get purchase invoice: %w", err)
	}

	// Delete invoice
	err = s.purchaseInvoiceRepo.Delete(ctx, id, deletedBy)
	if err != nil {
		return fmt.Errorf("failed to delete purchase invoice: %w", err)
	}

	// Log to audit trail
	entry := AuditLogEntry{
		UserID:    &deletedBy,
		Username:  fmt.Sprintf("user_%d", deletedBy),
		Action:    "purchase_invoice.deleted",
		IPAddress: ipAddress,
		Outcome:   "success",
		Reason:    fmt.Sprintf("Deleted purchase invoice: %s (ID: %d)", invoice.InvoiceNumber, id),
	}
	if err := s.auditService.LogLoginAttempt(ctx, entry); err != nil {
		// Log error but don't fail the operation
	}

	// Story 10.7: Log to supplier audit trail for Badan POM compliance
	if s.supplierAuditService != nil {
		auditLog := &models.SupplierAuditTrail{
			TransactionType:   "purchase_invoice",
			EntityType:       "purchase_invoice",
			EntityID:         invoice.ID,
			UserID:           deletedBy,
			UserRole:         "Admin",
			ActionType:       "DELETE",
			ActionDescription: fmt.Sprintf("Menghapus faktur pembelian: %s", invoice.InvoiceNumber),
			Reason:           "Faktur pembelian dihapus",
			TransactionAmount: &invoice.TotalAmount,
			AffectedItemsCount: len(invoice.Items),
			IPAddress:        ipAddress,
			BranchID:         invoice.BranchID,
		}
		if err := s.supplierAuditService.LogSupplierOperation(ctx, auditLog); err != nil {
			fmt.Printf("WARN: Supplier audit logging failed for purchase_invoice.deleted (invoice ID: %d): %v\n", invoice.ID, err)
		}
	}

	return nil
}

// GetSuggestedPrice retrieves the suggested purchase price from supplier catalog
// Story 10.5, AC1: Returns catalog price for supplier-product combination, or error if not found
func (s *purchaseInvoiceServiceImpl) GetSuggestedPrice(ctx context.Context, supplierID uint, productID uint, branchID uint) (float64, error) {
	// Validate inputs
	if supplierID == 0 {
		return 0, fmt.Errorf("supplier ID is required")
	}
	if productID == 0 {
		return 0, fmt.Errorf("product ID is required")
	}
	if branchID == 0 {
		return 0, fmt.Errorf("branch ID is required")
	}

	// Check if supplier product catalog service is available
	if s.supplierProductCatalogService == nil {
		return 0, fmt.Errorf("supplier catalog service not available")
	}

	// Get current price from catalog
	filter := &SupplierProductCatalogListFilter{
		SupplierID:  &supplierID,
		ProductID:   &productID,
		BranchID:    &branchID,
		Page:        1,
		Limit:       1,
		SortBy:      "price_effective_from",
		SortOrder:   "desc",
	}

	catalogs, _, err := s.supplierProductCatalogService.ListProductCatalogs(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("catalog price not found for supplier-product combination: %w", err)
	}

	if len(catalogs) == 0 {
		return 0, fmt.Errorf("no catalog price found for this supplier-product combination")
	}

	return catalogs[0].PurchasePrice, nil
}
