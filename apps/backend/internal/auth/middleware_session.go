package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// SessionAuthMiddleware creates a middleware that validates JWT tokens and tracks session activity
// Story 1.8, Task 6: Modify JWTAuthMiddleware to track session activity, check blocklist
// Story 1.8, AC2: Last activity is updated on each authenticated API request
// Story 1.8, AC7: Return RFC 7807 formatted errors for expired sessions
func SessionAuthMiddleware(authService Service, sessionManager SessionManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader(AuthorizationHeader)
		if authHeader == "" {
			sendRFC7807Error(c, http.StatusUnauthorized, "Unauthorized", "authorization header required", c.Request.URL.Path)
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			sendRFC7807Error(c, http.StatusUnauthorized, "Unauthorized", "invalid authorization header format", c.Request.URL.Path)
			c.Abort()
			return
		}

		tokenString := parts[1]
		if tokenString == "" {
			sendRFC7807Error(c, http.StatusUnauthorized, "Unauthorized", "token cannot be empty", c.Request.URL.Path)
			c.Abort()
			return
		}

		// Validate token and extract claims
		claims, err := authService.ValidateToken(tokenString)
		if err != nil {
			if err == ErrExpiredToken {
				// Story 1.8, AC7: Expired session returns 401 with RFC 7807 format
				sendRFC7807Error(c, http.StatusUnauthorized, "Session Expired", "Your session has expired. Please log in again.", c.Request.URL.Path)
				c.Abort()
				return
			}
			sendRFC7807Error(c, http.StatusUnauthorized, "Unauthorized", "invalid token", c.Request.URL.Path)
			c.Abort()
			return
		}

		// SECURITY FIX: Reject tokens without TokenID claim
		// Tokens must have a TokenID for session tracking and revocation to work
		if claims.TokenID == "" {
			sendRFC7807Error(c, http.StatusUnauthorized, "Invalid Token", "Token is missing required identifier. Please log in again.", c.Request.URL.Path)
			c.Abort()
			return
		}

		// Story 1.8, Task 6: Check token blocklist during token validation
		if sessionManager != nil {
			revoked, err := sessionManager.IsTokenRevoked(c.Request.Context(), claims.TokenID)
			if err != nil {
				sendRFC7807Error(c, http.StatusInternalServerError, "Internal Server Error", "failed to validate token", c.Request.URL.Path)
				c.Abort()
				return
			}
			if revoked {
				// Story 1.8, AC7: Revoked token returns 401 with RFC 7807 format
				sendRFC7807Error(c, http.StatusUnauthorized, "Token Revoked", "This token has been revoked. Please log in again.", c.Request.URL.Path)
				c.Abort()
				return
			}

			// Story 1.8, AC2: Update last activity on each authenticated request
			// Store session data in Redis for tracking
			if err := sessionManager.UpdateLastActivity(c.Request.Context(), claims.UserID, claims.TokenID); err != nil {
				// Log error but don't fail the request - session tracking is best-effort
				// The session will be recreated on the next request if Redis is available
			}

			// Store session info in context for use in handlers (logout, etc.)
			c.Set("token_id", claims.TokenID)
		}

		c.Set(KeyUser, claims)
		c.Next()
	}
}

// RFC7807Error represents an RFC 7807 compliant error response
// Story 1.8, AC7: Error response format for expired/revoked tokens
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
		Type:     getTypeForStatus(status),
		Title:    title,
		Status:   status,
		Detail:   detail,
		Instance: instance,
	})
}

// getTypeForStatus returns error type string for HTTP status
func getTypeForStatus(status int) string {
	switch status {
	case http.StatusUnauthorized:
		return "https://api.simpo.com/errors/session-expired"
	case http.StatusForbidden:
		return "https://api.simpo.com/errors/forbidden"
	case http.StatusBadRequest:
		return "https://api.simpo.com/errors/bad-request"
	default:
		return "https://api.simpo.com/errors/internal-error"
	}
}

// GetTokenID extracts token ID from gin context
// Story 1.8: Helper for handlers to access the current token ID
func GetTokenID(c *gin.Context) string {
	if tokenID, exists := c.Get("token_id"); exists {
		if id, ok := tokenID.(string); ok {
			return id
		}
	}
	return ""
}
