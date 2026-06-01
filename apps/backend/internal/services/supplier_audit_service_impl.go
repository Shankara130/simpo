package services

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/dto"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/models"
	"gorm.io/gorm"
)

// Story 10.7: Implement Supplier Transaction Audit Trail

// supplierAuditServiceImpl implements SupplierAuditService
type supplierAuditServiceImpl struct {
	db *gorm.DB
}

// NewSupplierAuditService creates a new instance of SupplierAuditService
func NewSupplierAuditService(db *gorm.DB) SupplierAuditService {
	return &supplierAuditServiceImpl{
		db: db,
	}
}

// LogSupplierOperation logs a supplier operation to the audit trail
// AC: Automatically creates audit entry with Who, When, What, Why, How much
func (s *supplierAuditServiceImpl) LogSupplierOperation(ctx context.Context, auditLog *models.SupplierAuditTrail) error {
	slog.InfoContext(ctx, "logging_supplier_operation",
		"transaction_type", auditLog.TransactionType,
		"entity_type", auditLog.EntityType,
		"entity_id", auditLog.EntityID,
		"user_id", auditLog.UserID,
		"action_type", auditLog.ActionType,
		"branch_id", auditLog.BranchID,
	)

	// Set created_at to current time (UTC for consistency)
	auditLog.CreatedAt = time.Now().UTC()

	// Append-only: No UPDATE or DELETE operations allowed
	// Only INSERT permitted for Badan POM compliance
	if err := s.db.Create(auditLog).Error; err != nil {
		slog.ErrorContext(ctx, "failed_to_create_audit_log",
			"error", err,
			"transaction_type", auditLog.TransactionType,
			"entity_id", auditLog.EntityID,
		)
		return fmt.Errorf("failed to create audit log: %w", err)
	}

	slog.InfoContext(ctx, "supplier_audit_log_created",
		"audit_id", auditLog.ID,
		"transaction_type", auditLog.TransactionType,
		"entity_id", auditLog.EntityID,
	)

	return nil
}

// QueryAuditTrail retrieves audit trail entries based on filters
// AC: Queryable for 5 years per Badan POM requirements
func (s *supplierAuditServiceImpl) QueryAuditTrail(ctx context.Context, request *SupplierAuditQueryRequest) (*SupplierAuditTrailResponse, error) {
	slog.InfoContext(ctx, "querying_supplier_audit_trail",
		"start_date", request.StartDate,
		"end_date", request.EndDate,
		"transaction_type", request.TransactionType,
		"entity_type", request.EntityType,
		"entity_id", request.EntityID,
		"user_id", request.UserID,
		"branch_id", request.BranchID,
	)

	// Build query with filters
	query := s.db.WithContext(ctx).Model(&models.SupplierAuditTrail{})

	// Date range filter
	if request.StartDate != nil {
		query = query.Where("created_at >= ?", *request.StartDate)
	}
	if request.EndDate != nil {
		query = query.Where("created_at <= ?", *request.EndDate)
	}

	// Transaction type filter
	if request.TransactionType != nil {
		query = query.Where("transaction_type = ?", *request.TransactionType)
	}

	// Entity type filter
	if request.EntityType != nil {
		query = query.Where("entity_type = ?", *request.EntityType)
	}

	// Entity ID filter
	if request.EntityID != nil {
		query = query.Where("entity_id = ?", *request.EntityID)
	}

	// User filter
	if request.UserID != nil {
		query = query.Where("user_id = ?", *request.UserID)
	}

	// Branch filter
	if request.BranchID != nil {
		query = query.Where("branch_id = ?", *request.BranchID)
	}

	// Pagination
	page := 1
	if request.Page != nil {
		page = *request.Page
	}

	limit := 20 // Default limit
	if request.Limit != nil {
		limit = *request.Limit
	}

	offset := (page - 1) * limit

	// Get total count for pagination
	var total int64
	if err := query.Count(&total).Error; err != nil {
		slog.ErrorContext(ctx, "failed_to_count_audit_trail", "error", err)
		return nil, fmt.Errorf("failed to count audit trail: %w", err)
	}

	// Execute query with pagination
	var audits []models.SupplierAuditTrail
	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&audits).Error; err != nil {
		slog.ErrorContext(ctx, "failed_to_query_audit_trail", "error", err)
		return nil, fmt.Errorf("failed to query audit trail: %w", err)
	}

	// Build response
	response := &SupplierAuditTrailResponse{
		Data: make([]models.SupplierAuditTrail, 0, len(audits)),
	}

	// Convert to DTO items
	for _, audit := range audits {
		response.Data = append(response.Data, audit)
	}

	// Add pagination metadata if there are results
	if len(audits) > 0 {
		response.Pagination = dto.PaginationMeta{
			Page:     page,
			Limit:    limit,
			Total:    total,
			TotalPages: int((total + int64(limit) - 1) / int64(limit)),
		}
	}

	slog.InfoContext(ctx, "supplier_audit_trail_queried",
		"count", len(audits),
		"total", total,
	)

	return response, nil
}

// ExportAuditTrail exports audit trail data for compliance inspections
// AC: Audit logs can be exported for compliance inspections
// Returns the audit trail data for export by the handler layer
func (s *supplierAuditServiceImpl) ExportAuditTrail(ctx context.Context, request *SupplierAuditExportRequest) ([]models.SupplierAuditTrail, error) {
	slog.InfoContext(ctx, "exporting_supplier_audit_trail",
		"start_date", request.StartDate,
		"end_date", request.EndDate,
		"format", request.Format,
		"transaction_type", request.TransactionType,
		"branch_id", request.BranchID,
	)

	// Build query with filters
	query := s.db.WithContext(ctx).Model(&models.SupplierAuditTrail{})

	// Date range filter (required)
	query = query.Where("created_at >= ?", request.StartDate)
	query = query.Where("created_at <= ?", request.EndDate)

	// Transaction type filter
	if request.TransactionType != nil {
		query = query.Where("transaction_type = ?", *request.TransactionType)
	}

	// Branch filter
	if request.BranchID != nil {
		query = query.Where("branch_id = ?", *request.BranchID)
	}

	// Execute query ordered by date ascending for export
	var audits []models.SupplierAuditTrail
	if err := query.Order("created_at ASC").Find(&audits).Error; err != nil {
		slog.ErrorContext(ctx, "failed_to_query_audit_for_export", "error", err)
		return nil, fmt.Errorf("failed to query audit trail for export: %w", err)
	}

	slog.InfoContext(ctx, "supplier_audit_trail_export_data_queried",
		"record_count", len(audits),
		"format", request.Format,
	)

	return audits, nil
}

// GetAuditByEntityID retrieves all audit entries for a specific entity
// Used for displaying complete history of supplier, invoice, or payment
func (s *supplierAuditServiceImpl) GetAuditByEntityID(ctx context.Context, entityType string, entityID uint) ([]models.SupplierAuditTrail, error) {
	slog.InfoContext(ctx, "getting_audit_by_entity",
		"entity_type", entityType,
		"entity_id", entityID,
	)

	var audits []models.SupplierAuditTrail
	if err := s.db.Where("entity_type = ? AND entity_id = ?", entityType, entityID).
		Order("created_at DESC").
		Find(&audits).Error; err != nil {
		slog.ErrorContext(ctx, "failed_to_get_audit_by_entity",
			"error", err,
			"entity_type", entityType,
			"entity_id", entityID,
		)
		return nil, fmt.Errorf("failed to get audit by entity: %w", err)
	}

	slog.InfoContext(ctx, "audit_by_entity_retrieved",
		"count", len(audits),
		"entity_type", entityType,
		"entity_id", entityID,
	)

	return audits, nil
}

// GetAuditByUserID retrieves audit entries for a specific user within a date range
// Used for user activity tracking and compliance reporting
func (s *supplierAuditServiceImpl) GetAuditByUserID(ctx context.Context, userID uint, startDate, endDate time.Time) ([]models.SupplierAuditTrail, error) {
	slog.InfoContext(ctx, "getting_audit_by_user",
		"user_id", userID,
		"start_date", startDate,
		"end_date", endDate,
	)

	var audits []models.SupplierAuditTrail
	query := s.db.Where("user_id = ?", userID)

	// Date range filter
	if !startDate.IsZero() {
		query = query.Where("created_at >= ?", startDate)
	}
	if !endDate.IsZero() {
		query = query.Where("created_at <= ?", endDate)
	}

	if err := query.Order("created_at DESC").Find(&audits).Error; err != nil {
		slog.ErrorContext(ctx, "failed_to_get_audit_by_user",
			"error", err,
			"user_id", userID,
		)
		return nil, fmt.Errorf("failed to get audit by user: %w", err)
	}

	slog.InfoContext(ctx, "audit_by_user_retrieved",
		"count", len(audits),
		"user_id", userID,
	)

	return audits, nil
}
