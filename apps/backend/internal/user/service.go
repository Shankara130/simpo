package user

import (
	"context"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// isValidRoleForCreate checks if role is valid for admin user creation (Story 1.7)
// Wrapper for role.IsValidRoleForCreate for service layer use
func isValidRoleForCreate(role string) bool {
	return IsValidRoleForCreate(role)
}

const (
	// BcryptCost is the cost factor for bcrypt password hashing (Story 1.5, AC3, Decision 5)
	BcryptCost = 12
)

var (
	// ErrUserNotFound is returned when user is not found
	ErrUserNotFound = errors.New("user not found")
	// ErrEmailExists is returned when email already exists
	ErrEmailExists = errors.New("email already exists")
	// ErrInvalidCredentials is returned when credentials are invalid
	ErrInvalidCredentials = errors.New("invalid credentials")
	// ErrInvalidRole is returned when role is invalid
	ErrInvalidRole = errors.New("invalid role")
	// Story 1.7: Username already exists
	ErrUsernameExists = errors.New("username already exists")
	// Story 1.7: Invalid role for user creation
	ErrInvalidRoleForCreate = errors.New("invalid role for user creation")
	// Story 1.7: Branch ID required for CASHIER role
	ErrBranchIDRequired = errors.New("branch_id is required for CASHIER role")
)

// Service defines user service interface
type Service interface {
	RegisterUser(ctx context.Context, req RegisterRequest) (*User, error)
	RegisterUserForAdmin(ctx context.Context, req CreateUserRequest, adminID uint) (*User, error)
	AuthenticateUser(ctx context.Context, req LoginRequest) (*User, error)
	GetUserByID(ctx context.Context, id uint) (*User, error)
	UpdateUser(ctx context.Context, id uint, req UpdateUserRequest) (*User, error)
	DeleteUser(ctx context.Context, id uint) error
	ListUsers(ctx context.Context, filters UserFilterParams, page, perPage int) ([]User, int64, error)
	PromoteToAdmin(ctx context.Context, userID uint) error
}

type service struct {
	repo Repository
}

// NewService creates a new user service
func NewService(repo Repository) Service {
	return &service{
		repo: repo,
	}
}

// RegisterUser registers a new user
func (s *service) RegisterUser(ctx context.Context, req RegisterRequest) (*User, error) {
	existingUser, err := s.repo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing email: %w", err)
	}
	if existingUser != nil {
		return nil, ErrEmailExists
	}

	hashedPassword, err := hashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &User{
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: hashedPassword,
	}

	// Use transaction to ensure atomic user creation and role assignment
	err = s.repo.Transaction(ctx, func(txCtx context.Context) error {
		if err := s.repo.Create(txCtx, user); err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}

		if err := s.repo.AssignRole(txCtx, user.ID, RoleUser); err != nil {
			return fmt.Errorf("failed to assign default role: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Reload user with roles after successful transaction
	user, err = s.repo.FindByID(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to reload user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("failed to reload user: user not found after creation")
	}

	return user, nil
}

// RegisterUserForAdmin registers a new user with admin-specified role (Story 1.7)
func (s *service) RegisterUserForAdmin(ctx context.Context, req CreateUserRequest, adminID uint) (*User, error) {
	// Use transaction to ensure atomic user creation and prevent race conditions
	var createdUser *User
	err := s.repo.Transaction(ctx, func(txCtx context.Context) error {
		// 1. Validate role
		if !IsValidRoleForCreate(req.Role) {
			return ErrInvalidRoleForCreate
		}

		// 2. Check if username already exists
		usernameExists, err := s.repo.CheckUsernameExists(txCtx, req.Username)
		if err != nil {
			return fmt.Errorf("failed to check username existence: %w", err)
		}
		if usernameExists {
			return ErrUsernameExists
		}

		// 3. Check if email already exists
		emailExists, err := s.repo.CheckEmailExists(txCtx, req.Email)
		if err != nil {
			return fmt.Errorf("failed to check email existence: %w", err)
		}
		if emailExists {
			return ErrEmailExists
		}

		// 4. Validate branch_id for CASHIER role (ensure branch exists)
		if req.Role == RoleCashier {
			if req.BranchID == nil {
				return ErrBranchIDRequired
			}
			// Verify branch exists in database
			branchExists, err := s.repo.CheckBranchExists(txCtx, *req.BranchID)
			if err != nil {
				return fmt.Errorf("failed to check branch existence: %w", err)
			}
			if !branchExists {
				return fmt.Errorf("branch with ID %d does not exist", *req.BranchID)
			}
		}

		// 5. Hash password using bcrypt (cost factor 12)
		hashedPassword, err := hashPassword(req.Password)
		if err != nil {
			return fmt.Errorf("failed to hash password: %w", err)
		}

		// 6. Create user with ACTIVE status and specified role
		user := &User{
			Username:     req.Username,
			Name:         req.Username, // Set Name to Username for consistency
			Email:        req.Email,
			PasswordHash: hashedPassword,
			Role:         req.Role,
			BranchID:     req.BranchID,
			Status:       UserStatusActive,
		}

		// 7. Create user in database
		if err := s.repo.Create(txCtx, user); err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}

		createdUser = user
		return nil
	})

	if err != nil {
		return nil, err
	}

	// Reload user after creation
	user, err := s.repo.FindByID(ctx, createdUser.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to reload user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("failed to reload user: user not found after creation")
	}

	return user, nil
}

// AuthenticateUser authenticates a user with email and password
func (s *service) AuthenticateUser(ctx context.Context, req LoginRequest) (*User, error) {
	user, err := s.repo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return nil, ErrInvalidCredentials
	}

	if err := verifyPassword(user.PasswordHash, req.Password); err != nil {
		return nil, ErrInvalidCredentials
	}

	return user, nil
}

// GetUserByID retrieves a user by ID
func (s *service) GetUserByID(ctx context.Context, id uint) (*User, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

// UpdateUser updates a user's information
func (s *service) UpdateUser(ctx context.Context, id uint, req UpdateUserRequest) (*User, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Email != "" {
		existingUser, err := s.repo.FindByEmail(ctx, req.Email)
		if err != nil {
			return nil, fmt.Errorf("failed to check existing email: %w", err)
		}
		if existingUser != nil && existingUser.ID != user.ID {
			return nil, ErrEmailExists
		}
		user.Email = req.Email
	}

	if err := s.repo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return user, nil
}

// DeleteUser deletes a user
func (s *service) DeleteUser(ctx context.Context, id uint) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}

// ListUsers retrieves paginated list of users with filtering
func (s *service) ListUsers(ctx context.Context, filters UserFilterParams, page, perPage int) ([]User, int64, error) {
	// Validate pagination parameters
	if page < 1 {
		return nil, 0, fmt.Errorf("page must be >= 1")
	}
	if perPage < 1 {
		return nil, 0, fmt.Errorf("perPage must be >= 1")
	}
	if perPage > 100 {
		return nil, 0, fmt.Errorf("perPage must be <= 100")
	}

	if filters.Role != "" && !IsValidRole(filters.Role) {
		return nil, 0, ErrInvalidRole
	}

	users, total, err := s.repo.ListAllUsers(ctx, filters, page, perPage)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}

	return users, total, nil
}

// PromoteToAdmin promotes a user to admin role
func (s *service) PromoteToAdmin(ctx context.Context, userID uint) error {
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return ErrUserNotFound
	}

	if user.HasRole(RoleAdmin) {
		return nil
	}

	if err := s.repo.AssignRole(ctx, userID, RoleAdmin); err != nil {
		return fmt.Errorf("failed to assign admin role: %w", err)
	}

	return nil
}

// hashPassword hashes a plain text password using bcrypt (Story 1.5, AC3, Decision 5: cost factor 12)
func hashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), BcryptCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

// verifyPassword verifies a password against a hash
func verifyPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
