package middleware

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/permissions"
)

// RBACMiddleware creates middleware that enforces role-based access control
// Story 1.6, AC3: Role-Based Endpoint Access Control
// Story 1.6, AC7: Role Permission Mapping
//
// This middleware checks if the authenticated user has permission to access
// the requested endpoint. It uses the permissions package to determine
// access based on the user's role.
//
// Usage:
//   router.Use(RBACMiddleware())
//
// The middleware must be used AFTER JWTAuthMiddleware to ensure user context
// is available. Middleware order: CORS → Rate Limit → Auth → RBAC → Handler
func RBACMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract user role from context (set by auth middleware)
		// Story 1.6, AC2: Role is extracted from JWT token claims and stored in request context
		userRole := GetUserRole(c)
		if userRole == "" {
			// Story 1.6, AC6: Log authorization failures
			slog.Info("AUDIT",
				"action", "AUTH_FAILURE",
				"user_id", 0,
				"username", "unknown",
				"role", "",
				"endpoint", getRequestPath(c),
				"ip_address", c.ClientIP(),
				"outcome", "denied",
				"reason", "user role not found in request context",
			)
			sendRFC7807Error(c, http.StatusForbidden, "Forbidden", "user role not found in request context", getRequestPath(c))
			c.Abort()
			return
		}

		// Get the request path for permission checking
		requestPath := getRequestPath(c)

		// Check if user has permission to access this endpoint
		// Story 1.6, AC3: Role-Based Endpoint Access Control
		// Story 1.6, AC7: Role Permission Mapping (code-based for MVP)
		if !permissions.CanAccessEndpoint(userRole, requestPath) {
			// Story 1.6, AC6: Log authorization failures (403 responses)
			// Audit log includes: user_id, role, endpoint, reason, timestamp, IP address
			userID := GetUserID(c)
			username := GetUsername(c)

			// Log authorization failure to audit trail
			// Story 1.6, AC6: Audit trail is append-only per NFR-SEC-004
			slog.Info("AUDIT",
				"action", "FORBIDDEN_ACCESS",
				"user_id", userID,
				"username", username,
				"role", userRole,
				"endpoint", requestPath,
				"ip_address", c.ClientIP(),
				"outcome", "denied",
				"reason", "user role '"+userRole+"' cannot access endpoint '"+requestPath+"'",
			)

			// Story 1.6, AC1: Return 403 Forbidden with RFC 7807 format
			// Story 1.6, AC3: Access denied returns 403 Forbidden with RFC 7807 error format
			sendRFC7807Error(
				c,
				http.StatusForbidden,
				"Forbidden",
				"user role '"+userRole+"' cannot access endpoint '"+requestPath+"'",
				requestPath,
			)
			c.Abort()
			return
		}

		// User has permission - continue to next middleware/handler
		c.Next()
	}
}

// RequirePermission creates a middleware that requires specific permission
// This is an alternative to RBACMiddleware for more granular control
// Usage: router.GET("/admin", RequirePermission(permissions.PermAdmin))
//
// Note: For MVP, we use the simpler RBACMiddleware which checks endpoint access.
// This function is provided for future enhancement when more granular
// permission checks are needed (e.g., specific DELETE permission on certain endpoints)
func RequirePermission(perm permissions.Permission) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole := GetUserRole(c)
		if userRole == "" {
			sendRFC7807Error(c, http.StatusForbidden, "Forbidden", "user role not found in request context", getRequestPath(c))
			c.Abort()
			return
		}

		// Check if user has the required permission
		if !permissions.HasPermission(userRole, perm) {
			sendRFC7807Error(
				c,
				http.StatusForbidden,
				"Forbidden",
				"user role '"+userRole+"' does not have required permission: "+string(perm),
				getRequestPath(c),
			)
			c.Abort()
			return
		}

		c.Next()
	}
}

// Legacy GRAB boilerplate functions - retained for backward compatibility
// Deprecated: Use RBACMiddleware() instead for simpo RBAC system
// RequireRole returns a middleware that checks if the user has the specified role
func RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole := GetUserRole(c)
		if userRole != role {
			sendRFC7807Error(c, http.StatusForbidden, "Forbidden", "insufficient permissions", getRequestPath(c))
			c.Abort()
			return
		}
		c.Next()
	}
}

// RequireAdmin returns a middleware that checks if the user is an admin
// Deprecated: Use RBACMiddleware() instead for simpo RBAC system
func RequireAdmin() gin.HandlerFunc {
	return RequireRole("admin")
}
