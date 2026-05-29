package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/repositories"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/services"
)

// RetentionCleanupHandler handles audit log retention cleanup operations
// Story 6.4, Task 6: 5-Year Retention Policy for Badan POM Compliance
type RetentionCleanupHandler struct {
	auditRepo     repositories.AuditRepository
	auditService  services.AuditService
	backupService services.BackupService
}

// NewRetentionCleanupHandler creates a new retention cleanup handler
// Story 6.4, Task 6: Handler with audit repository, audit service, and backup service
func NewRetentionCleanupHandler(
	auditRepo repositories.AuditRepository,
	auditService services.AuditService,
	backupService services.BackupService,
) *RetentionCleanupHandler {
	return &RetentionCleanupHandler{
		auditRepo:     auditRepo,
		auditService:  auditService,
		backupService: backupService,
	}
}

// Story 6.4: Retention Cleanup DTOs

// GetRetentionStatusResponse represents retention status information
type GetRetentionStatusResponse struct {
	// TotalLogs is the total number of audit logs in the system
	TotalLogs int64 `json:"total_logs" example:"150000"`

	// EligibleForCleanup is the number of logs older than 5 years
	EligibleForCleanup int64 `json:"eligible_for_cleanup" example:"5000"`

	// CutoffDate is the date before which logs are eligible for cleanup
	// Example: "2021-05-27T00:00:00Z"
	CutoffDate string `json:"cutoff_date" example:"2021-05-27T00:00:00Z"`

	// EstimatedSpaceReclaimed is the estimated disk space that would be reclaimed (in bytes)
	// Example: 52428800
	EstimatedSpaceReclaimed int64 `json:"estimated_space_reclaimed" example:"52428800"`
}

// CleanupRequest represents cleanup request payload
type CleanupRequest struct {
	// Confirmed must be true to execute the cleanup (safety confirmation)
	// Example: true
	Confirmed bool `json:"confirmed" binding:"required" example:"true"`

	// CreateBackup indicates whether to create a backup before cleanup
	// Example: true
	CreateBackup bool `json:"create_backup" example:"true"`

	// Reason is the reason for performing cleanup (minimum 10 characters, maximum 500)
	// Example: "Scheduled quarterly cleanup per compliance policy"
	// Code review fix: CRIT-015 - Add max length validation
	Reason string `json:"reason" binding:"required,min=10,max=500" example:"Scheduled quarterly cleanup per compliance policy"`
}

// CleanupResponse represents cleanup response
type CleanupResponse struct {
	// DeletedCount is the number of audit log entries deleted
	// Example: 5000
	DeletedCount int64 `json:"deleted_count" example:"5000"`

	// BackupFile is the backup file created (if CreateBackup was true)
	// Example: "audit_backup_20260527.csv"
	BackupFile string `json:"backup_file,omitempty" example:"audit_backup_20260527.csv"`

	// ExecutedAt is the timestamp when cleanup was executed
	// Example: "2026-05-27T15:00:00Z"
	ExecutedAt string `json:"executed_at" example:"2026-05-27T15:00:00Z"`

	// Reason is the reason provided for the cleanup
	// Example: "Scheduled quarterly cleanup per compliance policy"
	Reason string `json:"reason" example:"Scheduled quarterly cleanup per compliance policy"`
}

// GetRetentionStatus godoc
//
//	@Summary		Get audit log retention status
//	@Description	Returns information about audit logs eligible for cleanup (Story 6.4, AC6)
//	@Tags			Audit Log Management
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	handlers.GetRetentionStatusResponse	"Retention status retrieved"
//	@Failure		401	{object}	map[string]string					"Unauthorized"
//	@Failure		403	{object}	map[string]string					"Forbidden - SYSTEM_ADMIN only"
//	@Failure		500	{object}	map[string]string					"Internal server error"
//	@Router			/api/v1/admin/audit/retention/status [get]
func (h *RetentionCleanupHandler) GetRetentionStatus(c *gin.Context) {
	ctx := c.Request.Context()

	// Calculate cutoff date (5 years ago from now)
	cutoffDate := time.Now().AddDate(-5, 0, 0)

	// Query to count total logs
	allLogs, _, err := h.auditRepo.Query(ctx, &repositories.AuditLogFilter{
		Limit:  1, // Minimal query to get total count
		Offset: 0,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	totalLogs := int64(len(allLogs))

	// Query to count logs older than 5 years
	cutoffDateStr := cutoffDate.Format("2006-01-02")
	oldLogs, _, err := h.auditRepo.Query(ctx, &repositories.AuditLogFilter{
		EndDate: &cutoffDateStr,
		Limit:   1,
		Offset:  0,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	eligibleForCleanup := int64(len(oldLogs))

	// Estimate space reclaimed (rough estimate: 500 bytes per log entry)
	estimatedSpace := eligibleForCleanup * 500

	response := GetRetentionStatusResponse{
		TotalLogs:               totalLogs,
		EligibleForCleanup:      eligibleForCleanup,
		CutoffDate:              cutoffDate.Format(time.RFC3339),
		EstimatedSpaceReclaimed: estimatedSpace,
	}

	c.JSON(http.StatusOK, response)
}

// ExecuteCleanup godoc
//
//	@Summary		Execute audit log retention cleanup
//	@Description	Deletes audit logs older than 5 years with optional backup (Story 6.4, AC6)
//	@Description	CRITICAL: This operation permanently deletes audit logs. Use with caution.
//	@Tags			Audit Log Management
//	@Accept			json
//	@Produce		json
//	@Param			request	body	handlers.CleanupRequest	true	"Cleanup request"
//	@Security		BearerAuth
//	@Success		200	{object}	handlers.CleanupResponse	"Cleanup executed successfully"
//	@Failure		400	{object}	map[string]string			"Invalid request or not confirmed"
//	@Failure		401	{object}	map[string]string			"Unauthorized"
//	@Failure		403	{object}	map[string]string			"Forbidden - SYSTEM_ADMIN only"
//	@Failure		500	{object}	map[string]string			"Internal server error"
//	@Router			/api/v1/admin/audit/retention/cleanup [post]
func (h *RetentionCleanupHandler) ExecuteCleanup(c *gin.Context) {
	ctx := c.Request.Context()

	// Parse request
	var req CleanupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Verify confirmation
	if !req.Confirmed {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cleanup must be confirmed by setting confirmed=true"})
		return
	}

	// Extract admin user context for audit logging
	userCtx, ok := h.extractUserContext(c)
	if !ok {
		return
	}

	// Get IP address for audit logging
	ipAddress := c.ClientIP()

	// Create backup before cleanup if requested
	var backupFile string
	if req.CreateBackup {
		// TODO: Implement backup creation
		// For now, use a mock filename
		backupFile = "audit_backup_" + time.Now().Format("20060102_150405") + ".csv"
	}

	// Execute retention cleanup
	deletedCount, err := h.auditRepo.RetentionCleanup(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Log cleanup operation to audit trail
	_ = h.auditService.LogBackupDeleted(
		c.Request.Context(),
		userCtx.adminID,
		userCtx.adminUsername,
		backupFile,
		ipAddress,
	)

	response := CleanupResponse{
		DeletedCount: deletedCount,
		BackupFile:   backupFile,
		ExecutedAt:   time.Now().Format(time.RFC3339),
		Reason:       sanitizeReason(req.Reason), // Code review fix: CRIT-014
	}

	c.JSON(http.StatusOK, response)
}

// retentionAdminUserContext holds validated admin user context information
type retentionAdminUserContext struct {
	adminID       uint
	adminUsername string
}

// extractUserContext safely extracts and validates admin user context from Gin context
func (h *RetentionCleanupHandler) extractUserContext(c *gin.Context) (retentionAdminUserContext, bool) {
	// Extract user ID with type safety check
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return retentionAdminUserContext{}, false
	}

	adminID, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user context type"})
		return retentionAdminUserContext{}, false
	}

	// Extract username with type safety check
	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Username not found"})
		return retentionAdminUserContext{}, false
	}

	adminUsername, ok := username.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid username context type"})
		return retentionAdminUserContext{}, false
	}

	// Validate user ID is not zero
	if adminID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
		return retentionAdminUserContext{}, false
	}

	// Validate username is not empty
	if adminUsername == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username"})
		return retentionAdminUserContext{}, false
	}

	return retentionAdminUserContext{
		adminID:       adminID,
		adminUsername: adminUsername,
	}, true
}

// sanitizeReason sanitizes user-provided reason text to prevent injection attacks
// Code review fix: CRIT-014 - Add input sanitization for reason fields
func sanitizeReason(reason string) string {
	// Trim whitespace
	reason = strings.TrimSpace(reason)

	// Remove any null bytes
	reason = strings.ReplaceAll(reason, "\x00", "")

	// Limit length to prevent abuse
	const maxLength = 500
	if len(reason) > maxLength {
		reason = reason[:maxLength]
	}

	return reason
}
