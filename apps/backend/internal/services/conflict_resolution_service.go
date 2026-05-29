package services

import (
	"context"
	"fmt"
	"time"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/dto"
)

// ConflictResolutionService handles conflict resolution for offline transactions
type ConflictResolutionService struct {
	productRepo ProductRepository
	auditSvc    ConflictAuditService
}

// ProductRepository defines the interface for product data access
type ProductRepository interface {
	GetStockForProduct(ctx context.Context, productID uint) (int64, error)
}

// ConflictAuditService defines the interface for conflict audit logging
type ConflictAuditService interface {
	LogConflictResolution(ctx context.Context, log ConflictResolutionLog) error
}

// NewConflictResolutionService creates a new conflict resolution service
func NewConflictResolutionService(repo ProductRepository, auditSvc ConflictAuditService) *ConflictResolutionService {
	return &ConflictResolutionService{
		productRepo: repo,
		auditSvc:    auditSvc,
	}
}

// OfflineTransaction represents a transaction from mobile sync
type OfflineTransaction struct {
	ID              uint
	TransactionNumber string
	Timestamp       time.Time
	CashierID       uint
	PaymentMethod  string
	Total           string
	Items           []TransactionItem
}

// TransactionItem represents items in a transaction
type TransactionItem struct {
	ProductID       uint
	ProductSKU      string
	ProductName     string
	Quantity        int
	UnitPrice       string
	Subtotal        string
}

// BatchProcessResult contains results from batch processing
type BatchProcessResult struct {
	SuccessfulTransactions []OfflineTransaction
	FailedTransactions     []OfflineTransaction
	ConflictErrors         []dto.ConflictErrorResponse
}

// ConflictResolutionLog represents audit log entry
type ConflictResolutionLog struct {
	ID              uint      `gorm:"primaryKey"`
	EventType       string
	TransactionID   string
	OriginalError   string
	ResolutionType  string
	ResolvedBy      string
	ResolvedAt      time.Time
	ConflictDetails string    `gorm:"type:json"`
	CreatedAt       time.Time
}

// ProcessBatchWithValidation processes multiple offline transactions with conflict resolution
func (s *ConflictResolutionService) ProcessBatchWithValidation(ctx context.Context, transactions []OfflineTransaction) (*BatchProcessResult, []dto.ConflictErrorResponse) {
	result := &BatchProcessResult{
		SuccessfulTransactions: []OfflineTransaction{},
		FailedTransactions:     []OfflineTransaction{},
		ConflictErrors:         []dto.ConflictErrorResponse{},
	}

	if len(transactions) == 0 {
		return result, nil
	}

	// Sort transactions chronologically by timestamp (oldest first)
	sortedTransactions := s.sortTransactionsChronologically(transactions)

	// Initialize batch stock counter for products in this batch
	batchStock := make(map[uint]int64)

	// Process each transaction sequentially
	for _, tx := range sortedTransactions {
		// Validate stock availability with batch context
		sufficient, details := s.ValidateStockAvailability(ctx, tx, batchStock)

		if !sufficient {
			// Build RFC 8807 error response
			errorResp := s.BuildConflictErrorResponse(*details, tx.TransactionNumber)

			// Add to failed transactions and errors
			result.FailedTransactions = append(result.FailedTransactions, tx)
			result.ConflictErrors = append(result.ConflictErrors, errorResp)

			// Log automatic failure
			s.logAutomaticFailure(ctx, tx, details)
		} else {
			// Transaction is valid - update batch stock counter
			s.updateBatchStockCounter(tx, batchStock)
			result.SuccessfulTransactions = append(result.SuccessfulTransactions, tx)
		}
	}

	return result, result.ConflictErrors
}

// ValidateStockAvailability checks if sufficient stock is available for a transaction
func (s *ConflictResolutionService) ValidateStockAvailability(ctx context.Context, tx OfflineTransaction, batchStock map[uint]int64) (bool, *dto.ConflictDetails) {
	for _, item := range tx.Items {
		// Get current stock from database
		dbStock, err := s.productRepo.GetStockForProduct(ctx, item.ProductID)
		if err != nil {
			// If we can't get stock, assume insufficient (safe default)
			return false, &dto.ConflictDetails{
				ProductID:         item.ProductID,
				ProductSKU:        item.ProductSKU,
				RequestedQuantity: item.Quantity,
				AvailableStock:    0,
				Shortfall:          item.Quantity,
			}
		}

		// Add batch adjustments (stock consumed by previously processed transactions)
		batchAdjustment := batchStock[item.ProductID]
		availableStock := dbStock - batchAdjustment

		if availableStock < int64(item.Quantity) {
			return false, &dto.ConflictDetails{
				ProductID:         item.ProductID,
				ProductSKU:        item.ProductSKU,
				RequestedQuantity: item.Quantity,
				AvailableStock:    availableStock,
				Shortfall:          item.Quantity - int(availableStock),
			}
		}
	}

	return true, nil
}

// BuildConflictErrorResponse creates an RFC 8807 formatted error response
func (s *ConflictResolutionService) BuildConflictErrorResponse(details dto.ConflictDetails, transactionID string) dto.ConflictErrorResponse {
	return dto.BuildConflictErrorResponse(details, transactionID, "/api/v1/sync")
}

// sortTransactionsChronologically sorts transactions by timestamp (oldest first)
func (s *ConflictResolutionService) sortTransactionsChronologically(transactions []OfflineTransaction) []OfflineTransaction {
	// Create a copy to avoid modifying the input
	sorted := make([]OfflineTransaction, len(transactions))
	copy(sorted, transactions)

	// Sort by timestamp (oldest first)
	for i := 0; i < len(sorted)-1; i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i].Timestamp.After(sorted[j].Timestamp) {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	return sorted
}

// updateBatchStockCounter updates the batch stock counter with consumed items
func (s *ConflictResolutionService) updateBatchStockCounter(tx OfflineTransaction, batchStock map[uint]int64) {
	for _, item := range tx.Items {
		batchStock[item.ProductID] += int64(item.Quantity)
	}
}

// logAutomaticFailure logs an automatic conflict failure
func (s *ConflictResolutionService) logAutomaticFailure(ctx context.Context, tx OfflineTransaction, details *dto.ConflictDetails) {
	// This would log to the audit service in a real implementation
	// For now, we'll create a log entry
	log := ConflictResolutionLog{
		EventType:       "conflict_resolution",
		TransactionID:   tx.TransactionNumber,
		OriginalError:   fmt.Sprintf("Insufficient stock for product %s", details.ProductSKU),
		ResolutionType:  "automatic_failure",
		ResolvedBy:      "system",
		ResolvedAt:      time.Now().UTC(),
		ConflictDetails: fmt.Sprintf(`{"product_id":%d,"product_sku":"%s","requested_qty":%d,"available_stock":%d,"shortfall":%d}`,
			details.ProductID, details.ProductSKU, details.RequestedQuantity, details.AvailableStock, details.Shortfall),
		CreatedAt:       time.Now().UTC(),
	}

	// In real implementation, we'd call auditSvc.LogConflictResolution(ctx, log)
	// For now, this is a placeholder
	_ = log
}

// ProcessTransactionWithOverride processes a transaction with override flag (allows negative stock)
func (s *ConflictResolutionService) ProcessTransactionWithOverride(ctx context.Context, tx OfflineTransaction, adminUserID uint) error {
	// Get current stock
	for _, item := range tx.Items {
		stock, err := s.productRepo.GetStockForProduct(ctx, item.ProductID)
		if err != nil {
			return fmt.Errorf("failed to get stock for product %d: %w", item.ProductID, err)
		}

		// Allow negative stock (override flag)
		newStock := stock - int64(item.Quantity)

		// In real implementation, update stock in database
		_ = newStock

		// If stock became negative, trigger alert
		if newStock < 0 {
			s.triggerCriticalStockAlert(ctx, item.ProductID, newStock, tx.TransactionNumber)
		}
	}

	// Log the manual override
	s.logManualOverride(ctx, tx, adminUserID)

	return nil
}

// triggerCriticalStockAlert publishes a Redis pub/sub alert for negative stock
func (s *ConflictResolutionService) triggerCriticalStockAlert(ctx context.Context, productID uint, newStock int64, transactionNumber string) {
	// In real implementation, this would publish to Redis pub/sub channel 'stock.critical'
	// alert := CriticalStockAlert{
	// 	ProductSKU:    productSKU,
	// 	CurrentStock:   newStock,
	// 	TransactionID: transactionNumber,
	// 	Timestamp:     time.Now().UTC(),
	// }
	// publish to 'stock.critical' channel

	// Placeholder for now
	_ = productID
	_ = newStock
	_ = transactionNumber
}

// logManualOverride logs a manual override action
func (s *ConflictResolutionService) logManualOverride(ctx context.Context, tx OfflineTransaction, adminUserID uint) {
	log := ConflictResolutionLog{
		EventType:       "conflict_resolution",
		TransactionID:   tx.TransactionNumber,
		OriginalError:   "Insufficient stock",
		ResolutionType:  "manual_override",
		ResolvedBy:      fmt.Sprintf("admin_user_%d", adminUserID),
		ResolvedAt:      time.Now().UTC(),
		ConflictDetails: fmt.Sprintf(`{"override_by":"admin_%d","transaction_id":"%s"}`, adminUserID, tx.TransactionNumber),
		CreatedAt:       time.Now().UTC(),
	}

	// In real implementation, we'd call auditSvc.LogConflictResolution(ctx, log)
	_ = log
}
