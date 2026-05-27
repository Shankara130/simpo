package dto

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBackupStatusValues(t *testing.T) {
	// AC: Backup status constants are properly defined
	tests := []struct {
		status   BackupStatus
		expected string
	}{
		{BackupStatusPending, "pending"},
		{BackupStatusRunning, "running"},
		{BackupStatusSuccess, "success"},
		{BackupStatusFailed, "failed"},
		{BackupStatusCorrupted, "corrupted"},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.status))
		})
	}
}

func TestBackupInfo(t *testing.T) {
	// AC2: Backup files with timestamp, size, status
	now := time.Now()
	backup := BackupInfo{
		Filename:    "simpo_20260527_020000.dump",
		Size:        1024000,
		CreatedAt:   now,
		Status:      BackupStatusSuccess,
		Checksum:    "a1b2c3d4e5f6",
		Duration:    120,
		Description: "Scheduled backup",
	}

	assert.Equal(t, "simpo_20260527_020000.dump", backup.Filename)
	assert.Equal(t, int64(1024000), backup.Size)
	assert.Equal(t, BackupStatusSuccess, backup.Status)
	assert.Equal(t, "a1b2c3d4e5f6", backup.Checksum)
	assert.Equal(t, int64(120), backup.Duration)
	assert.Equal(t, "Scheduled backup", backup.Description)
}

func TestBackupListResponse(t *testing.T) {
	// AC5: GET /api/v1/admin/backups response structure
	now := time.Now()
	backups := []BackupInfo{
		{
			Filename:  "simpo_20260527_020000.dump",
			Size:      1024000,
			CreatedAt: now,
			Status:    BackupStatusSuccess,
		},
		{
			Filename:  "simpo_20260526_020000.dump",
			Size:      1023000,
			CreatedAt: now.Add(-24 * time.Hour),
			Status:    BackupStatusSuccess,
		},
	}

	response := BackupListResponse{
		Backups:       backups,
		RetentionDays: 30,
		TotalSize:     2047000,
		LastBackup:    &backups[0],
	}

	assert.Len(t, response.Backups, 2)
	assert.Equal(t, 30, response.RetentionDays)
	assert.Equal(t, int64(2047000), response.TotalSize)
	assert.NotNil(t, response.LastBackup)
	assert.Equal(t, "simpo_20260527_020000.dump", response.LastBackup.Filename)
}

func TestCreateBackupRequest(t *testing.T) {
	// AC5: Manual backup request structure
	req := CreateBackupRequest{
		Description: "Manual backup before system update",
	}

	assert.Equal(t, "Manual backup before system update", req.Description)
}

func TestCreateBackupResponse(t *testing.T) {
	// AC5: Manual backup response returns 202 Accepted
	resp := CreateBackupResponse{
		Status:       "started",
		Filename:     "simpo_20260527_143000.dump",
		EstimatedTime: "2-5 min",
		Message:      "Backup operation started",
	}

	assert.Equal(t, "started", resp.Status)
	assert.Equal(t, "simpo_20260527_143000.dump", resp.Filename)
	assert.Equal(t, "2-5 min", resp.EstimatedTime)
	assert.Equal(t, "Backup operation started", resp.Message)
}

func TestRestoreBackupRequest(t *testing.T) {
	// AC6: Restore request requires confirmation and reason
	req := RestoreBackupRequest{
		Confirmed: true,
		Reason:    "Data corruption incident",
	}

	assert.True(t, req.Confirmed)
	assert.Equal(t, "Data corruption incident", req.Reason)
}

func TestRestoreBackupResponse(t *testing.T) {
	// AC6: Restore response returns 202 Accepted
	resp := RestoreBackupResponse{
		Status:   "started",
		Filename: "simpo_20260527_020000.dump",
		Message:  "Restore operation started. This may take several minutes.",
	}

	assert.Equal(t, "started", resp.Status)
	assert.Equal(t, "simpo_20260527_020000.dump", resp.Filename)
}

func TestBackupConfig(t *testing.T) {
	// AC8: Configurable backup schedule and retention
	config := BackupConfig{
		Schedule:      "0 2 * * *",
		RetentionDays: 30,
		StoragePath:   "/backups",
		Enabled:       true,
	}

	assert.Equal(t, "0 2 * * *", config.Schedule)
	assert.Equal(t, 30, config.RetentionDays)
	assert.Equal(t, "/backups", config.StoragePath)
	assert.True(t, config.Enabled)
}

func TestBackupJobStatus(t *testing.T) {
	// AC4: Backup job status tracking
	now := time.Now()
	status := BackupJobStatus{
		IsRunning:     true,
		CurrentBackup: "simpo_20260527_020000.dump",
		LastBackup:    now.Add(-24 * time.Hour),
		LastStatus:    BackupStatusSuccess,
		NextBackup:    now.Add(24 * time.Hour),
		SuccessRate:   95.5,
	}

	assert.True(t, status.IsRunning)
	assert.Equal(t, "simpo_20260527_020000.dump", status.CurrentBackup)
	assert.Equal(t, BackupStatusSuccess, status.LastStatus)
	assert.Equal(t, 95.5, status.SuccessRate)
}

func TestBackupValidationError(t *testing.T) {
	// Validation error structure for backup operations
	err := BackupValidationError{
		Field:   "filename",
		Message: "Invalid backup filename format",
	}

	assert.Equal(t, "filename", err.Field)
	assert.Equal(t, "Invalid backup filename format", err.Message)
}
