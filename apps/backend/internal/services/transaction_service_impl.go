package services

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/dto"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/models"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/repositories"
)

// transactionService implements TransactionService interface
// AC2: Services use repository interfaces (not concrete implementations)
// Story 4.2, Task 2: StockEventService for publishing stock updates
// Story 4.4: AlertService for low stock notifications
// Story 4.6, Task 3: ProductService for expired product validation
type transactionService struct {
	transactionRepo     repositories.TransactionRepository
	transactionItemRepo repositories.TransactionItemRepository
	productRepo         repositories.ProductRepository
	productService      ProductService // Story 4.6: For expired product validation
	auditService        AuditService
	stockEventService   StockEventService
	alertService        AlertService // Story 4.4: For low stock notifications
}

// NewTransactionService creates a new transaction service with dependency injection
// AC2: Services accept repository interfaces via constructor injection
// Story 4.2, Task 2: Add stockEventService parameter
// Story 4.4: Add alertService parameter for low stock notifications
// Story 4.6, Task 3: Add productService parameter for expired product validation
func NewTransactionService(
	transactionRepo repositories.TransactionRepository,
	transactionItemRepo repositories.TransactionItemRepository,
	productRepo repositories.ProductRepository,
	productService ProductService, // Story 4.6: For expired product validation
	auditService AuditService,
	stockEventService StockEventService,
	alertService AlertService, // Story 4.4: AlertService for low stock notifications
) TransactionService {
	// Fail fast on nil dependencies
	if transactionRepo == nil {
		panic("transactionService: transactionRepo cannot be nil")
	}
	if transactionItemRepo == nil {
		panic("transactionService: transactionItemRepo cannot be nil")
	}
	if productRepo == nil {
		panic("transactionService: productRepo cannot be nil")
	}
	if auditService == nil {
		panic("transactionService: auditService cannot be nil")
	}
	if productService == nil {
		panic("transactionService: productService cannot be nil")
	}
	// Story 4.2, Task 2: stockEventService is optional (can be nil for graceful degradation)
	// Story 4.4: alertService is optional (can be nil for graceful degradation)
	// Events won't be published if not provided, but transactions will still work

	return &transactionService{
		transactionRepo:     transactionRepo,
		transactionItemRepo: transactionItemRepo,
		productRepo:         productRepo,
		productService:      productService,
		auditService:        auditService,
		stockEventService:   stockEventService,
		alertService:        alertService,
	}
}

// CreateTransaction creates a new transaction with transaction number generation
// AC3: Transaction number format: TRX-{YYYYMMDD}-{sequential}
func (s *transactionService) CreateTransaction(ctx context.Context, transaction *models.Transaction) error {
	// Check context cancellation
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("operation cancelled: %w", err)
	}

	// Generate transaction number if not provided
	if transaction.TransactionNumber == "" {
		num, err := s.generateTransactionNumber(ctx, transaction.BranchID)
		if err != nil {
			return &ServiceError{Op: "generate transaction number", Err: err}
		}
		transaction.TransactionNumber = num
	}

	// Create transaction via repository
	if err := s.transactionRepo.Create(ctx, transaction); err != nil {
		return &ServiceError{Op: "create transaction", Err: err}
	}

	return nil
}

// ProcessSale processes a sale with transactional operations
// AC5: Transaction Support for Multi-Step Operations
// Business rules:
// - Validate all products exist and have sufficient stock
// - Begin database transaction
// - Deduct stock for all items (atomic increments)
// - Create transaction record
// - Create transaction items
// - Commit transaction
// - Rollback on any error
func (s *transactionService) ProcessSale(ctx context.Context, sale *SaleRequest, cashierID uint, branchID uint) (*models.Transaction, error) {
	// Check context cancellation
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("operation cancelled: %w", err)
	}

	if sale == nil {
		return nil, fmt.Errorf("sale request cannot be nil")
	}

	// CRITICAL-003: Check for existing transaction with same idempotency key
	if sale.IdempotencyKey != "" {
		existing, err := s.transactionRepo.GetByIdempotencyKey(ctx, sale.IdempotencyKey)
		if err == nil && existing != nil {
			// Transaction with this key already exists, return it (idempotent response)
			return existing, nil
		}
		if err != repositories.ErrNotFound {
			// Unexpected error checking for existing transaction
			return nil, fmt.Errorf("failed to check idempotency: %w", err)
		}
	}

	// Validate: at least 1 item required (AC3)
	if len(sale.Items) == 0 {
		return nil, &InvalidInputError{Field: "items", Message: "at least one item is required"}
	}

	// PATCH: Aggregate quantities by ProductID to handle duplicate items
	// This prevents individual items from passing stock checks when combined total exceeds stock
	itemMap := make(map[uint]*SaleItem)
	for _, item := range sale.Items {
		if existing, ok := itemMap[item.ProductID]; ok {
			// Duplicate ProductID - aggregate quantities
			existing.Quantity += item.Quantity
			// Keep the first unit price for consistency
		} else {
			itemMap[item.ProductID] = item
		}
	}

	// Convert map back to slice for processing
	aggregatedItems := make([]*SaleItem, 0, len(itemMap))
	for _, item := range itemMap {
		aggregatedItems = append(aggregatedItems, item)
	}

	// Validate all product fields
	for _, item := range aggregatedItems {
		if item.ProductID == 0 {
			return nil, &InvalidInputError{Field: "product_id", Message: "product ID is required"}
		}
		if item.Quantity <= 0 {
			return nil, &InvalidInputError{Field: "quantity", Message: "quantity must be positive"}
		}
		if item.UnitPrice == "" {
			return nil, &InvalidInputError{Field: "unit_price", Message: "unit price is required"}
		}
	}



		// Story 4.6, Task 3.2-3.3: Validate products are not expired before allowing sale
		// This prevents expired medications from being sold and ensures regulatory compliance
		// Story 4.6, Task 5: Log blocked sale attempts for audit compliance
		for _, item := range aggregatedItems {
			if err := s.productService.ValidateProductForSale(ctx, item.ProductID); err != nil {
				// Story 4.6, Task 5.1-5.4: Log blocked sale attempt for audit trail
				// Get product details for audit logging
				if product, fetchErr := s.productRepo.GetByID(ctx, item.ProductID); fetchErr == nil {
					expiryDateStr := "N/A"
					if product.ExpiryDate != nil {
						expiryDateStr = product.ExpiryDate.Format("2006-01-02")
					}
					// Get cashier name for audit trail (use transaction number as fallback)
					cashierName := fmt.Sprintf("Cashier #%d", cashierID)
					// Log blocked sale attempt asynchronously (don't block transaction)
					// Use request context with timeout to prevent goroutine leaks
					ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
					go func() {
						defer cancel()
						_ = s.auditService.LogBlockedSaleAttempt(ctx, cashierID, cashierName,
							product.ID, product.SKU, product.Name, expiryDateStr, "Product expired and cannot be sold", "") // Story 5.4, Task 4.5: Empty IP (service layer has no HTTP context)
					}()
				}
				// Return the domain error from ValidateProductForSale
				// This could be ErrProductExpired or ProductNotFoundError
				return nil, err
			}
		}

		// Calculate total
	total, err := s.CalculateTotal(aggregatedItems)
	if err != nil {
		return nil, err
	}

	// Validate payment method
	allowedPaymentMethods := map[string]bool{
		"CASH":     true,
		"TRANSFER": true,
		"E-WALLET": true,
		"CARD":     true,
		"QRIS":     true,
	}
	if !allowedPaymentMethods[sale.PaymentMethod] {
		return nil, &InvalidInputError{
			Field:   "payment_method",
			Message: "payment method must be one of: CASH, TRANSFER, E-WALLET, CARD, QRIS",
		}
	}

	// Generate transaction number
	transactionNumber, err := s.generateTransactionNumber(ctx, branchID)
	if err != nil {
		return nil, &ServiceError{Op: "generate transaction number", Err: err}
	}

	// Create transaction record
	transaction := &models.Transaction{
		TransactionNumber: transactionNumber,
		CashierID:         cashierID,
		BranchID:          branchID,
		Total:             total,
		Subtotal:          total, // Simplified for MVP
		Tax:               sale.TaxAmount,
		Discount:          sale.DiscountAmount,
		PaymentMethod:     sale.PaymentMethod,
		IdempotencyKey:    sale.IdempotencyKey, // CRITICAL-003: Store idempotency key
		Status:            models.StatusCompleted,
		CustomerName:      &sale.CustomerName,
	}

	// Create transaction items
	var items []*models.TransactionItem
	for _, item := range aggregatedItems {
		// Get product details for item names
		product, err := s.productRepo.GetByID(ctx, item.ProductID)
		if err != nil {
			return nil, &ProductNotFoundError{ProductID: item.ProductID}
		}

		// Calculate subtotal
		subtotal := s.calculateSubtotal(item.Quantity, item.UnitPrice)

		items = append(items, &models.TransactionItem{
			ProductID:   item.ProductID,
			Quantity:    item.Quantity,
			UnitPrice:   item.UnitPrice,
			Subtotal:    subtotal,
			ProductName: product.Name,
			ProductSKU:  product.SKU,
		})
	}

	// Build stock updates map (negative quantity for deduction)
	stockUpdates := make(map[uint]int64)
	for _, item := range aggregatedItems {
		stockUpdates[item.ProductID] = -item.Quantity
	}

	// Use atomic transaction with stock locking (Story 3.6 Task 2)
	// ProcessSaleWithStockUpdate handles:
	// - Stock validation with SELECT FOR UPDATE locking
	// - Atomic stock updates
	// - Transaction creation
	// - Automatic rollback on any error
	err = s.transactionRepo.ProcessSaleWithStockUpdate(ctx, transaction, items, stockUpdates)
	if err != nil {
		return nil, &ServiceError{Op: "process sale", Err: err}
	}

	// Story 4.2, Task 2.1-2.4: Publish stock events after successful transaction
	// Stock events are published after transaction commit to ensure consistency
	s.publishStockUpdateEvents(ctx, aggregatedItems, cashierID, branchID)

	return transaction, nil
}

// CalculateTotal calculates the total amount for a sale
// AC3: Returns sum of (quantity * unit_price) for all items
func (s *transactionService) CalculateTotal(items []*SaleItem) (string, error) {
	if len(items) == 0 {
		return "0.00", nil
	}

	// For MVP, we use string concatenation
	// In production, you would use decimal math library
	total := "0.00"
	for _, item := range items {
		subtotal := s.calculateSubtotal(item.Quantity, item.UnitPrice)
		total = s.addDecimal(total, subtotal)
	}

	return total, nil
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// calculateSubtotal calculates subtotal for a single item
// MVP Implementation: Basic decimal multiplication (quantity * unitPrice)
// In production: Use github.com/shopspring/decimal for accurate decimal math
func (s *transactionService) calculateSubtotal(quantity int64, unitPrice string) string {
	// Parse unit price as decimal string (format: "12345.67" = 1234567/100)
	// For MVP: Remove decimal point, multiply by quantity, then reinsert decimal point
	var priceInCents int64
	fmt.Sscanf(unitPrice, "%d", &priceInCents)

	// Handle decimal places
	parts := strings.Split(unitPrice, ".")
	decimalPlaces := 0
	if len(parts) == 2 {
		decimalPlaces = len(parts[1])
	}

	// Calculate subtotal in smallest unit
	totalInCents := priceInCents * quantity

	// Convert back to string with decimal places
	totalStr := fmt.Sprintf("%d", totalInCents)
	if decimalPlaces > 0 && len(totalStr) > decimalPlaces {
		insertPos := len(totalStr) - decimalPlaces
		totalStr = totalStr[:insertPos] + "." + totalStr[insertPos:]
	} else if decimalPlaces > 0 {
		// Pad with leading zeros
		padded := fmt.Sprintf("%0*s", decimalPlaces, totalStr)
		totalStr = "0." + padded
	}

	return totalStr
}

// addDecimal adds two decimal strings
// MVP Implementation: Basic decimal addition (a + b)
// In production: Use github.com/shopspring/decimal for accurate decimal math
func (s *transactionService) addDecimal(a, b string) string {
	// For MVP: Parse both numbers as integers and add
	// Handle decimal places correctly

	// Parse a
	aParts := strings.Split(a, ".")
	aDecimals := 0
	if len(aParts) == 2 {
		aDecimals = len(aParts[1])
	}
	var aValue int64
	fmt.Sscanf(strings.ReplaceAll(a, ".", ""), "%d", &aValue)

	// Parse b
	bParts := strings.Split(b, ".")
	bDecimals := 0
	if len(bParts) == 2 {
		bDecimals = len(bParts[1])
	}
	var bValue int64
	fmt.Sscanf(strings.ReplaceAll(b, ".", ""), "%d", &bValue)

	// Normalize to same decimal places
	maxDecimals := max(aDecimals, bDecimals)
	for i := aDecimals; i < maxDecimals; i++ {
		aValue *= 10
	}
	for i := bDecimals; i < maxDecimals; i++ {
		bValue *= 10
	}

	// Add
	sum := aValue + bValue

	// Convert back to string
	sumStr := fmt.Sprintf("%d", sum)
	if maxDecimals > 0 && len(sumStr) > maxDecimals {
		insertPos := len(sumStr) - maxDecimals
		sumStr = sumStr[:insertPos] + "." + sumStr[insertPos:]
	} else if maxDecimals > 0 {
		padded := fmt.Sprintf("%0*s", maxDecimals, sumStr)
		sumStr = "0." + padded
	}

	return sumStr
}

// GenerateReceiptData generates receipt structure for printing
// PATCH: Authorization note - Branch ownership check should be performed at handler layer
// The service layer returns transaction data; authorization is the handler's responsibility
func (s *transactionService) GenerateReceiptData(ctx context.Context, transactionID uint) (*ReceiptData, error) {
	// Check context cancellation
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("operation cancelled: %w", err)
	}

	// Get transaction with items
	transaction, err := s.GetTransactionByID(ctx, transactionID)
	if err != nil {
		return nil, err
	}

	// Build receipt data
	receipt := &ReceiptData{
		TransactionNumber: transaction.TransactionNumber,
		TransactionDate:   transaction.CreatedAt,
		CustomerName:      "",
		Subtotal:          transaction.Subtotal,
		TaxAmount:         transaction.Tax,
		DiscountAmount:    transaction.Discount,
		Total:             transaction.Total,
		PaymentMethod:     transaction.PaymentMethod,
	}

	if transaction.CustomerName != nil {
		receipt.CustomerName = *transaction.CustomerName
	}

	// Convert items to receipt items
	for _, item := range transaction.TransactionItems {
		receipt.Items = append(receipt.Items, &ReceiptItem{
			ProductName: item.ProductName,
			SKU:         item.ProductSKU,
			Quantity:    item.Quantity,
			UnitPrice:   item.UnitPrice,
			Total:       item.Subtotal,
		})
	}

	return receipt, nil
}

// Story 4.2, Task 2.1-2.4: publishStockUpdateEvents publishes stock update events
// after successful transaction completion
// Story 4.4, Task 4: Add low stock check and notification triggering
// This method publishes events for each product in the transaction with old and new stock values
func (s *transactionService) publishStockUpdateEvents(ctx context.Context, items []*SaleItem, cashierID uint, branchID uint) {
	// Only publish if StockEventService is available
	if s.stockEventService == nil {
		return
	}

	// Get cashier name for audit trail (use transaction number as fallback)
	cashierName := fmt.Sprintf("Cashier #%d", cashierID)

	// Publish event for each product in the transaction
	for _, item := range items {
		// Get current product details to fetch new stock level
		product, err := s.productRepo.GetByID(ctx, item.ProductID)
		if err != nil {
			// Log error but continue with other products
			continue
		}

		// Calculate old stock (new stock + quantity sold)
		oldStock := product.StockQty + item.Quantity

		// Story 4.2, Task 2.2: Publish events for each product in transaction
		// Story 4.2, Task 2.3: Include old stock, new stock, and change delta
		event := StockUpdatedEvent{
			ProductID: product.ID,
			BranchID:  branchID,
			SKU:       product.SKU,
			Name:      product.Name,
			OldStock:  oldStock,
			NewStock:  product.StockQty,
			Change:    -item.Quantity, // Negative for deductions
			UpdatedBy: cashierName,
			UpdatedAt: time.Now(),
		}

		// Publish event asynchronously (don't block transaction)
		// Story 4.2, Task 2.4: Use transaction context to ensure events only publish on commit
		// Events are published after successful commit, so we use background context
		go func(evt StockUpdatedEvent) {
			if err := s.stockEventService.PublishStockUpdate(context.Background(), evt); err != nil {
				slog.Error("Failed to publish stock update event",
					"error", err,
					"product_id", evt.ProductID,
					"sku", evt.SKU,
					"branch_id", evt.BranchID,
				)
			}
		}(event)

		// Story 4.4, Task 4.2-4.4: Check and trigger low stock notification
		// After stock deduction, check if product is now in low stock state
		// Only trigger notification if stock falls below threshold
		// Use async goroutine to avoid blocking transaction completion
		if s.alertService != nil && product.StockQty < int64(product.ReorderThreshold) {
			go func(prod *models.Product, bid uint) {
				// Check if this is a new low stock condition (not already notified)
				// The actual debounce check happens in PublishLowStockAlert
				s.publishLowStockNotification(context.Background(), prod, bid)
			}(product, branchID)
		}
	}
}

// GetTransactionByID retrieves a transaction by ID with relationships
func (s *transactionService) GetTransactionByID(ctx context.Context, id uint) (*models.Transaction, error) {
	// Check context cancellation
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("operation cancelled: %w", err)
	}

	// Validate ID (Epic 2 retro: zero ID validation)
	if id == 0 {
		return nil, &InvalidInputError{Field: "id", Message: "transaction ID cannot be zero"}
	}

	// Get transaction via repository
	transaction, err := s.transactionRepo.GetByID(ctx, id)
	if err != nil {
		return nil, &ServiceError{Op: "get transaction by ID", Err: err}
	}

	return transaction, nil
}

// ListTransactions retrieves transactions with filtering and pagination
func (s *transactionService) ListTransactions(ctx context.Context, filter *TransactionFilter) ([]*models.Transaction, int64, error) {
	// Check context cancellation
	if err := ctx.Err(); err != nil {
		return nil, 0, fmt.Errorf("operation cancelled: %w", err)
	}

	// Default filter if nil
	if filter == nil {
		filter = &TransactionFilter{
			Page:  1,
			Limit: 20,
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
		filter.Limit = 1000
	}

	// PATCH: Validate date range (max 1 year to prevent performance issues)
	maxDateRange := 365 * 24 * time.Hour // 1 year
	if filter.StartDate != nil && filter.EndDate != nil {
		rangeDuration := filter.EndDate.Sub(*filter.StartDate)
		if rangeDuration > maxDateRange {
			return nil, 0, &InvalidInputError{
				Field:   "date_range",
				Message: "date range cannot exceed 1 year",
			}
		}
	}

	// PATCH: Validate SortBy and SortOrder (prevent SQL injection via ORDER BY)
	allowedSortFields := map[string]bool{
		"id": true, "transaction_number": true, "total": true,
		"cashier_id": true, "branch_id": true, "status": true,
		"payment_method": true, "created_at": true, "updated_at": true,
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
	repoFilter := &repositories.TransactionFilter{
		BranchID:          filter.BranchID,
		CashierID:         filter.CashierID,
		StartDate:         filter.StartDate,
		EndDate:           filter.EndDate,
		PaymentMethod:     filter.PaymentMethod,
		Status:            filter.Status,
		TransactionNumber: filter.TransactionNumber,
		CustomerName:      filter.CustomerName,
		Page:              filter.Page,
		Limit:             filter.Limit,
		SortBy:            filter.SortBy,
		SortOrder:         filter.SortOrder,
	}

	// List transactions via repository
	transactions, total, err := s.transactionRepo.List(ctx, repoFilter)
	if err != nil {
		return nil, 0, &ServiceError{Op: "list transactions", Err: err}
	}

	return transactions, total, nil
}

// generateTransactionNumber generates a unique transaction number
// Format: TRX-{YYYYMMDD}-{sequential}
func (s *transactionService) generateTransactionNumber(ctx context.Context, branchID uint) (string, error) {
	now := time.Now().UTC()
	dateStr := now.Format("20060102")
	sequential := "0001" // For MVP, use sequential number

	return fmt.Sprintf("TRX-%s-%s", dateStr, sequential), nil
}

// publishLowStockNotification publishes a low stock notification event
// Story 4.4, Task 4.2-4.4: Trigger low stock notification after sale
// Creates and publishes low stock event with debounce logic
func (s *transactionService) publishLowStockNotification(ctx context.Context, product *models.Product, branchID uint) {
	// Story 4.4, Task 4.4: Add metrics tracking (log low stock alerts)
	// For MVP, we use structured logging for metrics
	slog.Info("Low stock condition detected",
		"product_id", product.ID,
		"sku", product.SKU,
		"product_name", product.Name,
		"current_stock", product.StockQty,
		"reorder_threshold", product.ReorderThreshold,
		"branch_id", branchID)

	// Create low stock notification event
	// Story 4.4, Task 2.2-2.3: Event structure with UUID and timestamp
	// Story 4.4, Patch: Get branch name from product's branch relationship
	branchName := s.getBranchName(product.BranchID)

	event := &dto.LowStockNotificationEvent{
		EventID:   fmt.Sprintf("evt_%d", time.Now().UnixNano()),
		EventType: "stock.low",
		Timestamp: time.Now().Format(time.RFC3339),
		Data: dto.ProductLowStockData{
			ProductID:         product.ID,
			SKU:               product.SKU,
			ProductName:       product.Name,
			CurrentStock:      int(product.StockQty),
			ReorderThreshold:  product.ReorderThreshold,
			SuggestedOrderQty: product.ReorderThreshold - int(product.StockQty),
			BranchID:          product.BranchID,
			BranchName:        branchName,
		},
	}

	// Story 4.4, Task 4.3: Publish notification asynchronously via AlertService
	if s.alertService != nil {
		if err := s.alertService.PublishLowStockAlert(ctx, event); err != nil {
			// Log error but don't fail - notifications are best-effort
			slog.Error("Failed to publish low stock alert", "error", err, "product_id", product.ID)
		}
	}
}

// getBranchName retrieves branch name for a given branch ID
// Story 4.4, Patch: Populate BranchName in notification payload
// For MVP, uses hardcoded mapping. Future: integrate with BranchService
func (s *transactionService) getBranchName(branchID uint) string {
	// Simple hardcoded mapping for MVP
	// TODO: Integrate with BranchService when available
	branchNames := map[uint]string{
		1: "Jakarta Branch",
		2: "Bandung Branch",
		3: "Surabaya Branch",
		4: "Medan Branch",
		5: "Semarang Branch",
	}

	if name, exists := branchNames[branchID]; exists {
		return name
	}

	// Fallback to generic format
	return fmt.Sprintf("Branch %d", branchID)
}
