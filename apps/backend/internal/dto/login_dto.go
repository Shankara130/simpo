package dto

import (
	"github.com/go-playground/validator/v10"
)

// LoginRequest represents user login request payload
// Uses username instead of email for authentication (Story 1.5, AC1, AC5)
type LoginRequest struct {
	Username string `json:"username" binding:"required,min=3" validate:"required,min=3" example:"admin"`
	Password string `json:"password" binding:"required,min=8" validate:"required,min=8" example:"SecurePassword123!"`
}

// ValidateStruct provides validation helper for LoginRequest
func (lr *LoginRequest) ValidateStruct(v *validator.Validate) error {
	return v.Struct(lr)
}

// UserInfo represents user information in authentication response (Story 1.5, AC4)
type UserInfo struct {
	ID       uint    `json:"id" example:"1"`
	Username string  `json:"username" example:"admin"`
	Email    string  `json:"email" example:"admin@simpo.pharmacy"`
	Role     string  `json:"role" example:"SYSTEM_ADMIN"`
	BranchID *uint  `json:"branch_id,omitempty" example:"null"`
}

// LoginResponse represents successful authentication response (Story 1.5, AC4)
// Follows RFC 7807 error format for failures
type LoginResponse struct {
	AccessToken string   `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	TokenType   string   `json:"token_type" example:"Bearer"`
	ExpiresIn   int      `json:"expires_in" example:"28800"` // 8 hours in seconds (NFR-SEC-002)
	User        UserInfo `json:"user"`
}
