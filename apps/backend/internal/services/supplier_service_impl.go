package services

import (
	"context"
	"fmt"
	"strings"

	"golang.org/x/text/unicode/norm"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/models"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/repositories"
)

// supplierServiceImpl implements SupplierService interface
// Story 10.1: Service layer with business logic and audit logging
type supplierServiceImpl struct {
	supplierRepo repositories.SupplierRepository
	auditService AuditService
}

// NewSupplierService creates a new supplier service
// Story 10.1: Factory function with dependency injection
func NewSupplierService(supplierRepo repositories.SupplierRepository, auditService AuditService) SupplierService {
	return &supplierServiceImpl{
		supplierRepo: supplierRepo,
		auditService: auditService,
	}
}

// CreateSupplier creates a new supplier with validation and audit logging
// Story 10.1, AC1: Validates required fields (name, phone) and logs creation
func (s *supplierServiceImpl) CreateSupplier(ctx context.Context, supplier *models.Supplier, createdBy uint, ipAddress string) (*models.Supplier, error) {
	// Validate required fields
	if supplier == nil {
		return nil, fmt.Errorf("supplier cannot be nil")
	}

	// Normalize and trim supplier name
	supplier.Name = norm.NFKC.String(strings.TrimSpace(supplier.Name))
	if supplier.Name == "" {
		return nil, fmt.Errorf("supplier name is required")
	}

	supplier.Phone = strings.TrimSpace(supplier.Phone)
	if supplier.Phone == "" {
		return nil, fmt.Errorf("supplier phone is required")
	}
	if createdBy == 0 {
		return nil, fmt.Errorf("createdBy user ID is required")
	}

	// Check for duplicate supplier name (with normalized name)
	existing, err := s.supplierRepo.GetByName(ctx, supplier.Name)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("supplier with name '%s' already exists", supplier.Name)
	}

	// Set default active status
	supplier.IsActive = true

	// Create supplier (database unique constraint provides final safety against race condition)
	err = s.supplierRepo.Create(ctx, supplier, createdBy)
	if err != nil {
		// Handle unique constraint violation from race condition
		if strings.Contains(err.Error(), "unique constraint") || strings.Contains(err.Error(), "duplicate") {
			return nil, fmt.Errorf("supplier with name '%s' already exists", supplier.Name)
		}
		return nil, fmt.Errorf("failed to create supplier: %w", err)
	}

	// Log to audit trail
	// Story 10.1, AC1: Logs supplier creation with admin user ID
	// Note: Using AuditActionBranchCreated as placeholder - will add supplier-specific actions in Task 9
	entry := AuditLogEntry{
		UserID:    &createdBy,
		Username:  fmt.Sprintf("user_%d", createdBy), // Will be resolved from context in handler
		Action:    models.AuditActionBranchCreated, // Placeholder - will be supplier.created in Task 9
		IPAddress: ipAddress,
		Outcome:   "success",
		Reason:    fmt.Sprintf("Created supplier: %s (ID: %d)", supplier.Name, supplier.ID),
	}
	if err := s.auditService.LogLoginAttempt(ctx, entry); err != nil {
		// Log error but don't fail the operation
		// In production, this should be logged to a separate error logger
	}

	return supplier, nil
}

// GetSupplierByID retrieves a supplier by ID
// Story 10.1, AC1: Returns supplier details or error if not found
func (s *supplierServiceImpl) GetSupplierByID(ctx context.Context, id uint) (*models.Supplier, error) {
	if id == 0 {
		return nil, fmt.Errorf("invalid supplier ID")
	}

	supplier, err := s.supplierRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get supplier: %w", err)
	}

	return supplier, nil
}

// ListSuppliers retrieves suppliers with filtering and pagination
// Story 10.1, AC2: Supports search, active status filter, and pagination
func (s *supplierServiceImpl) ListSuppliers(ctx context.Context, filter *SupplierListFilter) ([]*models.Supplier, int64, error) {
	// Set defaults if filter is nil
	if filter == nil {
		filter = &SupplierListFilter{
			Page:  1,
			Limit: 20,
		}
	}

	// Convert to repository filter
	repoFilter := &repositories.SupplierFilter{
		SearchQuery: filter.SearchQuery,
		IsActive:     filter.IsActive,
		Page:         filter.Page,
		Limit:        filter.Limit,
		SortBy:       filter.SortBy,
		SortOrder:    filter.SortOrder,
	}

	suppliers, total, err := s.supplierRepo.List(ctx, repoFilter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list suppliers: %w", err)
	}

	return suppliers, total, nil
}

// UpdateSupplier updates an existing supplier with validation and audit logging
// Story 10.1, AC2: Validates changes and logs update with reason
func (s *supplierServiceImpl) UpdateSupplier(ctx context.Context, id uint, updates *UpdateSupplierRequest, updatedBy uint, ipAddress string) (*models.Supplier, error) {
	// Validate inputs
	if id == 0 {
		return nil, fmt.Errorf("invalid supplier ID")
	}
	if updates == nil {
		return nil, fmt.Errorf("updates cannot be nil")
	}
	if strings.TrimSpace(updates.Reason) == "" {
		return nil, fmt.Errorf("reason is required for update")
	}
	if updatedBy == 0 {
		return nil, fmt.Errorf("updatedBy user ID is required")
	}

	// Validate required fields in updates
	if strings.TrimSpace(updates.Name) == "" {
		return nil, fmt.Errorf("supplier name is required")
	}
	if strings.TrimSpace(updates.Phone) == "" {
		return nil, fmt.Errorf("supplier phone is required")
	}

	// Get existing supplier
	existing, err := s.supplierRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get supplier: %w", err)
	}

	// Check for duplicate name if name is being changed
	if updates.Name != existing.Name {
		duplicate, err := s.supplierRepo.GetByName(ctx, updates.Name)
		if err == nil && duplicate != nil && duplicate.ID != id {
			return nil, fmt.Errorf("supplier with name '%s' already exists", updates.Name)
		}
	}

	// Apply updates
	existing.Name = updates.Name
	existing.ContactPerson = updates.ContactPerson
	existing.Phone = updates.Phone
	existing.Email = updates.Email
	existing.Address = updates.Address

	// Update supplier
	err = s.supplierRepo.Update(ctx, existing, updatedBy)
	if err != nil {
		return nil, fmt.Errorf("failed to update supplier: %w", err)
	}

	// Log to audit trail
	// Story 10.1, AC2: Logs supplier update with reason
	// Note: Using AuditActionBranchUpdated as placeholder - will add supplier-specific actions in Task 9
	entry := AuditLogEntry{
		UserID:    &updatedBy,
		Username:  fmt.Sprintf("user_%d", updatedBy), // Will be resolved from context in handler
		Action:    models.AuditActionBranchUpdated, // Placeholder - will be supplier.updated in Task 9
		IPAddress: ipAddress,
		Outcome:   "success",
		Reason:    fmt.Sprintf("Updated supplier: %s (ID: %d). Reason: %s", existing.Name, id, updates.Reason),
	}
	if err := s.auditService.LogLoginAttempt(ctx, entry); err != nil {
		// Log error but don't fail the operation
	}

	return existing, nil
}

// DeactivateSupplier deactivates a supplier with audit logging
// Story 10.1, AC3: Soft deletes supplier and logs deactivation with reason
func (s *supplierServiceImpl) DeactivateSupplier(ctx context.Context, id uint, reason string, deactivatedBy uint, ipAddress string) error {
	// Validate inputs
	if id == 0 {
		return fmt.Errorf("invalid supplier ID")
	}
	if strings.TrimSpace(reason) == "" {
		return fmt.Errorf("reason is required for deactivation")
	}
	if deactivatedBy == 0 {
		return fmt.Errorf("deactivatedBy user ID is required")
	}

	// Check if supplier exists
	supplier, err := s.supplierRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get supplier: %w", err)
	}

	// Deactivate supplier
	err = s.supplierRepo.Deactivate(ctx, id, deactivatedBy)
	if err != nil {
		return fmt.Errorf("failed to deactivate supplier: %w", err)
	}

	// Log to audit trail
	// Story 10.1, AC3: Logs supplier deactivation with reason
	// Note: Using AuditActionBranchDeactivated as placeholder - will add supplier-specific actions in Task 9
	entry := AuditLogEntry{
		UserID:    &deactivatedBy,
		Username:  fmt.Sprintf("user_%d", deactivatedBy), // Will be resolved from context in handler
		Action:    models.AuditActionBranchDeactivated, // Placeholder - will be supplier.deactivated in Task 9
		IPAddress: ipAddress,
		Outcome:   "success",
		Reason:    fmt.Sprintf("Deactivated supplier: %s (ID: %d). Reason: %s", supplier.Name, id, reason),
	}
	if err := s.auditService.LogLoginAttempt(ctx, entry); err != nil {
		// Log error but don't fail the operation
	}

	return nil
}
