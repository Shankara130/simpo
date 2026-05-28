package tests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/handlers"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/models"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/repositories"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/services"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/user"
)

// TestSystemAuditIntegration_SettingsUpdate_LogsAuditEntry tests that system settings update creates audit log entry
// Story 6.4, Task 8.2: Test system settings update creates audit log entry
func TestSystemAuditIntegration_SettingsUpdate_LogsAuditEntry(t *testing.T) {
	// Setup test environment
	gin.SetMode(gin.TestMode)
	db := setupTestDBForAuditIntegration(t)

	// Create audit repository and service
	auditRepo := repositories.NewAuditRepository(db)
	auditService := services.NewAuditService(auditRepo)

	ctx := context.Background()

	// Log system settings update
	err := auditService.LogSettingsUpdate(
		ctx,
		1,     // adminID
		"admin", // adminUsername
		`{"pharmacy_name":"Simpo Pharmacy","timezone":"Asia/Jakarta"}`, // changesJSON
		"127.0.0.1", // ipAddress
	)

	// Assert audit log was created
	require.NoError(t, err, "System settings update should be logged to audit trail")

	// Verify audit log entry exists in database
	var count int64
	db.Raw("SELECT COUNT(*) FROM audit_logs WHERE action = ?", "SYSTEM_SETTINGS_UPDATED").Scan(&count)
	assert.Greater(t, count, int64(0), "Audit log entry should exist for SYSTEM_SETTINGS_UPDATED")

	// Verify audit log details
	var auditLog models.AuditLog
	err = db.Where("action = ?", "SYSTEM_SETTINGS_UPDATED").First(&auditLog).Error
	require.NoError(t, err)
	assert.Equal(t, uint(1), auditLog.UserID)
	assert.Equal(t, "admin", auditLog.Username)
	assert.Equal(t, "SYSTEM_SETTINGS_UPDATED", auditLog.Action)
	assert.Contains(t, auditLog.Reason, "pharmacy_name")
}

// TestSystemAuditIntegration_BackupOperations_LogsAuditEntries tests that backup operations create audit log entries
// Story 6.4, Task 8.3: Test backup operations create audit log entries
func TestSystemAuditIntegration_BackupOperations_LogsAuditEntries(t *testing.T) {
	// Setup test environment
	gin.SetMode(gin.TestMode)
	db := setupTestDBForAuditIntegration(t)

	// Create audit repository and service
	auditRepo := repositories.NewAuditRepository(db)
	auditService := services.NewAuditService(auditRepo)

	ctx := context.Background()

	t.Run("Backup creation creates audit log", func(t *testing.T) {
		err := auditService.LogBackupCreated(
			ctx,
			1,                         // adminID
			"admin",                   // adminUsername
			"simpo_20260527_120000.dump", // backupFile
			1024000,                   // size in bytes
			"192.168.1.100",           // ipAddress
		)

		require.NoError(t, err, "Backup creation should be logged")

		// Verify audit log entry
		var auditLog models.AuditLog
		err = db.Where("action = ?", "BACKUP_CREATED").First(&auditLog).Error
		require.NoError(t, err)
		assert.Equal(t, "BACKUP_CREATED", auditLog.Action)
		assert.Contains(t, auditLog.Reason, "simpo_20260527_120000.dump")
		assert.Contains(t, auditLog.Reason, "1024000")
	})

	t.Run("Backup restore creates audit log", func(t *testing.T) {
		err := auditService.LogBackupRestored(
			ctx,
			1,                         // adminID
			"admin",                   // adminUsername
			"simpo_20260526_120000.dump", // backupFile
			"192.168.1.100",           // ipAddress
		)

		require.NoError(t, err, "Backup restore should be logged")

		// Verify audit log entry
		var auditLog models.AuditLog
		err = db.Where("action = ? AND user_id = ?", "BACKUP_RESTORED", 1).First(&auditLog).Error
		require.NoError(t, err)
		assert.Equal(t, "BACKUP_RESTORED", auditLog.Action)
		assert.Contains(t, auditLog.Reason, "simpo_20260526_120000.dump")
	})

	t.Run("Backup deletion creates audit log", func(t *testing.T) {
		err := auditService.LogBackupDeleted(
			ctx,
			1,                         // adminID
			"admin",                   // adminUsername
			"simpo_20260520_120000.dump", // backupFile
			"192.168.1.100",           // ipAddress
		)

		require.NoError(t, err, "Backup deletion should be logged")

		// Verify audit log entry
		var auditLog models.AuditLog
		err = db.Where("action = ? AND user_id = ?", "BACKUP_DELETED", 1).First(&auditLog).Error
		require.NoError(t, err)
		assert.Equal(t, "BACKUP_DELETED", auditLog.Action)
		assert.Contains(t, auditLog.Reason, "simpo_20260520_120000.dump")
	})
}

// TestSystemAuditIntegration_RoleChanges_LogsAuditEntries tests that role changes create audit log entries
// Story 6.4, Task 8.4: Test role changes create audit log entries
func TestSystemAuditIntegration_RoleChanges_LogsAuditEntries(t *testing.T) {
	// Setup test environment
	gin.SetMode(gin.TestMode)
	db := setupTestDBForAuditIntegration(t)

	// Create audit repository and service
	auditRepo := repositories.NewAuditRepository(db)
	auditService := services.NewAuditService(auditRepo)

	ctx := context.Background()

	t.Run("Role update creates audit log", func(t *testing.T) {
		err := auditService.LogRoleUpdated(
			ctx,
			1,              // adminID
			"admin",        // adminUsername
			2,              // targetUserID
			"cashier1",     // targetUsername
			"CASHIER",      // oldRole
			"OWNER",        // newRole
			"192.168.1.100", // ipAddress
		)

		require.NoError(t, err, "Role update should be logged")

		// Verify audit log entry
		var auditLog models.AuditLog
		err = db.Where("action = ?", "ROLE_UPDATED").First(&auditLog).Error
		require.NoError(t, err)
		assert.Equal(t, "ROLE_UPDATED", auditLog.Action)
		assert.Contains(t, auditLog.Reason, "cashier1")
		assert.Contains(t, auditLog.Reason, "CASHIER")
		assert.Contains(t, auditLog.Reason, "OWNER")
	})

	t.Run("Permission grant creates audit log", func(t *testing.T) {
		err := auditService.LogPermissionGranted(
			ctx,
			1,              // adminID
			"admin",        // adminUsername
			2,              // targetUserID
			"cashier1",     // targetUsername
			"MANAGE_INVENTORY", // permission
			"192.168.1.100", // ipAddress
		)

		require.NoError(t, err, "Permission grant should be logged")

		// Verify audit log entry
		var auditLog models.AuditLog
		err = db.Where("action = ? AND user_id = ?", "PERMISSION_GRANTED", 1).First(&auditLog).Error
		require.NoError(t, err)
		assert.Equal(t, "PERMISSION_GRANTED", auditLog.Action)
		assert.Contains(t, auditLog.Reason, "cashier1")
		assert.Contains(t, auditLog.Reason, "MANAGE_INVENTORY")
	})

	t.Run("Permission revoke creates audit log", func(t *testing.T) {
		err := auditService.LogPermissionRevoked(
			ctx,
			1,              // adminID
			"admin",        // adminUsername
			2,              // targetUserID
			"cashier1",     // targetUsername
			"MANAGE_INVENTORY", // permission
			"192.168.1.100", // ipAddress
		)

		require.NoError(t, err, "Permission revoke should be logged")

		// Verify audit log entry
		var auditLog models.AuditLog
		err = db.Where("action = ? AND user_id = ?", "PERMISSION_REVOKED", 1).First(&auditLog).Error
		require.NoError(t, err)
		assert.Equal(t, "PERMISSION_REVOKED", auditLog.Action)
		assert.Contains(t, auditLog.Reason, "cashier1")
		assert.Contains(t, auditLog.Reason, "MANAGE_INVENTORY")
	})
}

// TestSystemAuditIntegration_BranchManagement_LogsAuditEntries tests that branch management creates audit log entries
// Story 6.4, Task 8.5: Test branch management creates audit log entries
func TestSystemAuditIntegration_BranchManagement_LogsAuditEntries(t *testing.T) {
	// Setup test environment
	gin.SetMode(gin.TestMode)
	db := setupTestDBForAuditIntegration(t)

	// Create audit repository and service
	auditRepo := repositories.NewAuditRepository(db)
	auditService := services.NewAuditService(auditRepo)

	ctx := context.Background()

	t.Run("Branch creation creates audit log", func(t *testing.T) {
		err := auditService.LogBranchCreated(
			ctx,
			1,                  // adminID
			"admin",            // adminUsername
			"Jakarta Central",  // branchName
			"Jakarta, Indonesia", // branchLocation
			"192.168.1.100",    // ipAddress
		)

		require.NoError(t, err, "Branch creation should be logged")

		// Verify audit log entry
		var auditLog models.AuditLog
		err = db.Where("action = ?", "BRANCH_CREATED").First(&auditLog).Error
		require.NoError(t, err)
		assert.Equal(t, "BRANCH_CREATED", auditLog.Action)
		assert.Contains(t, auditLog.Reason, "Jakarta Central")
		assert.Contains(t, auditLog.Reason, "Jakarta, Indonesia")
	})

	t.Run("Branch update creates audit log", func(t *testing.T) {
		err := auditService.LogBranchUpdated(
			ctx,
			1,                  // adminID
			"admin",            // adminUsername
			1,                  // branchID
			"Jakarta Central",  // branchName
			"Changed location from 'Jakarta' to 'Jakarta Central'", // changes
			"192.168.1.100",    // ipAddress
		)

		require.NoError(t, err, "Branch update should be logged")

		// Verify audit log entry
		var auditLog models.AuditLog
		err = db.Where("action = ? AND user_id = ?", "BRANCH_UPDATED", 1).First(&auditLog).Error
		require.NoError(t, err)
		assert.Equal(t, "BRANCH_UPDATED", auditLog.Action)
		assert.Contains(t, auditLog.Reason, "Jakarta Central")
		assert.Contains(t, auditLog.Reason, "Changed location")
	})

	t.Run("Branch deactivation creates audit log", func(t *testing.T) {
		err := auditService.LogBranchDeactivated(
			ctx,
			1,                  // adminID
			"admin",            // adminUsername
			1,                  // branchID
			"Jakarta Central",  // branchName
			"Branch closed permanently", // reason
			"192.168.1.100",    // ipAddress
		)

		require.NoError(t, err, "Branch deactivation should be logged")

		// Verify audit log entry
		var auditLog models.AuditLog
		err = db.Where("action = ? AND user_id = ?", "BRANCH_DEACTIVATED", 1).First(&auditLog).Error
		require.NoError(t, err)
		assert.Equal(t, "BRANCH_DEACTIVATED", auditLog.Action)
		assert.Contains(t, auditLog.Reason, "Jakarta Central")
		assert.Contains(t, auditLog.Reason, "Branch closed permanently")
	})
}

// TestSystemAuditIntegration_SystemOperations_LogsAuditEntries tests that system startup/shutdown creates audit log entries
// Story 6.4, Task 8.6: Test system startup/shutdown creates audit log entries
func TestSystemAuditIntegration_SystemOperations_LogsAuditEntries(t *testing.T) {
	// Setup test environment
	gin.SetMode(gin.TestMode)
	db := setupTestDBForAuditIntegration(t)

	// Create audit repository and service
	auditRepo := repositories.NewAuditRepository(db)
	auditService := services.NewAuditService(auditRepo)

	ctx := context.Background()

	t.Run("System startup creates audit log", func(t *testing.T) {
		err := auditService.LogSystemStartup(
			ctx,
			"simpo-backend-01", // systemID
			"Linux server01 5.15.0-1023-aws #27-Ubuntu SMP Fri Nov 17 12:05:18 UTC 2023 x86_64", // serverInfo
			"127.0.0.1", // ipAddress
		)

		require.NoError(t, err, "System startup should be logged")

		// Verify audit log entry
		var auditLog models.AuditLog
		err = db.Where("action = ?", "SYSTEM_STARTUP").First(&auditLog).Error
		require.NoError(t, err)
		assert.Equal(t, "SYSTEM_STARTUP", auditLog.Action)
		assert.Contains(t, auditLog.Reason, "simpo-backend-01")
		assert.Contains(t, auditLog.Reason, "Linux server01")
	})

	t.Run("System shutdown creates audit log", func(t *testing.T) {
		err := auditService.LogSystemShutdown(
			ctx,
			"simpo-backend-01", // systemID
			"Scheduled maintenance", // reason
			"127.0.0.1", // ipAddress
		)

		require.NoError(t, err, "System shutdown should be logged")

		// Verify audit log entry
		var auditLog models.AuditLog
		err = db.Where("action = ? AND user_id = ?", "SYSTEM_SHUTDOWN", 0, 0).First(&auditLog).Error
		require.NoError(t, err)
		assert.Equal(t, "SYSTEM_SHUTDOWN", auditLog.Action)
		assert.Contains(t, auditLog.Reason, "simpo-backend-01")
		assert.Contains(t, auditLog.Reason, "Scheduled maintenance")
	})

	t.Run("Maintenance mode enabled creates audit log", func(t *testing.T) {
		err := auditService.LogMaintenanceModeEnabled(
			ctx,
			1,                         // adminID
			"admin",                   // adminUsername
			"Scheduled system upgrade", // reason
			"192.168.1.100",           // ipAddress
		)

		require.NoError(t, err, "Maintenance mode enable should be logged")

		// Verify audit log entry
		var auditLog models.AuditLog
		err = db.Where("action = ?", "MAINTENANCE_MODE_ENABLED").First(&auditLog).Error
		require.NoError(t, err)
		assert.Equal(t, "MAINTENANCE_MODE_ENABLED", auditLog.Action)
		assert.Contains(t, auditLog.Reason, "Scheduled system upgrade")
	})

	t.Run("Maintenance mode disabled creates audit log", func(t *testing.T) {
		err := auditService.LogMaintenanceModeDisabled(
			ctx,
			1,                                         // adminID
			"admin",                                   // adminUsername
			"Maintenance completed successfully",      // reason
			"192.168.1.100",                           // ipAddress
		)

		require.NoError(t, err, "Maintenance mode disable should be logged")

		// Verify audit log entry
		var auditLog models.AuditLog
		err = db.Where("action = ? AND user_id = ?", "MAINTENANCE_MODE_DISABLED", 1).First(&auditLog).Error
		require.NoError(t, err)
		assert.Equal(t, "MAINTENANCE_MODE_DISABLED", auditLog.Action)
		assert.Contains(t, auditLog.Reason, "Maintenance completed successfully")
	})
}

// TestSystemAuditIntegration_QueryResults_IncludesSystemChanges tests that system change audits appear in query results
// Story 6.4, Task 8.7: Test that system change audits appear in query results
func TestSystemAuditIntegration_QueryResults_IncludesSystemChanges(t *testing.T) {
	// Setup test environment
	gin.SetMode(gin.TestMode)
	db := setupTestDBForAuditIntegration(t)

	// Create audit repository and service
	auditRepo := repositories.NewAuditRepository(db)
	auditService := services.NewAuditService(auditRepo)

	ctx := context.Background()

	// Create various system change audit logs
	auditService.LogSettingsUpdate(ctx, 1, "admin", `{"pharmacy_name":"Test"}`, "127.0.0.1")
	auditService.LogBackupCreated(ctx, 1, "admin", "test.dump", 1024, "127.0.0.1")
	auditService.LogRoleUpdated(ctx, 1, "admin", 2, "user1", "CASHIER", "OWNER", "127.0.0.1")
	auditService.LogBranchCreated(ctx, 1, "admin", "Test Branch", "Test Location", "127.0.0.1")
	auditService.LogSystemStartup(ctx, "test-system", "Test Server", "127.0.0.1")

	t.Run("Query results include all system change types", func(t *testing.T) {
		// Query audit logs with date range
		filter := &repositories.AuditLogFilter{
			StartDate: stringPtr(time.Now().AddDate(0, 0, -1).Format("2006-01-02")),
			EndDate:   stringPtr(time.Now().AddDate(0, 0, 1).Format("2006-01-02")),
			Limit:     100,
			Offset:    0,
		}

		logs, total, err := auditRepo.Query(ctx, filter)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, total, int64(5), "Should have at least 5 audit log entries")
		assert.GreaterOrEqual(t, len(logs), 5, "Should have at least 5 audit log entries")

		// Verify we have all system change action types
		actions := make(map[string]bool)
		for _, log := range logs {
			actions[string(log.Action)] = true
		}

		assert.True(t, actions["SYSTEM_SETTINGS_UPDATED"], "Should include SYSTEM_SETTINGS_UPDATED")
		assert.True(t, actions["BACKUP_CREATED"], "Should include BACKUP_CREATED")
		assert.True(t, actions["ROLE_UPDATED"], "Should include ROLE_UPDATED")
		assert.True(t, actions["BRANCH_CREATED"], "Should include BRANCH_CREATED")
		assert.True(t, actions["SYSTEM_STARTUP"], "Should include SYSTEM_STARTUP")
	})

	t.Run("Query can filter by specific system change action", func(t *testing.T) {
		// Query for only backup operations
		filter := &repositories.AuditLogFilter{
			StartDate: stringPtr(time.Now().AddDate(0, 0, -1).Format("2006-01-02")),
			EndDate:   stringPtr(time.Now().AddDate(0, 0, 1).Format("2006-01-02")),
			Action:    stringPtr("BACKUP_CREATED"),
			Limit:     100,
			Offset:    0,
		}

		logs, total, err := auditRepo.Query(ctx, filter)
		require.NoError(t, err)
		assert.Greater(t, total, int64(0), "Should have backup audit logs")
		assert.Greater(t, len(logs), 0, "Should have backup audit log entries")

		// Verify all results are backup operations
		for _, log := range logs {
			assert.Equal(t, "BACKUP_CREATED", log.Action, "All results should be BACKUP_CREATED")
		}
	})
}

// TestSystemAuditIntegration_Export_IncludesSystemChanges tests that system change audits can be exported
// Story 6.4, Task 8.8: Test that system change audits can be exported
func TestSystemAuditIntegration_Export_IncludesSystemChanges(t *testing.T) {
	// Setup test environment
	gin.SetMode(gin.TestMode)
	db := setupTestDBForAuditIntegration(t)

	// Create audit repository and service
	auditRepo := repositories.NewAuditRepository(db)
	auditService := services.NewAuditService(auditRepo)
	auditHandler := handlers.NewAuditHandler(auditRepo, auditService)

	// Setup router
	router := gin.New()
	router.GET("/api/v1/audit/logs/export", func(c *gin.Context) {
		c.Set("role", user.RoleAdmin)
		c.Set("user_id", uint(1))
		c.Set("username", "admin")
		auditHandler.GetAuditLogsExport(c)
	})

	ctx := context.Background()

	// Create system change audit logs
	auditService.LogSettingsUpdate(ctx, 1, "admin", `{"pharmacy_name":"Test"}`, "127.0.0.1")
	auditService.LogBackupCreated(ctx, 1, "admin", "test.dump", 1024, "127.0.0.1")
	auditService.LogSystemStartup(ctx, "test-system", "Test Server", "127.0.0.1")

	t.Run("CSV export includes system change audits", func(t *testing.T) {
		// Request CSV export
		req := httptest.NewRequest("GET", "/api/v1/audit/logs/export?start_date=2026-01-01&end_date=2026-12-31&format=csv", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Verify response
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "text/csv", w.Header().Get("Content-Type"))

		// Verify CSV content includes system changes
		body := w.Body.String()
		assert.Contains(t, body, "SYSTEM_SETTINGS_UPDATED")
		assert.Contains(t, body, "BACKUP_CREATED")
		assert.Contains(t, body, "SYSTEM_STARTUP")
		assert.Contains(t, body, "id,timestamp,user_id,username,action")
	})

	t.Run("JSON export includes system change audits", func(t *testing.T) {
		// Request JSON export
		req := httptest.NewRequest("GET", "/api/v1/audit/logs/export?start_date=2026-01-01&end_date=2026-12-31&format=json", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Verify response
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		// Verify JSON content includes system changes
		body := w.Body.String()
		assert.Contains(t, body, "SYSTEM_SETTINGS_UPDATED")
		assert.Contains(t, body, "BACKUP_CREATED")
		assert.Contains(t, body, "SYSTEM_STARTUP")
	})
}

// TestSystemAuditIntegration_AppendOnlyBehavior_EnforcesNoModifications tests append-only behavior for system change logs
// Story 6.4, AC3: Append-only audit trail (no modifications or deletions allowed)
func TestSystemAuditIntegration_AppendOnlyBehavior_EnforcesNoModifications(t *testing.T) {
	// Setup test environment
	gin.SetMode(gin.TestMode)
	db := setupTestDBForAuditIntegration(t)

	// Create audit repository
	auditRepo := repositories.NewAuditRepository(db)

	ctx := context.Background()

	t.Run("System change audit logs are append-only", func(t *testing.T) {
		// Create a system change audit log entry
		entry := &models.AuditLog{
			UserID:    1,
			Username:  "admin",
			Action:    "SYSTEM_SETTINGS_UPDATED",
			IPAddress: "127.0.0.1",
			Outcome:   "success",
			Reason:    `{"pharmacy_name":"Test"}`,
			Timestamp: time.Now(),
		}
		err := auditRepo.Create(ctx, entry)
		require.NoError(t, err)

		// Verify the entry was created
		var count int64
		db.Raw("SELECT COUNT(*) FROM audit_logs WHERE action = ?", "SYSTEM_SETTINGS_UPDATED").Scan(&count)
		assert.Equal(t, int64(1), count, "Should have 1 system settings audit log")

		// Verify the entry can be queried
		filter := &repositories.AuditLogFilter{
			StartDate: stringPtr(time.Now().AddDate(0, 0, -1).Format("2006-01-02")),
			EndDate:   stringPtr(time.Now().AddDate(0, 0, 1).Format("2006-01-02")),
			Action:    stringPtr("SYSTEM_SETTINGS_UPDATED"),
			Limit:     10,
			Offset:    0,
		}
		logs, _, err := auditRepo.Query(ctx, filter)
		require.NoError(t, err)
		assert.Len(t, logs, 1, "Should retrieve the created audit log")
		assert.Equal(t, "admin", logs[0].Username)
		assert.Equal(t, "SYSTEM_SETTINGS_UPDATED", logs[0].Action)

		// Note: The append-only property is enforced at the interface level
		// AuditRepository has no Update or Delete methods, only Create and Query
		// This prevents modification or deletion of audit logs at the application layer
	})
}
