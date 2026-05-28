package whitelist

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	apiErrors "github.com/vahiiiid/go-rest-api-boilerplate/internal/errors"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/middleware"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/models"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/services"
)

// Handler handles whitelist HTTP requests
// Story 1.9, Task 3: Whitelist management handlers
type Handler struct {
	service      Service
	auditService services.AuditService
}

// NewHandler creates a new whitelist handler
func NewHandler(service Service) *Handler {
	return &Handler{
		service: service,
	}
}

// SetAuditService injects the audit service dependency
func (h *Handler) SetAuditService(auditService services.AuditService) {
	h.auditService = auditService
}

// AddDomain godoc
// @Summary Add email domain to whitelist
// @Description Add a new email domain to the whitelist for staff self-registration (SYSTEM_ADMIN only)
// @Tags whitelist
// @Accept json
// @Produce json
// @Param request body AddWhitelistEntryRequest true "Add whitelist entry request"
// @Success 201 {object} apiErrors.Response{success=bool,data=WhitelistEntryResponse} "Whitelist entry created"
// @Failure 400 {object} apiErrors.Response{success=bool,error=errors.ErrorInfo} "Validation error"
// @Failure 401 {object} apiErrors.Response{success=bool,error=errors.ErrorInfo} "Unauthorized"
// @Failure 403 {object} apiErrors.Response{success=bool,error=errors.ErrorInfo} "Forbidden - SYSTEM_ADMIN only"
// @Failure 409 {object} apiErrors.Response{success=bool,error=errors.ErrorInfo} "Domain already exists"
// @Failure 500 {object} apiErrors.Response{success=bool,error=errors.ErrorInfo} "Internal server error"
// @Router /api/v1/whitelist [post]
// @Security BearerAuth
func (h *Handler) AddDomain(c *gin.Context) {
	var req AddWhitelistEntryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apiErrors.FromGinValidation(err))
		return
	}

	entry, err := h.service.AddDomain(c.Request.Context(), req)
	if err != nil {
		if err == ErrDomainRequired {
			_ = c.Error(apiErrors.BadRequest("Domain is required"))
			return
		}
		if err == ErrInvalidRole {
			_ = c.Error(apiErrors.BadRequest("Invalid role. Must be one of: SYSTEM_ADMIN, OWNER, CASHIER"))
			return
		}
		if err == ErrDomainAlreadyExists {
			_ = c.Error(apiErrors.Conflict("Domain already exists in whitelist"))
			return
		}
		_ = c.Error(apiErrors.InternalServerError(err))
		return
	}

	// Story 1.9, AC8: Log whitelist domain addition to audit trail
	if h.auditService != nil {
		userID := middleware.GetUserID(c)
		username := middleware.GetUsername(c)
		if err := h.auditService.LogWhitelistChange(c.Request.Context(), userID, username, req.Domain, models.AuditActionWhitelistDomainAdded, c.ClientIP()); err != nil {
			slog.Error("Failed to log whitelist change", "error", err)
		}
	}

	c.JSON(http.StatusCreated, apiErrors.Success(ToWhitelistEntryResponse(entry)))
}

// ListDomains godoc
// @Summary List all whitelisted domains
// @Description Retrieve all email domains in the whitelist (SYSTEM_ADMIN only)
// @Tags whitelist
// @Produce json
// @Success 200 {object} apiErrors.Response{success=bool,data=[]WhitelistEntryResponse} "List of whitelist entries"
// @Failure 401 {object} apiErrors.Response{success=bool,error=errors.ErrorInfo} "Unauthorized"
// @Failure 403 {object} apiErrors.Response{success=bool,error=errors.ErrorInfo} "Forbidden - SYSTEM_ADMIN only"
// @Failure 500 {object} apiErrors.Response{success=bool,error=errors.ErrorInfo} "Internal server error"
// @Router /api/v1/whitelist [get]
// @Security BearerAuth
func (h *Handler) ListDomains(c *gin.Context) {
	entries, err := h.service.ListDomains(c.Request.Context())
	if err != nil {
		_ = c.Error(apiErrors.InternalServerError(err))
		return
	}

	// Convert to response DTO
	responses := make([]WhitelistEntryResponse, len(entries))
	for i := range entries {
		responses[i] = ToWhitelistEntryResponse(&entries[i])
	}

	c.JSON(http.StatusOK, apiErrors.Success(responses))
}

// GetDomain godoc
// @Summary Get whitelist entry by ID
// @Description Retrieve a specific whitelist entry by ID (SYSTEM_ADMIN only)
// @Tags whitelist
// @Produce json
// @Param id path int true "Whitelist entry ID"
// @Success 200 {object} apiErrors.Response{success=bool,data=WhitelistEntryResponse} "Whitelist entry"
// @Failure 400 {object} apiErrors.Response{success=bool,error=errors.ErrorInfo} "Invalid ID"
// @Failure 401 {object} apiErrors.Response{success=bool,error=errors.ErrorInfo} "Unauthorized"
// @Failure 403 {object} apiErrors.Response{success=bool,error=errors.ErrorInfo} "Forbidden - SYSTEM_ADMIN only"
// @Failure 404 {object} apiErrors.Response{success=bool,error=errors.ErrorInfo} "Whitelist entry not found"
// @Failure 500 {object} apiErrors.Response{success=bool,error=errors.ErrorInfo} "Internal server error"
// @Router /api/v1/whitelist/{id} [get]
// @Security BearerAuth
func (h *Handler) GetDomain(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		_ = c.Error(apiErrors.BadRequest("Invalid whitelist entry ID"))
		return
	}

	entry, err := h.service.GetDomain(c.Request.Context(), uint(id))
	if err != nil {
		if err == ErrWhitelistEntryNotFound {
			_ = c.Error(apiErrors.NotFound("Whitelist entry not found"))
			return
		}
		_ = c.Error(apiErrors.InternalServerError(err))
		return
	}

	c.JSON(http.StatusOK, apiErrors.Success(ToWhitelistEntryResponse(entry)))
}

// UpdateDomain godoc
// @Summary Update whitelist entry
// @Description Update a whitelist entry's default role and/or description (SYSTEM_ADMIN only)
// @Tags whitelist
// @Accept json
// @Produce json
// @Param id path int true "Whitelist entry ID"
// @Param request body UpdateWhitelistEntryRequest true "Update whitelist entry request"
// @Success 200 {object} apiErrors.Response{success=bool,data=WhitelistEntryResponse} "Updated whitelist entry"
// @Failure 400 {object} apiErrors.Response{success=bool,error=errors.ErrorInfo} "Validation error"
// @Failure 401 {object} apiErrors.Response{success=bool,error=errors.ErrorInfo} "Unauthorized"
// @Failure 403 {object} apiErrors.Response{success=bool,error=errors.ErrorInfo} "Forbidden - SYSTEM_ADMIN only"
// @Failure 404 {object} apiErrors.Response{success=bool,error=errors.ErrorInfo} "Whitelist entry not found"
// @Failure 500 {object} apiErrors.Response{success=bool,error=errors.ErrorInfo} "Internal server error"
// @Router /api/v1/whitelist/{id} [put]
// @Security BearerAuth
func (h *Handler) UpdateDomain(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		_ = c.Error(apiErrors.BadRequest("Invalid whitelist entry ID"))
		return
	}

	var req UpdateWhitelistEntryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apiErrors.FromGinValidation(err))
		return
	}

	entry, err := h.service.UpdateDomain(c.Request.Context(), uint(id), req)
	if err != nil {
		if err == ErrWhitelistEntryNotFound {
			_ = c.Error(apiErrors.NotFound("Whitelist entry not found"))
			return
		}
		if err == ErrInvalidRole {
			_ = c.Error(apiErrors.BadRequest("Invalid role. Must be one of: SYSTEM_ADMIN, OWNER, CASHIER"))
			return
		}
		_ = c.Error(apiErrors.InternalServerError(err))
		return
	}

	// Story 1.9, AC8: Log whitelist domain update to audit trail
	if h.auditService != nil {
		userID := middleware.GetUserID(c)
		username := middleware.GetUsername(c)
		if err := h.auditService.LogWhitelistChange(c.Request.Context(), userID, username, entry.Domain, models.AuditActionWhitelistDomainUpdated, c.ClientIP()); err != nil {
			slog.Error("Failed to log whitelist change", "error", err)
		}
	}

	c.JSON(http.StatusOK, apiErrors.Success(ToWhitelistEntryResponse(entry)))
}

// DeleteDomain godoc
// @Summary Delete whitelist entry
// @Description Remove a domain from the whitelist (SYSTEM_ADMIN only)
// @Tags whitelist
// @Produce json
// @Param id path int true "Whitelist entry ID"
// @Success 204 "Whitelist entry deleted"
// @Failure 400 {object} apiErrors.Response{success=bool,error=errors.ErrorInfo} "Invalid ID"
// @Failure 401 {object} apiErrors.Response{success=bool,error=errors.ErrorInfo} "Unauthorized"
// @Failure 403 {object} apiErrors.Response{success=bool,error=errors.ErrorInfo} "Forbidden - SYSTEM_ADMIN only"
// @Failure 404 {object} apiErrors.Response{success=bool,error=errors.ErrorInfo} "Whitelist entry not found"
// @Failure 500 {object} apiErrors.Response{success=bool,error=errors.ErrorInfo} "Internal server error"
// @Router /api/v1/whitelist/{id} [delete]
// @Security BearerAuth
func (h *Handler) DeleteDomain(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		_ = c.Error(apiErrors.BadRequest("Invalid whitelist entry ID"))
		return
	}

	err = h.service.DeleteDomain(c.Request.Context(), uint(id))
	if err != nil {
		if err == ErrWhitelistEntryNotFound {
			_ = c.Error(apiErrors.NotFound("Whitelist entry not found"))
			return
		}
		_ = c.Error(apiErrors.InternalServerError(err))
		return
	}

	// Story 1.9, AC8: Log whitelist domain deletion to audit trail
	// Note: We need to get the entry before deletion, but the service.DeleteDomain doesn't return it
	// For now, we'll use the ID to reference the deleted entry
	if h.auditService != nil {
		userID := middleware.GetUserID(c)
		username := middleware.GetUsername(c)
		// Use ID as domain reference since we don't have the domain after deletion
		if err := h.auditService.LogWhitelistChange(c.Request.Context(), userID, username, fmt.Sprintf("ID:%d", id), models.AuditActionWhitelistDomainDeleted, c.ClientIP()); err != nil {
			slog.Error("Failed to log whitelist change", "error", err)
		}
	}

	c.Status(http.StatusNoContent)
}
