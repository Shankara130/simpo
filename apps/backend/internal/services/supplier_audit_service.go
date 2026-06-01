package services

import (
	"context"
	"time"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/dto"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/models"
)

// Story 10.7: Implement Supplier Transaction Audit Trail

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

// SupplierAuditTrailResponse represents the response for audit trail queries
type SupplierAuditTrailResponse struct {
	Data       []models.SupplierAuditTrail `json:"data"`
	Pagination dto.PaginationMeta          `json:"pagination,omitempty"`
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

// SupplierAuditService defines the interface for supplier audit trail operations
// AC: Maintain append-only audit trail for all supplier transactions
type SupplierAuditService interface {
	// LogSupplierOperation logs a supplier operation to the audit trail
	// AC: Automatically creates audit entry with Who, When, What, Why, How much
	LogSupplierOperation(ctx context.Context, auditLog *models.SupplierAuditTrail) error

	// QueryAuditTrail retrieves audit trail entries based on filters
	// AC: Queryable for 5 years per Badan POM requirements
	QueryAuditTrail(ctx context.Context, request *SupplierAuditQueryRequest) (*SupplierAuditTrailResponse, error)

	// ExportAuditTrail exports audit trail data for compliance inspections
	// AC: Audit logs can be exported for compliance inspections
	// Returns the audit trail data for export by the handler layer
	ExportAuditTrail(ctx context.Context, request *SupplierAuditExportRequest) ([]models.SupplierAuditTrail, error)

	// GetAuditByEntityID retrieves all audit entries for a specific entity
	// Used for displaying complete history of supplier, invoice, or payment
	GetAuditByEntityID(ctx context.Context, entityType string, entityID uint) ([]models.SupplierAuditTrail, error)

	// GetAuditByUserID retrieves audit entries for a specific user within a date range
	// Used for user activity tracking and compliance reporting
	GetAuditByUserID(ctx context.Context, userID uint, startDate, endDate time.Time) ([]models.SupplierAuditTrail, error)
}
