package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/dto"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/services"
)

// BackupHandler handles backup management endpoints
// Story 6.3, Task 4: Create Backup Admin API Endpoints
type BackupHandler struct {
	backupService services.BackupService
}

// NewBackupHandler creates a new backup handler
// Story 6.3, Task 4: Handler initialization with service dependency
func NewBackupHandler(backupService services.BackupService) *BackupHandler {
	return &BackupHandler{
		backupService: backupService,
	}
}

// GetBackupService returns the underlying backup service
// Story 6.3, Task 6: Expose service for health monitoring integration
func (h *BackupHandler) GetBackupService() any {
	return h.backupService
}

// CreateBackup godoc
// @Summary      Trigger manual backup
// @Description  Manually triggers a database backup operation (Story 6.3, AC5)
// @Tags         Admin Backups
// @Accept       json
// @Produce      json
// @Param        request body dto.CreateBackupRequest true "Backup request"
// @Security     BearerAuth
// @Success      202  {object}  dto.CreateBackupResponse  "Backup started"
// @Failure      400  {object}  map[string]string  "Invalid request"
// @Failure      401  {object}  map[string]string  "Unauthorized"
// @Failure      403  {object}  map[string]string  "Forbidden - Admin only"
// @Failure      409  {object}  map[string]string  "Backup already in progress"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /api/v1/admin/backups [post]
func (h *BackupHandler) CreateBackup(c *gin.Context) {
	ctx := c.Request.Context()

	// Parse request
	var req dto.CreateBackupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Set default description if not provided
	description := req.Description
	if description == "" {
		description = "Manual backup"
	}

	// Create backup
	backupInfo, err := h.backupService.CreateBackup(ctx, description)
	if err != nil {
		if strings.Contains(err.Error(), "already in progress") {
			c.JSON(http.StatusConflict, gin.H{"error": "Backup already in progress"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := dto.CreateBackupResponse{
		Status:        "started",
		Filename:      backupInfo.Filename,
		EstimatedTime: "2-5 min",
	}

	c.JSON(http.StatusAccepted, response)
}

// ListBackups godoc
// @Summary      List all backups
// @Description  Returns list of all available backups with metadata (Story 6.3, AC5)
// @Tags         Admin Backups
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  dto.BackupListResponse  "Backups retrieved successfully"
// @Failure      401  {object}  map[string]string  "Unauthorized"
// @Failure      403  {object}  map[string]string  "Forbidden - Admin only"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /api/v1/admin/backups [get]
func (h *BackupHandler) ListBackups(c *gin.Context) {
	ctx := c.Request.Context()

	response, err := h.backupService.ListBackups(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// DownloadBackup godoc
// @Summary      Download backup file
// @Description  Downloads the specified backup file (Story 6.3, AC5)
// @Tags         Admin Backups
// @Accept       json
// @Produce      application/octet-stream
// @Param        filename path string true "Backup filename"
// @Security     BearerAuth
// @Success      200  {file}  file  "Backup file"
// @Failure      400  {object}  map[string]string  "Invalid filename"
// @Failure      401  {object}  map[string]string  "Unauthorized"
// @Failure      403  {object}  map[string]string  "Forbidden - Admin only"
// @Failure      404  {object}  map[string]string  "Backup not found"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /api/v1/admin/backups/{filename} [get]
func (h *BackupHandler) DownloadBackup(c *gin.Context) {
	ctx := c.Request.Context()

	filename := c.Param("filename")
	if filename == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Filename is required"})
		return
	}

	// Validate filename has .dump extension
	if !strings.HasSuffix(filename, ".dump") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid backup file"})
		return
	}

	reader, err := h.backupService.GetBackupFile(ctx, filename)
	if err != nil {
		if strings.Contains(err.Error(), "invalid") || strings.Contains(err.Error(), "format") {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"error": "Backup file not found"})
		return
	}
	defer reader.Close()

	// Set headers for file download
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Disposition", "attachment; filename=\""+filename+"\"")
	c.Header("Content-Transfer-Encoding", "binary")

	// Stream file to response
	c.DataFromReader(http.StatusOK, -1, "application/octet-stream", reader, nil)
}

// RestoreBackup godoc
// @Summary      Restore from backup
// @Description  Restores database from the specified backup file (Story 6.3, AC6)
// @Tags         Admin Backups
// @Accept       json
// @Produce      json
// @Param        filename path string true "Backup filename"
// @Param        request body dto.RestoreBackupRequest true "Restore confirmation"
// @Security     BearerAuth
// @Success      202  {object}  map[string]string  "Restore started"
// @Failure      400  {object}  map[string]string  "Invalid request"
// @Failure      401  {object}  map[string]string  "Unauthorized"
// @Failure      403  {object}  map[string]string  "Forbidden - Admin only"
// @Failure      404  {object}  map[string]string  "Backup not found"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /api/v1/admin/backups/{filename}/restore [post]
func (h *BackupHandler) RestoreBackup(c *gin.Context) {
	ctx := c.Request.Context()

	filename := c.Param("filename")
	if filename == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Filename is required"})
		return
	}

	// Parse request
	var req dto.RestoreBackupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Require confirmation
	if !req.Confirmed {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Restore requires confirmation"})
		return
	}

	// Require reason for restore operation
	if req.Reason == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Restore reason is required"})
		return
	}

	// Validate backup before restore
	validationErrors, err := h.backupService.ValidateBackup(ctx, filename)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if len(validationErrors) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Backup validation failed",
			"details": validationErrors,
		})
		return
	}

	// Perform restore
	if err := h.backupService.RestoreBackup(ctx, filename, req.Reason); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"status":  "started",
		"message": "Database restore operation started",
		"warning": "This operation may take several minutes",
	})
}

// DeleteBackup godoc
// @Summary      Delete backup file
// @Description  Manually deletes the specified backup file (Story 6.3, Task 4)
// @Tags         Admin Backups
// @Accept       json
// @Produce      json
// @Param        filename path string true "Backup filename"
// @Param        request body dto.DeleteBackupRequest true "Delete confirmation"
// @Security     BearerAuth
// @Success      200  {object}  map[string]string  "Backup deleted"
// @Failure      400  {object}  map[string]string  "Invalid request"
// @Failure      401  {object}  map[string]string  "Unauthorized"
// @Failure      403  {object}  map[string]string  "Forbidden - Admin only"
// @Failure      404  {object}  map[string]string  "Backup not found"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /api/v1/admin/backups/{filename} [delete]
func (h *BackupHandler) DeleteBackup(c *gin.Context) {
	ctx := c.Request.Context()

	filename := c.Param("filename")
	if filename == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Filename is required"})
		return
	}

	// Parse request
	var req dto.DeleteBackupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Require confirmation
	if !req.Confirmed {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Deletion requires confirmation"})
		return
	}

	// Require reason for deletion
	if req.Reason == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Deletion reason is required"})
		return
	}

	// Delete backup file
	if err := h.backupService.DeleteBackup(ctx, filename); err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if strings.Contains(err.Error(), "invalid") {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "deleted",
		"message": "Backup file deleted successfully",
		"warning": "This operation cannot be undone",
	})
}

// GetBackupStatus godoc
// @Summary      Get backup status
// @Description  Returns current backup job status and statistics (Story 6.3, AC4)
// @Tags         Admin Backups
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  dto.BackupJobStatus  "Backup status retrieved"
// @Failure      401  {object}  map[string]string  "Unauthorized"
// @Failure      403  {object}  map[string]string  "Forbidden - Admin only"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /api/v1/admin/backups/status [get]
func (h *BackupHandler) GetBackupStatus(c *gin.Context) {
	ctx := c.Request.Context()

	status, err := h.backupService.GetBackupStatus(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, status)
}

// GetBackupConfig godoc
// @Summary      Get backup configuration
// @Description  Returns current backup configuration (Story 6.3, AC8)
// @Tags         Admin Backups
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  dto.BackupConfig  "Backup configuration"
// @Failure      401  {object}  map[string]string  "Unauthorized"
// @Failure      403  {object}  map[string]string  "Forbidden - Admin only"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /api/v1/admin/backups/config [get]
func (h *BackupHandler) GetBackupConfig(c *gin.Context) {
	ctx := c.Request.Context()

	config, err := h.backupService.GetConfig(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, config)
}

// UpdateBackupConfig godoc
// @Summary      Update backup configuration
// @Description  Updates backup configuration settings (Story 6.3, AC8)
// @Tags         Admin Backups
// @Accept       json
// @Produce      json
// @Param        request body dto.BackupConfig true "Backup configuration"
// @Security     BearerAuth
// @Success      200  {object}  map[string]string  "Configuration updated"
// @Failure      400  {object}  map[string]string  "Invalid request"
// @Failure      401  {object}  map[string]string  "Unauthorized"
// @Failure      403  {object}  map[string]string  "Forbidden - Admin only"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /api/v1/admin/backups/config [put]
func (h *BackupHandler) UpdateBackupConfig(c *gin.Context) {
	ctx := c.Request.Context()

	var config dto.BackupConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Validate configuration
	if config.Schedule == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Schedule cannot be empty"})
		return
	}

	if config.RetentionDays < 1 || config.RetentionDays > 365 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Retention days must be between 1 and 365"})
		return
	}

	if config.StoragePath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Storage path cannot be empty"})
		return
	}

	if err := h.backupService.UpdateConfig(ctx, &config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "updated",
		"message": "Backup configuration updated successfully",
		"config":  config,
	})
}

// ValidateBackup godoc
// @Summary      Validate backup file
// @Description  Validates the specified backup file for restoration (Story 6.3, AC6)
// @Tags         Admin Backups
// @Accept       json
// @Produce      json
// @Param        filename path string true "Backup filename"
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}  "Backup is valid"
// @Failure      400  {object}  map[string]string  "Invalid filename"
// @Failure      401  {object}  map[string]string  "Unauthorized"
// @Failure      403  {object}  map[string]string  "Forbidden - Admin only"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /api/v1/admin/backups/{filename}/validate [get]
func (h *BackupHandler) ValidateBackup(c *gin.Context) {
	ctx := c.Request.Context()

	filename := c.Param("filename")
	if filename == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Filename is required"})
		return
	}

	errors, err := h.backupService.ValidateBackup(ctx, filename)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if len(errors) > 0 {
		c.JSON(http.StatusOK, gin.H{
			"valid":   false,
			"errors":  errors,
			"message": "Backup validation failed",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"valid":   true,
		"message": "Backup file is valid for restoration",
	})
}

// CleanupOldBackups godoc
// @Summary      Cleanup old backups
// @Description  Manually triggers cleanup of backups older than retention period (Story 6.3, AC3)
// @Tags         Admin Backups
// @Accept       json
// @Produce      json
// @Param        retention_days query int false "30" "Retention period in days"
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}  "Cleanup completed"
// @Failure      401  {object}  map[string]string  "Unauthorized"
// @Failure      403  {object}  map[string]string  "Forbidden - Admin only"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /api/v1/admin/backups/cleanup [post]
func (h *BackupHandler) CleanupOldBackups(c *gin.Context) {
	ctx := c.Request.Context()

	// Parse retention days parameter
	retentionDays := 30 // default
	if rdParam := c.Query("retention_days"); rdParam != "" {
		if parsed, err := strconv.Atoi(rdParam); err == nil {
			if parsed >= 1 && parsed <= 365 {
				retentionDays = parsed
			}
		}
	}

	deletedCount, err := h.backupService.DeleteOldBackups(ctx, retentionDays)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":         "completed",
		"deleted_count":  deletedCount,
		"retention_days": retentionDays,
		"message":        "Backup cleanup completed successfully",
	})
}
