package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/dto"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/repositories"
)

// alertService implements AlertService interface
// AC2: Services use repository interfaces (not concrete implementations)
// Story 4.4: Add Redis client for pub/sub notifications
type alertService struct {
	productRepo  repositories.ProductRepository
	auditService AuditService
	redisClient  *redis.Client // Story 4.4: Redis for pub/sub and debounce tracking
}

// NewAlertService creates a new alert service with dependency injection
// AC2: Services accept repository interfaces via constructor injection
// Story 4.4: Add redisClient parameter for low stock notifications
func NewAlertService(productRepo repositories.ProductRepository, auditService AuditService, redisClient *redis.Client) AlertService {
	// Fail fast on nil dependencies
	if productRepo == nil {
		panic("alertService: productRepo cannot be nil")
	}
	if auditService == nil {
		panic("alertService: auditService cannot be nil")
	}
	// Story 4.4: Redis client is optional for graceful degradation
	// If nil, notifications won't be published but system continues to work

	return &alertService{
		productRepo:  productRepo,
		auditService: auditService,
		redisClient:  redisClient,
	}
}

// CheckLowStockAlerts checks for products with stock below reorder threshold
// AC3: Business rule: stock_qty <= reorder_threshold
func (s *alertService) CheckLowStockAlerts(ctx context.Context, branchID uint) ([]*LowStockAlert, error) {
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

	// Build alerts
	alerts := make([]*LowStockAlert, 0, len(products))
	for _, product := range products {
		// Determine severity based on stock level
		severity := "LOW"
		if product.StockQty == 0 {
			severity = "HIGH"
			// PATCH: Fixed division by zero issue - check threshold before division
			} else if product.ReorderThreshold > 2 && product.StockQty < int64(product.ReorderThreshold/2) {
			severity = "MEDIUM"
		}

		alerts = append(alerts, &LowStockAlert{
			ProductID:        product.ID,
			ProductName:      product.Name,
			SKU:              product.SKU,
			CurrentStock:     product.StockQty,
			ReorderThreshold: product.ReorderThreshold,
			BranchID:         product.BranchID,
			BranchName:       "", // Would need branch repo to get name
			Severity:         severity,
		})
	}

	return alerts, nil
}

// CheckExpiryAlerts checks for products expiring soon
// AC3: Business rule: expiry within 30/14/7 days
func (s *alertService) CheckExpiryAlerts(ctx context.Context, branchID uint, daysThreshold int) ([]*ExpiryAlert, error) {
	// Check context cancellation
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("operation cancelled: %w", err)
	}

	// Validate branch ID
	if branchID == 0 {
		return nil, &InvalidInputError{Field: "branch_id", Message: "branch ID is required"}
	}

	// Validate days threshold
	if daysThreshold <= 0 {
		return nil, &InvalidInputError{Field: "days_threshold", Message: "days threshold must be positive"}
	}

	// Get expired products via repository
	products, err := s.productRepo.GetExpiredProducts(ctx, branchID)
	if err != nil {
		return nil, &ServiceError{Op: "get expired products", Err: err}
	}

	// Filter products expiring within threshold and build alerts
	alerts := make([]*ExpiryAlert, 0)
	for _, product := range products {
		if product.ExpiryDate == nil {
			continue
		}

		// PATCH: Safe days calculation with overflow protection
		hoursUntilExpiry := time.Until(*product.ExpiryDate).Hours()
		daysUntilExpiry := int(hoursUntilExpiry / 24)

		// Bound the value to prevent overflow and handle already-expired products
		if daysUntilExpiry < -36500 {
			daysUntilExpiry = -36500 // Cap at -100 years
		} else if daysUntilExpiry > 36500 {
			daysUntilExpiry = 36500 // Cap at 100 years
		}

		// Only include products expiring within threshold
		if daysUntilExpiry <= daysThreshold && daysUntilExpiry >= 0 {
			// Determine severity
			severity := "INFO"
			if daysUntilExpiry <= 7 {
				severity = "CRITICAL"
			} else if daysUntilExpiry <= 14 {
				severity = "WARNING"
			}

			alerts = append(alerts, &ExpiryAlert{
				ProductID:       product.ID,
				ProductName:     product.Name,
				SKU:             product.SKU,
				ExpiryDate:      *product.ExpiryDate,
				DaysUntilExpiry: daysUntilExpiry,
				BranchID:        product.BranchID,
				BranchName:      "", // Would need branch repo to get name
				Severity:        severity,
			})
		}
	}

	return alerts, nil
}

// SendNotification sends alert notifications
// Stub for future Redis pub/sub story
func (s *alertService) SendNotification(ctx context.Context, alert interface{}) error {
	// Check context cancellation
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("operation cancelled: %w", err)
	}

	// Stub implementation for future story
	return &InvalidInputError{
		Field:   "notification",
		Message: "alert notifications not implemented - scheduled for future Redis pub/sub story",
	}
}

// PublishLowStockAlert publishes low stock notification to Redis pub/sub
// Story 4.4, AC2, AC3, AC6: Publish stock.low event with debounce logic
// Story 4.4, Task 3.3-3.5: Implement Redis pub/sub with debounce tracking
func (s *alertService) PublishLowStockAlert(ctx context.Context, event *dto.LowStockNotificationEvent) error {
	// Check context cancellation
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("operation cancelled: %w", err)
	}

	// Validate event
	if event.Data.ProductID == 0 {
		return &InvalidInputError{Field: "product_id", Message: "product ID is required"}
	}
	if event.Data.BranchID == 0 {
		return &InvalidInputError{Field: "branch_id", Message: "branch ID is required"}
	}

	// Story 4.4, AC7: Debounce logic - check if already in low stock state
	// Story 4.4, Task 3.4: Use Redis Set to track products in low stock state
	if s.redisClient != nil {
		// Key format: low_stock:{product_id}:{branch_id}
		key := fmt.Sprintf("low_stock:%d:%d", event.Data.ProductID, event.Data.BranchID)

		// Check if already in low stock state (debounce)
		alreadyLowStock, err := s.redisClient.Exists(ctx, key).Result()
		if err != nil {
			// Log error but continue - don't fail notification on Redis error
			slog.Warn("Failed to check low stock state", "error", err, "product_id", event.Data.ProductID)
		} else if alreadyLowStock > 0 {
			// Already in low stock state - skip notification (AC7: debounce)
			slog.Debug("Low stock notification suppressed (already notified)",
				"product_id", event.Data.ProductID,
				"branch_id", event.Data.BranchID)
			return nil
		}

		// Set low stock state with 24-hour TTL (auto-expire)
		// This allows re-notification after 24 hours if stock stays low
		if err := s.redisClient.Set(ctx, key, "1", 24*time.Hour).Err(); err != nil {
			slog.Warn("Failed to set low stock state", "error", err, "product_id", event.Data.ProductID)
		}

		// Story 4.4, Task 3.3: Publish to Redis pub/sub channel: stock.low
		// Marshal event to JSON
		eventJSON, err := json.Marshal(event)
		if err != nil {
			return fmt.Errorf("failed to marshal low stock event: %w", err)
		}

		// Publish to stock.low channel
		channel := "stock.low"
		if err := s.redisClient.Publish(ctx, channel, eventJSON).Err(); err != nil {
			// Story 4.4, Error Handling: Log but don't fail operation
			// Graceful degradation - notifications are best-effort
			slog.Error("Failed to publish low stock alert", "error", err, "product_id", event.Data.ProductID)
			return nil // Don't fail the operation
		}

		// Story 4.4, Task 3.3: Log publication with structured logging
		slog.Info("Low stock alert published",
			"event_id", event.EventID,
			"product_id", event.Data.ProductID,
			"sku", event.Data.SKU,
			"current_stock", event.Data.CurrentStock,
			"threshold", event.Data.ReorderThreshold,
			"branch_id", event.Data.BranchID)
	}

	return nil
}

// ClearLowStockState clears low stock state when stock returns to normal
// Story 4.4, AC7: Debounce logic - remove tracking when stock >= threshold
// Called when stock is replenished and returns to normal levels
func (s *alertService) ClearLowStockState(ctx context.Context, productID uint, branchID uint) error {
	// Check context cancellation
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("operation cancelled: %w", err)
	}

	// Story 4.4, Task 3.5: Remove Redis key when stock returns to normal
	if s.redisClient != nil {
		key := fmt.Sprintf("low_stock:%d:%d", productID, branchID)
		if err := s.redisClient.Del(ctx, key).Err(); err != nil {
			// Log error but don't fail - state cleanup is best-effort
			slog.Warn("Failed to clear low stock state", "error", err, "product_id", productID)
		} else {
			slog.Info("Low stock state cleared",
				"product_id", productID,
				"branch_id", branchID)
		}
	}

	return nil
}

// PublishExpiryAlert publishes expiry notification to Redis pub/sub
// Story 4.5, AC4: Event published to Redis pub/sub with event type "product.expiry"
// Story 4.5, Task 3.1-3.5: Implement Redis pub/sub with debounce tracking
func (s *alertService) PublishExpiryAlert(ctx context.Context, event *dto.ExpiryAlertEvent) error {
	// Check context cancellation
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("operation cancelled: %w", err)
	}

	// Validate event
	if event.Data.ProductID == 0 {
		return &InvalidInputError{Field: "product_id", Message: "product ID is required"}
	}
	if event.Data.BranchID == 0 {
		return &InvalidInputError{Field: "branch_id", Message: "branch ID is required"}
	}

	// Story 4.5, Task 3.3: Publish to Redis pub/sub channel: product.expiry
	if s.redisClient != nil {
		// Marshal event to JSON
		eventJSON, err := json.Marshal(event)
		if err != nil {
			return fmt.Errorf("failed to marshal expiry alert event: %w", err)
		}

		// Publish to product.expiry channel
		channel := "product.expiry"
		if err := s.redisClient.Publish(ctx, channel, eventJSON).Err(); err != nil {
			// Return error so caller knows NOT to update debounce tracking
			// If publish fails, we should allow retry in next check cycle
			slog.Error("Failed to publish expiry alert", "error", err, "product_id", event.Data.ProductID)
			return fmt.Errorf("failed to publish expiry alert: %w", err)
		}

		// Story 4.5, Task 3.3: Log publication with structured logging
		slog.Info("Expiry alert published",
			"event_id", event.EventID,
			"product_id", event.Data.ProductID,
			"sku", event.Data.SKU,
			"days_remaining", event.Data.DaysRemaining,
			"alert_level", event.Data.AlertLevel,
			"branch_id", event.Data.BranchID)
	}

	return nil
}
