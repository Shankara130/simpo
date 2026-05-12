package user

import (
	"time"

	"gorm.io/gorm"
)

// User Status Constants (Story 1.5, AC6)
const (
	UserStatusActive   = "ACTIVE"
	UserStatusInactive = "INACTIVE"
	UserStatusPending  = "PENDING" // Story 1.9: For self-registered users awaiting email verification
)

// User represents a user in the system
// Updated for Story 1.5: Added Username, Status, Role, BranchID fields
// Story 1.10: Added DeactivatedAt, DeactivatedBy, DeactivationReason fields
type User struct {
	ID                  uint           `gorm:"primaryKey" json:"id"`
	Name                string         `gorm:"not null" json:"name"`
	Username            string         `gorm:"uniqueIndex;not null" json:"username"` // Story 1.5: Username for login
	Email               string         `gorm:"uniqueIndex;not null" json:"email"`
	PasswordHash        string         `gorm:"not null;column:password_hash" json:"-"` // Story 1.5: bcrypt hash
	Status              string         `gorm:"not null;default:ACTIVE" json:"status"` // Story 1.5: ACTIVE/INACTIVE/PENDING
	Role                string         `gorm:"not null;default:CASHIER" json:"role"`  // Story 1.5: Single role (not many-to-many)
	BranchID            *uint          `gorm:"index" json:"branch_id,omitempty"`      // Story 1.5: Nullable for system admin
	// Story 1.10: Deactivation tracking fields
	DeactivatedAt       *time.Time     `gorm:"column:deactivated_at" json:"deactivated_at,omitempty"`
	DeactivatedBy       *uint          `gorm:"column:deactivated_by" json:"deactivated_by,omitempty"`
	DeactivationReason  string         `gorm:"column:deactivation_reason" json:"deactivation_reason,omitempty"`
	Roles               []Role         `gorm:"many2many:user_roles;joinForeignKey:UserID;joinReferences:RoleID" json:"-"` // Legacy: GRAB compatibility (deprecated)
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
	DeletedAt           gorm.DeletedAt `gorm:"index" json:"-"`
}

// UserRole represents the many-to-many relationship between users and roles
// Includes assigned_at timestamp for audit trail
type UserRole struct {
	UserID     uint      `gorm:"primaryKey;column:user_id" json:"user_id"`
	RoleID     uint      `gorm:"primaryKey;column:role_id" json:"role_id"`
	AssignedAt time.Time `gorm:"column:assigned_at;not null;default:CURRENT_TIMESTAMP" json:"assigned_at"`
}

// TableName specifies the table name for UserRole model
func (UserRole) TableName() string {
	return "user_roles"
}

// TableName specifies the table name for User model
func (User) TableName() string {
	return "users"
}

// HasRole checks if user has specific role
func (u *User) HasRole(roleName string) bool {
	for _, role := range u.Roles {
		if role.Name == roleName {
			return true
		}
	}
	return false
}

// IsAdmin checks if user has admin role
func (u *User) IsAdmin() bool {
	return u.HasRole(RoleAdmin)
}

// GetRoleNames returns list of role names
func (u *User) GetRoleNames() []string {
	roleNames := make([]string, len(u.Roles))
	for i, role := range u.Roles {
		roleNames[i] = role.Name
	}
	return roleNames
}
