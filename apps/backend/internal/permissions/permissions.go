package permissions

import (
	"strings"
)

// Role constants for simpo RBAC system
// Defined here to avoid import cycle with user and middleware packages
const (
	RoleSystemAdmin = "SYSTEM_ADMIN"
	RoleOwner       = "OWNER"
	RoleCashier     = "CASHIER"
	// Legacy roles for GRAB boilerplate compatibility
	RoleAdmin = "admin"
	RoleUser  = "user"
)

// Permission represents a specific action capability
type Permission string

const (
	// PermRead allows read-only access
	PermRead Permission = "read"
	// PermWrite allows write/create access
	PermWrite Permission = "write"
	// PermDelete allows delete access
	PermDelete Permission = "delete"
	// PermAdmin allows full administrative access
	PermAdmin Permission = "admin"
)

// RolePermissions defines which permissions each role has
// Story 1.6, AC7: Role permissions defined in code (no database table for MVP)
type RolePermissions struct {
	Role              string
	Permissions       []Permission
	AllowedEndpoints  []string // Endpoint prefixes for whitelist approach
	AllBranchesAccess bool     // Whether role can access all branches
}

// GetRolePermissions returns the permission configuration for a given role
// Story 1.6, AC7: SYSTEM_ADMIN, OWNER, CASHIER role definitions
func GetRolePermissions(role string) RolePermissions {
	switch role {
	case RoleSystemAdmin:
		return RolePermissions{
			Role:              RoleSystemAdmin,
			Permissions:       []Permission{PermAdmin, PermRead, PermWrite, PermDelete},
			AllowedEndpoints:  []string{"*"}, // Wildcard: all endpoints
			AllBranchesAccess: true,
		}

	case RoleOwner:
		return RolePermissions{
			Role:              RoleOwner,
			Permissions:       []Permission{PermRead, PermWrite},
			AllowedEndpoints: []string{
				"/api/v1/products",
				"/api/v1/transactions",
				"/api/v1/reports",
				"/api/v1/users/:id",		// Story 1.7: Can view specific users, not create/list
				"/api/v1/inventory",
				"/api/v1/branches",
			},
			AllBranchesAccess: true, // Owner can see all branches
		}

	case RoleCashier:
		return RolePermissions{
			Role:              RoleCashier,
			Permissions:       []Permission{PermRead, PermWrite},
			AllowedEndpoints: []string{
				"/api/v1/transactions", // Can process transactions
				"/api/v1/products",     // Can check stock (read-only)
			},
			AllBranchesAccess: false, // Cashier: assigned branch only
		}

	default:
		// Default deny: unknown roles get no permissions
		return RolePermissions{
			Role:              role,
			Permissions:       []Permission{},
			AllowedEndpoints:  []string{},
			AllBranchesAccess: false,
		}
	}
}

// HasPermission checks if a role has a specific permission
func HasPermission(role string, perm Permission) bool {
	rolePerms := GetRolePermissions(role)
	for _, p := range rolePerms.Permissions {
		if p == perm || p == PermAdmin {
			return true
		}
	}
	return false
}

// CanAccessEndpoint checks if a role can access a specific endpoint
// Story 1.6, AC3: Role-based endpoint access control
func CanAccessEndpoint(role, endpoint string) bool {
	rolePerms := GetRolePermissions(role)

	// Wildcard access for SYSTEM_ADMIN
	if len(rolePerms.AllowedEndpoints) > 0 && rolePerms.AllowedEndpoints[0] == "*" {
		return true
	}

	// Check if endpoint is in allowed list (prefix match)
	for _, allowed := range rolePerms.AllowedEndpoints {
		if strings.HasPrefix(endpoint, allowed) {
			return true
		}
	}

	return false
}

// CanAccessAllBranches checks if a role can access data from all branches
// Story 1.6, AC4: Branch-level data isolation
func CanAccessAllBranches(role string) bool {
	rolePerms := GetRolePermissions(role)
	return rolePerms.AllBranchesAccess
}

// IsValidRole checks if a role is valid for simpo system
// Story 1.6, AC7: Role validation for three-tier RBAC system
func IsValidRole(role string) bool {
	switch role {
	case RoleAdmin, RoleSystemAdmin, RoleOwner, RoleCashier, RoleUser:
		return true
	default:
		return false
	}
}
