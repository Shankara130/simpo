package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/models"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/services"
)

// BranchManagementHandler handles branch management operations with audit logging
// Story 6.4, Task 5: Branch Management Audit Integration
type BranchManagementHandler struct {
	auditService services.AuditService
}

// NewBranchManagementHandler creates a new branch management handler
// Story 6.4, Task 5: Handler with audit service dependency
func NewBranchManagementHandler(auditService services.AuditService) *BranchManagementHandler {
	return &BranchManagementHandler{
		auditService: auditService,
	}
}

// Story 6.4: Branch Management DTOs

// CreateBranchRequest represents branch creation request payload (Story 6.4, AC1, AC2)
type CreateBranchRequest struct {
	// Name is the branch name (unique, maximum 100 characters)
	// Example: "Jakarta Central"
	Name string `json:"name" binding:"required,max=100" example:"Jakarta Central"`

	// Location is the branch address
	// Example: "Jl. Sudirman No. 123, Jakarta Pusat"
	Location string `json:"location" binding:"required" example:"Jl. Sudirman No. 123, Jakarta Pusat"`

	// Phone is the branch phone number
	// Example: "+62-21-1234-5678"
	Phone string `json:"phone" binding:"omitempty" example:"+62-21-1234-5678"`

	// Email is the branch email
	// Example: "jakarta@simpo.pharmacy"
	Email string `json:"email" binding:"omitempty,email" example:"jakarta@simpo.pharmacy"`
}

// CreateBranchResponse represents branch creation response (Story 6.4, AC1, AC2)
type CreateBranchResponse struct {
	// ID is the unique identifier for the created branch
	// Example: 5
	ID uint `json:"id" example:"5"`

	// Name is the branch name
	// Example: "Jakarta Central"
	Name string `json:"name" example:"Jakarta Central"`

	// Location is the branch address
	// Example: "Jl. Sudirman No. 123, Jakarta Pusat"
	Location string `json:"location" example:"Jl. Sudirman No. 123, Jakarta Pusat"`

	// Phone is the branch phone number
	// Example: "+62-21-1234-5678"
	Phone string `json:"phone,omitempty" example:"+62-21-1234-5678"`

	// Email is the branch email
	// Example: "jakarta@simpo.pharmacy"
	Email string `json:"email,omitempty" example:"jakarta@simpo.pharmacy"`

	// CreatedAt is the timestamp when branch was created
	// Example: "2026-05-27T12:00:00Z"
	CreatedAt string `json:"created_at" example:"2026-05-27T12:00:00Z"`
}

// UpdateBranchRequest represents branch update request payload (Story 6.4, AC1, AC2)
type UpdateBranchRequest struct {
	// Name is the new branch name
	// Example: "Jakarta Central Branch"
	Name string `json:"name" binding:"omitempty,max=100" example:"Jakarta Central Branch"`

	// Location is the new branch address
	// Example: "Jl. Sudirman No. 456, Jakarta Pusat"
	Location string `json:"location" binding:"omitempty" example:"Jl. Sudirman No. 456, Jakarta Pusat"`

	// Phone is the new branch phone number
	// Example: "+62-21-9876-5432"
	Phone string `json:"phone" binding:"omitempty" example:"+62-21-9876-5432"`

	// Email is the new branch email
	// Example: "jakarta-central@simpo.pharmacy"
	Email string `json:"email" binding:"omitempty,email" example:"jakarta-central@simpo.pharmacy"`

	// Reason is the reason for the update (minimum 5 characters, maximum 500)
	// Example: "Address correction, phone number updated"
	// Code review fix: CRIT-015 - Add max length validation
	Reason string `json:"reason" binding:"required,min=5,max=500" example:"Address correction, phone number updated"`
}

// UpdateBranchResponse represents branch update response (Story 6.4, AC1, AC2)
type UpdateBranchResponse struct {
	// ID is the unique identifier for the branch
	// Example: 5
	ID uint `json:"id" example:"5"`

	// Name is the updated branch name
	// Example: "Jakarta Central Branch"
	Name string `json:"name" example:"Jakarta Central Branch"`

	// Changes is a summary of what was changed
	// Example: "Name: Jakarta Central → Jakarta Central Branch, Location updated"
	Changes string `json:"changes" example:"Name: Jakarta Central → Jakarta Central Branch, Location updated"`

	// UpdatedAt is the timestamp when branch was updated
	// Example: "2026-05-27T12:30:00Z"
	UpdatedAt string `json:"updated_at" example:"2026-05-27T12:30:00Z"`
}

// DeactivateBranchRequest represents branch deactivation request payload (Story 6.4, AC1, AC2)
type DeactivateBranchRequest struct {
	// Reason is the reason for deactivation (minimum 5 characters)
	// Examples: "Branch closed permanently", "Relocated to new address", "Temporary closure for renovation"
	Reason string `json:"reason" binding:"required,min=5" example:"Branch closed permanently"`
}

// DeactivateBranchResponse represents branch deactivation response (Story 6.4, AC1, AC2)
type DeactivateBranchResponse struct {
	// ID is the unique identifier for the deactivated branch
	// Example: 5
	ID uint `json:"id" example:"5"`

	// Name is the branch name
	// Example: "Jakarta Central"
	Name string `json:"name" example:"Jakarta Central"`

	// DeactivatedAt is the timestamp when branch was deactivated
	// Example: "2026-05-27T14:00:00Z"
	DeactivatedAt string `json:"deactivated_at" example:"2026-05-27T14:00:00Z"`

	// Reason is the reason for deactivation
	// Example: "Branch closed permanently"
	Reason string `json:"reason" example:"Branch closed permanently"`
}

// CreateBranch godoc
// @Summary      Create new branch
// @Description  Creates a new pharmacy branch with audit logging (Story 6.4, AC1, AC2)
// @Tags         Branch Management
// @Accept       json
// @Produce      json
// @Param        request body handlers.CreateBranchRequest true "Branch creation request"
// @Security     BearerAuth
// @Success      201  {object}  handlers.CreateBranchResponse  "Branch created successfully"
// @Failure      400  {object}  map[string]string  "Invalid request"
// @Failure      401  {object}  map[string]string  "Unauthorized"
// @Failure      403  {object}  map[string]string  "Forbidden - OWNER/ADMIN only"
// @Failure      409  {object}  map[string]string  "Branch name already exists"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /api/v1/admin/branches [post]
func (h *BranchManagementHandler) CreateBranch(c *gin.Context) {
	// Parse request
	var req CreateBranchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Extract admin user context for audit logging
	userCtx, ok := h.extractUserContext(c)
	if !ok {
		return
	}

	// Get IP address for audit logging
	ipAddress := c.ClientIP()

	// TODO: Implement actual branch creation via branch service/repository
	// For now, create a mock response
	branch := &models.Branch{
		Name:    req.Name,
		Address: req.Location,
		Phone:   req.Phone,
		Email:   req.Email,
	}

	// Log branch creation to audit trail (Story 6.4, AC1, AC2)
	_ = h.auditService.LogBranchCreated(
		c.Request.Context(),
		userCtx.adminID,
		userCtx.adminUsername,
		req.Name,
		req.Location,
		ipAddress,
	)

	response := CreateBranchResponse{
		ID:        1, // Mock ID
		Name:      branch.Name,
		Location:  branch.Address,
		Phone:     branch.Phone,
		Email:     branch.Email,
		CreatedAt: time.Now().Format(time.RFC3339),
	}

	c.JSON(http.StatusCreated, response)
}

// UpdateBranch godoc
// @Summary      Update branch information
// @Description  Updates a branch's information with audit logging (Story 6.4, AC1, AC2)
// @Tags         Branch Management
// @Accept       json
// @Produce      json
// @Param        branch_id path int true "Branch ID"
// @Param        request body handlers.UpdateBranchRequest true "Branch update request"
// @Security     BearerAuth
// @Success      200  {object}  handlers.UpdateBranchResponse  "Branch updated successfully"
// @Failure      400  {object}  map[string]string  "Invalid request"
// @Failure      401  {object}  map[string]string  "Unauthorized"
// @Failure      403  {object}  map[string]string  "Forbidden - OWNER/ADMIN only"
// @Failure      404  {object}  map[string]string  "Branch not found"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /api/v1/admin/branches/{branch_id} [put]
func (h *BranchManagementHandler) UpdateBranch(c *gin.Context) {
	// Extract branch ID from path
	branchIDStr := c.Param("branch_id")
	branchID, err := strconv.ParseUint(branchIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid branch ID"})
		return
	}

	// Parse request
	var req UpdateBranchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Extract admin user context for audit logging
	userCtx, ok := h.extractUserContext(c)
	if !ok {
		return
	}

	// Get IP address for audit logging
	ipAddress := c.ClientIP()

	// TODO: Implement actual branch update via branch service/repository
	// For now, create mock old and new values
	oldName := "Jakarta Central"    // Mock old name
	oldLocation := "Old Address"     // Mock old location
	newName := req.Name
	if newName == "" {
		newName = oldName
	}
	newLocation := req.Location
	if newLocation == "" {
		newLocation = oldLocation
	}

	// Build changes summary for audit log
	changes := ""
	if newName != oldName {
		changes += fmt.Sprintf("Name: %s → %s, ", oldName, newName)
	}
	if newLocation != oldLocation {
		changes += fmt.Sprintf("Location: %s → %s, ", oldLocation, newLocation)
	}
	if req.Phone != "" {
		changes += fmt.Sprintf("Phone updated, ")
	}
	if req.Email != "" {
		changes += fmt.Sprintf("Email updated, ")
	}
	if changes == "" {
		changes = "No changes made"
	} else {
		changes = changes[:len(changes)-2] // Remove trailing comma
	}

	// Log branch update to audit trail (Story 6.4, AC1, AC2)
	_ = h.auditService.LogBranchUpdated(
		c.Request.Context(),
		userCtx.adminID,
		userCtx.adminUsername,
		uint(branchID),
		oldName,
		changes,
		ipAddress,
	)

	response := UpdateBranchResponse{
		ID:        uint(branchID),
		Name:      newName,
		Changes:   changes,
		UpdatedAt: time.Now().Format(time.RFC3339),
	}

	c.JSON(http.StatusOK, response)
}

// DeactivateBranch godoc
// @Summary      Deactivate branch
// @Description  Deactivates a pharmacy branch with audit logging (Story 6.4, AC1, AC2)
// @Tags         Branch Management
// @Accept       json
// @Produce      json
// @Param        branch_id path int true "Branch ID"
// @Param        request body handlers.DeactivateBranchRequest true "Deactivation request"
// @Security     BearerAuth
// @Success      200  {object}  handlers.DeactivateBranchResponse  "Branch deactivated successfully"
// @Failure      400  {object}  map[string]string  "Invalid request"
// @Failure      401  {object}  map[string]string  "Unauthorized"
// @Failure      403  {object}  map[string]string  "Forbidden - OWNER only"
// @Failure      404  {object}  map[string]string  "Branch not found"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /api/v1/admin/branches/{branch_id} [delete]
func (h *BranchManagementHandler) DeactivateBranch(c *gin.Context) {
	// Extract branch ID from path
	branchIDStr := c.Param("branch_id")
	branchID, err := strconv.ParseUint(branchIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid branch ID"})
		return
	}

	// Parse request
	var req DeactivateBranchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Extract admin user context for audit logging
	userCtx, ok := h.extractUserContext(c)
	if !ok {
		return
	}

	// Get IP address for audit logging
	ipAddress := c.ClientIP()

	// TODO: Implement actual branch deactivation via branch service/repository
	// For now, create a mock response
	branchName := "Jakarta Central" // Mock branch name

	// Log branch deactivation to audit trail (Story 6.4, AC1, AC2)
	_ = h.auditService.LogBranchDeactivated(
		c.Request.Context(),
		userCtx.adminID,
		userCtx.adminUsername,
		uint(branchID),
		branchName,
		sanitizeReasonBranch(req.Reason), // Code review fix: CRIT-014
		ipAddress,
	)

	response := DeactivateBranchResponse{
		ID:            uint(branchID),
		Name:          branchName,
		DeactivatedAt: time.Now().Format(time.RFC3339),
		Reason:        sanitizeReasonBranch(req.Reason), // Code review fix: CRIT-014
	}

	c.JSON(http.StatusOK, response)
}

// branchAdminUserContext holds validated admin user context information
type branchAdminUserContext struct {
	adminID      uint
	adminUsername string
}

// extractUserContext safely extracts and validates admin user context from Gin context
func (h *BranchManagementHandler) extractUserContext(c *gin.Context) (branchAdminUserContext, bool) {
	// Extract user ID with type safety check
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return branchAdminUserContext{}, false
	}

	adminID, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user context type"})
		return branchAdminUserContext{}, false
	}

	// Extract username with type safety check
	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Username not found"})
		return branchAdminUserContext{}, false
	}

	adminUsername, ok := username.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid username context type"})
		return branchAdminUserContext{}, false
	}

	// Validate user ID is not zero
	if adminID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
		return branchAdminUserContext{}, false
	}

	// Validate username is not empty
	if adminUsername == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username"})
		return branchAdminUserContext{}, false
	}

	return branchAdminUserContext{
		adminID:      adminID,
		adminUsername: adminUsername,
	}, true
}

// sanitizeReason sanitizes user-provided reason text to prevent injection attacks
// Code review fix: CRIT-014 - Add input sanitization for reason fields
func sanitizeReasonBranch(reason string) string {
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
