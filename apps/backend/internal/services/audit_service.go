package services

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

// AuditAction represents the type of audit action
type AuditAction string

const (
	AuditActionLoginSuccess       AuditAction = "LOGIN_SUCCESS"
	AuditActionLoginFailure       AuditAction = "LOGIN_FAILURE"
	AuditActionLogout             AuditAction = "LOGOUT"
	AuditActionPasswordReset      AuditAction = "PASSWORD_RESET"
	// Story 1.6, AC6: Authorization audit logging
	AuditActionAuthFailure        AuditAction = "AUTH_FAILURE"
	AuditActionForbiddenAccess    AuditAction = "FORBIDDEN_ACCESS"
	// Story 1.7: User creation audit logging
	AuditActionUserCreated        AuditAction = "USER_CREATED"
	// Story 1.9: Whitelist and self-registration audit logging
	AuditActionWhitelistDomainAdded   AuditAction = "WHITELIST_DOMAIN_ADDED"
	AuditActionWhitelistDomainUpdated AuditAction = "WHITELIST_DOMAIN_UPDATED"
	AuditActionWhitelistDomainDeleted AuditAction = "WHITELIST_DOMAIN_DELETED"
	AuditActionSelfRegistration       AuditAction = "SELF_REGISTRATION"
	AuditActionEmailVerified          AuditAction = "EMAIL_VERIFIED"
	// Story 1.10: User deactivation audit logging
	AuditActionUserDeactivated        AuditAction = "USER_DEACTIVATED"
	// Story 4.3: Stock adjustment audit logging (AC5: append-only audit trail)
	AuditActionStockAdjustment        AuditAction = "STOCK_ADJUSTMENT"
	// Story 4.6: Blocked sale attempt audit logging (AC6: regulatory compliance)
	AuditActionBlockedSaleAttempt     AuditAction = "BLOCKED_SALE_ATTEMPT"
	// Story 5.3: Report export audit logging (regulatory compliance)
	AuditActionExportReport            AuditAction = "EXPORT_REPORT"
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
	// Logs admin_user_id, product_id, product_sku, old_qty, new_qty, reason, timestamp
	// Append-only audit trail for Badan POM compliance (NFR-SEC-004, NFR-SEC-009)
	LogStockAdjustment(ctx context.Context, adminID uint, adminUsername string, productID uint, productSKU string, oldQty int64, newQty int64, reason string) error

	// LogBlockedSaleAttempt logs blocked sale attempts for expired products (Story 4.6, AC6)
	// Logs user_id, username, product_id, product_sku, product_name, expiry_date, reason, timestamp
	// Append-only audit trail for Badan POM regulatory compliance (NFR-SEC-004, NFR-SEC-009, NFR-SEC-011)
	LogBlockedSaleAttempt(ctx context.Context, userID uint, username string, productID uint, productSKU string, productName string, expiryDate string, reason string) error

	// LogReportExport logs report export events (Story 5.3, Task 4.7)
	// Logs user_id, username, report_type, format, date_range, timestamp
	// Append-only audit trail for Badan POM regulatory compliance (NFR-SEC-004, NFR-SEC-009)
	LogReportExport(ctx context.Context, userID uint, username string, reportType string, format string, dateRange string, outcome string) error
}

// auditService implements AuditService with in-memory logging for MVP
// TODO: Future story should implement persistent storage (database or log file)
type auditService struct {
	// For MVP, logs are written to stdout
	// Future: Add database repository or log file writer
}

// NewAuditService creates a new audit service instance
func NewAuditService() AuditService {
	return &auditService{}
}

// LogLoginAttempt logs login attempt to append-only audit trail (Story 1.5, AC7)
func (s *auditService) LogLoginAttempt(ctx context.Context, entry AuditLogEntry) error {
	// Set timestamp if not provided
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now()
	}

	// Log to stdout in structured format for MVP (Story 1.5, AC7, NFR-SEC-004)
	// In production, this should write to an append-only database table or log file
	// Format: AUDIT | timestamp | action | username | ip_address | outcome | reason
	slog.Info("AUDIT",
		"timestamp", entry.Timestamp.Format(time.RFC3339),
		"action", string(entry.Action),
		"username", entry.Username,
		"user_id", entry.UserID,
		"ip_address", entry.IPAddress,
		"outcome", entry.Outcome,
		"reason", entry.Reason,
	)

	// TODO: Future story - Add persistent storage (database or log file)
	// Per NFR-SEC-004: audit trail must be append-only (no delete/update)
	return nil
}

// LogAuthorizationFailure logs authorization failures to append-only audit trail
// Story 1.6, AC6: All authorization failures are logged with user_id, role, endpoint, reason
func (s *auditService) LogAuthorizationFailure(ctx context.Context, entry AuditLogEntry) error {
	// Set timestamp if not provided
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now()
	}

	// Log to stdout in structured format for MVP (Story 1.6, AC6, NFR-SEC-004)
	// Format: AUDIT | timestamp | action | username | role | endpoint | ip_address | outcome | reason
	slog.Info("AUDIT",
		"timestamp", entry.Timestamp.Format(time.RFC3339),
		"action", string(entry.Action),
		"username", entry.Username,
		"user_id", entry.UserID,
		"ip_address", entry.IPAddress,
		"outcome", entry.Outcome,
		"reason", entry.Reason,
	)

	// TODO: Future story - Add persistent storage (database or log file)
	// Per NFR-SEC-004: audit trail must be append-only (no delete/update)
	return nil
}

// LogAuditEntry is a helper function that can be called from services
// This is a placeholder for the actual logging implementation
func LogAuditEntry(entry AuditLogEntry) error {
	// For MVP: Return nil (no-op)
	// For production: Write to append-only storage
	return nil
}

// LogUserCreation logs user creation actions to append-only audit trail (Story 1.7, AC7)
func (s *auditService) LogUserCreation(ctx context.Context, adminID uint, createdUserID uint, adminUsername string, createdUsername string, ipAddress string) error {
	// Create audit log entry
	entry := AuditLogEntry{
		UserID:    &adminID,
		Username:  adminUsername,
		Action:    AuditActionUserCreated,
		IPAddress: ipAddress,
		Outcome:   "success",
		Reason:    fmt.Sprintf("Created user '%s' (ID: %d)", createdUsername, createdUserID),
		Timestamp: time.Now(),
	}

	// Log to stdout in structured format for MVP (Story 1.7, AC7, NFR-SEC-004)
	// Format: AUDIT | timestamp | action | admin_user_id | admin_username | created_user_id | created_username | ip_address | outcome
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

	// TODO: Future story - Add persistent storage (database or log file)
	// Per NFR-SEC-004: audit trail must be append-only (no delete/update)
	return nil
}

// LogWhitelistChange logs whitelist domain management actions to append-only audit trail (Story 1.9, AC8)
func (s *auditService) LogWhitelistChange(ctx context.Context, adminID uint, adminUsername string, domain string, action AuditAction, ipAddress string) error {
	// Create audit log entry
	entry := AuditLogEntry{
		UserID:    &adminID,
		Username:  adminUsername,
		Action:    action,
		IPAddress: ipAddress,
		Outcome:   "success",
		Reason:    fmt.Sprintf("Domain '%s'", domain),
		Timestamp: time.Now(),
	}

	// Log to stdout in structured format for MVP (Story 1.9, AC8, NFR-SEC-004)
	// Format: AUDIT | timestamp | action | admin_user_id | admin_username | domain | ip_address | outcome
	slog.Info("AUDIT",
		"timestamp", entry.Timestamp.Format(time.RFC3339),
		"action", string(entry.Action),
		"admin_user_id", adminID,
		"admin_username", adminUsername,
		"domain", domain,
		"ip_address", ipAddress,
		"outcome", entry.Outcome,
	)

	// TODO: Future story - Add persistent storage (database or log file)
	// Per NFR-SEC-004: audit trail must be append-only (no delete/update)
	return nil
}

// LogSelfRegistration logs staff self-registration actions to append-only audit trail (Story 1.9, AC8)
func (s *auditService) LogSelfRegistration(ctx context.Context, userID uint, email string, domain string, ipAddress string) error {
	// Create audit log entry
	entry := AuditLogEntry{
		UserID:    &userID,
		Username:  email, // Use email as username for self-registration
		Action:    AuditActionSelfRegistration,
		IPAddress: ipAddress,
		Outcome:   "success",
		Reason:    fmt.Sprintf("Self-registered from domain '%s'", domain),
		Timestamp: time.Now(),
	}

	// Log to stdout in structured format for MVP (Story 1.9, AC8, NFR-SEC-004)
	// Format: AUDIT | timestamp | action | user_id | email | domain | ip_address | outcome
	slog.Info("AUDIT",
		"timestamp", entry.Timestamp.Format(time.RFC3339),
		"action", string(entry.Action),
		"user_id", userID,
		"email", email,
		"domain", domain,
		"ip_address", ipAddress,
		"outcome", entry.Outcome,
	)

	// TODO: Future story - Add persistent storage (database or log file)
	// Per NFR-SEC-004: audit trail must be append-only (no delete/update)
	return nil
}

// LogEmailVerification logs email verification actions to append-only audit trail (Story 1.9, AC8)
func (s *auditService) LogEmailVerification(ctx context.Context, userID uint, email string, ipAddress string) error {
	// Create audit log entry
	entry := AuditLogEntry{
		UserID:    &userID,
		Username:  email, // Use email as username for email verification
		Action:    AuditActionEmailVerified,
		IPAddress: ipAddress,
		Outcome:   "success",
		Reason:    "Email verified and account activated",
		Timestamp: time.Now(),
	}

	// Log to stdout in structured format for MVP (Story 1.9, AC8, NFR-SEC-004)
	// Format: AUDIT | timestamp | action | user_id | email | ip_address | outcome
	slog.Info("AUDIT",
		"timestamp", entry.Timestamp.Format(time.RFC3339),
		"action", string(entry.Action),
		"user_id", userID,
		"email", email,
		"ip_address", ipAddress,
		"outcome", entry.Outcome,
	)

	// TODO: Future story - Add persistent storage (database or log file)
	// Per NFR-SEC-004: audit trail must be append-only (no delete/update)
	return nil
}

// LogUserDeactivation logs user deactivation actions to append-only audit trail (Story 1.10, AC5)
func (s *auditService) LogUserDeactivation(ctx context.Context, adminID uint, deactivatedUserID uint, adminUsername string, deactivatedUsername string, reason string, ipAddress string) error {
	// Create audit log entry
	entry := AuditLogEntry{
		UserID:    &adminID,
		Username:  adminUsername,
		Action:    AuditActionUserDeactivated,
		IPAddress: ipAddress,
		Outcome:   "success",
		Reason:    fmt.Sprintf("Deactivated user '%s' (ID: %d) - Reason: %s", deactivatedUsername, deactivatedUserID, reason),
		Timestamp: time.Now(),
	}

	// Log to stdout in structured format for MVP (Story 1.10, AC5, NFR-SEC-004)
	// Format: AUDIT | timestamp | action | admin_user_id | admin_username | deactivated_user_id | deactivated_username | reason | ip_address | outcome
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

	// TODO: Future story - Add persistent storage (database or log file)
	// Per NFR-SEC-004: audit trail must be append-only (no delete/update)
	return nil
}

// LogStockAdjustment logs manual stock adjustment actions to append-only audit trail (Story 4.3, AC5)
func (s *auditService) LogStockAdjustment(ctx context.Context, adminID uint, adminUsername string, productID uint, productSKU string, oldQty int64, newQty int64, reason string) error {
	// Create audit log entry
	entry := AuditLogEntry{
		UserID:    &adminID,
		Username:  adminUsername,
		Action:    AuditActionStockAdjustment,
		IPAddress: "", // IP address will be extracted from request context in production
		Outcome:   "success",
		Reason:    fmt.Sprintf("Adjusted stock for product '%s' (ID: %d): %d → %d - Reason: %s", productSKU, productID, oldQty, newQty, reason),
		Timestamp: time.Now(),
	}

	// Log to stdout in structured format for MVP (Story 4.3, AC5, NFR-SEC-004, NFR-SEC-009)
	// Format: AUDIT | timestamp | STOCK_ADJUSTMENT | admin_user_id | product_sku | old_qty | new_qty | reason
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

	// TODO: Future story - Add persistent storage (database or log file)
	// Per NFR-SEC-004: audit trail must be append-only (no delete/update)
	// Per NFR-SEC-009: 5-year minimum retention for Badan POM compliance
	return nil
}

// LogReportExport logs report export events for regulatory compliance
// Story 5.3, Task 4.7: Add export event logging to audit trail
func (s *auditService) LogReportExport(ctx context.Context, userID uint, username string, reportType string, format string, dateRange string, outcome string) error {
	// Create audit log entry
	entry := AuditLogEntry{
		UserID:    &userID,
		Username:  username,
		Action:    AuditActionExportReport,
		IPAddress: "", // IP address will be extracted from request context in production
		Outcome:   outcome,
		Reason:    fmt.Sprintf("Exported report: type=%s, format=%s, range=%s", reportType, format, dateRange),
		Timestamp: time.Now(),
	}

	// Log to stdout in structured format for MVP
	// Format: AUDIT | timestamp | EXPORT_REPORT | user_id | username | report_type | format | date_range | outcome
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

	// TODO: Future story - Add persistent storage (database or log file)
	return nil
}


// LogBlockedSaleAttempt logs blocked sale attempts for expired products to append-only audit trail
// Story 4.6, AC6: Audit trail logging for blocked sale attempts (regulatory compliance)
// Per NFR-SEC-004: audit trail must be append-only (no delete/update)
// Per NFR-SEC-009: 5-year minimum retention for Badan POM compliance
// Per NFR-SEC-011: Expiry date blocking is mandatory with audit logging
func (s *auditService) LogBlockedSaleAttempt(ctx context.Context, userID uint, username string, productID uint, productSKU string, productName string, expiryDate string, reason string) error {
	// Create audit log entry
	entry := AuditLogEntry{
		UserID:    &userID,
		Username:  username,
		Action:    AuditActionBlockedSaleAttempt,
		IPAddress: "", // IP address will be extracted from request context in production
		Outcome:   "blocked",
		Reason:    fmt.Sprintf("Blocked sale attempt for expired product '%s' (ID: %d, SKU: %s, Expiry: %s): %s", productName, productID, productSKU, expiryDate, reason),
		Timestamp: time.Now(),
	}

	// Log to stdout in structured format for MVP (Story 4.6, AC6, NFR-SEC-004, NFR-SEC-009, NFR-SEC-011)
	// Format: AUDIT | timestamp | BLOCKED_SALE_ATTEMPT | user_id | username | product_id | product_sku | product_name | expiry_date | reason | outcome
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

	// TODO: Future story - Add persistent storage (database or log file)
	// Per NFR-SEC-004: audit trail must be append-only (no delete/update)
	// Per NFR-SEC-009: 5-year minimum retention for Badan POM compliance
	return nil
}



