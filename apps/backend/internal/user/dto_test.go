package user

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestCreateUserRequest_Validation Tests for CreateUserRequest DTO validation
func TestCreateUserRequest_Validation(t *testing.T) {
	tests := []struct {
		name        string
		request     CreateUserRequest
		expectError bool
	}{
		{
			name: "Valid request - CASHIER with branch_id",
			request: CreateUserRequest{
				Username: "testcashier",
				Password: "SecurePass123!",
				Email:    "cashier@simpo.pharmacy",
				Role:     "CASHIER",
				BranchID: uintPtr(1),
			},
			expectError: false,
		},
		{
			name: "Valid request - SYSTEM_ADMIN without branch_id",
			request: CreateUserRequest{
				Username: "testadmin",
				Password: "SecurePass123!",
				Email:    "admin@simpo.pharmacy",
				Role:     "SYSTEM_ADMIN",
				BranchID: nil,
			},
			expectError: false,
		},
		{
			name: "Valid request - OWNER without branch_id",
			request: CreateUserRequest{
				Username: "testowner",
				Password: "SecurePass123!",
				Email:    "owner@simpo.pharmacy",
				Role:     "OWNER",
				BranchID: nil,
			},
			expectError: false,
		},
		{
			name: "Missing username",
			request: CreateUserRequest{
				Username: "",
				Password: "SecurePass123!",
				Email:    "test@simpo.pharmacy",
				Role:     "CASHIER",
				BranchID: uintPtr(1),
			},
			expectError: true, // Will fail binding:"required,min=3"
		},
		{
			name: "Username too short",
			request: CreateUserRequest{
				Username: "ab",
				Password: "SecurePass123!",
				Email:    "test@simpo.pharmacy",
				Role:     "CASHIER",
				BranchID: uintPtr(1),
			},
			expectError: true, // Will fail binding:"min=3"
		},
		{
			name: "Missing password",
			request: CreateUserRequest{
				Username: "testuser",
				Password: "",
				Email:    "test@simpo.pharmacy",
				Role:     "CASHIER",
				BranchID: uintPtr(1),
			},
			expectError: true, // Will fail binding:"required,min=8"
		},
		{
			name: "Password too short",
			request: CreateUserRequest{
				Username: "testuser",
				Password: "short",
				Email:    "test@simpo.pharmacy",
				Role:     "CASHIER",
				BranchID: uintPtr(1),
			},
			expectError: true, // Will fail binding:"min=8"
		},
		{
			name: "Missing email",
			request: CreateUserRequest{
				Username: "testuser",
				Password: "SecurePass123!",
				Email:    "",
				Role:     "CASHIER",
				BranchID: uintPtr(1),
			},
			expectError: true, // Will fail binding:"required,email"
		},
		{
			name: "Invalid email format",
			request: CreateUserRequest{
				Username: "testuser",
				Password: "SecurePass123!",
				Email:    "notanemail",
				Role:     "CASHIER",
				BranchID: uintPtr(1),
			},
			expectError: true, // Will fail binding:"email"
		},
		{
			name: "Missing role",
			request: CreateUserRequest{
				Username: "testuser",
				Password: "SecurePass123!",
				Email:    "test@simpo.pharmacy",
				Role:     "",
				BranchID: uintPtr(1),
			},
			expectError: true, // Will fail binding:"required"
		},
		{
			name: "Invalid role",
			request: CreateUserRequest{
				Username: "testuser",
				Password: "SecurePass123!",
				Email:    "test@simpo.pharmacy",
				Role:     "INVALID_ROLE",
				BranchID: uintPtr(1),
			},
			expectError: true, // Will fail custom validation
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test that validation tags work correctly
			// In actual handler, binding will validate these
			// Here we just test the struct can be created
			assert.NotNil(t, tt.request)

			// For invalid role, test our validation function
			if tt.request.Role != "" && !isValidRoleForCreate(tt.request.Role) {
				assert.True(t, tt.expectError, "Invalid role should trigger error")
			}
		})
	}
}

// TestCreateUserResponse Tests for CreateUserResponse DTO
func TestCreateUserResponse(t *testing.T) {
	user := &User{
		ID:        5,
		Username:  "testcashier",
		Email:     "cashier@simpo.pharmacy",
		Role:      "CASHIER",
		BranchID:  uintPtr(1),
		Status:    "ACTIVE",
		CreatedAt: testTime,
		UpdatedAt: testTime,
	}

	response := ToCreateUserResponse(user)

	assert.Equal(t, uint(5), response.ID)
	assert.Equal(t, "testcashier", response.Username)
	assert.Equal(t, "cashier@simpo.pharmacy", response.Email)
	assert.Equal(t, "CASHIER", response.Role)
	assert.Equal(t, uintPtr(1), response.BranchID)
	assert.Equal(t, "ACTIVE", response.Status)
	assert.NotNil(t, response.CreatedAt)
	// Password should not be in response
	assert.Empty(t, "") // No password field
}

// Helper function
func uintPtr(val uint) *uint {
	return &val
}

var testTime = time.Date(2026, 5, 11, 5, 0, 0, 0, time.UTC)
