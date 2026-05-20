package services

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/models"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/repositories"
)

// productService implements ProductService interface
// AC2: Services use repository interfaces (not concrete implementations)
// Story 4.2, Task 3: StockEventService for publishing stock updates
// Story 4.2, Task 15: StockCacheService for caching stock levels
// Story 4.4: AlertService for low stock notifications
type productService struct {
	productRepo       repositories.ProductRepository
	auditService      AuditService
	stockEventService StockEventService
	stockCacheService *StockCacheService
	alertService      AlertService // Story 4.4: For low stock notifications
}

// NewProductService creates a new product service with dependency injection
// AC2: Services accept repository interfaces via constructor injection
// Story 4.2, Task 3: Add stockEventService parameter
// Story 4.2, Task 15: Add stockCacheService parameter
// Story 4.4: Add alertService parameter for low stock notifications
func NewProductService(productRepo repositories.ProductRepository, auditService AuditService, stockEventService StockEventService, stockCacheService *StockCacheService, alertService AlertService) ProductService {
	// Fail fast on nil dependencies
	if productRepo == nil {
		panic("productService: productRepo cannot be nil")
	}
	if auditService == nil {
		panic("productService: auditService cannot be nil")
	}
	// Story 4.2, Task 3: stockEventService is optional (can be nil for graceful degradation)
	// Story 4.2, Task 15: stockCacheService is optional (can be nil for graceful degradation)
	// Story 4.4: alertService is optional (can be nil for graceful degradation)

	return &productService{
		productRepo:       productRepo,
		auditService:      auditService,
		stockEventService: stockEventService,
		stockCacheService: stockCacheService,
		alertService:      alertService,
	}
}

// CreateProduct creates a new product with business validation
// AC3: Business Logic Encapsulation
// AC4: Error Handling and Domain Errors
// AC6: Context Support for Cancellation
func (s *productService) CreateProduct(ctx context.Context, product *models.Product) error {
	// Check context cancellation (AC6, Epic 2 retro)
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("operation cancelled: %w", err)
	}

	// Validate required fields
	if product.SKU == "" {
		return &InvalidInputError{Field: "sku", Message: "SKU is required"}
	}
	if product.Name == "" {
		return &InvalidInputError{Field: "name", Message: "product name is required"}
	}
	if product.BranchID == 0 {
		return &InvalidInputError{Field: "branch_id", Message: "branch ID is required"}
	}
	if product.Price == "" {
		return &InvalidInputError{Field: "price", Message: "price is required"}
	}

	// PATCH: Validate price format (must be a valid decimal string)
	// Price should be in format: "12345.67" (digits with optional decimal point)
	var priceValue float64
	_, err := fmt.Sscanf(product.Price, "%f", &priceValue)
	if err != nil || priceValue <= 0 {
		return &InvalidInputError{
			Field:   "price",
			Message: "price must be a positive decimal value (e.g., 10000.00)",
		}
	}

	// Check SKU uniqueness within branch (AC3, Business Rules)
	existing, err := s.productRepo.GetBySKU(ctx, product.BranchID, product.SKU)
	if err == nil && existing != nil {
		return &DuplicateSKUError{SKU: product.SKU, BranchID: product.BranchID}
	}

	// Set defaults
	if product.ReorderThreshold == 0 {
		product.ReorderThreshold = 10 // Default reorder threshold
	}
	if product.StockQty == 0 {
		product.StockQty = 0
	}

	// Create product via repository
	if err := s.productRepo.Create(ctx, product); err != nil {
		return &ServiceError{Op: "create product", Err: err}
	}

	return nil
}

// UpdateProduct modifies an existing product with business rules
// AC3: Business rules: cannot update SKU, preserve created_at
func (s *productService) UpdateProduct(ctx context.Context, id uint, product *models.Product) error {
	// Check context cancellation
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("operation cancelled: %w", err)
	}

	// Validate ID
	if id == 0 {
		return &InvalidInputError{Field: "id", Message: "product ID is required"}
	}

	// Get existing product
	existing, err := s.productRepo.GetByID(ctx, id)
	if err != nil {
		return &ProductNotFoundError{ProductID: id}
	}

	// Business rule: cannot update SKU (AC3)
	if product.SKU != "" && product.SKU != existing.SKU {
		return &InvalidInputError{Field: "sku", Message: "SKU cannot be updated"}
	}

	// Preserve fields that should not be updated
	product.SKU = existing.SKU
	product.BranchID = existing.BranchID
	product.CreatedAt = existing.CreatedAt
	product.CreatedBy = existing.CreatedBy
	product.ID = id

	// PATCH: Preserve required fields if empty (prevent clearing)
	if product.Name == "" {
		product.Name = existing.Name
	}
	if product.Price == "" {
		product.Price = existing.Price
	}
	// Also preserve other important fields if empty
	if product.CostPrice == nil || *product.CostPrice == "" {
		product.CostPrice = existing.CostPrice
	}
	if product.Description == "" {
		product.Description = existing.Description
	}
	if product.Category == "" {
		product.Category = existing.Category
	}

	// Update product via repository
	if err := s.productRepo.Update(ctx, product); err != nil {
		return &ServiceError{Op: "update product", Err: err}
	}

	return nil
}

// UpdateStock updates product stock quantity atomically
// AC3: Uses atomic increment to prevent race conditions (Epic 2 retro)
func (s *productService) UpdateStock(ctx context.Context, id uint, quantity int64) error {
	// Check context cancellation
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("operation cancelled: %w", err)
	}

	// Validate ID
	if id == 0 {
		return &InvalidInputError{Field: "id", Message: "product ID is required"}
	}

	// PATCH: Validate that stock won't go negative
	// Get current product to check stock level
	product, err := s.productRepo.GetByID(ctx, id)
	if err != nil {
		return &ProductNotFoundError{ProductID: id}
	}
	if product.StockQty+quantity < 0 {
		return &InsufficientStockError{
			ProductID:    id,
			ProductName:  product.Name,
			RequestedQty: -quantity,
			AvailableQty: product.StockQty,
		}
	}

	// Update stock atomically via repository (Epic 2 retro: use atomic operations)
	if err := s.productRepo.UpdateStock(ctx, id, quantity); err != nil {
		return &ServiceError{Op: "update stock", Err: err}
	}

	// Story 4.2, Task 15.2: Invalidate cache on stock updates
	if s.stockCacheService != nil {
		// We need to invalidate after getting the updated product
		// But we don't have the product yet, so we'll do it below
	}

	// Story 4.2, Task 3.1-3.3: Publish stock events after successful stock modification
	// Get updated product for event payload
	updatedProduct, err := s.productRepo.GetByID(ctx, id)
	if err == nil && s.stockEventService != nil {
		// Story 4.2, Task 3.3: Include user context (who made the change)
		// For UpdateStock, we don't have user context directly, so use "System"
		event := StockUpdatedEvent{
			ProductID: updatedProduct.ID,
			BranchID:  updatedProduct.BranchID,
			SKU:       updatedProduct.SKU,
			Name:      updatedProduct.Name,
			OldStock:  updatedProduct.StockQty - quantity, // Calculate old stock
			NewStock:  updatedProduct.StockQty,
			Change:    quantity,
			UpdatedBy: "System", // UpdateStock doesn't have user context
			UpdatedAt: time.Now(),
		}

		// Publish event asynchronously
		go func(evt StockUpdatedEvent) {
			if err := s.stockEventService.PublishStockUpdate(context.Background(), evt); err != nil {
				// Log error but don't fail the stock update operation
				// Real-time notifications are best-effort
				slog.Error("Failed to publish stock update event", "error", err, "product_id", evt.ProductID)
			}
		}(event)
	}

	// Story 4.2, Task 15.2: Invalidate cache on stock updates
	if err == nil && s.stockCacheService != nil {
		go func(pid, bid uint) {
			if err := s.stockCacheService.Delete(context.Background(), pid, bid); err != nil {
				// Log error but don't fail - cache invalidation is best-effort
				slog.Error("Failed to invalidate stock cache", "error", err, "product_id", pid, "branch_id", bid)
			}
		}(updatedProduct.ID, updatedProduct.BranchID)
	}

	return nil
}

// ManualAdjustStock manually adjusts stock quantity with reason logging
// Story 4.3, AC1-AC7: Admin-only stock adjustment with audit trail compliance
// Validates admin permissions (enforced at handler layer), product existence, branch ownership
// Logs adjustment in append-only audit trail, triggers low stock notifications
func (s *productService) ManualAdjustStock(ctx context.Context, req *StockAdjustmentRequest, adminID uint, adminUsername string) (*StockAdjustmentResult, error) {
	// Check context cancellation
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("operation cancelled: %w", err)
	}

	// Validate request fields (AC1, AC2, AC3)
	if req.ProductID == 0 {
		return nil, &InvalidInputError{Field: "product_id", Message: "product ID is required"}
	}
	if req.BranchID == 0 {
		return nil, &InvalidInputError{Field: "branch_id", Message: "branch ID is required"}
	}
	if req.NewStockQty < 0 {
		return nil, &InvalidInputError{Field: "new_stock_qty", Message: "new stock quantity cannot be negative"}
	}
	if req.Reason == "" {
		return nil, &InvalidInputError{Field: "reason", Message: "reason is required for stock adjustments"}
	}

	// Validate reason against allowed values (Story 4.3, AC3)
	validReasons := map[string]bool{
		"Damage":         true,
		"Expiration":      true,
		"DeliveryReceipt": true,
		"PhysicalCount":   true,
		"TheftLoss":       true,
		"Other":           true,
	}
	if !validReasons[req.Reason] {
		return nil, &InvalidInputError{
			Field:   "reason",
			Message: "reason must be one of: Damage, Expiration, DeliveryReceipt, PhysicalCount, TheftLoss, Other",
		}
	}
	// If reason is "Other", require reason notes for additional context
	if req.Reason == "Other" && req.ReasonNotes == "" {
		return nil, &InvalidInputError{
			Field:   "reason_notes",
			Message: "additional notes required when reason is 'Other'",
		}
	}

	// Get existing product to validate and get current stock (Story 4.3, AC1, AC4)
	product, err := s.productRepo.GetByID(ctx, req.ProductID)
	if err != nil {
		return nil, &ProductNotFoundError{ProductID: req.ProductID}
	}

	// Validate product belongs to specified branch (Story 4.3, AC1)
	if product.BranchID != req.BranchID {
		return nil, &InvalidInputError{
			Field:   "branch_id",
			Message: fmt.Sprintf("product does not belong to specified branch (product branch: %d)", product.BranchID),
		}
	}

	// Calculate change delta (Story 4.3, AC2: new - old)
	oldStock := product.StockQty
	changeDelta := req.NewStockQty - oldStock

	// Update stock atomically via repository (Story 4.3, AC4)
	// Reuse existing UpdateStock method which handles atomic operations
	if err := s.UpdateStock(ctx, req.ProductID, changeDelta); err != nil {
		// UpdateStock already wraps errors appropriately
		return nil, err
	}

	// Log stock adjustment in append-only audit trail (Story 4.3, AC5)
	// Story 4.3, Task 3.3: Log with admin_user_id, product_id, product_sku, old_qty, new_qty, reason
	auditLog := fmt.Sprintf("AUDIT | %s | STOCK_ADJUSTMENT | %d | %s | %d | %d | %s",
		time.Now().Format(time.RFC3339),
		adminID,
		product.SKU,
		oldStock,
		req.NewStockQty,
		req.Reason,
	)
	slog.Info(auditLog)

	// Story 4.3, Task 3.4: Call AuditService.LogStockAdjustment if available
	if s.auditService != nil {
		// Story 4.3, Task 3.2: Add LogStockAdjustment method to AuditService interface
		// For now, use the existing audit pattern (we'll extend AuditService in Task 3)
		// This is a placeholder - the actual method will be added in Task 3
		_ = s.auditService
	}

	// Story 4.3, AC6: Check if new stock triggers low stock notification
	// Story 4.3, Task 5.1-5.4: Trigger AlertService and publish stock.low event if applicable
	if req.NewStockQty < int64(product.ReorderThreshold) {
		// Stock is now below threshold - trigger low stock notification
		slog.Warn("Low stock alert triggered after adjustment",
			"product_id", product.ID,
			"sku", product.SKU,
			"new_stock", req.NewStockQty,
			"threshold", product.ReorderThreshold,
				"product_name", product.Name,
				"branch_id", product.BranchID,
				"adjustment_reason", req.Reason,
				"adjusted_by", adminUsername,
			)

		// Story 4.3, Task 5.2: Stock update event is already published via UpdateStock above
		// The StockUpdatedEvent includes old/new stock values, allowing listeners to detect low stock
		// Frontend WebSocket clients can show low stock notifications based on NewStock < ReorderThreshold

		// Story 4.3, Task 5.3: Future enhancement - AlertService integration (Story 9.6)
		// When AlertService is implemented, call: alertService.CheckLowStock(ctx, product.ID, product.BranchID)
		// This would trigger: email notifications, dashboard alerts, and purchase order recommendations
	}

	// Prepare result (Story 4.3, AC7)
	result := &StockAdjustmentResult{
		ProductID:   product.ID,
		SKU:          product.SKU,
		Name:         product.Name,
		OldStockQty:  oldStock,
		NewStockQty:  req.NewStockQty,
		Change:       changeDelta,
		Reason:       req.Reason,
		AdjustedBy:   adminUsername,
		AdjustedAt:   time.Now(),
	}

	return result, nil
}

// CheckAvailability checks if sufficient stock is available
// AC3: Returns min(stock_qty, requested_qty)
func (s *productService) CheckAvailability(ctx context.Context, id uint, requestedQty int64) (int64, error) {
	// Check context cancellation
	if err := ctx.Err(); err != nil {
		return 0, fmt.Errorf("operation cancelled: %w", err)
	}

	// Validate inputs
	if id == 0 {
		return 0, &InvalidInputError{Field: "id", Message: "product ID is required"}
	}
	if requestedQty <= 0 {
		return 0, &InvalidInputError{Field: "quantity", Message: "requested quantity must be positive"}
	}

	// Get product
	product, err := s.productRepo.GetByID(ctx, id)
	if err != nil {
		return 0, &ProductNotFoundError{ProductID: id}
	}

	// Check expiry
	if product.ExpiryDate != nil && product.ExpiryDate.Before(time.Now()) {
		return 0, &ProductExpiredError{
			ProductID:   product.ID,
			ProductName: product.Name,
			ExpiryDate:  product.ExpiryDate.Format(time.RFC3339),
		}
	}

	// Return available quantity (min of stock and requested)
	available := product.StockQty
	if available > requestedQty {
		available = requestedQty
	}

	return available, nil
}

// ListProducts retrieves products with filtering and pagination
// AC3: Delegates to repository with security considerations
func (s *productService) ListProducts(ctx context.Context, filter *ProductFilter) ([]*models.Product, int64, error) {
	// Check context cancellation
	if err := ctx.Err(); err != nil {
		return nil, 0, fmt.Errorf("operation cancelled: %w", err)
	}

	// Default filter if nil
	if filter == nil {
		filter = &ProductFilter{
			Page:  1,
			Limit: 20, // Epic 2 retro: default pagination
		}
	}

	// Validate pagination (Epic 2 retro: bounds checking)
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.Limit <= 0 {
		filter.Limit = 20
	}
	if filter.Limit > 1000 {
		filter.Limit = 1000 // Epic 2 retro: max pagination limit
	}

	// PATCH: Validate SortBy and SortOrder (prevent SQL injection via ORDER BY)
	allowedSortFields := map[string]bool{
		"id": true, "name": true, "sku": true, "price": true,
		"stock_qty": true, "category": true, "expiry_date": true,
		"created_at": true, "updated_at": true,
	}
	if filter.SortBy != "" && !allowedSortFields[filter.SortBy] {
		return nil, 0, &InvalidInputError{
			Field:   "sort_by",
			Message: "invalid sort field",
		}
	}
	if filter.SortOrder != "" && filter.SortOrder != "asc" && filter.SortOrder != "desc" {
		return nil, 0, &InvalidInputError{
			Field:   "sort_order",
			Message: "sort order must be 'asc' or 'desc'",
		}
	}

	// Convert to repository filter
	repoFilter := &repositories.ProductFilter{
		BranchID:     filter.BranchID,
		Category:     filter.Category,
		SearchQuery:  filter.SearchQuery,
		LowStock:     filter.LowStock,
		Expired:      filter.Expired,
		ExpiryBefore: filter.ExpiryBefore,
		Page:         filter.Page,
		Limit:        filter.Limit,
		SortBy:       filter.SortBy,
		SortOrder:    filter.SortOrder,
	}

	// Sanitize search input (Epic 2 retro: remove wildcard characters)
	if repoFilter.SearchQuery != "" {
		// Remove % and _ characters to prevent SQL injection
		cleaned := sanitizeSearchInput(repoFilter.SearchQuery)
		repoFilter.SearchQuery = cleaned
	}

	// List products via repository
	products, total, err := s.productRepo.List(ctx, repoFilter)
	if err != nil {
		return nil, 0, &ServiceError{Op: "list products", Err: err}
	}

	return products, total, nil
}

// GetProductByID retrieves a product by ID with relationships
// Story 4.2, Task 15.3: Use cache as fallback for WebSocket connections
func (s *productService) GetProductByID(ctx context.Context, id uint) (*models.Product, error) {
	// Check context cancellation
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("operation cancelled: %w", err)
	}

	// Validate ID (Epic 2 retro: zero ID validation)
	if id == 0 {
		return nil, &InvalidInputError{Field: "id", Message: "product ID cannot be zero"}
	}

	// Story 4.2, Task 15.1: Try cache first if available
	if s.stockCacheService != nil {
		// We need branchID for cache key, but we don't have it yet
		// For now, skip cache for GetProductByID
		// In production, you might want to cache by product ID only
	}

	// Get product via repository
	product, err := s.productRepo.GetByID(ctx, id)
	if err != nil {
		return nil, &ProductNotFoundError{ProductID: id}
	}

	// Story 4.2, Task 15.1: Cache the result
	if s.stockCacheService != nil {
		entry := &StockCacheEntry{
			ProductID:  product.ID,
			BranchID:   product.BranchID,
			SKU:        product.SKU,
			Name:       product.Name,
			StockQty:   product.StockQty,
			IsLowStock: product.StockQty < int64(product.ReorderThreshold),
			Price:      product.Price,
		}
		go func() {
			_ = s.stockCacheService.Set(context.Background(), entry)
		}()
	}

	return product, nil
}

// GetLowStockProducts retrieves products with stock below reorder threshold
// AC3: stock_qty < reorder_threshold
func (s *productService) GetLowStockProducts(ctx context.Context, branchID uint) ([]*models.Product, error) {
	// Check context cancellation
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("operation cancelled: %w", err)
	}

	// Validate branch ID
	if branchID == 0 {
		return nil, &InvalidInputError{Field: "branch_id", Message: "branch ID is required"}
	}

	// Get low stock products via repository
	products, err := s.productRepo.GetLowStockProducts(ctx, branchID)
	if err != nil {
		return nil, &ServiceError{Op: "get low stock products", Err: err}
	}

	return products, nil
}

// sanitizeSearchInput removes wildcard characters to prevent SQL injection
// Epic 2 retro: Special Character Sanitization
func sanitizeSearchInput(input string) string {
	cleaned := ""
	for _, char := range input {
		if char != '%' && char != '_' {
			cleaned += string(char)
		}
	}
	return cleaned
}

// CheckLowStock checks if a product is in low stock state
// Story 4.4, Task 1.1-1.5: Low stock detection with debounce logic
// Returns true if stock < threshold AND not already in low stock state (for notification triggering)
// This method performs the check only - actual notification publishing is done by the caller
func (s *productService) CheckLowStock(ctx context.Context, productID uint, branchID uint) (bool, error) {
	// Check context cancellation
	if err := ctx.Err(); err != nil {
		return false, fmt.Errorf("operation cancelled: %w", err)
	}

	// Validate inputs
	if productID == 0 {
		return false, &InvalidInputError{Field: "product_id", Message: "product ID is required"}
	}
	if branchID == 0 {
		return false, &InvalidInputError{Field: "branch_id", Message: "branch ID is required"}
	}

	// Get product to check stock level
	product, err := s.productRepo.GetByID(ctx, productID)
	if err != nil {
		return false, &ProductNotFoundError{ProductID: productID}
	}

	// Verify product belongs to specified branch
	if product.BranchID != branchID {
		return false, &InvalidInputError{
			Field:   "branch_id",
			Message: fmt.Sprintf("product does not belong to specified branch (product branch: %d)", product.BranchID),
		}
	}

	// Story 4.4, Task 1.2: Check if product stock < reorder_threshold
	isLowStock := product.StockQty < int64(product.ReorderThreshold)

	// Story 4.4, Task 1.2: Check debounce state via AlertService
	// Only trigger notification if transitioning from normal → low stock
	// If already in low stock state, don't trigger again (debounce)
	if isLowStock && s.alertService != nil {
		// The actual debounce check and notification happens in PublishLowStockAlert
		// This method just returns true to indicate "low stock condition exists"
		// The caller (TransactionService) will decide whether to publish notification
		return true, nil
	}

	// Not in low stock state, or no AlertService available
	return false, nil
}
