package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/services"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/user"
)

// UserManagementHandler handles user role and permission management operations
// Story 6.4, Task 4: User Management Audit Integration
type UserManagementHandler struct {
	userService  user.Service
	auditService services.AuditService
}

// NewUserManagementHandler creates a new user management handler
// Story 6.4, Task 4: Handler with user and audit service dependencies
func NewUserManagementHandler(userService user.Service, auditService services.AuditService) *UserManagementHandler {
	return &UserManagementHandler{
		userService:  userService,
		auditService: auditService,
	}
}

// UpdateRole godoc
//
//	@Summary		Update user role
//	@Description	Updates a user's role with audit logging (Story 6.4, AC1, AC2)
//	@Tags			User Management
//	@Accept			json
//	@Produce		json
//	@Param			user_id	path	int						true	"User ID"
//	@Param			request	body	user.UpdateRoleRequest	true	"Role update request"
//	@Security		BearerAuth
//	@Success		200	{object}	user.UpdateRoleResponse	"Role updated successfully"
//	@Failure		400	{object}	map[string]string		"Invalid request"
//	@Failure		401	{object}	map[string]string		"Unauthorized"
//	@Failure		403	{object}	map[string]string		"Forbidden - SYSTEM_ADMIN only"
//	@Failure		404	{object}	map[string]string		"User not found"
//	@Failure		500	{object}	map[string]string		"Internal server error"
//	@Router			/api/v1/admin/users/{user_id}/role [put]
func (h *UserManagementHandler) UpdateRole(c *gin.Context) {
	ctx := c.Request.Context()

	// Extract target user ID from path
	userIDStr := c.Param("user_id")
	targetUserID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Parse request
	var req user.UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Validate role is valid
	if !user.IsValidRoleForCreate(req.Role) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role. Must be one of: SYSTEM_ADMIN, OWNER, CASHIER"})
		return
	}

	// Extract admin user context for audit logging
	userCtx, ok := h.extractUserContext(c)
	if !ok {
		return
	}

	// Get current user to check old role
	targetUser, err := h.userService.GetUserByID(ctx, uint(targetUserID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	oldRole := targetUser.Role

	// Update user role using UpdateUser with Role field
	updateReq := user.UpdateUserRequest{
		Role: req.Role,
	}

	updatedUser, err := h.userService.UpdateUser(ctx, uint(targetUserID), updateReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get IP address for audit logging
	ipAddress := c.ClientIP()

	// Log role update to audit trail (Story 6.4, AC1, AC2)
	_ = h.auditService.LogRoleUpdated(
		ctx,
		userCtx.adminID,
		userCtx.adminUsername,
		uint(targetUserID),
		targetUser.Username,
		oldRole,
		req.Role,
		ipAddress,
	)

	response := user.UpdateRoleResponse{
		ID:        updatedUser.ID,
		Username:  updatedUser.Username,
		OldRole:   oldRole,
		NewRole:   req.Role,
		UpdatedAt: time.Now().Format(time.RFC3339),
	}

	c.JSON(http.StatusOK, response)
}

// GrantPermission godoc
//
//	@Summary		Grant permission to user
//	@Description	Grants a specific permission to a user with audit logging (Story 6.4, AC1, AC2)
//	@Description	Note: Current implementation uses RBAC with pre-defined roles. This endpoint
//	@Description	is provided for future PBAC (Permission-Based Access Control) enhancement.
//	@Tags			User Management
//	@Accept			json
//	@Produce		json
//	@Param			user_id	path	int							true	"User ID"
//	@Param			request	body	user.GrantPermissionRequest	true	"Permission grant request"
//	@Security		BearerAuth
//	@Success		200	{object}	user.GrantPermissionResponse	"Permission granted successfully"
//	@Failure		400	{object}	map[string]string				"Invalid request"
//	@Failure		401	{object}	map[string]string				"Unauthorized"
//	@Failure		403	{object}	map[string]string				"Forbidden - SYSTEM_ADMIN only"
//	@Failure		404	{object}	map[string]string				"User not found"
//	@Failure		501	{object}	map[string]string				"Not Implemented - RBAC system in use"
//	@Router			/api/v1/admin/users/{user_id}/permissions [post]
func (h *UserManagementHandler) GrantPermission(c *gin.Context) {
	ctx := c.Request.Context()

	// Extract target user ID from path
	userIDStr := c.Param("user_id")
	targetUserID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Parse request
	var req user.GrantPermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Extract admin user context for audit logging
	userCtx, ok := h.extractUserContext(c)
	if !ok {
		return
	}

	// Get target user
	targetUser, err := h.userService.GetUserByID(ctx, uint(targetUserID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Get IP address for audit logging
	ipAddress := c.ClientIP()

	// Log permission grant to audit trail (Story 6.4, AC1, AC2)
	_ = h.auditService.LogPermissionGranted(
		ctx,
		userCtx.adminID,
		userCtx.adminUsername,
		uint(targetUserID),
		targetUser.Username,
		req.Permission,
		ipAddress,
	)

	response := user.GrantPermissionResponse{
		ID:         targetUser.ID,
		Username:   targetUser.Username,
		Permission: req.Permission,
		GrantedAt:  time.Now().Format(time.RFC3339),
	}

	c.JSON(http.StatusOK, response)
}

// RevokePermission godoc
//
//	@Summary		Revoke permission from user
//	@Description	Revokes a specific permission from a user with audit logging (Story 6.4, AC1, AC2)
//	@Description	Note: Current implementation uses RBAC with pre-defined roles. This endpoint
//	@Description	is provided for future PBAC (Permission-Based Access Control) enhancement.
//	@Tags			User Management
//	@Accept			json
//	@Produce		json
//	@Param			user_id	path	int								true	"User ID"
//	@Param			request	body	user.RevokePermissionRequest	true	"Permission revoke request"
//	@Security		BearerAuth
//	@Success		200	{object}	user.RevokePermissionResponse	"Permission revoked successfully"
//	@Failure		400	{object}	map[string]string				"Invalid request"
//	@Failure		401	{object}	map[string]string				"Unauthorized"
//	@Failure		403	{object}	map[string]string				"Forbidden - SYSTEM_ADMIN only"
//	@Failure		404	{object}	map[string]string				"User not found"
//	@Failure		501	{object}	map[string]string				"Not Implemented - RBAC system in use"
//	@Router			/api/v1/admin/users/{user_id}/permissions [delete]
func (h *UserManagementHandler) RevokePermission(c *gin.Context) {
	ctx := c.Request.Context()

	// Extract target user ID from path
	userIDStr := c.Param("user_id")
	targetUserID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Parse request
	var req user.RevokePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Extract admin user context for audit logging
	userCtx, ok := h.extractUserContext(c)
	if !ok {
		return
	}

	// Get target user
	targetUser, err := h.userService.GetUserByID(ctx, uint(targetUserID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Get IP address for audit logging
	ipAddress := c.ClientIP()

	// Log permission revoke to audit trail (Story 6.4, AC1, AC2)
	_ = h.auditService.LogPermissionRevoked(
		ctx,
		userCtx.adminID,
		userCtx.adminUsername,
		uint(targetUserID),
		targetUser.Username,
		req.Permission,
		ipAddress,
	)

	response := user.RevokePermissionResponse{
		ID:         targetUser.ID,
		Username:   targetUser.Username,
		Permission: req.Permission,
		RevokedAt:  time.Now().Format(time.RFC3339),
	}

	c.JSON(http.StatusOK, response)
}

// adminUserContext holds validated admin user context information
// Story 6.4, CRIT-002: Extract and validate user context safely
type adminUserContext struct {
	adminID       uint
	adminUsername string
}

// extractUserContext safely extracts and validates admin user context from Gin context
// Story 6.4, CRIT-002: Fix type assertion panics with comma-ok pattern
func (h *UserManagementHandler) extractUserContext(c *gin.Context) (adminUserContext, bool) {
	// Extract user ID with type safety check
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return adminUserContext{}, false
	}

	adminID, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user context type"})
		return adminUserContext{}, false
	}

	// Extract username with type safety check
	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Username not found"})
		return adminUserContext{}, false
	}

	adminUsername, ok := username.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid username context type"})
		return adminUserContext{}, false
	}

	// Validate user ID is not zero
	if adminID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
		return adminUserContext{}, false
	}

	// Validate username is not empty
	if adminUsername == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username"})
		return adminUserContext{}, false
	}

	return adminUserContext{
		adminID:       adminID,
		adminUsername: adminUsername,
	}, true
}
