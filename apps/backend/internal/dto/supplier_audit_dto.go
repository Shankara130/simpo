package dto

// Story 10.7: Implement Supplier Transaction Audit Trail

import (
	"time"
)

// SupplierAuditQueryRequest represents the request parameters for querying supplier audit trail
// AC: Query audit trail with filters for date range, transaction type, entity, user, branch
type SupplierAuditQueryRequest struct {
	// Date range filter
	StartDate *time.Time `json:"startDate,omitempty" form:"start_date"`
	EndDate   *time.Time `json:"endDate,omitempty" form:"end_date"`

	// Transaction type filter (supplier_operation, purchase_invoice, goods_receipt, payment, return)
	TransactionType *string `json:"transactionType,omitempty" form:"transaction_type"`

	// Entity type filter (supplier, purchase_invoice, supplier_payment)
	EntityType *string `json:"entityType,omitempty" form:"entity_type"`

	// Specific entity filter
	EntityID *uint `json:"entityId,omitempty" form:"entity_id"`

	// User filter
	UserID *uint `json:"userId,omitempty" form:"user_id"`

	// Branch filter
	BranchID *uint `json:"branchId,omitempty" form:"branch_id"`

	// Pagination
	Page  *int `json:"page,omitempty" form:"page" binding:"required,min=1"`
	Limit *int `json:"limit,omitempty" form:"limit" binding:"required,min=1,max=100"`
}

// SupplierAuditTrailItem represents a single audit trail entry
type SupplierAuditTrailItem struct {
	ID                   uint       `json:"id"`
	TransactionType      string     `json:"transactionType"`
	EntityType           string     `json:"entityType"`
	EntityID             uint       `json:"entityId"`
	UserID               uint       `json:"userId"`
	UserRole             string     `json:"userRole"`
	ActionType           string     `json:"actionType"`
	ActionDescription     string     `json:"actionDescription"`
	Reason               *string    `json:"reason,omitempty"`
	TransactionAmount    *float64   `json:"transactionAmount,omitempty"`
	AffectedItemsCount   int        `json:"affectedItemsCount"`
	IPAddress            string     `json:"ipAddress,omitempty"`
	UserAgent            string     `json:"userAgent,omitempty"`
	BranchID             uint       `json:"branchId"`
	CreatedAt            time.Time  `json:"createdAt"`
}

// SupplierAuditTrailResponse represents the response for audit trail queries
type SupplierAuditTrailResponse struct {
	Data       []SupplierAuditTrailItem `json:"data"`
	Pagination PaginationMeta            `json:"pagination,omitempty"`
}

// SupplierAuditExportRequest represents the request for exporting audit trail
// AC: Export audit trail for compliance inspections
type SupplierAuditExportRequest struct {
	// Date range filter (required)
	StartDate time.Time `json:"startDate" form:"start_date" binding:"required"`
	EndDate   time.Time `json:"endDate" form:"end_date" binding:"required"`

	// Transaction type filter (optional)
	TransactionType *string `json:"transactionType,omitempty" form:"transaction_type"`

	// Branch filter (optional)
	BranchID *uint `json:"branchId,omitempty" form:"branch_id"`

	// Export format (csv or pdf)
	Format string `json:"format" form:"format" binding:"required,oneof=csv pdf"`
}
