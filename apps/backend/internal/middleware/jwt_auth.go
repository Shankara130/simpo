package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// Role constants for simpo RBAC system
// Defined here to avoid import cycle with user package
const (
	RoleSystemAdmin = "SYSTEM_ADMIN"
	RoleOwner       = "OWNER"
	RoleCashier     = "CASHIER"
)

// TokenValidator defines interface for token validation (avoid import cycle)
// This allows us to avoid importing internal/services which creates circular dependency
type TokenValidator interface {
	ValidateToken(tokenString string) (*JWTClaims, error)
}

// JWTClaims represents JWT token claims with role and branch_id
// Story 1.6, AC2: System extracts user role and branch_id from JWT token claims
// Note: Defined here to avoid import cycle with services package
type JWTClaims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	BranchID *uint  `json:"branch_id,omitempty"`
}

// UserContext represents user information stored in request context
// Story 1.6, AC2: Role and branch_id are stored in request context for downstream use
type UserContext struct {
	UserID   uint
	Username string
	Email    string
	Role     string
	BranchID *uint
}

// Context keys for storing user information in Gin context
const (
	UserContextKey = "user_context"
)

// RFC7807Error represents an RFC 7807 compliant error response
// Story 1.6, AC1: Return 401/403 with RFC 7807 format
type RFC7807Error struct {
	Type     string `json:"type"`
	Title    string `json:"title"`
	Status   int    `json:"status"`
	Detail   string `json:"detail"`
	Instance string `json:"instance"`
}

// sendRFC7807Error sends an RFC 7807 compliant error response
func sendRFC7807Error(c *gin.Context, status int, title, detail, instance string) {
	c.JSON(status, RFC7807Error{
		Type:     fmt.Sprintf("https://api.simpo.com/errors/%s", getStatusType(status)),
		Title:    title,
		Status:   status,
		Detail:   detail,
		Instance: instance,
	})
}

// getStatusType returns error type string for HTTP status
func getStatusType(status int) string {
	switch status {
	case http.StatusUnauthorized:
		return "unauthorized"
	case http.StatusForbidden:
		return "forbidden"
	case http.StatusBadRequest:
		return "bad-request"
	case http.StatusNotFound:
		return "not-found"
	default:
		return "error"
	}
}

// getRequestPath returns the current request path
func getRequestPath(c *gin.Context) string {
	if c.Request == nil || c.Request.URL == nil {
		return ""
	}
	return c.Request.URL.Path
}

// JWTAuthMiddleware creates middleware that validates JWT tokens
// Story 1.6, AC1: System validates JWT token on every protected API request
// Story 1.6, AC2: Extracts user role and branch_id from token claims
func JWTAuthMiddleware(tokenValidator TokenValidator) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			sendRFC7807Error(c, http.StatusUnauthorized, "Unauthorized", "authorization header required", getRequestPath(c))
			c.Abort()
			return
		}

		// Check Bearer scheme format
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			sendRFC7807Error(c, http.StatusUnauthorized, "Unauthorized", "invalid authorization header format", getRequestPath(c))
			c.Abort()
			return
		}

		tokenString := parts[1]
		if tokenString == "" {
			sendRFC7807Error(c, http.StatusUnauthorized, "Unauthorized", "token cannot be empty", getRequestPath(c))
			c.Abort()
			return
		}

		// Validate token and extract claims
		claims, err := tokenValidator.ValidateToken(tokenString)
		if err != nil {
			sendRFC7807Error(c, http.StatusUnauthorized, "Unauthorized", "invalid or expired token", getRequestPath(c))
			c.Abort()
			return
		}

		// Story 1.6, AC2: Check that required claims are present
		if claims.Role == "" {
			sendRFC7807Error(c, http.StatusForbidden, "Forbidden", "missing role claim in token", getRequestPath(c))
			c.Abort()
			return
		}

		// Store user context for downstream middleware
		userCtx := &UserContext{
			UserID:   claims.UserID,
			Username: claims.Username,
			Email:    claims.Email,
			Role:     claims.Role,
			BranchID: claims.BranchID,
		}
		c.Set(UserContextKey, userCtx)

		c.Next()
	}
}

// GetUserContext retrieves user context from Gin context
// Returns nil if user context is not found
func GetUserContext(c *gin.Context) *UserContext {
	value, exists := c.Get(UserContextKey)
	if !exists {
		return nil
	}

	userCtx, ok := value.(*UserContext)
	if !ok {
		return nil
	}

	return userCtx
}

// GetUserID extracts user ID from context (helper function)
// Returns 0 if not found
func GetUserID(c *gin.Context) uint {
	userCtx := GetUserContext(c)
	if userCtx == nil {
		return 0
	}
	return userCtx.UserID
}

// GetUserRole extracts user role from context (helper function)
// Returns empty string if not found
func GetUserRole(c *gin.Context) string {
	userCtx := GetUserContext(c)
	if userCtx == nil {
		return ""
	}
	return userCtx.Role
}

// GetBranchID extracts branch ID from context (helper function)
// Returns nil if not found (for users with all-branch access)
func GetBranchID(c *gin.Context) *uint {
	userCtx := GetUserContext(c)
	if userCtx == nil {
		return nil
	}
	return userCtx.BranchID
}

// GetUsername extracts username from context (helper function)
// Returns empty string if not found
func GetUsername(c *gin.Context) string {
	userCtx := GetUserContext(c)
	if userCtx == nil {
		return ""
	}
	return userCtx.Username
}
