package services

import (
	"context"
	"log/slog"
	"time"
)

// AuditAction represents the type of audit action
type AuditAction string

const (
	AuditActionLoginSuccess  AuditAction = "LOGIN_SUCCESS"
	AuditActionLoginFailure  AuditAction = "LOGIN_FAILURE"
	AuditActionLogout        AuditAction = "LOGOUT"
	AuditActionPasswordReset AuditAction = "PASSWORD_RESET"
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

// LogAuditEntry is a helper function that can be called from services
// This is a placeholder for the actual logging implementation
func LogAuditEntry(entry AuditLogEntry) error {
	// For MVP: Return nil (no-op)
	// For production: Write to append-only storage
	return nil
}
