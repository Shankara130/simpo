package user

import "time"

// RegisterRequest represents registration request payload
type RegisterRequest struct {
	Name     string `json:"name" binding:"required,min=2,max=100"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// CreateUserRequest represents admin user creation request payload (Story 1.7, AC2, AC3)
type CreateUserRequest struct {
	// Username is the login username (minimum 3 characters, must be unique)
	// Example: "newcashier"
	Username string `json:"username" binding:"required,min=3" example:"newcashier"`

	// Password is the user's password (minimum 8 characters, will be hashed with bcrypt)
	// Example: "SecurePass123!"
	Password string `json:"password" binding:"required,min=8" example:"SecurePass123!"`

	// Email is the user's email address (valid format, must be unique)
	// Example: "cashier@example.com"
	Email string `json:"email" binding:"required,email" example:"cashier@example.com"`

	// Role is the user's role (must be one of: SYSTEM_ADMIN, OWNER, CASHIER)
	// Example: "CASHIER"
	// Enum: SYSTEM_ADMIN, OWNER, CASHIER
	Role string `json:"role" binding:"required" example:"CASHIER" enum:"SYSTEM_ADMIN,OWNER,CASHIER"`

	// BranchID is the branch assignment (required for CASHIER role, optional for others)
	// Must reference an existing branch ID
	// Example: 1
	BranchID *uint `json:"branch_id,omitempty" example:"1"`
}

// CreateUserResponse represents user creation response (Story 1.7, AC8)
type CreateUserResponse struct {
	// ID is the unique identifier for the created user
	// Example: 5
	ID uint `json:"id" example:"5"`

	// Username is the login username
	// Example: "newcashier"
	Username string `json:"username" example:"newcashier"`

	// Email is the user's email address
	// Example: "cashier@example.com"
	Email string `json:"email" example:"cashier@example.com"`

	// Role is the assigned user role
	// Example: "CASHIER"
	Role string `json:"role" example:"CASHIER"`

	// BranchID is the branch assignment (null for SYSTEM_ADMIN and OWNER roles)
	// Example: 1
	BranchID *uint `json:"branch_id,omitempty" example:"1"`

	// Status is the user account status (always "ACTIVE" for newly created users)
	// Example: "ACTIVE"
	Status string `json:"status" example:"ACTIVE"`

	// CreatedAt is the timestamp when the user was created (ISO 8601 format)
	// Example: "2026-05-11T22:00:00Z"
	CreatedAt time.Time `json:"created_at" example:"2026-05-11T22:00:00Z"`
}

// ToCreateUserResponse converts User model to CreateUserResponse DTO (Story 1.7)
func ToCreateUserResponse(user *User) CreateUserResponse {
	return CreateUserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Role:      user.Role,
		BranchID:  user.BranchID,
		Status:    user.Status,
		CreatedAt: user.CreatedAt,
	}
}

// LoginRequest represents login request payload
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// UpdateUserRequest represents user update request payload
type UpdateUserRequest struct {
	Name  string `json:"name" binding:"omitempty,min=2,max=100"`
	Email string `json:"email" binding:"omitempty,email"`
}

// UserResponse represents user response (without sensitive fields)
type UserResponse struct {
	ID        uint     `json:"id"`
	Name      string   `json:"name"`
	Email     string   `json:"email"`
	Roles     []string `json:"roles"`
	CreatedAt string   `json:"created_at"`
	UpdatedAt string   `json:"updated_at"`
}

// AuthResponse represents authentication response
type AuthResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	TokenType    string       `json:"token_type"`
	ExpiresIn    int64        `json:"expires_in"`
	User         UserResponse `json:"user"`
}

// LegacyAuthResponse represents legacy authentication response (deprecated)
type LegacyAuthResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}

// UserListResponse represents paginated user list response
type UserListResponse struct {
	Users      []UserResponse `json:"users"`
	Total      int64          `json:"total"`
	Page       int            `json:"page"`
	PerPage    int            `json:"per_page"`
	TotalPages int            `json:"total_pages"`
}

// ToUserResponse converts User model to UserResponse DTO
func ToUserResponse(user *User) UserResponse {
	return UserResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		Roles:     user.GetRoleNames(),
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: user.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}
