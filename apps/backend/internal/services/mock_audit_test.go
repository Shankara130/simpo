package services

import (
	"context"
	"time"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/models"
)

// MockAuditService is a mock implementation of AuditService for testing
type MockAuditService struct {
	LogLoginAttemptFunc           func(ctx context.Context, entry AuditLogEntry) error
	LogAuthorizationFailureFunc    func(ctx context.Context, entry AuditLogEntry) error
	LogUserCreationFunc            func(ctx context.Context, adminID uint, createdUserID uint, adminUsername string, createdUsername string, ipAddress string) error
	LogWhitelistChangeFunc        func(ctx context.Context, adminID uint, adminUsername string, domain string, action models.AuditAction, ipAddress string) error
	LogSelfRegistrationFunc       func(ctx context.Context, userID uint, email string, domain string, ipAddress string) error
	LogEmailVerificationFunc       func(ctx context.Context, userID uint, email string, ipAddress string) error
	LogUserDeactivationFunc       func(ctx context.Context, adminID uint, deactivatedUserID uint, adminUsername string, deactivatedUsername string, reason string, ipAddress string) error
	LogStockAdjustmentFunc        func(ctx context.Context, adminID uint, adminUsername string, productID uint, productSKU string, oldQty int64, newQty int64, reason string, ipAddress string) error
	LogBlockedSaleAttemptFunc     func(ctx context.Context, userID uint, username string, productID uint, productSKU string, productName string, expiryDate string, reason string, ipAddress string) error
	LogReportExportFunc           func(ctx context.Context, userID uint, username string, reportType string, format string, dateRange string, outcome string, ipAddress string) error
	LogSettingsUpdateFunc         func(ctx context.Context, adminID uint, adminUsername string, changesJSON string, ipAddress string) error

	// Story 6.4: Backup audit methods
	LogBackupCreatedFunc          func(ctx context.Context, adminID uint, adminUsername string, backupFile string, size int64, ipAddress string) error
	LogBackupRestoredFunc         func(ctx context.Context, adminID uint, adminUsername string, backupFile string, ipAddress string) error
	LogBackupDeletedFunc          func(ctx context.Context, adminID uint, adminUsername string, backupFile string, ipAddress string) error

	// Story 6.4: Role and permission management audit methods
	LogRoleUpdatedFunc             func(ctx context.Context, adminID uint, adminUsername string, targetUserID uint, targetUsername string, oldRole string, newRole string, ipAddress string) error
	LogPermissionGrantedFunc       func(ctx context.Context, adminID uint, adminUsername string, targetUserID uint, targetUsername string, permission string, ipAddress string) error
	LogPermissionRevokedFunc       func(ctx context.Context, adminID uint, adminUsername string, targetUserID uint, targetUsername string, permission string, ipAddress string) error

	// Story 6.4: Branch management audit methods
	LogBranchCreatedFunc           func(ctx context.Context, adminID uint, adminUsername string, branchName string, branchLocation string, ipAddress string) error
	LogBranchUpdatedFunc           func(ctx context.Context, adminID uint, adminUsername string, branchID uint, branchName string, changes string, ipAddress string) error
	LogBranchDeactivatedFunc       func(ctx context.Context, adminID uint, adminUsername string, branchID uint, branchName string, reason string, ipAddress string) error

	// Story 6.4: System operation audit methods
	LogSystemStartupFunc              func(ctx context.Context, systemID string, serverInfo string, ipAddress string) error
	LogSystemShutdownFunc             func(ctx context.Context, systemID string, reason string, ipAddress string) error
	LogMaintenanceModeEnabledFunc     func(ctx context.Context, adminID uint, adminUsername string, reason string, ipAddress string) error
	LogMaintenanceModeDisabledFunc    func(ctx context.Context, adminID uint, adminUsername string, reason string, ipAddress string) error

	// Story 8.5: Conflict resolution audit method
	LogConflictResolutionFunc func(ctx context.Context, eventType string, transactionID string, originalError string, resolutionType string, resolvedBy string, resolvedAt time.Time, conflictDetails string, ipAddress string) error

	LogCount                      int // Track how many times logging was called
}

func (m *MockAuditService) LogLoginAttempt(ctx context.Context, entry AuditLogEntry) error {
	m.LogCount++
	if m.LogLoginAttemptFunc != nil {
		return m.LogLoginAttemptFunc(ctx, entry)
	}
	return nil
}

func (m *MockAuditService) LogAuthorizationFailure(ctx context.Context, entry AuditLogEntry) error {
	m.LogCount++
	if m.LogAuthorizationFailureFunc != nil {
		return m.LogAuthorizationFailureFunc(ctx, entry)
	}
	return nil
}

// LogUserCreation logs user creation actions (Story 1.7, AC7)
func (m *MockAuditService) LogUserCreation(ctx context.Context, adminID uint, createdUserID uint, adminUsername string, createdUsername string, ipAddress string) error {
	m.LogCount++
	if m.LogUserCreationFunc != nil {
		return m.LogUserCreationFunc(ctx, adminID, createdUserID, adminUsername, createdUsername, ipAddress)
	}
	return nil
}

// LogWhitelistChange logs whitelist domain management actions (Story 1.9, AC8)
func (m *MockAuditService) LogWhitelistChange(ctx context.Context, adminID uint, adminUsername string, domain string, action models.AuditAction, ipAddress string) error {
	m.LogCount++
	if m.LogWhitelistChangeFunc != nil {
		return m.LogWhitelistChangeFunc(ctx, adminID, adminUsername, domain, action, ipAddress)
	}
	return nil
}

// LogSelfRegistration logs staff self-registration actions (Story 1.9, AC8)
func (m *MockAuditService) LogSelfRegistration(ctx context.Context, userID uint, email string, domain string, ipAddress string) error {
	m.LogCount++
	if m.LogSelfRegistrationFunc != nil {
		return m.LogSelfRegistrationFunc(ctx, userID, email, domain, ipAddress)
	}
	return nil
}

// LogEmailVerification logs email verification actions (Story 1.9, AC8)
func (m *MockAuditService) LogEmailVerification(ctx context.Context, userID uint, email string, ipAddress string) error {
	m.LogCount++
	if m.LogEmailVerificationFunc != nil {
		return m.LogEmailVerificationFunc(ctx, userID, email, ipAddress)
	}
	return nil
}

// LogUserDeactivation logs user deactivation actions (Story 1.10, AC5)
func (m *MockAuditService) LogUserDeactivation(ctx context.Context, adminID uint, deactivatedUserID uint, adminUsername string, deactivatedUsername string, reason string, ipAddress string) error {
	m.LogCount++
	if m.LogUserDeactivationFunc != nil {
		return m.LogUserDeactivationFunc(ctx, adminID, deactivatedUserID, adminUsername, deactivatedUsername, reason, ipAddress)
	}
	return nil
}

// LogStockAdjustment logs manual stock adjustment actions (Story 4.3, AC5)
// Story 5.4, Task 4.5: Added ipAddress parameter
func (m *MockAuditService) LogStockAdjustment(ctx context.Context, adminID uint, adminUsername string, productID uint, productSKU string, oldQty int64, newQty int64, reason string, ipAddress string) error {
	m.LogCount++
	if m.LogStockAdjustmentFunc != nil {
		return m.LogStockAdjustmentFunc(ctx, adminID, adminUsername, productID, productSKU, oldQty, newQty, reason, ipAddress)
	}
	return nil
}

// LogBlockedSaleAttempt logs blocked sale attempts for expired products (Story 4.6, AC6)
// Story 5.4, Task 4.5: Added ipAddress parameter
func (m *MockAuditService) LogBlockedSaleAttempt(ctx context.Context, userID uint, username string, productID uint, productSKU string, productName string, expiryDate string, reason string, ipAddress string) error {
	m.LogCount++
	if m.LogBlockedSaleAttemptFunc != nil {
		return m.LogBlockedSaleAttemptFunc(ctx, userID, username, productID, productSKU, productName, expiryDate, reason, ipAddress)
	}
	return nil
}

// LogReportExport logs report export actions (Story 5.3, AC4 - Code review fix: CRITICAL-004)
// Story 5.4, Task 4.5: Added ipAddress parameter
func (m *MockAuditService) LogReportExport(ctx context.Context, userID uint, username string, reportType string, format string, dateRange string, outcome string, ipAddress string) error {
	m.LogCount++
	if m.LogReportExportFunc != nil {
		return m.LogReportExportFunc(ctx, userID, username, reportType, format, dateRange, outcome, ipAddress)
	}
	return nil
}

// LogSettingsUpdate logs system settings changes (Story 6.1, AC7)
func (m *MockAuditService) LogSettingsUpdate(ctx context.Context, adminID uint, adminUsername string, changesJSON string, ipAddress string) error {
	m.LogCount++
	if m.LogSettingsUpdateFunc != nil {
		return m.LogSettingsUpdateFunc(ctx, adminID, adminUsername, changesJSON, ipAddress)
	}
	return nil
}

// LogBackupCreated logs backup creation operations (Story 6.4)
func (m *MockAuditService) LogBackupCreated(ctx context.Context, adminID uint, adminUsername string, backupFile string, size int64, ipAddress string) error {
	m.LogCount++
	if m.LogBackupCreatedFunc != nil {
		return m.LogBackupCreatedFunc(ctx, adminID, adminUsername, backupFile, size, ipAddress)
	}
	return nil
}

// LogBackupRestored logs backup restore operations (Story 6.4)
func (m *MockAuditService) LogBackupRestored(ctx context.Context, adminID uint, adminUsername string, backupFile string, ipAddress string) error {
	m.LogCount++
	if m.LogBackupRestoredFunc != nil {
		return m.LogBackupRestoredFunc(ctx, adminID, adminUsername, backupFile, ipAddress)
	}
	return nil
}

// LogBackupDeleted logs backup deletion operations (Story 6.4)
func (m *MockAuditService) LogBackupDeleted(ctx context.Context, adminID uint, adminUsername string, backupFile string, ipAddress string) error {
	m.LogCount++
	if m.LogBackupDeletedFunc != nil {
		return m.LogBackupDeletedFunc(ctx, adminID, adminUsername, backupFile, ipAddress)
	}
	return nil
}

// LogRoleUpdated logs role changes (Story 6.4)
func (m *MockAuditService) LogRoleUpdated(ctx context.Context, adminID uint, adminUsername string, targetUserID uint, targetUsername string, oldRole string, newRole string, ipAddress string) error {
	m.LogCount++
	if m.LogRoleUpdatedFunc != nil {
		return m.LogRoleUpdatedFunc(ctx, adminID, adminUsername, targetUserID, targetUsername, oldRole, newRole, ipAddress)
	}
	return nil
}

// LogPermissionGranted logs permission grant operations (Story 6.4)
func (m *MockAuditService) LogPermissionGranted(ctx context.Context, adminID uint, adminUsername string, targetUserID uint, targetUsername string, permission string, ipAddress string) error {
	m.LogCount++
	if m.LogPermissionGrantedFunc != nil {
		return m.LogPermissionGrantedFunc(ctx, adminID, adminUsername, targetUserID, targetUsername, permission, ipAddress)
	}
	return nil
}

// LogPermissionRevoked logs permission revoke operations (Story 6.4)
func (m *MockAuditService) LogPermissionRevoked(ctx context.Context, adminID uint, adminUsername string, targetUserID uint, targetUsername string, permission string, ipAddress string) error {
	m.LogCount++
	if m.LogPermissionRevokedFunc != nil {
		return m.LogPermissionRevokedFunc(ctx, adminID, adminUsername, targetUserID, targetUsername, permission, ipAddress)
	}
	return nil
}

// LogBranchCreated logs branch creation operations (Story 6.4)
func (m *MockAuditService) LogBranchCreated(ctx context.Context, adminID uint, adminUsername string, branchName string, branchLocation string, ipAddress string) error {
	m.LogCount++
	if m.LogBranchCreatedFunc != nil {
		return m.LogBranchCreatedFunc(ctx, adminID, adminUsername, branchName, branchLocation, ipAddress)
	}
	return nil
}

// LogBranchUpdated logs branch update operations (Story 6.4)
func (m *MockAuditService) LogBranchUpdated(ctx context.Context, adminID uint, adminUsername string, branchID uint, branchName string, changes string, ipAddress string) error {
	m.LogCount++
	if m.LogBranchUpdatedFunc != nil {
		return m.LogBranchUpdatedFunc(ctx, adminID, adminUsername, branchID, branchName, changes, ipAddress)
	}
	return nil
}

// LogBranchDeactivated logs branch deactivation operations (Story 6.4)
func (m *MockAuditService) LogBranchDeactivated(ctx context.Context, adminID uint, adminUsername string, branchID uint, branchName string, reason string, ipAddress string) error {
	m.LogCount++
	if m.LogBranchDeactivatedFunc != nil {
		return m.LogBranchDeactivatedFunc(ctx, adminID, adminUsername, branchID, branchName, reason, ipAddress)
	}
	return nil
}

// LogSystemStartup logs system startup operations (Story 6.4)
func (m *MockAuditService) LogSystemStartup(ctx context.Context, systemID string, serverInfo string, ipAddress string) error {
	m.LogCount++
	if m.LogSystemStartupFunc != nil {
		return m.LogSystemStartupFunc(ctx, systemID, serverInfo, ipAddress)
	}
	return nil
}

// LogSystemShutdown logs system shutdown operations (Story 6.4)
func (m *MockAuditService) LogSystemShutdown(ctx context.Context, systemID string, reason string, ipAddress string) error {
	m.LogCount++
	if m.LogSystemShutdownFunc != nil {
		return m.LogSystemShutdownFunc(ctx, systemID, reason, ipAddress)
	}
	return nil
}

// LogMaintenanceModeEnabled logs maintenance mode enable operations (Story 6.4)
func (m *MockAuditService) LogMaintenanceModeEnabled(ctx context.Context, adminID uint, adminUsername string, reason string, ipAddress string) error {
	m.LogCount++
	if m.LogMaintenanceModeEnabledFunc != nil {
		return m.LogMaintenanceModeEnabledFunc(ctx, adminID, adminUsername, reason, ipAddress)
	}
	return nil
}

// LogMaintenanceModeDisabled logs maintenance mode disable operations (Story 6.4)
func (m *MockAuditService) LogMaintenanceModeDisabled(ctx context.Context, adminID uint, adminUsername string, reason string, ipAddress string) error {
	m.LogCount++
	if m.LogMaintenanceModeDisabledFunc != nil {
		return m.LogMaintenanceModeDisabledFunc(ctx, adminID, adminUsername, reason, ipAddress)
	}
	return nil
}

// Shutdown gracefully shuts down the audit service (Story 6.4, CRIT-001)
func (m *MockAuditService) Shutdown(ctx context.Context) error {
	return nil
}

func (m *MockAuditService) LogConflictResolution(ctx context.Context, eventType string, transactionID string, originalError string, resolutionType string, resolvedBy string, resolvedAt time.Time, conflictDetails string, ipAddress string) error {
	m.LogCount++
	if m.LogConflictResolutionFunc != nil {
		return m.LogConflictResolutionFunc(ctx, eventType, transactionID, originalError, resolutionType, resolvedBy, resolvedAt, conflictDetails, ipAddress)
	}
	return nil
}
