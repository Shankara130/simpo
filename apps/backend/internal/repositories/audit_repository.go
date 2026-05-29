package repositories

import (
	"context"
	"io"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/models"
)

// AuditRepository defines the interface for audit log data operations
// Story 5.4: Implement Append-Only Audit Trail for Compliance
// CRITICAL: This interface intentionally has NO Update or Delete methods
// Append-only behavior is enforced at interface level for Badan POM compliance
type AuditRepository interface {
	// Create inserts a new audit log entry into the database
	// Story 5.4, Task 3.3: Implement CreateAuditLog(entry AuditLogEntry) error method
	// This is the ONLY write operation allowed - no Update/Delete
	Create(ctx context.Context, auditLog *models.AuditLog) error

	// Query retrieves audit logs with optional filtering and pagination
	// Story 5.4, Task 3.4: Implement QueryAuditLogs(filters AuditLogFilter) ([]AuditLogEntry, error) method
	// Story 5.4, Task 3.5: Implement ExportAuditLogs(filters AuditLogFilter, format string) ([]byte, error) method
	Query(ctx context.Context, filter *AuditLogFilter) ([]*models.AuditLog, int64, error)

	// Export generates audit log export in specified format (CSV or JSON)
	// Story 5.4, Task 3.5: Export audit logs for compliance inspections
	Export(ctx context.Context, filter *AuditLogFilter, format string, writer io.Writer) error

	// RetentionCleanup performs cleanup of audit logs older than 5 years
	// Story 5.4, Task 3.6: Implement RetentionCleanup() method for 5-year archival
	// Story 5.4, Task 7: Manual trigger only for retention cleanup (SystemAdmin only)
	RetentionCleanup(ctx context.Context) (int64, error)
}

// AuditLogFilter defines filtering options for audit log queries
// Story 5.4, Task 3.4: Query parameters for audit log retrieval
type AuditLogFilter struct {
	UserID    *uint   // Filter by specific user (optional)
	Action    *string // Filter by audit action (optional)
	StartDate *string // Filter by start date (optional, format: YYYY-MM-DD)
	EndDate   *string // Filter by end date (optional, format: YYYY-MM-DD)
	Limit     int     // Pagination limit (default: 20, max: 100)
	Offset    int     // Pagination offset (default: 0)
}

// AuditLogExportFormat defines supported export formats
type AuditLogExportFormat string

const (
	// AuditLogExportCSV exports audit logs in CSV format
	AuditLogExportCSV AuditLogExportFormat = "csv"
	// AuditLogExportJSON exports audit logs in JSON format
	AuditLogExportJSON AuditLogExportFormat = "json"
)
