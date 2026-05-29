package models

import (
	"time"

	"gorm.io/gorm"
)

// AuditAction represents the type of audit action
// Story 5.4: Audit action types for append-only audit trail
// Story 6.4: Extended with system change audit actions for Badan POM compliance
type AuditAction string

const (
	// Authentication actions
	AuditActionLoginSuccess    AuditAction = "LOGIN_SUCCESS"
	AuditActionLoginFailure    AuditAction = "LOGIN_FAILURE"
	AuditActionLogout          AuditAction = "LOGOUT"
	AuditActionPasswordReset   AuditAction = "PASSWORD_RESET"
	AuditActionAuthFailure     AuditAction = "AUTH_FAILURE"
	AuditActionForbiddenAccess AuditAction = "FORBIDDEN_ACCESS"

	// User management actions
	AuditActionUserCreated      AuditAction = "USER_CREATED"
	AuditActionUserDeactivated  AuditAction = "USER_DEACTIVATED"
	AuditActionSelfRegistration AuditAction = "SELF_REGISTRATION"
	AuditActionEmailVerified    AuditAction = "EMAIL_VERIFIED"

	// Whitelist management actions
	AuditActionWhitelistDomainAdded   AuditAction = "WHITELIST_DOMAIN_ADDED"
	AuditActionWhitelistDomainUpdated AuditAction = "WHITELIST_DOMAIN_UPDATED"
	AuditActionWhitelistDomainDeleted AuditAction = "WHITELIST_DOMAIN_DELETED"

	// Inventory actions
	AuditActionStockAdjustment    AuditAction = "STOCK_ADJUSTMENT"
	AuditActionBlockedSaleAttempt AuditAction = "BLOCKED_SALE_ATTEMPT"

	// Reporting actions
	AuditActionExportReport AuditAction = "EXPORT_REPORT"

	// System settings actions (Story 6.4)
	AuditActionSystemSettingsUpdated AuditAction = "SYSTEM_SETTINGS_UPDATED"
	AuditActionSystemConfigChanged   AuditAction = "SYSTEM_CONFIG_CHANGED"

	// Backup operations (Story 6.4)
	AuditActionBackupCreated  AuditAction = "BACKUP_CREATED"
	AuditActionBackupRestored AuditAction = "BACKUP_RESTORED"
	AuditActionBackupDeleted  AuditAction = "BACKUP_DELETED"

	// Role and permission management (Story 6.4)
	AuditActionRoleUpdated       AuditAction = "ROLE_UPDATED"
	AuditActionPermissionGranted AuditAction = "PERMISSION_GRANTED"
	AuditActionPermissionRevoked AuditAction = "PERMISSION_REVOKED"

	// Branch management (Story 6.4)
	AuditActionBranchCreated     AuditAction = "BRANCH_CREATED"
	AuditActionBranchUpdated     AuditAction = "BRANCH_UPDATED"
	AuditActionBranchDeactivated AuditAction = "BRANCH_DEACTIVATED"

	// System operations (Story 6.4)
	AuditActionSystemStartup           AuditAction = "SYSTEM_STARTUP"
	AuditActionSystemShutdown          AuditAction = "SYSTEM_SHUTDOWN"
	AuditActionMaintenanceModeEnabled  AuditAction = "MAINTENANCE_MODE_ENABLED"
	AuditActionMaintenanceModeDisabled AuditAction = "MAINTENANCE_MODE_DISABLED"

	// Conflict resolution actions (Story 8.5)
	AuditActionConflictResolutionAutomaticFailure AuditAction = "CONFLICT_RESOLUTION_AUTOMATIC_FAILURE"
	AuditActionConflictResolutionManualOverride   AuditAction = "CONFLICT_RESOLUTION_MANUAL_OVERRIDE"
)

// AuditLog represents an append-only audit log entry for Badan POM compliance
// Story 5.4: Implement Append-Only Audit Trail for Compliance
// Per NFR-SEC-004: Append-only audit trail with user identification, timestamp, and reason
// Per NFR-SEC-009: 5-year minimum data retention for Badan POM compliance
//
// IMPORTANT: This model intentionally does NOT have UpdatedAt field
// Audit logs are immutable once created - append-only behavior
type AuditLog struct {
	ID        uint        `gorm:"primaryKey" json:"id"`
	UserID    uint        `gorm:"not null; index" json:"user_id"`
	Username  string      `gorm:"not null; size:255" json:"username"`
	Action    AuditAction `gorm:"not null; size:100; index" json:"action"`
	IPAddress string      `gorm:"size:45" json:"ip_address,omitempty"`
	Outcome   string      `gorm:"not null; size:50" json:"outcome"`
	Reason    string      `gorm:"type:text" json:"reason,omitempty"`
	Timestamp time.Time   `gorm:"not null; default:now(); index" json:"timestamp"`
	// Note: No UpdatedAt field - audit entries are immutable
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// TableName specifies the database table name for AuditLog
func (AuditLog) TableName() string {
	return "audit_logs"
}

// BeforeCreate GORM hook - ensure immutability constraints
func (a *AuditLog) BeforeCreate(tx *gorm.DB) error {
	// Validate required fields for Badan POM compliance
	if a.Username == "" {
		return tx.AddError(ErrInvalidAuditLog{Field: "username", Message: "username is required"})
	}
	if a.Action == "" {
		return tx.AddError(ErrInvalidAuditLog{Field: "action", Message: "action is required"})
	}
	if a.Outcome == "" {
		return tx.AddError(ErrInvalidAuditLog{Field: "outcome", Message: "outcome is required"})
	}
	// Code review fix: CRIT-011 - Validate IP address length (max 45 chars for IPv6)
	if len(a.IPAddress) > 45 {
		return tx.AddError(ErrInvalidAuditLog{Field: "ip_address", Message: "ip_address exceeds maximum length of 45 characters"})
	}
	return nil
}

// ErrInvalidAuditLog is returned when audit log validation fails
type ErrInvalidAuditLog struct {
	Field   string
	Message string
}

func (e ErrInvalidAuditLog) Error() string {
	return e.Field + ": " + e.Message
}

// ErrAuditLogUpdateNotAllowed is returned when attempting to update an audit log
// Story 5.4, AC3: Append-only enforcement - no modifications allowed
type ErrAuditLogUpdateNotAllowed struct {
	AuditLogID uint
}

func (e ErrAuditLogUpdateNotAllowed) Error() string {
	return "audit log update not allowed: audit logs are append-only"
}

// ErrAuditLogDeleteNotAllowed is returned when attempting to delete an audit log
// Story 5.4, AC3: Append-only enforcement - no deletions allowed
type ErrAuditLogDeleteNotAllowed struct {
	AuditLogID uint
}

func (e ErrAuditLogDeleteNotAllowed) Error() string {
	return "audit log delete not allowed: audit logs are append-only"
}
