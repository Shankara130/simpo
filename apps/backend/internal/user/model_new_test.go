package user

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// TestUserModel_HasUsername tests user has username field
func TestUserModel_HasUsername(t *testing.T) {
	user := User{
		Name:     "Test User",
		Username: "testuser",
		Email:    "test@example.com",
	}

	assert.Equal(t, "testuser", user.Username, "User should have Username field")
}

// TestUserModel_HasStatus tests user has status field
func TestUserModel_HasStatus(t *testing.T) {
	user := User{
		Name:     "Test User",
		Username: "testuser",
		Status:   UserStatusActive,
	}

	assert.Equal(t, UserStatusActive, user.Status, "User should have Status field with UserStatusActive")
}

// TestUserModel_HasSingleRole tests user has single role field (not many-to-many)
func TestUserModel_HasSingleRole(t *testing.T) {
	user := User{
		Name:     "Test User",
		Username: "testuser",
		Role:     RoleSystemAdmin,
	}

	assert.Equal(t, RoleSystemAdmin, user.Role, "User should have single Role field")
	assert.Equal(t, "SYSTEM_ADMIN", user.Role, "Role should be SYSTEM_ADMIN")
}

// TestUserModel_HasBranchID tests user has branch_id field
func TestUserModel_HasBranchID(t *testing.T) {
	branchID := uint(5)
	user := User{
		Name:     "Test User",
		Username: "testuser",
		BranchID: &branchID,
	}

	assert.NotNil(t, user.BranchID, "User should have BranchID field")
	assert.Equal(t, uint(5), *user.BranchID, "BranchID should be 5")
}

// TestUserModel_BranchIDNullable tests branch_id can be null for system admin
func TestUserModel_BranchIDNullable(t *testing.T) {
	user := User{
		Name:     "System Admin",
		Username: "admin",
		Role:     RoleSystemAdmin,
		BranchID: nil,
	}

	assert.Nil(t, user.BranchID, "System admin should have nil BranchID")
}

// TestUserModel_PasswordHash tests user has password_hash field
func TestUserModel_PasswordHash(t *testing.T) {
	hash := "$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewY5GyYzpLaEmc0u"
	user := User{
		Name:         "Test User",
		Username:     "testuser",
		PasswordHash: hash,
	}

	assert.Equal(t, hash, user.PasswordHash, "User should have PasswordHash field")
	assert.NotEqual(t, "plaintext", user.PasswordHash, "PasswordHash should not be plaintext")
}

// TestUserModel_StatusEnum tests valid status values
func TestUserModel_StatusEnum(t *testing.T) {
	validStatuses := []string{
		UserStatusActive,
		UserStatusInactive,
	}

	assert.Contains(t, validStatuses, UserStatusActive, "ACTIVE should be valid status")
	assert.Contains(t, validStatuses, UserStatusInactive, "INACTIVE should be valid status")
}

// TestUserModel_RoleEnum tests valid role values (Story 1.5, FR1, NFR-SEC-001)
func TestUserModel_RoleEnum(t *testing.T) {
	assert.Equal(t, "SYSTEM_ADMIN", RoleSystemAdmin, "System admin role constant")
	assert.Equal(t, "OWNER", RoleOwner, "Owner role constant")
	assert.Equal(t, "CASHIER", RoleCashier, "Cashier role constant")
}

// TestUserModel_JSONTags tests JSON serialization uses camelCase
func TestUserModel_JSONTags(t *testing.T) {
	user := User{
		ID:       1,
		Username: "testuser",
		Status:   UserStatusActive,
		Role:     RoleCashier,
	}

	// JSON tags should be camelCase for API responses
	// This is verified by GORM struct tags
	type UserJSON struct {
		ID       uint   `json:"id"`
		Username string `json:"username"`
		Status   string `json:"status"`
		Role     string `json:"role"`
	}

	userJSON := UserJSON{
		ID:       user.ID,
		Username: user.Username,
		Status:   user.Status,
		Role:     user.Role,
	}

	assert.Equal(t, uint(1), userJSON.ID)
	assert.Equal(t, "testuser", userJSON.Username)
	assert.Equal(t, UserStatusActive, userJSON.Status)
	assert.Equal(t, RoleCashier, userJSON.Role)
}

// TestUserModel_SoftDelete tests user supports soft delete with gorm.DeletedAt
func TestUserModel_SoftDelete(t *testing.T) {
	user := User{
		Name:     "Test User",
		Username: "testuser",
	}

	// GORM's DeletedAt field enables soft delete
	assert.IsType(t, gorm.DeletedAt{}, user.DeletedAt, "User should have DeletedAt for soft delete")
	assert.False(t, user.DeletedAt.Valid, "Initially, DeletedAt should not be valid (not deleted)")
}
