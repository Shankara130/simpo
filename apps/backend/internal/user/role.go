package user

import "time"

// simpo Role System
// =================
// Three-tier RBAC system for Indonesian SME pharmacy management:
// - Admin: System administrator with full access
// - Owner: Pharmacy owner with business management access
// - Cashier: Staff with transaction processing access
//
// Note: "user" role retained for backward compatibility with GRAB boilerplate

const (
	RoleUser    = "user"         // Legacy role (GRAB boilerplate compatibility)
	RoleAdmin   = "admin"        // Legacy: System administrator (GRAB compatibility)
	RoleSystemAdmin = "SYSTEM_ADMIN" // Story 1.5: System administrator with full access
	RoleOwner   = "OWNER"        // Story 1.5: Pharmacy owner with business management access
	RoleCashier = "CASHIER"      // Story 1.5: POS staff with transaction processing access
)

// IsValidRole checks if a role string is valid for simpo system
func IsValidRole(role string) bool {
	switch role {
	case RoleAdmin, RoleSystemAdmin, RoleOwner, RoleCashier, RoleUser:
		return true
	default:
		return false
	}
}

// IsValidRoleForCreate checks if a role string is valid for user creation (Story 1.7)
// Excludes legacy "user" role, only allows SYSTEM_ADMIN, OWNER, CASHIER
func IsValidRoleForCreate(role string) bool {
	switch role {
	case RoleSystemAdmin, RoleOwner, RoleCashier:
		return true
	default:
		return false
	}
}

// Role represents a user role in the system
type Role struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"uniqueIndex;not null" json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TableName specifies the table name for Role model
func (Role) TableName() string {
	return "roles"
}
