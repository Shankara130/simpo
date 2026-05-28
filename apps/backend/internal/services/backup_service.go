package services

import (
	"context"
	"io"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/dto"
)

// BackupService defines the interface for backup operations
// Story 6.3, Task 1: Service interface for backup domain with clear business method signatures
type BackupService interface {
	// CreateBackup creates a full PostgreSQL database backup
	// Story 6.3, AC1: System automatically creates full PostgreSQL database backup
	// Story 6.3, AC2: Backup files stored with timestamp in configured location
	// Story 6.3, AC7: Backup operations maintain database consistency
	// Story 6.4: Added adminID, adminUsername, ipAddress for audit logging
	CreateBackup(ctx context.Context, description string, adminID uint, adminUsername string, ipAddress string) (*dto.BackupInfo, error)

	// RestoreBackup restores database from a backup file
	// Story 6.3, AC6: System supports restoration from any backup in 30-day window
	// Story 6.4: Added adminID, adminUsername, ipAddress for audit logging
	RestoreBackup(ctx context.Context, filename string, reason string, adminID uint, adminUsername string, ipAddress string) error

	// ListBackups returns all available backups with metadata
	// Story 6.3, AC3: Backups retained for 30 days with automatic cleanup
	ListBackups(ctx context.Context) (*dto.BackupListResponse, error)

	// DeleteOldBackups removes backups older than retention period
	// Story 6.3, AC3: Automatic cleanup of backups older than 30 days
	DeleteOldBackups(ctx context.Context, retentionDays int) (int, error)

	// DeleteBackup removes a specific backup file
	// Story 6.3, Task 4: Manual deletion of specific backup file
	// Story 6.4: Added adminID, adminUsername, ipAddress for audit logging
	DeleteBackup(ctx context.Context, filename string, adminID uint, adminUsername string, ipAddress string) error

	// GetBackupFile returns a reader for the backup file
	// Story 6.3, AC5: Support for downloading backup files
	GetBackupFile(ctx context.Context, filename string) (io.ReadCloser, error)

	// GetBackupStatus returns current backup job status
	// Story 6.3, AC4: Backup success/failure logged in system health log
	GetBackupStatus(ctx context.Context) (*dto.BackupJobStatus, error)

	// ValidateBackup checks if a backup file is valid for restoration
	// Story 6.3, AC6: Validation before restore operation
	ValidateBackup(ctx context.Context, filename string) ([]dto.BackupValidationError, error)

	// StartScheduler begins the automated backup scheduler
	// Story 6.3, AC1: Automated daily backups at scheduled time
	StartScheduler(ctx context.Context) error

	// StopScheduler gracefully stops the backup scheduler
	// Story 6.3, Task 2: Graceful shutdown handling
	StopScheduler(ctx context.Context) error

	// GetConfig returns current backup configuration
	// Story 6.3, AC8: Configurable backup schedule and retention
	GetConfig(ctx context.Context) (*dto.BackupConfig, error)

	// UpdateConfig updates backup configuration
	// Story 6.3, AC8: Configurable via system settings
	UpdateConfig(ctx context.Context, config *dto.BackupConfig) error
}

// BackupProgress represents progress information for a running backup
type BackupProgress struct {
	StartTime   int64  `json:"start_time"`
	Elapsed     int64  `json:"elapsed_seconds"`
	CurrentSize int64  `json:"current_size_bytes"`
	Estimated   int64  `json:"estimated_size_bytes"`
	Percentage  float64 `json:"percentage"`
}

// BackupResult represents the result of a completed backup operation
type BackupResult struct {
	Success    bool             `json:"success"`
	Filename   string           `json:"filename"`
	Size       int64            `json:"size_bytes"`
	Checksum   string           `json:"checksum"`
	Duration   int64            `json:"duration_seconds"`
	Error      string           `json:"error,omitempty"`
	Metadata   *BackupMetadata  `json:"metadata,omitempty"`
}

// BackupMetadata represents additional metadata about a backup
type BackupMetadata struct {
	CreatedAt     int64  `json:"created_at"`
	DatabaseSize  int64  `json:"database_size_bytes"`
	TableCount    int    `json:"table_count"`
	SchemaVersion string `json:"schema_version"`
	Description   string `json:"description,omitempty"`
}

// RestoreProgress represents progress information for a running restore
type RestoreProgress struct {
	StartTime   int64  `json:"start_time"`
	Elapsed     int64  `json:"elapsed_seconds"`
	TableCount  int    `json:"total_tables"`
	CurrentTable int   `json:"current_table"`
	Percentage  float64 `json:"percentage"`
}
