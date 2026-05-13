package services

import (
	"context"
	"time"
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
