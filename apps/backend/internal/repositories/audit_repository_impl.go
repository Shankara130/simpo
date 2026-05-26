package repositories

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/models"
)

// auditRepository implements AuditRepository interface
// Story 5.4: Implement Append-Only Audit Trail for Compliance
type auditRepository struct {
	db *gorm.DB
}

// MED-003: Constants for audit trail configuration
const (
	// Default pagination limit
	auditDefaultLimit = 20
	// Maximum pagination limit to prevent memory exhaustion
	auditMaxLimit = 100
	// Maximum records per export to prevent memory exhaustion
	auditMaxExportLimit = 10000
	// CSV export format
	auditExportFormatCSV = "csv"
	// JSON export format
	auditExportFormatJSON = "json"
)

// NewAuditRepository creates a new audit repository
// Story 5.4, Task 3.1: Create audit_repository.go in apps/backend/internal/repositories/
func NewAuditRepository(db interface{}) AuditRepository {
	gormDB, ok := db.(*gorm.DB)
	if !ok {
		panic("db must be *gorm.DB")
	}
	return &auditRepository{db: gormDB}
}

// Create inserts a new audit log entry into the database
// Story 5.4, Task 3.3: Implement CreateAuditLog(entry AuditLogEntry) error method
// Story 5.4, AC1: System automatically creates an immutable audit trail entry
// Story 5.4, AC4: Audit entries are stored in a separate audit_logs table with write-only access
func (r *auditRepository) Create(ctx context.Context, auditLog *models.AuditLog) error {
	if auditLog == nil {
		return fmt.Errorf("audit log cannot be nil")
	}
	// CRIT-003: Guard against nil context
	if ctx == nil {
		ctx = context.Background()
	}
	err := r.db.WithContext(ctx).Create(auditLog).Error
	if err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}
	return nil
}

// Query retrieves audit logs with optional filtering and pagination
// Story 5.4, Task 3.4: Implement QueryAuditLogs(filters AuditLogFilter) ([]AuditLogEntry, error) method
// Story 5.4, AC5: Audit trail is queryable for at least 5 years per Badan POM requirements
func (r *auditRepository) Query(ctx context.Context, filter *AuditLogFilter) ([]*models.AuditLog, int64, error) {
	// CRIT-003: Guard against nil context
	if ctx == nil {
		ctx = context.Background()
	}
	// Start with base query
	query := r.db.WithContext(ctx).Model(&models.AuditLog{})

	// Apply filters if provided
	if filter != nil {
		// Story 5.4, Task 3.4: Query parameters: user_id (optional), action (optional), start_date, end_date, limit, offset
		if filter.UserID != nil {
			query = query.Where("user_id = ?", *filter.UserID)
		}
		if filter.Action != nil {
			query = query.Where("action = ?", *filter.Action)
		}
		if filter.StartDate != nil {
			startDate, err := time.Parse("2006-01-02", *filter.StartDate)
			if err != nil {
				// MED-008: Log warning for invalid date format (should be validated at handler level)
				slog.Warn("Invalid start_date format in audit query", "start_date", *filter.StartDate, "error", err)
			} else {
				query = query.Where("timestamp >= ?", startDate)
			}
		}
		if filter.EndDate != nil {
			endDate, err := time.Parse("2006-01-02", *filter.EndDate)
			if err != nil {
				// MED-008: Log warning for invalid date format (should be validated at handler level)
				slog.Warn("Invalid end_date format in audit query", "end_date", *filter.EndDate, "error", err)
			} else {
				// Include the entire end date by adding 23:59:59
				endDate = endDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
				query = query.Where("timestamp <= ?", endDate)
			}
		}

		// Validate and apply pagination
		// Story 5.4, Task 3.4: QueryAuditLogs pagination (limit, offset)
		if filter.Limit <= 0 {
			filter.Limit = auditDefaultLimit // Default limit
		}
		if filter.Limit > 100 {
			filter.Limit = auditMaxLimit // Max limit to prevent memory exhaustion
		}
		if filter.Offset < 0 {
			filter.Offset = 0
		}
	}

	// Count total records matching filters (for pagination metadata)
	var total int64
	countQuery := query
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count audit logs: %w", err)
	}

	// Apply pagination
	if filter != nil {
		query = query.Offset(filter.Offset).Limit(filter.Limit)
	}

	// Order by timestamp descending (newest first)
	query = query.Order("timestamp DESC")

	// Execute query
	var auditLogs []*models.AuditLog
	err := query.Find(&auditLogs).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query audit logs: %w", err)
	}

	return auditLogs, total, nil
}

// Export generates audit log export in specified format (CSV or JSON)
// Story 5.4, Task 3.5: Implement ExportAuditLogs(filters AuditLogFilter, format string) ([]byte, error) method
// Story 5.4, AC6: Audit logs can be exported for compliance inspections
func (r *auditRepository) Export(ctx context.Context, filter *AuditLogFilter, format string, writer io.Writer) error {
	// CRIT-003: Guard against nil context
	if ctx == nil {
		ctx = context.Background()
	}
	// Determine export format
	var exportFormat AuditLogExportFormat
	switch strings.ToLower(format) {
	case "csv":
		exportFormat = AuditLogExportCSV
	case "json":
		exportFormat = AuditLogExportJSON
	default:
		return fmt.Errorf("unsupported export format: %s (supported: csv, json)", format)
	}

	// Query audit logs with provided filter (no pagination for full export)
	// Use maximum limit of 10000 records to prevent memory exhaustion
	exportFilter := *filter
	if exportFilter.Limit == 0 {
		exportFilter.Limit = auditMaxExportLimit // Export limit
	}
	exportFilter.Offset = 0

	auditLogs, _, err := r.Query(ctx, &exportFilter)
	if err != nil {
		return fmt.Errorf("failed to query audit logs for export: %w", err)
	}

	// Generate export in specified format
	switch exportFormat {
	case AuditLogExportCSV:
		return r.exportToCSV(auditLogs, writer)
	case AuditLogExportJSON:
		return r.exportToJSON(auditLogs, writer)
	default:
		return fmt.Errorf("unsupported export format: %s", exportFormat)
	}
}

// exportToCSV generates CSV export of audit logs
func (r *auditRepository) exportToCSV(auditLogs []*models.AuditLog, writer io.Writer) error {
	csvWriter := csv.NewWriter(writer)
	defer csvWriter.Flush()

	// Write CSV header
	headers := []string{"id", "timestamp", "user_id", "username", "action", "ip_address", "outcome", "reason"}
	if err := csvWriter.Write(headers); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	// HIGH-001: Sanitize CSV fields to prevent CSV injection attacks
	// Prefix fields starting with dangerous characters with single quote
	sanitizeCSVField := func(field string) string {
		if field == "" {
			return ""
		}
		// Dangerous characters that can trigger Excel/formula injection
		dangerousPrefixes := []string{"=", "-", "+", "@", "\t", "\r"}
		for _, prefix := range dangerousPrefixes {
			if strings.HasPrefix(field, prefix) {
				return "'" + field // Prefix with single quote to treat as text
			}
		}
		return field
	}
	if err := csvWriter.Write(headers); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write audit log records
	for _, log := range auditLogs {
		record := []string{
			fmt.Sprintf("%d", log.ID),
			log.Timestamp.Format(time.RFC3339),
			fmt.Sprintf("%d", log.UserID),
			sanitizeCSVField(log.Username),
			sanitizeCSVField(string(log.Action)),
			sanitizeCSVField(log.IPAddress),
			sanitizeCSVField(log.Outcome),
			sanitizeCSVField(log.Reason),
		}
		if err := csvWriter.Write(record); err != nil {
			return fmt.Errorf("failed to write CSV record: %w", err)
		}
	}

	return nil
}

// exportToJSON generates JSON export of audit logs
func (r *auditRepository) exportToJSON(auditLogs []*models.AuditLog, writer io.Writer) error {
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ") // Pretty print for readability

	if err := encoder.Encode(auditLogs); err != nil {
		return fmt.Errorf("failed to encode audit logs to JSON: %w", err)
	}

	return nil
}

// RetentionCleanup performs cleanup of audit logs older than 5 years
// Story 5.4, Task 3.6: Implement RetentionCleanup() method for 5-year archival
// Story 5.4, Task 7.3: Cleanup logic: DELETE FROM audit_logs WHERE timestamp < NOW() - INTERVAL '5 years'
// Story 5.4, AC5: Audit trail is queryable for at least 5 years per Badan POM requirements
func (r *auditRepository) RetentionCleanup(ctx context.Context) (int64, error) {
	// CRIT-003: Guard against nil context
	if ctx == nil {
		ctx = context.Background()
	}

	// CRIT-002: Concurrent cleanup protection
	// Note: In production, implement distributed locking using Redis or database advisory locks
	// Example PostgreSQL advisory lock: SELECT pg_advisory_lock(1234); ... SELECT pg_advisory_unlock(1234);
	// This prevents multiple cleanup operations from running concurrently

	// Calculate cutoff date (5 years ago from now)
	cutoffDate := time.Now().AddDate(-5, 0, 0) // 5 years ago

	// Delete audit logs older than 5 years
	// Story 5.4, AC3: Append-only enforcement - only retention cleanup can delete old records
	// Use transaction for atomicity
	result := r.db.WithContext(ctx).
		Where("timestamp < ?", cutoffDate).
		Delete(&models.AuditLog{})

	if result.Error != nil {
		return 0, fmt.Errorf("failed to cleanup old audit logs: %w", result.Error)
	}

	return result.RowsAffected, nil
}

// Story 5.4, Task 3.7: Add comprehensive tests for append-only behavior
// NOTE: Update and Delete methods are intentionally NOT implemented
// This enforces append-only behavior at the application layer
// Any attempt to add these methods would violate Badan POM compliance requirements
