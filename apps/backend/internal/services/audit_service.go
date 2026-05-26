package services

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/models"
)

// AuditAction represents the type of audit action
type AuditAction string

const (
	// Authentication actions
	AuditActionLoginSuccess       AuditAction = "LOGIN_SUCCESS"
	AuditActionLoginFailure       AuditAction = "LOGIN_FAILURE"
	AuditActionLogout             AuditAction = "LOGOUT"
	AuditActionPasswordReset      AuditAction = "PASSWORD_RESET"
	// Story 1.6, AC6: Authorization audit logging
	AuditActionAuthFailure        AuditAction = "AUTH_FAILURE"
	AuditActionForbiddenAccess    AuditAction = "FORBIDDEN_ACCESS"

	// User management actions
	AuditActionUserCreated        AuditAction = "USER_CREATED"
	AuditActionUserDeactivated    AuditAction = "USER_DEACTIVATED"
	AuditActionSelfRegistration   AuditAction = "SELF_REGISTRATION"
	AuditActionEmailVerified      AuditAction = "EMAIL_VERIFIED"

	// Whitelist management actions
	AuditActionWhitelistDomainAdded    AuditAction = "WHITELIST_DOMAIN_ADDED"
	AuditActionWhitelistDomainUpdated  AuditAction = "WHITELIST_DOMAIN_UPDATED"
	AuditActionWhitelistDomainDeleted  AuditAction = "WHITELIST_DOMAIN_DELETED"

	// Inventory actions
	AuditActionStockAdjustment     AuditAction = "STOCK_ADJUSTMENT"
	AuditActionBlockedSaleAttempt  AuditAction = "BLOCKED_SALE_ATTEMPT"

	// Reporting actions
	AuditActionExportReport        AuditAction = "EXPORT_REPORT"

	// Story 6.1, AC7: System settings actions
	AuditActionSettingsUpdated     AuditAction = "SETTINGS_UPDATED"
)

// AuditLogEntry represents an append-only audit log entry (Story 1.5, AC7, NFR-SEC-004)
type AuditLogEntry struct {
	UserID    *uint      `json:"user_id,omitempty"`
	Username  string     `json:"username"`
	Action    AuditAction `json:"action"`
	IPAddress string     `json:"ip_address"`
	Outcome   string     `json:"outcome"`
	Reason    string     `json:"reason,omitempty"`
	Timestamp time.Time  `json:"timestamp"`
}

// AuditRepository defines the interface for audit log persistence
// Story 5.4: Interface for persistent audit storage (replaces stdout logging)
type AuditRepository interface {
	Create(ctx context.Context, auditLog *models.AuditLog) error
}

// AuditService defines the audit logging interface (Story 1.5, AC7, NFR-SEC-004)
// Audit logs are append-only per NFR-SEC-004 compliance requirements
type AuditService interface {
	// LogLoginAttempt logs login attempts (success and failure) to append-only audit trail
	LogLoginAttempt(ctx context.Context, entry AuditLogEntry) error

	// LogAuthorizationFailure logs authorization failures (403 responses)
	// Story 1.6, AC6: All authorization failures are logged with user_id, role, endpoint, reason
	LogAuthorizationFailure(ctx context.Context, entry AuditLogEntry) error

	// LogUserCreation logs user creation actions (Story 1.7, AC7)
	// Logs admin_user_id, created_user_id, action, timestamp, ip_address
	LogUserCreation(ctx context.Context, adminID uint, createdUserID uint, adminUsername string, createdUsername string, ipAddress string) error

	// LogWhitelistChange logs whitelist domain management actions (Story 1.9, AC8)
	// Logs admin_user_id, domain, action, timestamp, ip_address
	LogWhitelistChange(ctx context.Context, adminID uint, adminUsername string, domain string, action AuditAction, ipAddress string) error

	// LogSelfRegistration logs staff self-registration actions (Story 1.9, AC8)
	// Logs user_id, email, domain, action, timestamp, ip_address
	LogSelfRegistration(ctx context.Context, userID uint, email string, domain string, ipAddress string) error

	// LogEmailVerification logs email verification actions (Story 1.9, AC8)
	// Logs user_id, email, action, timestamp, ip_address
	LogEmailVerification(ctx context.Context, userID uint, email string, ipAddress string) error

	// LogUserDeactivation logs user deactivation actions (Story 1.10, AC5)
	// Logs admin_user_id, deactivated_user_id, admin_username, deactivated_username, reason, action, timestamp, ip_address
	LogUserDeactivation(ctx context.Context, adminID uint, deactivatedUserID uint, adminUsername string, deactivatedUsername string, reason string, ipAddress string) error

	// LogStockAdjustment logs manual stock adjustment actions (Story 4.3, AC5)
	// Logs admin_user_id, product_id, product_sku, old_qty, new_qty, reason, timestamp, ip_address
	// Append-only audit trail for Badan POM compliance (NFR-SEC-004, NFR-SEC-009)
	// Story 5.4, Task 4.5: Added ipAddress parameter for IP address extraction from request context
	LogStockAdjustment(ctx context.Context, adminID uint, adminUsername string, productID uint, productSKU string, oldQty int64, newQty int64, reason string, ipAddress string) error

	// LogBlockedSaleAttempt logs blocked sale attempts for expired products (Story 4.6, AC6)
	// Logs user_id, username, product_id, product_sku, product_name, expiry_date, reason, timestamp, ip_address
	// Append-only audit trail for Badan POM regulatory compliance (NFR-SEC-004, NFR-SEC-009, NFR-SEC-011)
	// Story 5.4, Task 4.5: Added ipAddress parameter for IP address extraction from request context
	LogBlockedSaleAttempt(ctx context.Context, userID uint, username string, productID uint, productSKU string, productName string, expiryDate string, reason string, ipAddress string) error

	// LogReportExport logs report export events (Story 5.3, Task 4.7)
	// Logs user_id, username, report_type, format, date_range, timestamp, ip_address
	// Append-only audit trail for Badan POM regulatory compliance (NFR-SEC-004, NFR-SEC-009)
	// Story 5.4, Task 4.5: Added ipAddress parameter for IP address extraction from request context
	LogReportExport(ctx context.Context, userID uint, username string, reportType string, format string, dateRange string, outcome string, ipAddress string) error

	// LogSettingsUpdate logs system settings changes (Story 6.1, AC7)
	// Logs admin_user_id, changes (JSON), timestamp, ip_address
	// Append-only audit trail for Badan POM regulatory compliance
	LogSettingsUpdate(ctx context.Context, adminID uint, adminUsername string, changesJSON string, ipAddress string) error
}

// auditService implements AuditService with persistent audit trail storage
// Story 5.4: Implements persistent audit trail storage (replaces MVP stdout logging)
type auditService struct {
	repo AuditRepository // Story 5.4: Inject repository for persistent storage
}

// NewAuditService creates a new audit service instance
// Story 5.4, Task 4.1: Modify audit_service.go to inject AuditRepository dependency
func NewAuditService(repo AuditRepository) AuditService {
	if repo == nil {
		// For backward compatibility during migration, allow nil repo (falls back to stdout)
		slog.Warn("AuditRepository not provided - using stdout logging (should be temporary)")
		return &auditService{repo: nil}
	}
	return &auditService{repo: repo}
}

// createAuditLogEntry creates an AuditLog model from AuditLogEntry
func (s *auditService) createAuditLogEntry(entry AuditLogEntry) *models.AuditLog {
	userID := uint(0)
	if entry.UserID != nil {
		userID = *entry.UserID
	}

	return &models.AuditLog{
		UserID:    userID,
		Username:  entry.Username,
		Action:    models.AuditAction(entry.Action),
		IPAddress: entry.IPAddress,
			Outcome:   entry.Outcome,
		Reason:    entry.Reason,
		Timestamp: entry.Timestamp,
	}
}

// persistToDatabase persists audit log to database via repository
// Story 5.4, Task 4.2: Update all Log* methods to call repository.CreateAuditLog instead of stdout
// Story 5.4, Task 4.3: Add error handling for audit log failures (log to stderr, don't block operations)
func (s *auditService) persistToDatabase(ctx context.Context, entry AuditLogEntry) error {
	if s.repo == nil {
		return fmt.Errorf("audit repository not configured")
	}

	// Story 5.4, Task 4.4: Add context cancellation checks before database writes
	select {
	case <-ctx.Done():
		return fmt.Errorf("context cancelled: %w", ctx.Err())
	default:
		// Continue with database write
	}

	// Create audit log model from entry
	auditLog := s.createAuditLogEntry(entry)

	// Persist to database
	err := s.repo.Create(ctx, auditLog)
	if err != nil {
		// Story 5.4, Task 4.3: Log audit log failures to stderr (don't block operations)
		// Non-blocking: Log error but don't fail the operation
		slog.Error("AUDIT_LOG_FAILED",
			"timestamp", time.Now().Format(time.RFC3339),
			"action", string(entry.Action),
			"username", entry.Username,
			"user_id", entry.UserID,
			"error", err.Error(),
		)

		// CRIT-005: Audit write failure handling for compliance
		// In production, implement:
		// 1. Metrics tracking (Prometheus counter for audit failures)
		//    Example: auditWriteFailureCounter.Inc()
		// 2. Retry logic for transient failures (with exponential backoff)
		// 3. Dead letter queue for audit entries that failed after retries
		// 4. Alerting for high audit failure rates
		//
		// For MVP: Error is logged but doesn't block operations (non-blocking design)

		return err
	}

	return nil
}

// LogLoginAttempt logs login attempt to append-only audit trail (Story 1.5, AC7)
func (s *auditService) LogLoginAttempt(ctx context.Context, entry AuditLogEntry) error {
	// Set timestamp if not provided
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now()
	}

	// Story 5.4: Persist to database (replaces stdout-only logging)
	_ = s.persistToDatabase(ctx, entry)

	// Keep stdout logging as fallback/backup for visibility during development
	slog.Info("AUDIT",
		"timestamp", entry.Timestamp.Format(time.RFC3339),
		"action", string(entry.Action),
		"username", entry.Username,
		"user_id", entry.UserID,
		"ip_address", entry.IPAddress,
		"outcome", entry.Outcome,
		"reason", entry.Reason,
	)

	return nil
}

// LogAuthorizationFailure logs authorization failures to append-only audit trail
// Story 1.6, AC6: All authorization failures are logged with user_id, role, endpoint, reason
func (s *auditService) LogAuthorizationFailure(ctx context.Context, entry AuditLogEntry) error {
	// Set timestamp if not provided
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now()
	}

	// Story 5.4: Persist to database
	_ = s.persistToDatabase(ctx, entry)

	// Keep stdout logging as fallback
	slog.Info("AUDIT",
		"timestamp", entry.Timestamp.Format(time.RFC3339),
		"action", string(entry.Action),
		"username", entry.Username,
		"user_id", entry.UserID,
		"ip_address", entry.IPAddress,
		"outcome", entry.Outcome,
		"reason", entry.Reason,
	)

	return nil
}

// LogUserCreation logs user creation actions to append-only audit trail (Story 1.7, AC7)
func (s *auditService) LogUserCreation(ctx context.Context, adminID uint, createdUserID uint, adminUsername string, createdUsername string, ipAddress string) error {
	entry := AuditLogEntry{
		UserID:    &adminID,
		Username:  adminUsername,
		Action:    AuditActionUserCreated,
		IPAddress: ipAddress,
		Outcome:   "success",
		Reason:    fmt.Sprintf("Created user '%s' (ID: %d)", createdUsername, createdUserID),
		Timestamp: time.Now(),
	}

	// Story 5.4: Persist to database
	_ = s.persistToDatabase(ctx, entry)

	// Keep stdout logging as fallback
	slog.Info("AUDIT",
		"timestamp", entry.Timestamp.Format(time.RFC3339),
		"action", string(entry.Action),
		"admin_user_id", adminID,
		"admin_username", adminUsername,
		"created_user_id", createdUserID,
		"created_username", createdUsername,
		"ip_address", ipAddress,
		"outcome", entry.Outcome,
	)

	return nil
}

// LogWhitelistChange logs whitelist domain management actions to append-only audit trail (Story 1.9, AC8)
func (s *auditService) LogWhitelistChange(ctx context.Context, adminID uint, adminUsername string, domain string, action AuditAction, ipAddress string) error {
	entry := AuditLogEntry{
		UserID:    &adminID,
		Username:  adminUsername,
		Action:    action,
		IPAddress: ipAddress,
		Outcome:   "success",
		Reason:    fmt.Sprintf("Domain '%s'", domain),
		Timestamp: time.Now(),
	}

	// Story 5.4: Persist to database
	_ = s.persistToDatabase(ctx, entry)

	// Keep stdout logging as fallback
	slog.Info("AUDIT",
		"timestamp", entry.Timestamp.Format(time.RFC3339),
		"action", string(entry.Action),
		"admin_user_id", adminID,
		"admin_username", adminUsername,
		"domain", domain,
		"ip_address", ipAddress,
		"outcome", entry.Outcome,
	)

	return nil
}

// LogSelfRegistration logs staff self-registration actions to append-only audit trail (Story 1.9, AC8)
func (s *auditService) LogSelfRegistration(ctx context.Context, userID uint, email string, domain string, ipAddress string) error {
	entry := AuditLogEntry{
		UserID:    &userID,
		Username:  email, // Use email as username for self-registration
		Action:    AuditActionSelfRegistration,
		IPAddress: ipAddress,
		Outcome:   "success",
		Reason:    fmt.Sprintf("Self-registered from domain '%s'", domain),
		Timestamp: time.Now(),
	}

	// Story 5.4: Persist to database
	_ = s.persistToDatabase(ctx, entry)

	// Keep stdout logging as fallback
	slog.Info("AUDIT",
		"timestamp", entry.Timestamp.Format(time.RFC3339),
		"action", string(entry.Action),
		"user_id", userID,
		"email", email,
		"domain", domain,
		"ip_address", ipAddress,
		"outcome", entry.Outcome,
	)

	return nil
}

// LogEmailVerification logs email verification actions to append-only audit trail (Story 1.9, AC8)
func (s *auditService) LogEmailVerification(ctx context.Context, userID uint, email string, ipAddress string) error {
	entry := AuditLogEntry{
		UserID:    &userID,
		Username:  email, // Use email as username for email verification
		Action:    AuditActionEmailVerified,
		IPAddress: ipAddress,
		Outcome:   "success",
		Reason:    "Email verified and account activated",
		Timestamp: time.Now(),
	}

	// Story 5.4: Persist to database
	_ = s.persistToDatabase(ctx, entry)

	// Keep stdout logging as fallback
	slog.Info("AUDIT",
		"timestamp", entry.Timestamp.Format(time.RFC3339),
		"action", string(entry.Action),
		"user_id", userID,
		"email", email,
		"ip_address", ipAddress,
		"outcome", entry.Outcome,
	)

	return nil
}

// LogUserDeactivation logs user deactivation actions to append-only audit trail (Story 1.10, AC5)
func (s *auditService) LogUserDeactivation(ctx context.Context, adminID uint, deactivatedUserID uint, adminUsername string, deactivatedUsername string, reason string, ipAddress string) error {
	entry := AuditLogEntry{
		UserID:    &adminID,
		Username:  adminUsername,
		Action:    AuditActionUserDeactivated,
		IPAddress: ipAddress,
		Outcome:   "success",
		Reason:    fmt.Sprintf("Deactivated user '%s' (ID: %d) - Reason: %s", deactivatedUsername, deactivatedUserID, reason),
		Timestamp: time.Now(),
	}

	// Story 5.4: Persist to database
	_ = s.persistToDatabase(ctx, entry)

	// Keep stdout logging as fallback
	slog.Info("AUDIT",
		"timestamp", entry.Timestamp.Format(time.RFC3339),
		"action", string(entry.Action),
		"admin_user_id", adminID,
		"admin_username", adminUsername,
		"deactivated_user_id", deactivatedUserID,
		"deactivated_username", deactivatedUsername,
		"reason", reason,
		"ip_address", ipAddress,
		"outcome", entry.Outcome,
	)

	return nil
}

// LogStockAdjustment logs manual stock adjustment actions to append-only audit trail (Story 4.3, AC5)
// Story 5.4, Task 4.5: Added ipAddress parameter for IP address extraction from request context
func (s *auditService) LogStockAdjustment(ctx context.Context, adminID uint, adminUsername string, productID uint, productSKU string, oldQty int64, newQty int64, reason string, ipAddress string) error {
	entry := AuditLogEntry{
		UserID:    &adminID,
		Username:  adminUsername,
		Action:    AuditActionStockAdjustment,
		IPAddress: ipAddress, // Story 5.4, Task 4.5: IP address extracted from request context
		Outcome:   "success",
		Reason:    fmt.Sprintf("Adjusted stock for product '%s' (ID: %d): %d → %d - Reason: %s", productSKU, productID, oldQty, newQty, reason),
		Timestamp: time.Now(),
	}

	// Story 5.4: Persist to database
	_ = s.persistToDatabase(ctx, entry)

	// Keep stdout logging as fallback
	slog.Info("AUDIT",
		"timestamp", entry.Timestamp.Format(time.RFC3339),
		"action", string(entry.Action),
		"admin_user_id", adminID,
		"admin_username", adminUsername,
		"product_id", productID,
		"product_sku", productSKU,
		"old_qty", oldQty,
		"new_qty", newQty,
		"reason", reason,
		"outcome", entry.Outcome,
	)

	// Story 5.4, Task 4.6: Update existing TODO comments to reference this story (Story 5.4)
	// TODO comment cleanup done - now using persistent storage

	return nil
}

// LogReportExport logs report export events for regulatory compliance
// Story 5.3, Task 4.7: Add export event logging to audit trail
// Story 5.4, Task 4.5: Added ipAddress parameter for IP address extraction from request context
func (s *auditService) LogReportExport(ctx context.Context, userID uint, username string, reportType string, format string, dateRange string, outcome string, ipAddress string) error {
	entry := AuditLogEntry{
		UserID:    &userID,
		Username:  username,
		Action:    AuditActionExportReport,
		IPAddress: ipAddress, // Story 5.4, Task 4.5: IP address extracted from request context
		Outcome:   outcome,
		Reason:    fmt.Sprintf("Exported report: type=%s, format=%s, range=%s", reportType, format, dateRange),
		Timestamp: time.Now(),
	}

	// Story 5.4: Persist to database
	_ = s.persistToDatabase(ctx, entry)

	// Keep stdout logging as fallback for MVP
	slog.Warn("AUDIT",
		"timestamp", entry.Timestamp.Format(time.RFC3339),
		"action", string(entry.Action),
		"user_id", userID,
		"username", username,
		"report_type", reportType,
		"format", format,
		"date_range", dateRange,
		"outcome", entry.Outcome,
	)

	// Story 5.4, Task 4.6: Update existing TODO comments to reference this story (Story 5.4)
	// TODO comment cleanup done - now using persistent storage

	return nil
}


// LogBlockedSaleAttempt logs blocked sale attempts for expired products to append-only audit trail
// Story 4.6, AC6: Audit trail logging for blocked sale attempts (regulatory compliance)
// Per NFR-SEC-004: audit trail must be append-only (no delete/update)
// Per NFR-SEC-009: 5-year minimum retention for Badan POM compliance
// Per NFR-SEC-011: Expiry date blocking is mandatory with audit logging
// Story 5.4, Task 4.5: Added ipAddress parameter for IP address extraction from request context
func (s *auditService) LogBlockedSaleAttempt(ctx context.Context, userID uint, username string, productID uint, productSKU string, productName string, expiryDate string, reason string, ipAddress string) error {
	entry := AuditLogEntry{
		UserID:    &userID,
		Username:  username,
		Action:    AuditActionBlockedSaleAttempt,
		IPAddress: ipAddress, // Story 5.4, Task 4.5: IP address extracted from request context
		Outcome:   "blocked",
		Reason:    fmt.Sprintf("Blocked sale attempt for expired product '%s' (ID: %d, SKU: %s, Expiry: %s): %s", productName, productID, productSKU, expiryDate, reason),
		Timestamp: time.Now(),
	}

	// Story 5.4: Persist to database
	_ = s.persistToDatabase(ctx, entry)

	// Keep stdout logging as fallback for MVP
	// Use Warn level for audit logs to ensure they're not filtered in production
	slog.Warn("AUDIT",
		"timestamp", entry.Timestamp.Format(time.RFC3339),
		"action", string(entry.Action),
		"user_id", userID,
		"username", username,
		"product_id", productID,
		"product_sku", productSKU,
		"product_name", productName,
		"expiry_date", expiryDate,
		"reason", reason,
		"outcome", entry.Outcome,
	)

	// Story 5.4, Task 4.6: Update existing TODO comments to reference this story (Story 5.4)
	// TODO comment cleanup done - now using persistent storage

	return nil
}

// LogSettingsUpdate logs system settings changes to append-only audit trail
// Story 6.1, AC7: Audit trail for configuration changes
// Per NFR-SEC-004: audit trail must be append-only (no delete/update)
// Logs admin_user_id, username, changes (JSON), timestamp, ip_address
func (s *auditService) LogSettingsUpdate(ctx context.Context, adminID uint, adminUsername string, changesJSON string, ipAddress string) error {
	entry := AuditLogEntry{
		UserID:    &adminID,
		Username:  adminUsername,
		Action:    AuditActionSettingsUpdated,
		IPAddress: ipAddress,
		Outcome:   "success",
		Reason:    changesJSON, // Store changes as JSON in reason field
		Timestamp: time.Now(),
	}

	// Story 5.4: Persist to database
	_ = s.persistToDatabase(ctx, entry)

	// Keep stdout logging as fallback for MVP
	slog.Info("AUDIT",
		"timestamp", entry.Timestamp.Format(time.RFC3339),
		"action", string(entry.Action),
		"admin_user_id", adminID,
		"admin_username", adminUsername,
		"changes", changesJSON,
		"ip_address", ipAddress,
		"outcome", entry.Outcome,
	)

	return nil
}
