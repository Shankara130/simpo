package dto

import "time"

// BackupStatus represents the current status of a backup operation
type BackupStatus string

const (
	BackupStatusPending   BackupStatus = "pending"
	BackupStatusRunning   BackupStatus = "running"
	BackupStatusSuccess   BackupStatus = "success"
	BackupStatusFailed    BackupStatus = "failed"
	BackupStatusCorrupted BackupStatus = "corrupted"
)

// BackupInfo represents metadata about a backup file
// Story 6.3, AC2: Backup files with timestamp, size, status
type BackupInfo struct {
	Filename    string       `json:"filename"`
	Size        int64        `json:"size"`
	CreatedAt   time.Time    `json:"created_at"`
	Status      BackupStatus `json:"status"`
	Checksum    string       `json:"checksum,omitempty"`
	Duration    int64        `json:"duration_seconds,omitempty"`
	Description string       `json:"description,omitempty"`
}

// BackupListResponse represents the response for listing backups
// Story 6.3, AC5: GET /api/v1/admin/backups endpoint
type BackupListResponse struct {
	Backups       []BackupInfo `json:"backups"`
	RetentionDays int          `json:"retention_days"`
	TotalSize     int64        `json:"total_size"`
	LastBackup    *BackupInfo  `json:"last_backup,omitempty"`
}

// CreateBackupRequest represents a request to create a manual backup
// Story 6.3, AC5: POST /api/v1/admin/backups endpoint
type CreateBackupRequest struct {
	Description string `json:"description,omitempty" example:"Manual backup before system update"`
}

// CreateBackupResponse represents the response when initiating a backup
// Story 6.3, AC5: Returns 202 Accepted for async operation
type CreateBackupResponse struct {
	Status       string `json:"status" example:"started"`
	Filename     string `json:"filename" example:"simpo_20260527_143000.dump"`
	EstimatedTime string `json:"estimated_time" example:"2-5 min"`
	Message      string `json:"message,omitempty" example:"Backup operation started"`
}

// RestoreBackupRequest represents a request to restore from a backup
// Story 6.3, AC6: Restore from backup with confirmation
type RestoreBackupRequest struct {
	Confirmed bool   `json:"confirmed" binding:"required" example:"true"`
	Reason    string `json:"reason" binding:"required" example:"Data corruption incident"`
}

// RestoreBackupResponse represents the response when initiating a restore
// Story 6.3, AC6: Returns 202 Accepted for async operation
type RestoreBackupResponse struct {
	Status    string `json:"status" example:"started"`
	Filename  string `json:"filename" example:"simpo_20260527_020000.dump"`
	Message   string `json:"message" example:"Restore operation started. This may take several minutes."`
}

// BackupConfig represents backup configuration settings
// Story 6.3, AC8: Configurable backup schedule and retention
type BackupConfig struct {
	Schedule      string `json:"schedule" example:"0 2 * * *"` // Cron expression
	RetentionDays int    `json:"retention_days" example:"30"`
	StoragePath   string `json:"storage_path" example:"/backups"`
	Enabled       bool   `json:"enabled" example:"true"`
}

// BackupJobStatus represents the current status of the backup job
type BackupJobStatus struct {
	IsRunning      bool      `json:"is_running"`
	CurrentBackup  string    `json:"current_backup,omitempty"`
	LastBackup     time.Time `json:"last_backup,omitempty"`
	LastStatus     BackupStatus `json:"last_status"`
	NextBackup     time.Time `json:"next_backup,omitempty"`
	SuccessRate    float64   `json:"success_rate"` // Percentage
}

// BackupValidationError represents backup validation errors
type BackupValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// DeleteBackupRequest represents a request to delete a backup file
// Story 6.3, Task 4: Manual backup deletion with confirmation
type DeleteBackupRequest struct {
	Confirmed bool   `json:"confirmed" binding:"required" example:"true"`
	Reason    string `json:"reason" binding:"required" example:"Corrupted backup file"`
}
