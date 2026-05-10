package dto

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLoginRequestValidation_Success tests valid login request
func TestLoginRequestValidation_Success(t *testing.T) {
	req := LoginRequest{
		Username: "admin",
		Password: "SecurePass123!",
	}

	validate := validator.New()
	err := validate.Struct(req)

	assert.NoError(t, err, "Valid login request should pass validation")
}

// TestLoginRequestValidation_MissingUsername tests username is required
func TestLoginRequestValidation_MissingUsername(t *testing.T) {
	req := LoginRequest{
		Username: "",
		Password: "SecurePass123!",
	}

	validate := validator.New()
	err := validate.Struct(req)

	require.Error(t, err, "Missing username should fail validation")
	validationErrors := err.(validator.ValidationErrors)
	assert.Contains(t, validationErrors[0].Field(), "Username", "Error should be about Username field")
}

// TestLoginRequestValidation_MissingPassword tests password is required
func TestLoginRequestValidation_MissingPassword(t *testing.T) {
	req := LoginRequest{
		Username: "admin",
		Password: "",
	}

	validate := validator.New()
	err := validate.Struct(req)

	require.Error(t, err, "Missing password should fail validation")
	validationErrors := err.(validator.ValidationErrors)
	assert.Contains(t, validationErrors[0].Field(), "Password", "Error should be about Password field")
}

// TestLoginRequestValidation_UsernameTooShort tests username minimum length
func TestLoginRequestValidation_UsernameTooShort(t *testing.T) {
	req := LoginRequest{
		Username: "ab", // Less than 3 characters
		Password: "SecurePass123!",
	}

	validate := validator.New()
	err := validate.Struct(req)

	require.Error(t, err, "Username too short should fail validation")
	validationErrors := err.(validator.ValidationErrors)
	assert.Contains(t, validationErrors[0].Field(), "Username", "Error should be about Username field")
	assert.Contains(t, validationErrors[0].Tag(), "min", "Should be a min length validation error")
}

// TestLoginRequestValidation_PasswordTooShort tests password minimum length
func TestLoginRequestValidation_PasswordTooShort(t *testing.T) {
	req := LoginRequest{
		Username: "admin",
		Password: "short", // Less than 8 characters
	}

	validate := validator.New()
	err := validate.Struct(req)

	require.Error(t, err, "Password too short should fail validation")
	validationErrors := err.(validator.ValidationErrors)
	assert.Contains(t, validationErrors[0].Field(), "Password", "Error should be about Password field")
	assert.Contains(t, validationErrors[0].Tag(), "min", "Should be a min length validation error")
}

// TestLoginResponse_Structure tests LoginResponse has required fields
func TestLoginResponse_Structure(t *testing.T) {
	userInfo := UserInfo{
		ID:       1,
		Username: "admin",
		Email:    "admin@simpo.pharmacy",
		Role:     "SYSTEM_ADMIN",
		BranchID: nil,
	}

	resp := LoginResponse{
		AccessToken: "eyJhbGciOiJIUzI1NiIs...",
		TokenType:   "Bearer",
		ExpiresIn:   28800,
		User:        userInfo,
	}

	assert.Equal(t, "eyJhbGciOiJIUzI1NiIs...", resp.AccessToken, "AccessToken should be set")
	assert.Equal(t, "Bearer", resp.TokenType, "TokenType should be Bearer")
	assert.Equal(t, 28800, resp.ExpiresIn, "ExpiresIn should be 28800 (8 hours)")
	assert.Equal(t, uint(1), resp.User.ID, "User ID should be set")
	assert.Equal(t, "admin", resp.User.Username, "Username should be set")
	assert.Equal(t, "SYSTEM_ADMIN", resp.User.Role, "Role should be set")
	assert.Nil(t, resp.User.BranchID, "BranchID should be nil for system admin")
}

// TestUserInfo_WithBranchID tests UserInfo with branch ID
func TestUserInfo_WithBranchID(t *testing.T) {
	branchID := uint(5)
	userInfo := UserInfo{
		ID:       2,
		Username: "cashier1",
		Email:    "cashier1@simpo.pharmacy",
		Role:     "CASHIER",
		BranchID: &branchID,
	}

	assert.Equal(t, uint(5), *userInfo.BranchID, "BranchID should be set for cashier")
}
