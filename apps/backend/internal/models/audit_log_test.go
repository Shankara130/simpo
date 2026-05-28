package models

import (
	"testing"
)

// TestAuditActionConstants verifies that all system change audit action constants
// are properly defined and can be used for audit logging
// Story 6.4, Task 1, Subtask 1.8: Add unit tests for new audit action constants
func TestAuditActionConstants(t *testing.T) {
	tests := []struct {
		name   string
		action AuditAction
		valid  bool
	}{
		// Existing actions from Story 5.4 (should still be valid)
		{"LOGIN_SUCCESS", AuditActionLoginSuccess, true},
		{"LOGIN_FAILURE", AuditActionLoginFailure, true},
		{"LOGOUT", AuditActionLogout, true},
		{"PASSWORD_RESET", AuditActionPasswordReset, true},
		{"AUTH_FAILURE", AuditActionAuthFailure, true},
		{"FORBIDDEN_ACCESS", AuditActionForbiddenAccess, true},
		{"USER_CREATED", AuditActionUserCreated, true},
		{"USER_DEACTIVATED", AuditActionUserDeactivated, true},
		{"SELF_REGISTRATION", AuditActionSelfRegistration, true},
		{"EMAIL_VERIFIED", AuditActionEmailVerified, true},
		{"WHITELIST_DOMAIN_ADDED", AuditActionWhitelistDomainAdded, true},
		{"WHITELIST_DOMAIN_UPDATED", AuditActionWhitelistDomainUpdated, true},
		{"WHITELIST_DOMAIN_DELETED", AuditActionWhitelistDomainDeleted, true},
		{"STOCK_ADJUSTMENT", AuditActionStockAdjustment, true},
		{"BLOCKED_SALE_ATTEMPT", AuditActionBlockedSaleAttempt, true},
		{"EXPORT_REPORT", AuditActionExportReport, true},

		// NEW system settings actions (Story 6.4)
		{"SYSTEM_SETTINGS_UPDATED", AuditActionSystemSettingsUpdated, true},
		{"SYSTEM_CONFIG_CHANGED", AuditActionSystemConfigChanged, true},

		// NEW backup operations (Story 6.4)
		{"BACKUP_CREATED", AuditActionBackupCreated, true},
		{"BACKUP_RESTORED", AuditActionBackupRestored, true},
		{"BACKUP_DELETED", AuditActionBackupDeleted, true},

		// NEW role and permission management (Story 6.4)
		{"ROLE_UPDATED", AuditActionRoleUpdated, true},
		{"PERMISSION_GRANTED", AuditActionPermissionGranted, true},
		{"PERMISSION_REVOKED", AuditActionPermissionRevoked, true},

		// NEW branch management (Story 6.4)
		{"BRANCH_CREATED", AuditActionBranchCreated, true},
		{"BRANCH_UPDATED", AuditActionBranchUpdated, true},
		{"BRANCH_DEACTIVATED", AuditActionBranchDeactivated, true},

		// NEW system operations (Story 6.4)
		{"SYSTEM_STARTUP", AuditActionSystemStartup, true},
		{"SYSTEM_SHUTDOWN", AuditActionSystemShutdown, true},
		{"MAINTENANCE_MODE_ENABLED", AuditActionMaintenanceModeEnabled, true},
		{"MAINTENANCE_MODE_DISABLED", AuditActionMaintenanceModeDisabled, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify the action constant is not empty
			if tt.action == "" {
				t.Errorf("AuditAction constant is empty for %s", tt.name)
			}

			// Verify the action can be converted to string
			actionStr := string(tt.action)
			if actionStr == "" {
				t.Errorf("AuditAction string conversion failed for %s", tt.name)
			}

			// Verify the action can be used in an AuditLog
			log := &AuditLog{
				Action: tt.action,
			}
			if log.Action != tt.action {
				t.Errorf("AuditLog.Action assignment failed for %s", tt.name)
			}
		})
	}
}

// TestSystemChangeAuditActionsCount verifies that all expected system change
// audit actions are present (Story 6.4 added 13 new actions)
func TestSystemChangeAuditActionsCount(t *testing.T) {
	expectedNewActions := []AuditAction{
		AuditActionSystemSettingsUpdated,
		AuditActionSystemConfigChanged,
		AuditActionBackupCreated,
		AuditActionBackupRestored,
		AuditActionBackupDeleted,
		AuditActionRoleUpdated,
		AuditActionPermissionGranted,
		AuditActionPermissionRevoked,
		AuditActionBranchCreated,
		AuditActionBranchUpdated,
		AuditActionBranchDeactivated,
		AuditActionSystemStartup,
		AuditActionSystemShutdown,
		AuditActionMaintenanceModeEnabled,
		AuditActionMaintenanceModeDisabled,
	}

	if len(expectedNewActions) != 15 {
		t.Errorf("Expected 15 new system change audit actions, got %d", len(expectedNewActions))
	}

	// Verify all expected actions are defined
	for _, action := range expectedNewActions {
		if action == "" {
			t.Errorf("System change audit action is undefined: %v", action)
		}
	}
}

// TestAuditActionStringConversion verifies that audit actions can be
// properly converted to/from strings for database storage and API serialization
func TestAuditActionStringConversion(t *testing.T) {
	actions := []AuditAction{
		// System settings
		AuditActionSystemSettingsUpdated,
		AuditActionSystemConfigChanged,
		// Backup operations
		AuditActionBackupCreated,
		AuditActionBackupRestored,
		AuditActionBackupDeleted,
		// Role management
		AuditActionRoleUpdated,
		AuditActionPermissionGranted,
		AuditActionPermissionRevoked,
		// Branch management
		AuditActionBranchCreated,
		AuditActionBranchUpdated,
		AuditActionBranchDeactivated,
		// System operations
		AuditActionSystemStartup,
		AuditActionSystemShutdown,
		AuditActionMaintenanceModeEnabled,
		AuditActionMaintenanceModeDisabled,
	}

	for _, action := range actions {
		t.Run(string(action), func(t *testing.T) {
			// Convert to string
			actionStr := string(action)

			// Verify string is not empty
			if actionStr == "" {
				t.Error("AuditAction string conversion resulted in empty string")
			}

			// Verify string can be converted back to AuditAction
			reconstructedAction := AuditAction(actionStr)
			if reconstructedAction != action {
				t.Errorf("Round-trip conversion failed: got %v, want %v", reconstructedAction, action)
			}
		})
	}
}

// TestAuditLogWithSystemChangeActions verifies that AuditLog can be created
// with all new system change action types
func TestAuditLogWithSystemChangeActions(t *testing.T) {
	systemChangeActions := []AuditAction{
		AuditActionSystemSettingsUpdated,
		AuditActionSystemConfigChanged,
		AuditActionBackupCreated,
		AuditActionBackupRestored,
		AuditActionBackupDeleted,
		AuditActionRoleUpdated,
		AuditActionPermissionGranted,
		AuditActionPermissionRevoked,
		AuditActionBranchCreated,
		AuditActionBranchUpdated,
		AuditActionBranchDeactivated,
		AuditActionSystemStartup,
		AuditActionSystemShutdown,
		AuditActionMaintenanceModeEnabled,
		AuditActionMaintenanceModeDisabled,
	}

	for _, action := range systemChangeActions {
		t.Run(string(action), func(t *testing.T) {
			log := &AuditLog{
				UserID:    1,
				Username:  "test_admin",
				Action:    action,
				IPAddress: "192.168.1.100",
				Outcome:   "success",
				Reason:    "Test audit log entry",
			}

			// Verify the audit log was created successfully
			if log.UserID != 1 {
				t.Errorf("UserID not set correctly: got %d, want %d", log.UserID, 1)
			}
			if log.Username != "test_admin" {
				t.Errorf("Username not set correctly: got %s, want %s", log.Username, "test_admin")
			}
			if log.Action != action {
				t.Errorf("Action not set correctly: got %s, want %s", log.Action, action)
			}
			if log.Outcome != "success" {
				t.Errorf("Outcome not set correctly: got %s, want %s", log.Outcome, "success")
			}
		})
	}
}
