package user

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/auth"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/contextutil"
	apiErrors "github.com/vahiiiid/go-rest-api-boilerplate/internal/errors"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/middleware"
)

// AuditLogger defines the interface for audit logging (Story 1.7, AC7)
// This interface is defined here to avoid import cycles with services package
type AuditLogger interface {
	LogUserCreation(ctx context.Context, adminID uint, createdUserID uint, adminUsername string, createdUsername string, ipAddress string) error
	// Story 1.9: Audit logging for whitelist and self-registration
	LogSelfRegistration(ctx context.Context, userID uint, email string, domain string, ipAddress string) error
	LogEmailVerification(ctx context.Context, userID uint, email string, ipAddress string) error
}

// Handler handles user-related HTTP requests
type Handler struct {
	userService     Service
	authService     auth.Service
	auditLogger     AuditLogger
	sessionManager  *middleware.SessionManager // Story 1.8: Session tracking and blocklist
}

// NewHandler creates a new user handler
func NewHandler(userService Service, authService auth.Service, auditLogger AuditLogger) *Handler {
	return &Handler{
		userService:    userService,
		authService:    authService,
		auditLogger:    auditLogger,
		sessionManager: nil, // Will be set after creation if needed
	}
}

// SetSessionManager sets the session manager (used for dependency injection)
// Story 1.8, Task 1: Session tracking mechanism
func (h *Handler) SetSessionManager(sessionManager *middleware.SessionManager) {
	h.sessionManager = sessionManager
}

// P4 FIX: ValidateSessionManager checks if session manager is initialized
// This is called by handlers that require session tracking (logout, refresh)
func (h *Handler) ValidateSessionManager() error {
	if h.sessionManager == nil {
		return errors.New("session manager not available - session tracking is required")
	}
	return nil
}

// RefreshResponse represents token refresh response (Story 1.8, AC4)
type RefreshResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
}

// Register godoc
// @Summary Register a new user
// @Description Register a new user with name, email and password, returns access and refresh tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Registration request"
// @Success 200 {object} errors.Response{success=bool,data=AuthResponse} "Success response with user data and tokens"
// @Failure 400 {object} errors.Response{success=bool,error=errors.ErrorInfo} "Validation error"
// @Failure 409 {object} errors.Response{success=bool,error=errors.ErrorInfo} "Email already exists"
// @Failure 500 {object} errors.Response{success=bool,error=errors.ErrorInfo} "Failed to register user or generate token"
// @Router /api/v1/auth/register [post]
func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apiErrors.FromGinValidation(err))
		return
	}

	user, err := h.userService.RegisterUser(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, ErrEmailExists) {
			_ = c.Error(apiErrors.Conflict("Email already exists"))
			return
		}
		_ = c.Error(apiErrors.InternalServerError(err))
		return
	}

	tokenPair, err := h.authService.GenerateTokenPair(c.Request.Context(), user.ID, user.Email, user.Name)
	if err != nil {
		_ = c.Error(apiErrors.InternalServerError(err))
		return
	}

	c.JSON(http.StatusOK, apiErrors.Success(AuthResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		TokenType:    tokenPair.TokenType,
		ExpiresIn:    tokenPair.ExpiresIn,
		User:         ToUserResponse(user),
	}))
}

// Login godoc
// @Summary Login user
// @Description Authenticate user with email and password, returns access and refresh tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login request"
// @Success 200 {object} errors.Response{success=bool,data=AuthResponse} "Success response with user data and tokens"
// @Failure 400 {object} errors.Response{success=bool,error=errors.ErrorInfo} "Validation error"
// @Failure 401 {object} errors.Response{success=bool,error=errors.ErrorInfo} "Invalid email or password"
// @Failure 500 {object} errors.Response{success=bool,error=errors.ErrorInfo} "Failed to authenticate user or generate token"
// @Router /api/v1/auth/login [post]
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apiErrors.FromGinValidation(err))
		return
	}

	user, err := h.userService.AuthenticateUser(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			_ = c.Error(apiErrors.Unauthorized("Invalid email or password"))
			return
		}
		_ = c.Error(apiErrors.InternalServerError(err))
		return
	}

	tokenPair, err := h.authService.GenerateTokenPair(c.Request.Context(), user.ID, user.Email, user.Name)
	if err != nil {
		_ = c.Error(apiErrors.InternalServerError(err))
		return
	}

	c.JSON(http.StatusOK, apiErrors.Success(AuthResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		TokenType:    tokenPair.TokenType,
		ExpiresIn:    tokenPair.ExpiresIn,
		User:         ToUserResponse(user),
	}))
}

// GetUser godoc
// @Summary Get user by ID
// @Description Get a user by their ID (requires authentication)
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Security BearerAuth
// @Success 200 {object} errors.Response{success=bool,data=UserResponse} "Success response with user data"
// @Failure 400 {object} errors.Response{success=bool,error=errors.ErrorInfo} "Invalid user ID"
// @Failure 403 {object} errors.Response{success=bool,error=errors.ErrorInfo} "Forbidden user ID"
// @Failure 404 {object} errors.Response{success=bool,error=errors.ErrorInfo} "User not found"
// @Failure 429 {object} errors.Response{success=bool,error=errors.ErrorInfo} "Rate limit exceeded"
// @Failure 500 {object} errors.Response{success=bool,error=errors.ErrorInfo} "Failed to get user"
// @Router /api/v1/users/{id} [get]
func (h *Handler) GetUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		_ = c.Error(apiErrors.BadRequest("Invalid user ID"))
		return
	}

	// Story 1.9: Validate ID is within reasonable range
	if !isValidID(uint(id)) {
		_ = c.Error(apiErrors.BadRequest("Invalid user ID"))
		return
	}

	if !contextutil.CanAccessUser(c, uint(id)) {
		_ = c.Error(apiErrors.Forbidden("Forbidden user ID"))
		return
	}

	user, err := h.userService.GetUserByID(c.Request.Context(), uint(id))
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			_ = c.Error(apiErrors.NotFound("User not found"))
			return
		}
		_ = c.Error(apiErrors.InternalServerError(err))
		return
	}

	c.JSON(http.StatusOK, apiErrors.Success(ToUserResponse(user)))
}

// UpdateUser godoc
// @Summary Update user
// @Description Update user information (requires authentication)
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param request body UpdateUserRequest true "Update request"
// @Security BearerAuth
// @Success 200 {object} errors.Response{success=bool,data=UserResponse} "Success response with updated user data"
// @Failure 400 {object} errors.Response{success=bool,error=errors.ErrorInfo} "Invalid user ID or Validation error"
// @Failure 403 {object} errors.Response{success=bool,error=errors.ErrorInfo} "Forbidden user ID"
// @Failure 404 {object} errors.Response{success=bool,error=errors.ErrorInfo} "User not found"
// @Failure 409 {object} errors.Response{success=bool,error=errors.ErrorInfo} "Email already exists"
// @Failure 429 {object} errors.Response{success=bool,error=errors.ErrorInfo} "Rate limit exceeded"
// @Failure 500 {object} errors.Response{success=bool,error=errors.ErrorInfo} "Failed to update user"
// @Router /api/v1/users/{id} [put]
func (h *Handler) UpdateUser(c *gin.Context) {
	// Parse ID from URL
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		_ = c.Error(apiErrors.BadRequest("Invalid user ID"))
		return
	}

	// Story 1.9: Validate ID is within reasonable range
	if !isValidID(uint(id)) {
		_ = c.Error(apiErrors.BadRequest("Invalid user ID"))
		return
	}

	// Authorization check
	if !contextutil.CanAccessUser(c, uint(id)) {
		_ = c.Error(apiErrors.Forbidden("Forbidden user ID"))
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apiErrors.FromGinValidation(err))
		return
	}

	user, err := h.userService.UpdateUser(c.Request.Context(), uint(id), req)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			_ = c.Error(apiErrors.NotFound("User not found"))
			return
		}
		if errors.Is(err, ErrEmailExists) {
			_ = c.Error(apiErrors.Conflict("Email already exists"))
			return
		}
		_ = c.Error(apiErrors.InternalServerError(err))
		return
	}

	c.JSON(http.StatusOK, apiErrors.Success(ToUserResponse(user)))
}

// DeleteUser godoc
// @Summary Delete user
// @Description Delete a user by ID (requires authentication)
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Security BearerAuth
// @Success 204
// @Failure 400 {object} errors.Response{success=bool,error=errors.ErrorInfo} "Invalid user ID"
// @Failure 403 {object} errors.Response{success=bool,error=errors.ErrorInfo} "Forbidden user ID"
// @Failure 404 {object} errors.Response{success=bool,error=errors.ErrorInfo} "User not found"
// @Failure 429 {object} errors.Response{success=bool,error=errors.ErrorInfo} "Rate limit exceeded"
// @Failure 500 {object} errors.Response{success=bool,error=errors.ErrorInfo} "Failed to delete user"
// @Router /api/v1/users/{id} [delete]
func (h *Handler) DeleteUser(c *gin.Context) {
	// Parse ID from URL
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		_ = c.Error(apiErrors.BadRequest("Invalid user ID"))
		return
	}

	// Story 1.9: Validate ID is within reasonable range
	if !isValidID(uint(id)) {
		_ = c.Error(apiErrors.BadRequest("Invalid user ID"))
		return
	}

	// Authorization check
	if !contextutil.CanAccessUser(c, uint(id)) {
		_ = c.Error(apiErrors.Forbidden("Forbidden user ID"))
		return
	}

	if err := h.userService.DeleteUser(c.Request.Context(), uint(id)); err != nil {
		if errors.Is(err, ErrUserNotFound) {
			_ = c.Error(apiErrors.NotFound("User not found"))
			return
		}
		_ = c.Error(apiErrors.InternalServerError(err))
		return
	}

	c.Status(http.StatusNoContent)
}

// RefreshToken godoc
// @Summary Refresh access token
// @Description Exchange refresh token for new access and refresh tokens with automatic rotation
// @Tags auth
// @Accept json
// @Produce json
// @Param request body auth.RefreshTokenRequest true "Refresh token request"
// @Success 200 {object} errors.Response{success=bool,data=auth.TokenPairResponse} "Success response with new token pair"
// @Failure 400 {object} errors.Response{success=bool,error=errors.ErrorInfo} "Validation error"
// @Failure 401 {object} errors.Response{success=bool,error=errors.ErrorInfo} "Invalid or expired refresh token"
// @Failure 403 {object} errors.Response{success=bool,error=errors.ErrorInfo} "Token reuse detected - all tokens revoked"
// @Failure 500 {object} errors.Response{success=bool,error=errors.ErrorInfo} "Failed to refresh token"
// @Router /api/v1/auth/refresh [post]
func (h *Handler) RefreshToken(c *gin.Context) {
	// Story 1.8, AC4: Token refresh uses current JWT token from Authorization header
	// Extract token from context (set by SessionAuthMiddleware)
	claims, exists := c.Get("user")
	if !exists {
		_ = c.Error(apiErrors.Unauthorized("user not authenticated"))
		return
	}

	userClaims, ok := claims.(*auth.Claims)
	if !ok {
		_ = c.Error(apiErrors.Unauthorized("invalid user claims"))
		return
	}

	// Get token ID from context (set by SessionAuthMiddleware)
	tokenID := auth.GetTokenID(c)
	if tokenID == "" {
		_ = c.Error(apiErrors.Unauthorized("token ID not found"))
		return
	}

	// SECURITY FIX: Revoke old token FIRST to prevent race condition
	// This prevents simultaneous refresh requests from both succeeding
	if h.sessionManager == nil {
		_ = c.Error(apiErrors.InternalServerError(errors.New("session manager not available")))
		return
	}

	// P1 FIX: Calculate actual remaining TTL from JWT expiration claim
	// This prevents setting wrong TTL on revoked tokens
	var oldTokenTTL time.Duration
	if !userClaims.ExpiresAt.IsZero() {
		oldTokenTTL = time.Until(userClaims.ExpiresAt)
		if oldTokenTTL < 0 {
			oldTokenTTL = 0 // Token already expired
		}
	} else {
		oldTokenTTL = 8 * time.Hour // Fallback to default TTL
	}

	// Revoke the old token before generating new one
	if err := h.sessionManager.RevokeToken(c.Request.Context(), tokenID, oldTokenTTL); err != nil {
		// Log error but don't fail - the token will expire naturally
		slog.Warn("Failed to revoke token during refresh", "error", err, "token_id", tokenID)
	}

	// P3/P6 FIX: Delete old session data to prevent orphaned sessions
	if err := h.sessionManager.DeleteSession(c.Request.Context(), userClaims.UserID, tokenID); err != nil {
		// Log error but don't fail - session will expire via TTL
		slog.Warn("Failed to delete session during refresh", "error", err, "user_id", userClaims.UserID, "token_id", tokenID)
	}

	// Story 1.8, AC4: Generate new token with same user info and updated expiration
	// Use authService to generate new token for the user
	newToken, err := h.authService.GenerateToken(userClaims.UserID, userClaims.Email, userClaims.Name)
	if err != nil {
		// Rollback: If token generation fails, we've already revoked the old token
		// The user will need to re-authenticate - this is acceptable security behavior
		_ = c.Error(apiErrors.InternalServerError(err))
		return
	}

	// P5 FIX: Calculate ExpiresIn from actual JWT TTL configuration
	// Parse the new token to get actual expiration time
	expiresIn := int64(8 * time.Hour.Seconds()) // Default fallback
	newClaims, err := h.authService.ValidateToken(newToken)
	if err == nil && !newClaims.ExpiresAt.IsZero() {
		expiresIn = int64(time.Until(newClaims.ExpiresAt).Seconds())
		if expiresIn < 0 {
			expiresIn = 0 // Shouldn't happen, but safety check
		}
	}

	// Story 1.8, AC4: Return new token to client
	c.JSON(http.StatusOK, apiErrors.Success(RefreshResponse{
		AccessToken: newToken,
		TokenType:   "Bearer",
		ExpiresIn:   expiresIn,
	}))
}

// Logout godoc
// @Summary Logout user
// @Description Invalidate current JWT token (Story 1.8, AC5)
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} errors.Response{success=bool,data=object} "Successfully logged out"
// @Failure 401 {object} errors.Response{success=bool,error=errors.ErrorInfo} "Unauthorized"
// @Failure 500 {object} errors.Response{success=bool,error=errors.ErrorInfo} "Failed to logout"
// @Router /api/v1/auth/logout [post]
func (h *Handler) Logout(c *gin.Context) {
	// Story 1.8, AC5: Logout invalidates the current JWT token immediately
	// Extract token ID from context (set by SessionAuthMiddleware)
	tokenID := auth.GetTokenID(c)
	if tokenID == "" {
		_ = c.Error(apiErrors.Unauthorized("token ID not found"))
		return
	}

	// Story 1.8, AC6: Get user info for audit logging
	userID := contextutil.GetUserID(c)
	if userID == 0 {
		_ = c.Error(apiErrors.Unauthorized("user not authenticated"))
		return
	}

	// SECURITY FIX: Require session manager for logout - session tracking is mandatory
	if h.sessionManager == nil {
		_ = c.Error(apiErrors.InternalServerError(errors.New("session manager not available")))
		return
	}

	// P1 FIX: Calculate remaining token lifetime from JWT claims
	// Get user claims to extract expiration time
	var tokenTTL time.Duration
	if claims, exists := c.Get("user"); exists {
		if userClaims, ok := claims.(*auth.Claims); ok {
			// P6 FIX: Delete session data to prevent orphaned sessions
			if err := h.sessionManager.DeleteSession(c.Request.Context(), userClaims.UserID, tokenID); err != nil {
				// Log error but don't fail - session will expire via TTL
				slog.Warn("Failed to delete session during logout", "error", err, "user_id", userClaims.UserID, "token_id", tokenID)
			}

			if !userClaims.ExpiresAt.IsZero() {
				tokenTTL = time.Until(userClaims.ExpiresAt)
				if tokenTTL < 0 {
					tokenTTL = 0 // Token already expired
				}
			}

			// Story 1.8, AC6: Log logout action to audit trail
			ipAddress := c.ClientIP()
			// Log logout action using structured logging
			// Format: action=LOGOUT user_id=<id> username=<user> token_id=<token> ip=<ip>
			slog.Info("AUDIT", "action", "LOGOUT", "user_id", userClaims.UserID, "username", userClaims.Email, "token_id", tokenID, "ip", ipAddress, "outcome", "success")
		}
	}
	// Fallback to 8 hours if we can't determine expiration
	if tokenTTL == 0 {
		tokenTTL = 8 * time.Hour
	}

	// Revoke the token
	if err := h.sessionManager.RevokeToken(c.Request.Context(), tokenID, tokenTTL); err != nil {
		_ = c.Error(apiErrors.InternalServerError(err))
		return
	}

	// Story 1.8, AC5: Return 200 OK on success
	c.JSON(http.StatusOK, apiErrors.Success(gin.H{"message": "Successfully logged out"}))
}

// RegisterStaff godoc
// @Summary Register a new staff member via self-registration
// @Description Self-registration for staff with whitelisted email domains (Public endpoint)
// @Tags auth
// @Accept json
// @Produce json
// @Param request body StaffRegistrationRequest true "Staff self-registration request"
// @Success 201 {object} apiErrors.Response{success=bool,data=StaffRegistrationResponse} "User registered with PENDING status"
// @Failure 400 {object} apiErrors.Response{success=bool,error=errors.ErrorInfo} "Validation error or duplicate username/email"
// @Failure 403 {object} apiErrors.Response{success=bool,error=errors.ErrorInfo} "Email domain not whitelisted"
// @Failure 500 {object} apiErrors.Response{success=bool,error=errors.ErrorInfo} "Internal server error"
// @Router /api/v1/auth/register-staff [post]
func (h *Handler) RegisterStaff(c *gin.Context) {
	var req StaffRegistrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apiErrors.FromGinValidation(err))
		return
	}

	user, token, err := h.userService.RegisterStaff(c.Request.Context(), req)
	if err != nil {
		if err == ErrDomainNotWhitelisted {
			_ = c.Error(apiErrors.Forbidden("Email domain is not approved for self-registration. Please contact your system administrator."))
			return
		}
		if err == ErrUsernameExists {
			_ = c.Error(apiErrors.Conflict("Username already exists"))
			return
		}
		if err == ErrEmailExists {
			_ = c.Error(apiErrors.Conflict("Email already exists"))
			return
		}
		_ = c.Error(apiErrors.InternalServerError(err))
		return
	}

	// Story 1.9, AC9: Return response with verification_sent flag
	response := StaffRegistrationResponse{
		ID:              user.ID,
		Username:        user.Username,
		Email:           user.Email,
		Role:            user.Role,
		Status:          user.Status,
		VerificationSent: token != "",
		CreatedAt:       user.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}

	// Story 1.9, AC8: Log self-registration to audit trail
	if h.auditLogger != nil {
		// Extract domain from email for audit logging
		domain := extractDomainFromEmail(user.Email)
		if err := h.auditLogger.LogSelfRegistration(c.Request.Context(), user.ID, user.Email, domain, c.ClientIP()); err != nil {
			// Log error but don't fail the request
			slog.Warn("Failed to log self-registration", "error", err, "user_id", user.ID, "email", user.Email, "domain", domain, "ip", c.ClientIP())
		}
	}

	c.JSON(http.StatusCreated, apiErrors.Success(response))
}

// VerifyEmail godoc
// @Summary Verify email and activate account
// @Description Verify email verification token and activate user account (Public endpoint)
// @Tags auth
// @Accept json
// @Produce json
// @Param request body VerifyEmailRequest true "Email verification request"
// @Success 200 {object} apiErrors.Response{success=bool,data=UserResponse} "Account activated successfully"
// @Failure 400 {object} apiErrors.Response{success=bool,error=errors.ErrorInfo} "Invalid or expired verification token"
// @Failure 500 {object} apiErrors.Response{success=bool,error=errors.ErrorInfo} "Internal server error"
// @Router /api/v1/auth/verify-email [post]
func (h *Handler) VerifyEmail(c *gin.Context) {
	var req VerifyEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apiErrors.FromGinValidation(err))
		return
	}

	user, err := h.userService.VerifyEmail(c.Request.Context(), req.Token)
	if err != nil {
		if err == ErrInvalidVerificationToken {
			_ = c.Error(apiErrors.BadRequest("Invalid or expired verification token. Please request a new verification email."))
			return
		}
		_ = c.Error(apiErrors.InternalServerError(err))
		return
	}

	// Story 1.9, AC8: Log email verification to audit trail
	if h.auditLogger != nil {
		if err := h.auditLogger.LogEmailVerification(c.Request.Context(), user.ID, user.Email, c.ClientIP()); err != nil {
			// Log error but don't fail the request
			slog.Warn("Failed to log email verification", "error", err, "user_id", user.ID, "email", user.Email, "ip", c.ClientIP())
		}
	}

	c.JSON(http.StatusOK, apiErrors.Success(ToUserResponse(user)))
}

// GetMe godoc
// @Summary Get current user
// @Description Get the currently authenticated user's information with roles
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} errors.Response{success=bool,data=UserResponse} "Success response with current user data"
// @Failure 401 {object} errors.Response{success=bool,error=errors.ErrorInfo} "Unauthorized"
// @Failure 500 {object} errors.Response{success=bool,error=errors.ErrorInfo} "Failed to get user"
// @Router /api/v1/auth/me [get]
func (h *Handler) GetMe(c *gin.Context) {
	userID := contextutil.GetUserID(c)
	if userID == 0 {
		_ = c.Error(apiErrors.Unauthorized("User not authenticated"))
		return
	}

	user, err := h.userService.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			_ = c.Error(apiErrors.NotFound("User not found"))
			return
		}
		_ = c.Error(apiErrors.InternalServerError(err))
		return
	}

	c.JSON(http.StatusOK, apiErrors.Success(ToUserResponse(user)))
}

// ListUsers godoc
// @Summary List all users (Admin only)
// @Description Get paginated list of all users with optional filtering (requires admin role)
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page (max 100)" default(20)
// @Param role query string false "Filter by role (user or admin)"
// @Param search query string false "Search by name or email"
// @Param sort query string false "Sort by field (created_at, updated_at, name, email)" default(created_at)
// @Param order query string false "Sort order (asc or desc)" default(desc)
// @Success 200 {object} errors.Response{success=bool,data=UserListResponse} "Success response with paginated user list"
// @Failure 400 {object} errors.Response{success=bool,error=errors.ErrorInfo} "Invalid parameters"
// @Failure 403 {object} errors.Response{success=bool,error=errors.ErrorInfo} "Admin access required"
// @Failure 500 {object} errors.Response{success=bool,error=errors.ErrorInfo} "Failed to list users"
// @Router /api/v1/admin/users [get]
func (h *Handler) ListUsers(c *gin.Context) {
	pagination := middleware.ParsePaginationParams(c)
	filters := ParseUserFilters(c)

	users, total, err := h.userService.ListUsers(c.Request.Context(), filters, pagination.Page, pagination.PerPage)
	if err != nil {
		if errors.Is(err, ErrInvalidRole) {
			_ = c.Error(apiErrors.BadRequest("Invalid role filter"))
			return
		}
		_ = c.Error(apiErrors.InternalServerError(err))
		return
	}

	userResponses := make([]UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = ToUserResponse(&user)
	}

	totalPages := int(total) / pagination.PerPage
	if int(total)%pagination.PerPage > 0 {
		totalPages++
	}

	response := UserListResponse{
		Users:      userResponses,
		Total:      total,
		Page:       pagination.Page,
		PerPage:    pagination.PerPage,
		TotalPages: totalPages,
	}

	c.JSON(http.StatusOK, apiErrors.Success(response))
}

// CreateUser godoc
// @Summary Create a new user (Admin only)
// @Description Create a new user with role assignment. Only users with SYSTEM_ADMIN role can access this endpoint.
// Required fields: username (min 3 chars), password (min 8 chars), email (valid format), role (SYSTEM_ADMIN, OWNER, or CASHIER).
// Optional field: branch_id (required for CASHIER role, must reference existing branch).
// User is created with ACTIVE status. Password is hashed using bcrypt (cost factor 12).
// Story 1.7, AC1-8: User registration with admin approval.
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateUserRequest true "User creation request" SchemaExample(true, "Example SYSTEM_ADMIN user", "{\"username\":\"newadmin\",\"password\":\"SecurePass123!\",\"email\":\"newadmin@example.com\",\"role\":\"SYSTEM_ADMIN\"}") SchemaExample(true, "Example CASHIER user", "{\"username\":\"newcashier\",\"password\":\"SecurePass123!\",\"email\":\"cashier@example.com\",\"role\":\"CASHIER\",\"branch_id\":1}")
// @Success 201 {object} errors.Response{success=bool,data=CreateUserResponse} "User created successfully"
// @Failure 400 {object} errors.Response{success=bool,error=errors.ErrorInfo} "Invalid request or validation error (missing fields, invalid format, invalid role, missing branch_id for CASHIER)"
// @Failure 401 {object} errors.Response{success=bool,error=errors.ErrorInfo} "Unauthorized - user not authenticated"
// @Failure 403 {object} errors.Response{success=bool,error=errors.ErrorInfo} "Forbidden - insufficient permissions (SYSTEM_ADMIN role required)"
// @Failure 409 {object} errors.Response{success=bool,error=errors.ErrorInfo} "Conflict - username or email already exists"
// @Failure 500 {object} errors.Response{success=bool,error=errors.ErrorInfo} "Internal server error"
// @Router /api/v1/users [post]
func (h *Handler) CreateUser(c *gin.Context) {
	// Extract admin user ID from JWT context
	adminID := contextutil.GetUserID(c)
	if adminID == 0 {
		_ = c.Error(apiErrors.Unauthorized("User not authenticated"))
		return
	}

	// Bind request to DTO
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apiErrors.FromGinValidation(err))
		return
	}

	// Call service layer
	user, err := h.userService.RegisterUserForAdmin(c.Request.Context(), req, adminID)
	if err != nil {
		// Handle specific errors
		if errors.Is(err, ErrInvalidRoleForCreate) {
			_ = c.Error(apiErrors.BadRequest("Invalid role. Must be one of: SYSTEM_ADMIN, OWNER, CASHIER"))
			return
		}
		if errors.Is(err, ErrUsernameExists) {
			_ = c.Error(apiErrors.Conflict("Username already exists"))
			return
		}
		if errors.Is(err, ErrEmailExists) {
			_ = c.Error(apiErrors.Conflict("Email already exists"))
			return
		}
		if errors.Is(err, ErrBranchIDRequired) {
			_ = c.Error(apiErrors.BadRequest("branch_id is required for CASHIER role"))
			return
		}
		// Generic internal server error
		_ = c.Error(apiErrors.InternalServerError(err))
		return
	}

	// Story 1.7, AC7: Log user creation action to audit trail
	if h.auditLogger != nil {
		// Get admin username from context for audit log
		adminUsername := contextutil.GetUserName(c)
		if adminUsername == "" {
			// Log warning and return error - admin username is required for audit
			_ = c.Error(apiErrors.InternalServerError(errors.New("admin username not found in context for audit logging")))
			return
		}
		ipAddress := c.ClientIP()
		if err := h.auditLogger.LogUserCreation(c.Request.Context(), adminID, user.ID, adminUsername, user.Username, ipAddress); err != nil {
			// Log the audit error but don't fail the request - audit is asynchronous
			// The user was already created successfully
			_ = c.Error(apiErrors.InternalServerError(err))
		}
	}

	// Return 201 Created with user response
	c.JSON(http.StatusCreated, apiErrors.Success(ToCreateUserResponse(user)))
}

// extractDomainFromEmail extracts the domain from an email address
// Story 1.9: Helper for audit logging
func extractDomainFromEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return ""
	}
	return parts[1]
}

// isValidID checks if an ID is within a reasonable range
// Story 1.9: ID validation to prevent potential issues with extremely large IDs
func isValidID(id uint) bool {
	// IDs should be within reasonable database range
	// Most databases use uint32 for IDs, so we check against that
	return id > 0 && id <= 4294967295 // Max uint32 value
}
