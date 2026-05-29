package services

import (
	"context"
	"fmt"
	"math"
	"time"

	"log/slog"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/dto"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/models"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/repositories"
)

// ExpiryCheckService checks for products approaching expiry
// Story 4.5, AC1, AC2, AC3: Scheduled job to generate expiry alerts
type ExpiryCheckService struct {
	productRepo  repositories.ProductRepository
	alertService AlertService
	redisClient  *redis.Client
	logger       *slog.Logger
}

// NewExpiryCheckService creates a new expiry check service
// Story 4.5, Subtask 2.5: Constructor with dependencies
func NewExpiryCheckService(
	productRepo repositories.ProductRepository,
	alertService AlertService,
	redisClient *redis.Client,
	logger *slog.Logger,
) *ExpiryCheckService {
	return &ExpiryCheckService{
		productRepo:  productRepo,
		alertService: alertService,
		redisClient:  redisClient,
		logger:       logger,
	}
}

// CheckExpiringProducts checks for products expiring within 30 days and generates alerts
// Story 4.5, AC1, AC2, AC3: 30-day, 14-day, and 7-day alerts
// Story 4.5, Subtask 2.3: Query, categorize, filter, debounce
func (s *ExpiryCheckService) CheckExpiringProducts(ctx context.Context) ([]*dto.ExpiryAlertEvent, error) {
	now := time.Now().UTC()
	thirtyDaysOut := now.AddDate(0, 0, 30)

	// Get all products expiring within 30 days
	// Story 4.5, Subtask 2.3: Query products where expiry_date BETWEEN (NOW + 7 days) AND (NOW + 30 days)
	products, err := s.productRepo.GetExpiringProducts(ctx, 0, now, thirtyDaysOut)
	if err != nil {
		s.logger.Error("failed to get expiring products", "error", err)
		return nil, fmt.Errorf("failed to get expiring products: %w", err)
	}

	var events []*dto.ExpiryAlertEvent

	for _, product := range products {
		if product.ExpiryDate == nil {
			continue
		}

		// PATCH: Use ceiling for precision - product expiring at 23:59 should count as 1 day remaining
		hoursUntilExpiry := product.ExpiryDate.Sub(now).Hours()
		daysRemaining := int(math.Ceil(hoursUntilExpiry / 24))

		// Story 4.5, Subtask 2.3: Categorize by alert level
		var alertLevel string

		// PATCH: Improved alert level calculation with boundary handling
		// Go switch doesn't fall through, so cases are evaluated in order:
		// 0-7 days: urgent (first match)
		// 8-14 days: critical (second match)
		// 15-30 days: warning (third match)
		// >30 or <0: skip
		switch {
		case daysRemaining < 0:
			// Product already expired, skip
			continue
		case daysRemaining <= 7:
			alertLevel = "urgent"
		case daysRemaining <= 14:
			alertLevel = "critical"
		case daysRemaining <= 30:
			alertLevel = "warning"
		default:
			continue
		}

		// Story 4.5, Subtask 2.4: Debounce logic - check if we've already alerted for this threshold
		alreadyAlerted, err := s.checkDebounce(ctx, product.ID, product.BranchID, alertLevel)
		if err != nil {
			s.logger.Warn("failed to check debounce, continuing", "productID", product.ID, "error", err)
			// Continue without debounce check on Redis failure - fail open
		} else if alreadyAlerted {
			// Skip if we've already alerted for this threshold within 24 hours
			continue
		}

		// Create the expiry alert event
		event := &dto.ExpiryAlertEvent{
			EventID:   uuid.New().String(),
			EventType: "product.expiry",
			Timestamp: now.Format(time.RFC3339),
			Data: dto.ProductExpiryData{
				ProductID:     product.ID,
				SKU:           product.SKU,
				ProductName:   product.Name,
				ExpiryDate:    product.ExpiryDate.Format(time.RFC3339),
				DaysRemaining: daysRemaining,
				AlertLevel:    alertLevel,
				BranchID:      product.BranchID,
				BranchName:    s.getBranchName(product),
			},
		}

		// Publish the alert
		if err := s.alertService.PublishExpiryAlert(ctx, event); err != nil {
			s.logger.Error("failed to publish expiry alert", "productID", product.ID, "error", err)
			// Continue processing other products
			continue
		}

		// PATCH: Update debounce tracking after successful publish
		// If update fails, log error but consider publish successful (alert was sent)
		// Debounce key will expire naturally via TTL, allowing re-alert after 24 hours
		if err := s.updateDebounce(ctx, product.ID, product.BranchID, alertLevel); err != nil {
			s.logger.Warn("failed to update debounce tracking", "productID", product.ID, "error", err)
			// Non-critical error - alert was sent successfully, just debounce tracking failed
			// Next check in 6 hours may send duplicate alert, but this is preferable to missing alerts
		}

		events = append(events, event)
	}

	s.logger.Info("expiry check completed", "alertsGenerated", len(events))
	return events, nil
}

// checkDebounce checks if an alert was already sent for this product/threshold within 24 hours
// Story 4.5, Subtask 2.4: Debounce logic using Redis Sorted Set
func (s *ExpiryCheckService) checkDebounce(ctx context.Context, productID uint, branchID uint, alertLevel string) (bool, error) {
	if s.redisClient == nil {
		return false, nil // No debounce if Redis not available
	}

	key := fmt.Sprintf("expiry_alerts:%d:%d:%s", productID, branchID, alertLevel)

	// Check if the key exists (alert sent within last 24 hours)
	exists, err := s.redisClient.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}

	return exists > 0, nil
}

// updateDebounce updates the debounce tracking with a 24-hour TTL
// Story 4.5, Subtask 3.4: Update debounce tracking in Redis
func (s *ExpiryCheckService) updateDebounce(ctx context.Context, productID uint, branchID uint, alertLevel string) error {
	if s.redisClient == nil {
		return nil // No debounce tracking if Redis not available
	}

	key := fmt.Sprintf("expiry_alerts:%d:%d:%s", productID, branchID, alertLevel)

	// Set key with 24-hour TTL
	err := s.redisClient.Set(ctx, key, time.Now().Format(time.RFC3339), 24*time.Hour).Err()
	if err != nil {
		return fmt.Errorf("failed to set debounce key: %w", err)
	}

	return nil
}

// ClearExpiryAlertState clears the debounce state for a product
// Story 4.5, Subtask 3.5: Add ClearExpiryAlertState method
func (s *ExpiryCheckService) ClearExpiryAlertState(ctx context.Context, productID uint, branchID uint) error {
	if s.redisClient == nil {
		return nil
	}

	// Clear all alert level keys for this product/branch
	pattern := fmt.Sprintf("expiry_alerts:%d:%d:*", productID, branchID)
	iter := s.redisClient.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		if err := s.redisClient.Del(ctx, iter.Val()).Err(); err != nil {
			s.logger.Warn("failed to clear expiry alert state", "key", iter.Val(), "error", err)
		}
	}
	if err := iter.Err(); err != nil {
		return fmt.Errorf("failed to scan expiry alert keys: %w", err)
	}

	return nil
}

// getBranchName extracts the branch name from a product
// PATCH: Added fallback for when Branch is not preloaded
func (s *ExpiryCheckService) getBranchName(product *models.Product) string {
	if product.Branch != nil {
		return product.Branch.Name
	}
	// PATCH: Log warning when Branch is not preloaded - this should not happen
	// as GetExpiringProducts uses Preload("Branch")
	s.logger.Warn("Branch not preloaded for product", "product_id", product.ID, "branch_id", product.BranchID)
	return fmt.Sprintf("Branch %d", product.BranchID)
}
