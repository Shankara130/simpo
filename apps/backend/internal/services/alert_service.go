package services

import (
	"context"
	"time"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/dto"
)

// AlertService defines the interface for alert business operations
// AC1: Service interface for alert domain with clear business method signatures
type AlertService interface {
	// CheckLowStockAlerts checks for products with stock below reorder threshold
	// Business rule: stock_qty <= reorder_threshold
	CheckLowStockAlerts(ctx context.Context, branchID uint) ([]*LowStockAlert, error)

	// CheckExpiryAlerts checks for products expiring soon
	// Business rule: expiry within 30/14/7 days
	CheckExpiryAlerts(ctx context.Context, branchID uint, daysThreshold int) ([]*ExpiryAlert, error)

	// SendNotification sends alert notifications
	// Stub for future Redis pub/sub story
	SendNotification(ctx context.Context, alert interface{}) error

	// PublishLowStockAlert publishes low stock notification to Redis pub/sub
	// Story 4.4, AC2, AC3, AC6: Publish stock.low event with debounce logic
	PublishLowStockAlert(ctx context.Context, event *dto.LowStockNotificationEvent) error

	// PublishExpiryAlert publishes expiry notification to Redis pub/sub
	// Story 4.5, AC4, AC6: Publish product.expiry event with debounce logic
	PublishExpiryAlert(ctx context.Context, event *dto.ExpiryAlertEvent) error

	// ClearLowStockState clears low stock state when stock returns to normal
	// Story 4.4, AC7: Debounce logic - remove tracking when stock >= threshold
	ClearLowStockState(ctx context.Context, productID uint, branchID uint) error
}

// LowStockAlert represents a low stock alert
type LowStockAlert struct {
	ProductID       uint
	ProductName     string
	SKU             string
	CurrentStock    int64
	ReorderThreshold int
	BranchID        uint
	BranchName      string
	Severity        string // HIGH, MEDIUM, LOW
}

// ExpiryAlert represents an expiry alert
type ExpiryAlert struct {
	ProductID    uint
	ProductName  string
	SKU          string
	ExpiryDate   time.Time
	DaysUntilExpiry int
	BranchID     uint
	BranchName   string
	Severity     string // CRITICAL (7 days), WARNING (14 days), INFO (30 days)
}
