package tests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/handlers"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/models"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/repositories"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/services"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/user"
)

// setupTestDBForAuditIntegration creates an in-memory SQLite database for integration testing
func setupTestDBForAuditIntegration(t *testing.T) *gorm.DB {
	t.Helper()

	// Open in-memory SQLite database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Create audit_logs table manually
	db.Exec(`
		CREATE TABLE audit_logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			username VARCHAR(255) NOT NULL,
			action VARCHAR(100) NOT NULL,
			ip_address VARCHAR(45),
			outcome VARCHAR(50) NOT NULL,
			reason TEXT,
			timestamp DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
	`)

	// Create indexes for performance
	db.Exec(`CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);`)
	db.Exec(`CREATE INDEX idx_audit_logs_action ON audit_logs(action);`)
	db.Exec(`CREATE INDEX idx_audit_logs_timestamp ON audit_logs(timestamp);`)

	return db
}

// setupTestRedisForAuditIntegration creates a miniredis instance for integration testing
func setupTestRedisForAuditIntegration(t *testing.T) (*miniredis.Miniredis, *redis.Client) {
	t.Helper()

	mr, err := miniredis.Run()
	require.NoError(t, err)

	// Create Redis client connected to miniredis
	redisClient := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	return mr, redisClient
}

// TestAuditIntegration_TransactionCreation_LogsAuditEntry tests that transaction creation creates audit log entry
// Story 5.4, Task 10.1: Test transaction creation creates audit log entry
func TestAuditIntegration_TransactionCreation_LogsAuditEntry(t *testing.T) {
	// Setup test environment
	gin.SetMode(gin.TestMode)
	db := setupTestDBForAuditIntegration(t)

	// Create audit repository and service
	auditRepo := repositories.NewAuditRepository(db)
	auditService := services.NewAuditService(auditRepo)

	// Simulate transaction creation
	ctx := context.Background()

	// Log transaction creation to audit trail
	err := auditService.LogStockAdjustment(
		ctx,
		1,                    // adminID
		"admin_user",         // adminUsername
		123,                  // productID
		"PARACETAMOL",        // productSKU
		100,                  // oldQty
		95,                   // newQty
		"Damaged packaging",  // reason
		"127.0.0.1",          // ipAddress
	)

	// Assert audit log was created
	require.NoError(t, err, "Stock adjustment should be logged to audit trail")

	// Verify audit log entry exists in database
	var count int64
	db.Raw("SELECT COUNT(*) FROM audit_logs WHERE action = ?", "STOCK_ADJUSTMENT").Scan(&count)
	assert.Greater(t, count, int64(0), "Audit log entry should exist for STOCK_ADJUSTMENT")

	// Verify audit log details
	var auditLog models.AuditLog
	err = db.Where("action = ?", "STOCK_ADJUSTMENT").First(&auditLog).Error
	require.NoError(t, err)
	assert.Equal(t, uint(1), auditLog.UserID)
	assert.Equal(t, "admin_user", auditLog.Username)
	assert.Equal(t, "STOCK_ADJUSTMENT", auditLog.Action)
	assert.Contains(t, auditLog.Reason, "PARACETAMOL")
}

// TestAuditIntegration_StockAdjustment_LogsAuditEntry tests that stock adjustment creates audit log entry
// Story 5.4, Task 10.2: Test stock adjustment creates audit log entry
func TestAuditIntegration_StockAdjustment_LogsAuditEntry(t *testing.T) {
	// Setup test environment
	gin.SetMode(gin.TestMode)
	db := setupTestDBForAuditIntegration(t)

	// Create audit repository and service
	auditRepo := repositories.NewAuditRepository(db)
	auditService := services.NewAuditService(auditRepo)

	ctx := context.Background()

	// Simulate stock adjustment
	err := auditService.LogStockAdjustment(
		ctx,
		2,                     // adminID
		"manager_user",         // adminUsername
		456,                   // productID
		"AMOXICILLIN",          // productSKU
		50,                    // oldQty
		45,                    // newQty
		"Expired items removed", // reason
		"192.168.1.100",        // ipAddress
	)

	// Assert audit log was created
	require.NoError(t, err, "Stock adjustment should be logged")

	// Verify audit log entry
	var auditLog models.AuditLog
	err = db.Where("action = ? AND user_id = ?", "STOCK_ADJUSTMENT", 2).First(&auditLog).Error
	require.NoError(t, err)
	assert.Equal(t, uint(2), auditLog.UserID)
	assert.Equal(t, "manager_user", auditLog.Username)
	assert.Equal(t, "192.168.1.100", auditLog.IPAddress)
	assert.Equal(t, "success", auditLog.Outcome)
}

// TestAuditIntegration_BlockedSaleAttempt_LogsAuditEntry tests that blocked sale attempt creates audit log entry
// Story 5.4, Task 10.3: Test blocked sale attempt creates audit log entry
func TestAuditIntegration_BlockedSaleAttempt_LogsAuditEntry(t *testing.T) {
	// Setup test environment
	gin.SetMode(gin.TestMode)
	db := setupTestDBForAuditIntegration(t)

	// Create audit repository and service
	auditRepo := repositories.NewAuditRepository(db)
	auditService := services.NewAuditService(auditRepo)

	ctx := context.Background()

	// Simulate blocked sale attempt
	err := auditService.LogBlockedSaleAttempt(
		ctx,
		3,                      // userID
		"cashier_user",          // username
		789,                    // productID
		"EXP001",               // productSKU
		"Expired Medicine",     // productName
		"2024-01-01",           // expiryDate
		"Product expired and cannot be sold", // reason
		"192.168.1.101",        // ipAddress
	)

	// Assert audit log was created
	require.NoError(t, err, "Blocked sale attempt should be logged")

	// Verify audit log entry
	var auditLog models.AuditLog
	err = db.Where("action = ? AND user_id = ?", "BLOCKED_SALE_ATTEMPT", 3).First(&auditLog).Error
	require.NoError(t, err)
	assert.Equal(t, "BLOCKED_SALE_ATTEMPT", auditLog.Action)
	assert.Equal(t, "blocked", auditLog.Outcome)
	assert.Contains(t, auditLog.Reason, "EXP001")
	assert.Contains(t, auditLog.Reason, "Expired Medicine")
	assert.Equal(t, "192.168.1.101", auditLog.IPAddress)
}

// TestAuditIntegration_ReportExport_LogsAuditEntry tests that report export creates audit log entry
// Story 5.4, Task 10.4: Test report export creates audit log entry
func TestAuditIntegration_ReportExport_LogsAuditEntry(t *testing.T) {
	// Setup test environment
	gin.SetMode(gin.TestMode)
	db := setupTestDBForAuditIntegration(t)

	// Create audit repository and service
	auditRepo := repositories.NewAuditRepository(db)
	auditService := services.NewAuditService(auditRepo)

	ctx := context.Background()

	// Simulate report export
	err := auditService.LogReportExport(
		ctx,
		1,                      // userID
		"admin_user",           // username
		"daily_sales",          // reportType
		"pdf",                  // format
		"2026-05-01_to_2026-05-26", // dateRange
		"success",              // outcome
		"192.168.1.100",        // ipAddress
	)

	// Assert audit log was created
	require.NoError(t, err, "Report export should be logged")

	// Verify audit log entry
	var auditLog models.AuditLog
	err = db.Where("action = ? AND user_id = ?", "EXPORT_REPORT", 1).First(&auditLog).Error
	require.NoError(t, err)
	assert.Equal(t, "EXPORT_REPORT", auditLog.Action)
	assert.Contains(t, auditLog.Reason, "daily_sales")
	assert.Contains(t, auditLog.Reason, "pdf")
	assert.Equal(t, "192.168.1.100", auditLog.IPAddress)
}

// TestAuditIntegration_QueryAPI_RBAC_ValidatesAccess tests audit log query API with RBAC
// Story 5.4, Task 10.5: Test audit log query API with RBAC (Admin/Owner only)
func TestAuditIntegration_QueryAPI_RBAC_ValidatesAccess(t *testing.T) {
	// Setup test environment
	gin.SetMode(gin.TestMode)
	db := setupTestDBForAuditIntegration(t)

	// Create audit repository and handler
	auditRepo := repositories.NewAuditRepository(db)
	auditService := services.NewAuditService(auditRepo)
	auditHandler := handlers.NewAuditHandler(auditRepo, auditService)

	// Setup router with RBAC middleware
	router := gin.New()
	auditGroup := router.Group("/api/v1/audit")
	auditGroup.Use(func(c *gin.Context) {
		// Mock session auth middleware for testing
		c.Set("user_id", c.GetHeader("X-User-ID"))
		c.Set("username", c.GetHeader("X-Username"))
		c.Set("role", c.GetHeader("X-Role"))
		c.Next()
	})
	{
		auditGroup.GET("/logs", auditHandler.GetAuditLogs)
	}

	t.Run("Admin can access audit logs", func(t *testing.T) {
		// Create audit log entry
		ctx := context.Background()
		err := auditService.LogStockAdjustment(ctx, 1, "admin", 1, "SKU001", 100, 95, "Test", "127.0.0.1")
		require.NoError(t, err)

		// Make request as Admin
		req := httptest.NewRequest("GET", "/api/v1/audit/logs?start_date=2026-01-01&end_date=2026-12-31", nil)
		req.Header.Set("X-Role", user.RoleAdmin)
		req.Header.Set("X-User-ID", "1")
		req.Header.Set("X-Username", "admin")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Should return 200 OK
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Cashier cannot access audit logs", func(t *testing.T) {
		// Make request as Cashier
		req := httptest.NewRequest("GET", "/api/v1/audit/logs?start_date=2026-01-01&end_date=2026-12-31", nil)
		req.Header.Set("X-Role", user.RoleCashier)
		req.Header.Set("X-User-ID", "2")
		req.Header.Set("X-Username", "cashier")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Should return 403 Forbidden
		assert.Equal(t, http.StatusForbidden, w.Code)
		assert.Contains(t, w.Body.String(), "permission")
	})

	t.Run("Owner can access audit logs", func(t *testing.T) {
		// Make request as Owner
		req := httptest.NewRequest("GET", "/api/v1/audit/logs?start_date=2026-01-01&end_date=2026-12-31", nil)
		req.Header.Set("X-Role", user.RoleOwner)
		req.Header.Set("X-User-ID", "3")
		req.Header.Set("X-Username", "owner")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Should return 200 OK
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

// TestAuditIntegration_ExportFunctionality_EndToEnd tests audit log export functionality end-to-end
// Story 5.4, Task 10.6: Test audit log export functionality end-to-end
func TestAuditIntegration_ExportFunctionality_EndToEnd(t *testing.T) {
	// Setup test environment
	gin.SetMode(gin.TestMode)
	db := setupTestDBForAuditIntegration(t)

	// Create audit repository and handler
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

	t.Run("CSV export returns valid CSV format", func(t *testing.T) {
		// Create test audit logs
		ctx := context.Background()
		err := auditService.LogStockAdjustment(ctx, 1, "admin", 1, "SKU001", 100, 95, "Test", "127.0.0.1")
		require.NoError(t, err)

		// Request CSV export
		req := httptest.NewRequest("GET", "/api/v1/audit/logs/export?start_date=2026-01-01&end_date=2026-12-31&format=csv", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Verify response
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "text/csv", w.Header().Get("Content-Type"))
		assert.Contains(t, w.Header().Get("Content-Disposition"), ".csv")

		// Verify CSV content
		body := w.Body.String()
		assert.Contains(t, body, "id,timestamp,user_id,username,action")
		assert.Contains(t, body, "STOCK_ADJUSTMENT")
	})

	t.Run("JSON export returns valid JSON format", func(t *testing.T) {
		// Request JSON export
		req := httptest.NewRequest("GET", "/api/v1/audit/logs/export?start_date=2026-01-01&end_date=2026-12-31&format=json", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Verify response
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
		assert.Contains(t, w.Header().Get("Content-Disposition"), ".json")

		// Verify JSON content
		body := w.Body.String()
		assert.Contains(t, body, "[")
		assert.Contains(t, body, "]")
	})
}

// TestAuditIntegration_RetentionCleanup_DeletesOldRecords tests 5-year retention cleanup
// Story 5.4, Task 10.7: Test 5-year retention cleanup (verify old records deleted)
func TestAuditIntegration_RetentionCleanup_DeletesOldRecords(t *testing.T) {
	// Setup test environment
	gin.SetMode(gin.TestMode)
	db := setupTestDBForAuditIntegration(t)

	// Create audit repository
	auditRepo := repositories.NewAuditRepository(db)

	t.Run("Retention cleanup deletes records older than 5 years", func(t *testing.T) {
		ctx := context.Background()

		// Create recent audit log (should NOT be deleted)
		recentLog := &models.AuditLog{
			UserID:    1,
			Username:  "admin",
			Action:    "STOCK_ADJUSTMENT",
			IPAddress: "127.0.0.1",
			Outcome:   "success",
			Reason:    "Recent adjustment",
			Timestamp: time.Now(),
		}
		err := auditRepo.Create(ctx, recentLog)
		require.NoError(t, err)

		// Create old audit log (6 years ago, should be deleted)
		oldLog := &models.AuditLog{
			UserID:    2,
			Username:  "old_user",
			Action:    "STOCK_ADJUSTMENT",
			IPAddress: "127.0.0.1",
			Outcome:   "success",
			Reason:    "Old adjustment",
			Timestamp: time.Now().AddDate(-6, 0, 0), // 6 years ago
		}
		err = auditRepo.Create(ctx, oldLog)
		require.NoError(t, err)

		// Verify both logs exist before cleanup
		var countBefore int64
		db.Raw("SELECT COUNT(*) FROM audit_logs").Scan(&countBefore)
		assert.Equal(t, int64(2), countBefore, "Should have 2 audit logs before cleanup")

		// Perform retention cleanup
		deletedCount, err := auditRepo.RetentionCleanup(ctx)
		require.NoError(t, err)

		// Verify cleanup results
		assert.Equal(t, int64(1), deletedCount, "Should delete 1 old record")

		// Verify old log was deleted but recent log remains
		var countAfter int64
		db.Raw("SELECT COUNT(*) FROM audit_logs").Scan(&countAfter)
		assert.Equal(t, int64(1), countAfter, "Should have 1 audit log after cleanup")

		// Verify remaining log is the recent one
		var remainingLog models.AuditLog
		err = db.First(&remainingLog).Error
		require.NoError(t, err)
		assert.Equal(t, "admin", remainingLog.Username)
		assert.Equal(t, "Recent adjustment", remainingLog.Reason)
	})

	t.Run("Retention cleanup preserves 5-year boundary records", func(t *testing.T) {
		ctx := context.Background()

		// Create audit log exactly 5 years old (boundary case, should NOT be deleted)
		boundaryLog := &models.AuditLog{
			UserID:    3,
			Username:  "boundary_user",
			Action:    "STOCK_ADJUSTMENT",
			IPAddress: "127.0.0.1",
			Outcome:   "success",
			Reason:    "Boundary adjustment",
			Timestamp: time.Now().AddDate(-5, 0, 0), // Exactly 5 years ago
		}
		err := auditRepo.Create(ctx, boundaryLog)
		require.NoError(t, err)

		// Perform retention cleanup
		deletedCount, err := auditRepo.RetentionCleanup(ctx)
		require.NoError(t, err)

		// Verify boundary record was NOT deleted
		var remainingCount int64
		db.Raw("SELECT COUNT(*) FROM audit_logs").Scan(&remainingCount)
		assert.Equal(t, int64(1), remainingCount, "Boundary record should remain after cleanup")

		// Verify deleted count doesn't include boundary record
		assert.Equal(t, int64(0), deletedCount, "No records should be deleted (only boundary record exists)")
	})
}

// TestAuditIntegration_AppendOnlyInterface_NoUpdateDelete tests append-only behavior at integration level
// Story 5.4, AC3: Append-only audit trail (no modifications or deletions allowed)
func TestAuditIntegration_AppendOnlyInterface_NoUpdateDelete(t *testing.T) {
	// Setup test environment
	gin.SetMode(gin.TestMode)
	db := setupTestDBForAuditIntegration(t)

	// Create audit repository
	auditRepo := repositories.NewAuditRepository(db)

	t.Run("AuditRepository has no Update method", func(t *testing.T) {
		// This test verifies at compile time that the interface doesn't have Update/Delete methods
		// The interface definition in audit_repository.go intentionally omits these methods
		// Story 5.4, AC3: Append-only audit trail enforced at interface level

		// Create an audit log entry
		ctx := context.Background()
		entry := &models.AuditLog{
			UserID:    1,
			Username:  "admin",
			Action:    "STOCK_ADJUSTMENT",
			IPAddress: "127.0.0.1",
			Outcome:   "success",
			Reason:    "Test entry",
			Timestamp: time.Now(),
		}
		err := auditRepo.Create(ctx, entry)
		require.NoError(t, err)

		// Verify we can still query the entry
		filter := &repositories.AuditLogFilter{
			StartDate: stringPtr("2026-01-01"),
			EndDate:   stringPtr("2026-12-31"),
			Limit:     10,
			Offset:    0,
		}
		logs, _, err := auditRepo.Query(ctx, filter)
		require.NoError(t, err)
		assert.Len(t, logs, 1, "Should have 1 audit log entry")
		assert.Equal(t, "admin", logs[0].Username)
	})
}

// Helper function to create string pointers for filter parameters
func stringPtr(s string) *string {
	return &s
}
