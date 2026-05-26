package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/dto"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/errors"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/models"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/services"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/user"
)

// SystemSettingsHandler defines system settings handler interface
// Story 6.1, Task 5: Handler interface for system settings operations
type SystemSettingsHandler interface {
	GetSettings(c *gin.Context)         // GET /api/v1/settings - Get all settings (admin only)
	UpdateSettings(c *gin.Context)      // PUT /api/v1/settings - Update settings (admin only)
	GetPublicSettings(c *gin.Context)   // GET /api/v1/settings/public - Public settings (no auth)
}

// systemSettingsHandler implements SystemSettingsHandler
type systemSettingsHandler struct {
	systemService services.SystemService
}

// NewSystemSettingsHandler creates a new system settings handler
// Story 6.1, Task 5: Constructor with service dependency injection
func NewSystemSettingsHandler(systemService services.SystemService) SystemSettingsHandler {
	if systemService == nil {
		panic("systemService cannot be nil")
	}

	return &systemSettingsHandler{
		systemService: systemService,
	}
}

// GetSettings handles GET /api/v1/settings - Get all system settings
// Story 6.1, AC1-AC5: System Administrator can view all pharmacy settings
//
//	@Summary		Get system settings
//	@Description	Retrieve all pharmacy system settings (business name, address, phone, email). Requires System Admin role.
//	@Tags			settings
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	errors.Response{success=bool,data=dto.SystemSettingsResponse}	"Success response with system settings"
//	@Failure		401	{object}	errors.Response{success=bool,error=errors.ErrorInfo}	"Unauthorized - authentication required"
//	@Failure		403	{object}	errors.Response{success=bool,error=errors.ErrorInfo}	"Forbidden - insufficient permissions (System Admin only)"
//	@Failure		500	{object}	errors.Response{success=bool,error=errors.ErrorInfo}	"Server error"
//	@Router			/api/v1/settings [get]
//	@Security		BearerAuth
func (h *systemSettingsHandler) GetSettings(c *gin.Context) {
	// Extract user context for RBAC (AC8)
	userRole, exists := c.Get("user_role")
	if !exists {
		_ = c.Error(errors.Unauthorized("User role not found"))
		c.Status(http.StatusUnauthorized)
		return
	}

	role, ok := userRole.(string)
	if !ok {
		slog.Error("Invalid user role type in context", "type", fmt.Sprintf("%T", userRole))
		_ = c.Error(errors.InternalServerError(nil))
		c.Status(http.StatusInternalServerError)
		return
	}

	// Story 6.1, AC8: Only System Admin can access settings
	if role != user.RoleSystemAdmin {
		_ = c.Error(errors.Forbidden("Only System Administrators can access system settings"))
		c.Status(http.StatusForbidden)
		return
	}

	// Get settings from service
	settings, err := h.systemService.GetPharmacySettings(c.Request.Context())
	if err != nil {
		slog.Error("Failed to get pharmacy settings", "error", err)
		_ = c.Error(errors.InternalServerError(err))
		c.Status(http.StatusInternalServerError)
		return
	}

	// Build response
	resp := dto.SystemSettingsResponse{
		BusinessName: settings.Name,
		Address:      settings.Address,
		Phone:        settings.Phone,
		Email:        settings.Email,
		LogoURL:      settings.LogoURL,
		UpdatedAt:    time.Now(),
	}

	c.JSON(http.StatusOK, errors.Success(resp))
}

// UpdateSettings handles PUT /api/v1/settings - Update system settings
// Story 6.1, AC1-AC5, AC7: System Administrator can update pharmacy settings with audit logging
//
//	@Summary		Update system settings
//	@Description	Update pharmacy system settings (business name, address, phone, email). Requires System Admin role. Changes are logged to audit trail.
//	@Tags			settings
//	@Accept			json
//	@Produce		json
//	@Param			request	body	dto.SystemSettingsRequest	true	"Settings update request"
//	@Success		200	{object}	errors.Response{success=bool,data=dto.SettingsUpdateResponse}	"Success response with update confirmation"
//	@Failure		400	{object}	errors.Response{success=bool,error=errors.ErrorInfo}	"Validation error - invalid input"
//	@Failure		401	{object}	errors.Response{success=bool,error=errors.ErrorInfo}	"Unauthorized - authentication required"
//	@Failure		403	{object}	errors.Response{success=bool,error=errors.ErrorInfo}	"Forbidden - insufficient permissions (System Admin only)"
//	@Failure		500	{object}	errors.Response{success=bool,error=errors.ErrorInfo}	"Server error"
//	@Router			/api/v1/settings [put]
//	@Security		BearerAuth
func (h *systemSettingsHandler) UpdateSettings(c *gin.Context) {
	// Extract user context for RBAC (AC8)
	userRole, exists := c.Get("user_role")
	if !exists {
		_ = c.Error(errors.Unauthorized("User role not found"))
		c.Status(http.StatusUnauthorized)
		return
	}

	role, ok := userRole.(string)
	if !ok {
		_ = c.Error(errors.InternalServerError(nil))
		c.Status(http.StatusInternalServerError)
		return
	}

	// Story 6.1, AC8: Only System Admin can update settings
	if role != user.RoleSystemAdmin {
		_ = c.Error(errors.Forbidden("Only System Administrators can update system settings"))
		c.Status(http.StatusForbidden)
		return
	}

	// Bind and validate request body
	var req dto.SystemSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(errors.FromGinValidation(err))
		c.Status(http.StatusBadRequest)
		return
	}

	// Get user info for audit trail
	userID, exists := c.Get("user_id")
	if !exists {
		_ = c.Error(errors.Unauthorized("User ID not found"))
		c.Status(http.StatusUnauthorized)
		return
	}

	adminID, ok := userID.(uint)
	if !ok {
		slog.Error("Invalid user ID type in context", "type", fmt.Sprintf("%T", userID))
		_ = c.Error(errors.InternalServerError(nil))
		c.Status(http.StatusInternalServerError)
		return
	}

	username, exists := c.Get("username")
	if !exists {
		_ = c.Error(errors.Unauthorized("Username not found"))
		c.Status(http.StatusUnauthorized)
		return
	}

	adminUsername, ok := username.(string)
	if !ok {
		slog.Error("Invalid username type in context", "type", fmt.Sprintf("%T", username))
		_ = c.Error(errors.InternalServerError(nil))
		c.Status(http.StatusInternalServerError)
		return
	}

	// Get IP address for audit trail
	ipAddress := c.ClientIP()

	// Convert request to PharmacySettings
	settings := &models.PharmacySettings{
		Name:    req.BusinessName,
		Address: req.Address,
		Phone:   req.Phone,
		Email:   req.Email,
		LogoURL: req.LogoURL,
	}

	// Update settings via service (includes audit logging per AC7)
	if err := h.systemService.UpdateSettings(c.Request.Context(), settings, adminID, adminUsername, ipAddress); err != nil {
		slog.Error("Failed to update system settings", "error", err, "admin_id", adminID)

		// Check if it's a validation error
		if _, ok := err.(*services.InvalidInputError); ok {
			_ = c.Error(errors.BadRequest(err.Error()))
			c.Status(http.StatusBadRequest)
			return
		}

		_ = c.Error(errors.InternalServerError(err))
		c.Status(http.StatusInternalServerError)
		return
	}

	// Build response
	resp := dto.SettingsUpdateResponse{
		Message:   "Settings updated successfully",
		UpdatedAt: time.Now(),
		UpdatedBy: adminUsername,
	}

	c.JSON(http.StatusOK, errors.Success(resp))
}

// GetPublicSettings handles GET /api/v1/settings/public - Get public settings
// Story 6.1, AC6: Public settings for receipts and reports without authentication
//
//	@Summary		Get public settings
//	@Description	Retrieve public pharmacy settings (business name, address, phone, email) for receipts and reports. No authentication required.
//	@Tags			settings
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	errors.Response{success=bool,data=dto.PublicSettingsResponse}	"Success response with public settings"
//	@Failure		500	{object}	errors.Response{success=bool,error=errors.ErrorInfo}	"Server error"
//	@Router			/api/v1/settings/public [get]
func (h *systemSettingsHandler) GetPublicSettings(c *gin.Context) {
	// Get public settings from service
	settings, err := h.systemService.GetPublicSettings(c.Request.Context())
	if err != nil {
		slog.Error("Failed to get public settings", "error", err)
		_ = c.Error(errors.InternalServerError(err))
		c.Status(http.StatusInternalServerError)
		return
	}

	// Build response
	resp := dto.PublicSettingsResponse{
		BusinessName: settings.BusinessName,
		Address:      settings.Address,
		Phone:        settings.Phone,
		Email:        settings.Email,
	}

	c.JSON(http.StatusOK, errors.Success(resp))
}
