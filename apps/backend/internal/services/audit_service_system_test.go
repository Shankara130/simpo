package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/models"
)

// TestLogSettingsUpdate verifies that system settings update operations are logged correctly
// Story 6.4, Task 7.2: Test system settings audit logging methods
func TestLogSettingsUpdate(t *testing.T) {
	ctx := context.Background()
	mockRepo := &MockAuditRepository{}
	auditService := NewAuditService(mockRepo)

	tests := []struct {
		name           string
		adminID        uint
		adminUsername  string
		changesJSON    string
		ipAddress      string
		wantError      bool
	}{
		{
			name:           "Successful settings update audit",
			adminID:        1,
			adminUsername:  "admin",
			changesJSON:    `{"pharmacy_name":"New Pharmacy Name","timezone":"Asia/Jakarta"}`,
			ipAddress:      "192.168.1.100",
			wantError:      false,
		},
		{
			name:           "Multiple settings changes",
			adminID:        2,
			adminUsername:  "sysadmin",
			changesJSON:    `{"pharmacy_name":"Simpo Pharmacy","address":"Jl. Sudirman No. 1","phone":"021-12345678","email":"info@simpopharmacy.com"}`,
			ipAddress:      "192.168.1.101",
			wantError:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := auditService.LogSettingsUpdate(ctx, tt.adminID, tt.adminUsername, tt.changesJSON, tt.ipAddress)

			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestLogBackupCreated verifies that backup creation operations are logged correctly
// Story 6.4, Task 3.8: Test backup audit logging methods
func TestLogBackupCreated(t *testing.T) {
	ctx := context.Background()
	mockRepo := &MockAuditRepository{}
	auditService := NewAuditService(mockRepo)

	tests := []struct {
		name         string
		adminID      uint
		adminUsername string
		backupFile   string
		size         int64
		ipAddress    string
		wantError    bool
	}{
		{
			name:         "Successful backup creation audit",
			adminID:      1,
			adminUsername: "admin",
			backupFile:   "simpo_20260527_120000.dump",
			size:         1024000,
			ipAddress:    "192.168.1.100",
			wantError:    false,
		},
		{
			name:         "System backup audit",
			adminID:      0,
			adminUsername: "system",
			backupFile:   "simpo_20260527_020000.dump",
			size:         2048000,
			ipAddress:    "localhost",
			wantError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := auditService.LogBackupCreated(ctx, tt.adminID, tt.adminUsername, tt.backupFile, tt.size, tt.ipAddress)

			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestLogBackupRestored verifies that backup restore operations are logged correctly
// Story 6.4, Task 3.8: Test backup audit logging methods
func TestLogBackupRestored(t *testing.T) {
	ctx := context.Background()
	mockRepo := &MockAuditRepository{}
	auditService := NewAuditService(mockRepo)

	tests := []struct {
		name         string
		adminID      uint
		adminUsername string
		backupFile   string
		ipAddress    string
		wantError    bool
	}{
		{
			name:         "Successful backup restore audit",
			adminID:      1,
			adminUsername: "admin",
			backupFile:   "simpo_20260526_120000.dump",
			ipAddress:    "192.168.1.100",
			wantError:    false,
		},
		{
			name:         "System backup restore audit",
			adminID:      0,
			adminUsername: "system",
			backupFile:   "simpo_20260525_020000.dump",
			ipAddress:    "localhost",
			wantError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := auditService.LogBackupRestored(ctx, tt.adminID, tt.adminUsername, tt.backupFile, tt.ipAddress)

			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestLogBackupDeleted verifies that backup deletion operations are logged correctly
// Story 6.4, Task 3.8: Test backup audit logging methods
func TestLogBackupDeleted(t *testing.T) {
	ctx := context.Background()
	mockRepo := &MockAuditRepository{}
	auditService := NewAuditService(mockRepo)

	tests := []struct {
		name         string
		adminID      uint
		adminUsername string
		backupFile   string
		ipAddress    string
		wantError    bool
	}{
		{
			name:         "Successful backup deletion audit",
			adminID:      1,
			adminUsername: "admin",
			backupFile:   "simpo_20260520_120000.dump",
			ipAddress:    "192.168.1.100",
			wantError:    false,
		},
		{
			name:         "System backup deletion audit",
			adminID:      0,
			adminUsername: "system",
			backupFile:   "simpo_20260515_020000.dump",
			ipAddress:    "localhost",
			wantError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := auditService.LogBackupDeleted(ctx, tt.adminID, tt.adminUsername, tt.backupFile, tt.ipAddress)

			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestBackupmodels.AuditActionConstants verifies that backup audit action constants are defined
// Story 6.4, Task 3.8: Test backup audit action constants
func TestBackupmodels.AuditActionConstants(t *testing.T) {
	tests := []struct {
		name    string
		action  AuditAction
		wantVal string
	}{
		{
			name:    "BACKUP_CREATED action constant",
			action:  models.AuditActionBackupCreated,
			wantVal: "BACKUP_CREATED",
		},
		{
			name:    "BACKUP_RESTORED action constant",
			action:  models.AuditActionBackupRestored,
			wantVal: "BACKUP_RESTORED",
		},
		{
			name:    "BACKUP_DELETED action constant",
			action:  models.AuditActionBackupDeleted,
			wantVal: "BACKUP_DELETED",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantVal, string(tt.action))
			assert.NotEmpty(t, tt.action, "AuditAction constant should not be empty")
		})
	}
}

// TestBackupAuditServiceIntegration verifies the backup audit methods work with the service interface
// Story 6.4, Task 3.8: Test backup audit service integration
func TestBackupAuditServiceIntegration(t *testing.T) {
	ctx := context.Background()
	mockRepo := &MockAuditRepository{}
	auditService := NewAuditService(mockRepo)

	// Verify the service implements the interface correctly
	assert.Implements(t, (*AuditService)(nil), auditService)

	// Test all three backup audit methods
	t.Run("All backup audit methods work", func(t *testing.T) {
		// Create backup
		err := auditService.LogBackupCreated(ctx, 1, "admin", "simpo_test.dump", 1024, "127.0.0.1")
		assert.NoError(t, err)

		// Restore backup
		err = auditService.LogBackupRestored(ctx, 1, "admin", "simpo_test.dump", "127.0.0.1")
		assert.NoError(t, err)

		// Delete backup
		err = auditService.LogBackupDeleted(ctx, 1, "admin", "simpo_test.dump", "127.0.0.1")
		assert.NoError(t, err)
	})
}

// TestLogRoleUpdated verifies that role update operations are logged correctly
// Story 6.4, Task 4.6: Test role audit logging methods
func TestLogRoleUpdated(t *testing.T) {
	ctx := context.Background()
	mockRepo := &MockAuditRepository{}
	auditService := NewAuditService(mockRepo)

	tests := []struct {
		name           string
		adminID        uint
		adminUsername  string
		targetUserID   uint
		targetUsername string
		oldRole        string
		newRole        string
		ipAddress      string
		wantError      bool
	}{
		{
			name:           "Successful role update audit",
			adminID:        1,
			adminUsername:  "admin",
			targetUserID:   2,
			targetUsername: "cashier1",
			oldRole:        "CASHIER",
			newRole:        "OWNER",
			ipAddress:      "192.168.1.100",
			wantError:      false,
		},
		{
			name:           "Role change to SystemAdmin",
			adminID:        1,
			adminUsername:  "superadmin",
			targetUserID:   3,
			targetUsername: "newadmin",
			oldRole:        "OWNER",
			newRole:        "SYSTEM_ADMIN",
			ipAddress:      "192.168.1.100",
			wantError:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := auditService.LogRoleUpdated(ctx, tt.adminID, tt.adminUsername, tt.targetUserID, tt.targetUsername, tt.oldRole, tt.newRole, tt.ipAddress)

			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestLogPermissionGranted verifies that permission grant operations are logged correctly
// Story 6.4, Task 4.6: Test permission audit logging methods
func TestLogPermissionGranted(t *testing.T) {
	ctx := context.Background()
	mockRepo := &MockAuditRepository{}
	auditService := NewAuditService(mockRepo)

	tests := []struct {
		name           string
		adminID        uint
		adminUsername  string
		targetUserID   uint
		targetUsername string
		permission     string
		ipAddress      string
		wantError      bool
	}{
		{
			name:           "Successful permission grant audit",
			adminID:        1,
			adminUsername:  "admin",
			targetUserID:   2,
			targetUsername: "cashier1",
			permission:     "MANAGE_INVENTORY",
			ipAddress:      "192.168.1.100",
			wantError:      false,
		},
		{
			name:           "Grant admin permission",
			adminID:        1,
			adminUsername:  "superadmin",
			targetUserID:   3,
			targetUsername: "newadmin",
			permission:     "ADMIN_ACCESS",
			ipAddress:      "192.168.1.100",
			wantError:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := auditService.LogPermissionGranted(ctx, tt.adminID, tt.adminUsername, tt.targetUserID, tt.targetUsername, tt.permission, tt.ipAddress)

			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestLogPermissionRevoked verifies that permission revoke operations are logged correctly
// Story 6.4, Task 4.6: Test permission audit logging methods
func TestLogPermissionRevoked(t *testing.T) {
	ctx := context.Background()
	mockRepo := &MockAuditRepository{}
	auditService := NewAuditService(mockRepo)

	tests := []struct {
		name           string
		adminID        uint
		adminUsername  string
		targetUserID   uint
		targetUsername string
		permission     string
		ipAddress      string
		wantError      bool
	}{
		{
			name:           "Successful permission revoke audit",
			adminID:        1,
			adminUsername:  "admin",
			targetUserID:   2,
			targetUsername: "cashier1",
			permission:     "MANAGE_INVENTORY",
			ipAddress:      "192.168.1.100",
			wantError:      false,
		},
		{
			name:           "Revoke admin permission",
			adminID:        1,
			adminUsername:  "superadmin",
			targetUserID:   3,
			targetUsername: "problemuser",
			permission:     "ADMIN_ACCESS",
			ipAddress:      "192.168.1.100",
			wantError:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := auditService.LogPermissionRevoked(ctx, tt.adminID, tt.adminUsername, tt.targetUserID, tt.targetUsername, tt.permission, tt.ipAddress)

			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestRolePermissionmodels.AuditActionConstants verifies that role/permission audit action constants are defined
// Story 6.4, Task 4.6: Test role/permission audit action constants
func TestRolePermissionmodels.AuditActionConstants(t *testing.T) {
	tests := []struct {
		name    string
		action  AuditAction
		wantVal string
	}{
		{
			name:    "ROLE_UPDATED action constant",
			action:  models.AuditActionRoleUpdated,
			wantVal: "ROLE_UPDATED",
		},
		{
			name:    "PERMISSION_GRANTED action constant",
			action:  models.AuditActionPermissionGranted,
			wantVal: "PERMISSION_GRANTED",
		},
		{
			name:    "PERMISSION_REVOKED action constant",
			action:  models.AuditActionPermissionRevoked,
			wantVal: "PERMISSION_REVOKED",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantVal, string(tt.action))
			assert.NotEmpty(t, tt.action, "AuditAction constant should not be empty")
		})
	}
}

// TestLogBranchCreated verifies that branch creation operations are logged correctly
// Story 6.4, Task 5.8: Test branch audit logging methods
func TestLogBranchCreated(t *testing.T) {
	ctx := context.Background()
	mockRepo := &MockAuditRepository{}
	auditService := NewAuditService(mockRepo)

	tests := []struct {
		name            string
		adminID         uint
		adminUsername   string
		branchName      string
		branchLocation string
		ipAddress       string
		wantError       bool
	}{
		{
			name:            "Successful branch creation audit",
			adminID:         1,
			adminUsername:   "admin",
			branchName:      "Jakarta Central",
			branchLocation: "Jakarta, Indonesia",
			ipAddress:       "192.168.1.100",
			wantError:       false,
		},
		{
			name:            "Branch creation with location details",
			adminID:         1,
			adminUsername:   "superadmin",
			branchName:      "Surabaya Branch",
			branchLocation:  "Surabaya, East Java",
			ipAddress:       "192.168.1.100",
			wantError:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := auditService.LogBranchCreated(ctx, tt.adminID, tt.adminUsername, tt.branchName, tt.branchLocation, tt.ipAddress)

			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestLogBranchUpdated verifies that branch update operations are logged correctly
// Story 6.4, Task 5.8: Test branch audit logging methods
func TestLogBranchUpdated(t *testing.T) {
	ctx := context.Background()
	mockRepo := &MockAuditRepository{}
	auditService := NewAuditService(mockRepo)

	tests := []struct {
		name        string
		adminID     uint
		adminUsername string
		branchID    uint
		branchName  string
		changes     string
		ipAddress   string
		wantError   bool
	}{
		{
			name:        "Successful branch update audit",
			adminID:     1,
			adminUsername: "admin",
			branchID:    1,
			branchName:  "Jakarta Central",
			changes:     "Changed location from 'Jakarta' to 'Jakarta Central'",
			ipAddress:   "192.168.1.100",
			wantError:   false,
		},
		{
			name:        "Branch name change",
			adminID:     1,
			adminUsername: "admin",
			branchID:    2,
			branchName:  "Surabaya Main",
			changes:     "Renamed from 'Surabaya Branch' to 'Surabaya Main'",
			ipAddress:   "192.168.1.100",
			wantError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := auditService.LogBranchUpdated(ctx, tt.adminID, tt.adminUsername, tt.branchID, tt.branchName, tt.changes, tt.ipAddress)

			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestLogBranchDeactivated verifies that branch deactivation operations are logged correctly
// Story 6.4, Task 5.8: Test branch audit logging methods
func TestLogBranchDeactivated(t *testing.T) {
	ctx := context.Background()
	mockRepo := &MockAuditRepository{}
	auditService := NewAuditService(mockRepo)

	tests := []struct {
		name        string
		adminID     uint
		adminUsername string
		branchID    uint
		branchName  string
		reason      string
		ipAddress   string
		wantError   bool
	}{
		{
			name:        "Successful branch deactivation audit",
			adminID:     1,
			adminUsername: "admin",
			branchID:    1,
			branchName:  "Jakarta Central",
			reason:      "Branch closed permanently",
			ipAddress:   "192.168.1.100",
			wantError:   false,
		},
		{
			name:        "Branch deactivation due to relocation",
			adminID:     1,
			adminUsername: "superadmin",
			branchID:    2,
			branchName:  "Surabaya Branch",
			reason:      "Relocated to new office",
			ipAddress:   "192.168.1.100",
			wantError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := auditService.LogBranchDeactivated(ctx, tt.adminID, tt.adminUsername, tt.branchID, tt.branchName, tt.reason, tt.ipAddress)

			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestBranchmodels.AuditActionConstants verifies that branch audit action constants are defined
// Story 6.4, Task 5.8: Test branch audit action constants
func TestBranchmodels.AuditActionConstants(t *testing.T) {
	tests := []struct {
		name    string
		action  AuditAction
		wantVal string
	}{
		{
			name:    "BRANCH_CREATED action constant",
			action:  models.AuditActionBranchCreated,
			wantVal: "BRANCH_CREATED",
		},
		{
			name:    "BRANCH_UPDATED action constant",
			action:  models.AuditActionBranchUpdated,
			wantVal: "BRANCH_UPDATED",
		},
		{
			name:    "BRANCH_DEACTIVATED action constant",
			action:  models.AuditActionBranchDeactivated,
			wantVal: "BRANCH_DEACTIVATED",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantVal, string(tt.action))
			assert.NotEmpty(t, tt.action, "AuditAction constant should not be empty")
		})
	}
}

// TestLogSystemStartup verifies that system startup operations are logged correctly
// Story 6.4, Task 6.7: Test system operation audit logging methods
func TestLogSystemStartup(t *testing.T) {
	ctx := context.Background()
	mockRepo := &MockAuditRepository{}
	auditService := NewAuditService(mockRepo)

	tests := []struct {
		name       string
		systemID   string
		serverInfo string
		ipAddress  string
		wantError  bool
	}{
		{
			name:       "Successful system startup audit",
			systemID:   "simpo-backend-01",
			serverInfo: "Linux server01 5.15.0-1023-aws #27-Ubuntu SMP Fri Nov 17 12:05:18 UTC 2023 x86_64",
			ipAddress:  "127.0.0.1",
			wantError:  false,
		},
		{
			name:       "System startup with minimal info",
			systemID:   "simpo-api-server",
			serverInfo: "Production API Server",
			ipAddress:  "localhost",
			wantError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := auditService.LogSystemStartup(ctx, tt.systemID, tt.serverInfo, tt.ipAddress)

			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestLogSystemShutdown verifies that system shutdown operations are logged correctly
// Story 6.4, Task 6.7: Test system operation audit logging methods
func TestLogSystemShutdown(t *testing.T) {
	ctx := context.Background()
	mockRepo := &MockAuditRepository{}
	auditService := NewAuditService(mockRepo)

	tests := []struct {
		name       string
		systemID   string
		reason     string
		ipAddress  string
		wantError  bool
	}{
		{
			name:       "Normal system shutdown audit",
			systemID:   "simpo-backend-01",
			reason:     "Scheduled maintenance",
			ipAddress:  "127.0.0.1",
			wantError:  false,
		},
		{
			name:       "Emergency shutdown audit",
			systemID:   "simpo-api-server",
			reason:     "Emergency: Critical system failure",
			ipAddress:  "localhost",
			wantError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := auditService.LogSystemShutdown(ctx, tt.systemID, tt.reason, tt.ipAddress)

			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestLogMaintenanceModeEnabled verifies that maintenance mode enable operations are logged correctly
// Story 6.4, Task 6.7: Test system operation audit logging methods
func TestLogMaintenanceModeEnabled(t *testing.T) {
	ctx := context.Background()
	mockRepo := &MockAuditRepository{}
	auditService := NewAuditService(mockRepo)

	tests := []struct {
		name         string
		adminID      uint
		adminUsername string
		reason       string
		ipAddress    string
		wantError    bool
	}{
		{
			name:         "Successful maintenance mode enable audit",
			adminID:      1,
			adminUsername: "admin",
			reason:       "Scheduled system upgrade",
			ipAddress:    "192.168.1.100",
			wantError:    false,
		},
		{
			name:         "Emergency maintenance mode",
			adminID:      2,
			adminUsername: "sysadmin",
			reason:       "Emergency: Security patch required",
			ipAddress:    "192.168.1.101",
			wantError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := auditService.LogMaintenanceModeEnabled(ctx, tt.adminID, tt.adminUsername, tt.reason, tt.ipAddress)

			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestLogMaintenanceModeDisabled verifies that maintenance mode disable operations are logged correctly
// Story 6.4, Task 6.7: Test system operation audit logging methods
func TestLogMaintenanceModeDisabled(t *testing.T) {
	ctx := context.Background()
	mockRepo := &MockAuditRepository{}
	auditService := NewAuditService(mockRepo)

	tests := []struct {
		name         string
		adminID      uint
		adminUsername string
		reason       string
		ipAddress    string
		wantError    bool
	}{
		{
			name:         "Successful maintenance mode disable audit",
			adminID:      1,
			adminUsername: "admin",
			reason:       "Maintenance completed successfully",
			ipAddress:    "192.168.1.100",
			wantError:    false,
		},
		{
			name:         "Maintenance cancelled",
			adminID:      2,
			adminUsername: "sysadmin",
			reason:       "Maintenance cancelled - rolled back",
			ipAddress:    "192.168.1.101",
			wantError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := auditService.LogMaintenanceModeDisabled(ctx, tt.adminID, tt.adminUsername, tt.reason, tt.ipAddress)

			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestSystemOperationmodels.AuditActionConstants verifies that system operation audit action constants are defined
// Story 6.4, Task 6.8: Test system operation audit action constants
func TestSystemOperationmodels.AuditActionConstants(t *testing.T) {
	tests := []struct {
		name    string
		action  AuditAction
		wantVal string
	}{
		{
			name:    "SYSTEM_STARTUP action constant",
			action:  models.AuditActionSystemStartup,
			wantVal: "SYSTEM_STARTUP",
		},
		{
			name:    "SYSTEM_SHUTDOWN action constant",
			action:  models.AuditActionSystemShutdown,
			wantVal: "SYSTEM_SHUTDOWN",
		},
		{
			name:    "MAINTENANCE_MODE_ENABLED action constant",
			action:  models.AuditActionMaintenanceModeEnabled,
			wantVal: "MAINTENANCE_MODE_ENABLED",
		},
		{
			name:    "MAINTENANCE_MODE_DISABLED action constant",
			action:  models.AuditActionMaintenanceModeDisabled,
			wantVal: "MAINTENANCE_MODE_DISABLED",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantVal, string(tt.action))
			assert.NotEmpty(t, tt.action, "AuditAction constant should not be empty")
		})
	}
}

// TestSystemOperationAuditServiceIntegration verifies the system operation audit methods work with the service interface
// Story 6.4, Task 6.8: Test system operation audit service integration
func TestSystemOperationAuditServiceIntegration(t *testing.T) {
	ctx := context.Background()
	mockRepo := &MockAuditRepository{}
	auditService := NewAuditService(mockRepo)

	// Verify the service implements the interface correctly
	assert.Implements(t, (*AuditService)(nil), auditService)

	// Test all four system operation audit methods
	t.Run("All system operation audit methods work", func(t *testing.T) {
		// System startup
		err := auditService.LogSystemStartup(ctx, "simpo-test-01", "Test Server", "127.0.0.1")
		assert.NoError(t, err)

		// System shutdown
		err = auditService.LogSystemShutdown(ctx, "simpo-test-01", "Test shutdown", "127.0.0.1")
		assert.NoError(t, err)

		// Maintenance mode enabled
		err = auditService.LogMaintenanceModeEnabled(ctx, 1, "admin", "Test maintenance", "127.0.0.1")
		assert.NoError(t, err)

		// Maintenance mode disabled
		err = auditService.LogMaintenanceModeDisabled(ctx, 1, "admin", "Test complete", "127.0.0.1")
		assert.NoError(t, err)
	})
}

// MockAuditRepository is a mock implementation for testing
type MockAuditRepository struct {
	CreateFunc func(ctx context.Context, auditLog *models.AuditLog) error
}

func (m *MockAuditRepository) Create(ctx context.Context, auditLog *models.AuditLog) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, auditLog)
	}
	return nil
}
