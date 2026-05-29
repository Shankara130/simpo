package handlers

import (
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/dto"
	apiErrors "github.com/vahiiiid/go-rest-api-boilerplate/internal/errors"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/services"
)

// AuthHandler defines authentication handler interface (Story 1.5, Task 4)
type AuthHandler interface {
	Login(c *gin.Context)
}

// authHandler implements AuthHandler
type authHandler struct {
	authService services.AuthInterface
}

// NewAuthHandler creates a new authentication handler (Story 1.5, Task 4)
func NewAuthHandler(authService services.AuthInterface) AuthHandler {
	return &authHandler{
		authService: authService,
	}
}

// Login handles user login via username and password (Story 1.5, AC1, AC4, AC7)
//
//	@Summary		Login user
//	@Description	Authenticate user with username and password, returns JWT token
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.LoginRequest										true	"Login request with username and password"	SchemaExample({"username":"admin","password":"SecurePassword123!"})
//	@Success		200		{object}	errors.Response{success=bool,data=dto.LoginResponse}	"Success response with access token and user info"
//	@Failure		400		{object}	errors.Response{success=bool,error=errors.ErrorInfo}	"Validation error - missing or invalid input"
//	@Failure		401		{object}	errors.Response{success=bool,error=errors.ErrorInfo}	"Invalid credentials"
//	@Failure		403		{object}	errors.Response{success=bool,error=errors.ErrorInfo}	"User account inactive"
//	@Failure		500		{object}	errors.Response{success=bool,error=errors.ErrorInfo}	"Server error"
//	@Router			/api/v1/auth/login [post]
func (h *authHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Check for empty request body specifically (better UX)
		if errors.Is(err, io.EOF) {
			_ = c.Error(apiErrors.BadRequest("Request body cannot be empty"))
			c.Status(http.StatusBadRequest)
			return
		}
		// Other validation errors
		_ = c.Error(apiErrors.FromGinValidation(err))
		return
	}

	// Extract client IP address for audit logging (Story 1.5, AC7)
	ipAddress := c.ClientIP()

	// Call service layer for authentication
	result, err := h.authService.Login(c.Request.Context(), req.Username, req.Password, ipAddress)
	if err != nil {
		// Handle specific error types (Story 1.5, AC3, AC6)
		if errors.Is(err, services.ErrInvalidPassword) || errors.Is(err, services.ErrUserNotFound) {
			_ = c.Error(apiErrors.Unauthorized("Invalid username or password"))
			c.Status(http.StatusUnauthorized) // Explicit status for test compatibility
			return
		}
		if errors.Is(err, services.ErrUserInactive) {
			_ = c.Error(apiErrors.Forbidden("User account is inactive"))
			c.Status(http.StatusForbidden) // Explicit status for test compatibility
			return
		}
		if errors.Is(err, services.ErrEmptyUsername) || errors.Is(err, services.ErrEmptyPassword) {
			_ = c.Error(apiErrors.BadRequest("Username and password are required"))
			c.Status(http.StatusBadRequest) // Explicit status for test compatibility
			return
		}
		_ = c.Error(apiErrors.InternalServerError(err))
		c.Status(http.StatusInternalServerError) // Explicit status for test compatibility
		return
	}

	// Build login response (Story 1.5, AC4)
	response := dto.LoginResponse{
		AccessToken: result.AccessToken,
		TokenType:   "Bearer",
		ExpiresIn:   int(result.ExpiresIn),
		User: dto.UserInfo{
			ID:       result.User.ID,
			Username: result.User.Username,
			Email:    result.User.Email,
			Role:     result.User.Role,
			BranchID: result.User.BranchID,
		},
	}

	c.JSON(http.StatusOK, apiErrors.Success(response))
}
