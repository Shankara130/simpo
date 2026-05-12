package user

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// WhitelistEntry represents a whitelisted email domain entry
// Story 1.9: Used for domain validation in self-registration
type WhitelistEntry struct {
	Domain      string
	DefaultRole string
}

// WhitelistRepository defines the interface for whitelist operations
// Story 1.9: Used by user service to validate email domains
type WhitelistRepository interface {
	FindByDomain(ctx context.Context, domain string) (*WhitelistEntry, error)
}

// VerificationRepository defines the interface for verification token operations
// Story 1.9: Used by user service to manage email verification tokens
type VerificationRepository interface {
	CreateToken(ctx context.Context, token, email string, expiresAt time.Time) error
	FindByToken(ctx context.Context, token string) (*EmailVerificationToken, error)
	DeleteToken(ctx context.Context, token string) error
}

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
	// Story 1.9: Domain not whitelisted for self-registration
	ErrDomainNotWhitelisted = errors.New("email domain is not approved for self-registration")
	// Story 1.9: Invalid verification token
	ErrInvalidVerificationToken = errors.New("invalid or expired verification token")
	// Story 1.10: Cannot deactivate own account
	ErrCannotDeactivateSelf = errors.New("cannot deactivate own account")
	// Story 1.10: Account has been deactivated
	ErrAccountDeactivated = errors.New("account has been deactivated")
)

// Service defines user service interface
type Service interface {
	RegisterUser(ctx context.Context, req RegisterRequest) (*User, error)
	RegisterUserForAdmin(ctx context.Context, req CreateUserRequest, adminID uint) (*User, error)
	// Story 1.9: RegisterStaff - self-registration with email domain whitelist validation
	RegisterStaff(ctx context.Context, req StaffRegistrationRequest) (*User, string, error)
	// Story 1.9: VerifyEmail - verify email and activate account
	VerifyEmail(ctx context.Context, token string) (*User, error)
	AuthenticateUser(ctx context.Context, req LoginRequest) (*User, error)
	GetUserByID(ctx context.Context, id uint) (*User, error)
	UpdateUser(ctx context.Context, id uint, req UpdateUserRequest) (*User, error)
	DeleteUser(ctx context.Context, id uint) error
	ListUsers(ctx context.Context, filters UserFilterParams, page, perPage int) ([]User, int64, error)
	PromoteToAdmin(ctx context.Context, userID uint) error
	// Story 1.9: SetWhitelistRepo - inject whitelist repository for domain validation
	SetWhitelistRepo(whitelistRepo WhitelistRepository)
	// Story 1.9: SetVerificationRepo - inject verification repository for email tokens
	SetVerificationRepo(verificationRepo VerificationRepository)
	// Story 1.10: DeactivateUser - deactivate user account with reason
	DeactivateUser(ctx context.Context, targetUserID uint, adminID uint, reason string) (*User, error)
}

type service struct {
	repo               Repository
	whitelistRepo      WhitelistRepository
	verificationRepo   VerificationRepository
	sessionManager     SessionManager // Story 1.10: For token revocation on deactivation
}

// SessionManager defines the interface for session management operations (Story 1.8, Story 1.10)
type SessionManager interface {
	RevokeToken(ctx context.Context, tokenID string, ttl time.Duration) error
	RevokeAllUserTokens(ctx context.Context, userID uint) error
	DeleteSession(ctx context.Context, userID uint, tokenID string) error
}

// NewService creates a new user service
func NewService(repo Repository) Service {
	return &service{
		repo: repo,
	}
}

// SetWhitelistRepo injects the whitelist repository (dependency injection)
func (s *service) SetWhitelistRepo(whitelistRepo WhitelistRepository) {
	s.whitelistRepo = whitelistRepo
}

// SetVerificationRepo injects the verification repository (dependency injection)
func (s *service) SetVerificationRepo(verificationRepo VerificationRepository) {
	s.verificationRepo = verificationRepo
}

// SetSessionManager injects the session manager (dependency injection)
// Story 1.10: Required for token revocation on deactivation
func (s *service) SetSessionManager(sessionManager SessionManager) {
	s.sessionManager = sessionManager
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
// Story 1.10, AC1: Deactivated users cannot authenticate (returns 401 Unauthorized)
func (s *service) AuthenticateUser(ctx context.Context, req LoginRequest) (*User, error) {
	user, err := s.repo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return nil, ErrInvalidCredentials
	}

	// Story 1.10: Check if user account is deactivated
	if user.Status == UserStatusInactive {
		return nil, ErrAccountDeactivated
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

// WhitelistRepoAdapter adapts any repository with FindByDomain to user.WhitelistRepository
type WhitelistRepoAdapter struct {
	// The wrapped repository must have FindByDomain method
	findByDomain func(ctx context.Context, domain string) (interface{}, error)
}

// NewWhitelistRepoAdapter creates a new adapter from a repository with FindByDomain method
// The provided repo must implement: FindByDomain(ctx context.Context, domain string) (interface{}, error)
func NewWhitelistRepoAdapter(repo interface{}) *WhitelistRepoAdapter {
	// Use reflection or type assertion to get the method dynamically
	// For simplicity, we expect the repo to have a FindByDomain method
	return &WhitelistRepoAdapter{
		findByDomain: repo.(interface {
			FindByDomain(ctx context.Context, domain string) (interface{}, error)
		}).FindByDomain,
	}
}

// FindByDomain implements user.WhitelistRepository interface
func (a *WhitelistRepoAdapter) FindByDomain(ctx context.Context, domain string) (*WhitelistEntry, error) {
	entry, err := a.findByDomain(ctx, domain)
	if err != nil {
		return nil, err
	}
	if entry == nil {
		return nil, ErrDomainNotWhitelisted
	}

	// Extract Domain and DefaultRole using type assertions
	// Try common struct patterns
	type withFields interface {
		Domain() string
		DefaultRole() string
	}

	// Method-based access (if getters exist)
	if wf, ok := entry.(withFields); ok {
		return &WhitelistEntry{
			Domain:      wf.Domain(),
			DefaultRole: wf.DefaultRole(),
		}, nil
	}

	// Direct field access via map (for flexibility)
	if m, ok := entry.(map[string]interface{}); ok {
		domain, _ := m["Domain"].(string)
		role, _ := m["default_role"].(string)
		if domain == "" {
			domain, _ = m["domain"].(string)
		}
		if role == "" {
			role, _ = m["DefaultRole"].(string)
		}
		if domain != "" && role != "" {
			return &WhitelistEntry{
				Domain:      domain,
				DefaultRole: role,
			}, nil
		}
	}

	return nil, fmt.Errorf("invalid whitelist entry type: cannot extract Domain and DefaultRole")
}

// extractDomain extracts the domain from an email address
// Story 1.9: Helper for whitelist validation
func extractDomain(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return ""
	}
	return parts[1]
}

// isValidDomainFormat checks if a domain has a valid format
// Story 1.9: Domain format validation for whitelist and registration
func isValidDomainFormat(domain string) bool {
	if domain == "" {
		return false
	}

	// Basic domain format validation
	// Must contain at least one dot, no spaces, valid characters
	parts := strings.Split(domain, ".")
	if len(parts) < 2 {
		return false
	}

	for _, part := range parts {
		if part == "" {
			return false
		}
		// Check for valid hostname characters (alphanumeric and hyphens)
		for _, r := range part {
			if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
				(r >= '0' && r <= '9') || r == '-') {
				return false
			}
		}
		// Part cannot start or end with hyphen
		if strings.HasPrefix(part, "-") || strings.HasSuffix(part, "-") {
			return false
		}
	}

	return true
}

// RegisterStaff registers a new staff member via self-registration with email domain whitelist validation
// Story 1.9, AC5, AC7, AC10: Self-registration with whitelist validation, default role assignment, duplicate prevention
func (s *service) RegisterStaff(ctx context.Context, req StaffRegistrationRequest) (*User, string, error) {
	// 1. Extract email domain
	emailDomain := extractDomain(req.Email)
	if emailDomain == "" {
		return nil, "", errors.New("invalid email format")
	}

	// 2. Validate domain format
	if !isValidDomainFormat(emailDomain) {
		return nil, "", errors.New("invalid domain format in email address")
	}

	// 3. Normalize domain to lowercase for case-insensitive matching
	emailDomain = strings.ToLower(emailDomain)

	// 4. Validate domain against whitelist (whitelist repo must be configured)
	if s.whitelistRepo == nil {
		return nil, "", errors.New("whitelist repository not configured - self-registration is not available")
	}

	whitelistEntry, err := s.whitelistRepo.FindByDomain(ctx, emailDomain)
	if err != nil {
		return nil, "", ErrDomainNotWhitelisted
	}
	defaultRole := whitelistEntry.DefaultRole

	// 3. Use transaction for atomicity
	var createdUser *User
	var verificationToken string
	err = s.repo.Transaction(ctx, func(txCtx context.Context) error {
		// 4. Check username uniqueness
		usernameExists, err := s.repo.CheckUsernameExists(txCtx, req.Username)
		if err != nil {
			return fmt.Errorf("failed to check username existence: %w", err)
		}
		if usernameExists {
			return ErrUsernameExists
		}

		// 5. Check email uniqueness
		emailExists, err := s.repo.CheckEmailExists(txCtx, req.Email)
		if err != nil {
			return fmt.Errorf("failed to check email existence: %w", err)
		}
		if emailExists {
			return ErrEmailExists
		}

		// 6. Hash password using bcrypt (cost factor 12)
		hashedPassword, err := hashPassword(req.Password)
		if err != nil {
			return fmt.Errorf("failed to hash password: %w", err)
		}

		// 7. Create user with PENDING status and default role from whitelist
		createdUser = &User{
			Username:     req.Username,
			Name:         req.FullName,
			Email:        req.Email,
			PasswordHash: hashedPassword,
			Role:         defaultRole,
			Status:       UserStatusPending, // Story 1.9, AC7: PENDING until email verified
			BranchID:     nil, // Story 1.9, AC7: Branch assignment optional for self-registration
		}

		if err := s.repo.Create(txCtx, createdUser); err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}

		// 8. Generate email verification token (if verification repo is set)
		if s.verificationRepo != nil {
			verificationToken = uuid.New().String()
			expiresAt := time.Now().Add(24 * time.Hour) // Story 1.9, AC6: 24 hour expiration

			if err := s.verificationRepo.CreateToken(txCtx, verificationToken, req.Email, expiresAt); err != nil {
				return fmt.Errorf("failed to create verification token: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return nil, "", err
	}

	// 9. Send verification email (simulated for MVP - actual email service post-MVP)
	// Story 1.9, AC6: Token sent to user's email address
	// sendVerificationEmail(req.Email, verificationToken)

	return createdUser, verificationToken, nil
}

// VerifyEmail verifies an email verification token and activates the user account
// Story 1.9, AC6: Email verification token validation and account activation
func (s *service) VerifyEmail(ctx context.Context, token string) (*User, error) {
	if s.verificationRepo == nil {
		return nil, errors.New("verification repository not available")
	}

	// 1. Find verification token
	verificationToken, err := s.verificationRepo.FindByToken(ctx, token)
	if err != nil {
		if errors.Is(err, ErrVerificationTokenNotFound) {
			return nil, ErrInvalidVerificationToken
		}
		return nil, fmt.Errorf("failed to find verification token: %w", err)
	}

	// 2. Check token expiration
	if verificationToken.IsExpired() {
		return nil, ErrInvalidVerificationToken
	}

	// 3. Find user by email
	user, err := s.repo.FindByEmail(ctx, verificationToken.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	// 4. Delete verification token FIRST to prevent replay attacks (one-time use)
	if err := s.verificationRepo.DeleteToken(ctx, token); err != nil {
		return nil, fmt.Errorf("failed to delete verification token: %w", err)
	}

	// 5. Activate user account (PENDING → ACTIVE)
	user.Status = UserStatusActive
	if err := s.repo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to activate user: %w", err)
	}

	return user, nil
}

// DeactivateUser deactivates a user account with reason
// Story 1.10, AC1-AC8: User deactivation with token revocation and audit trail
func (s *service) DeactivateUser(ctx context.Context, targetUserID uint, adminID uint, reason string) (*User, error) {
	// AC4: Prevent self-deactivation
	if targetUserID == adminID {
		return nil, ErrCannotDeactivateSelf
	}

	var deactivatedUser *User
	err := s.repo.Transaction(ctx, func(txCtx context.Context) error {
		// 1. Find target user
		user, err := s.repo.FindByID(txCtx, targetUserID)
		if err != nil {
			return fmt.Errorf("failed to find user: %w", err)
		}
		if user == nil {
			return ErrUserNotFound
		}

		// AC7: Check if already inactive (idempotent operation)
		if user.Status == UserStatusInactive {
			// Return user without error (already deactivated)
			deactivatedUser = user
			return nil
		}

		// AC3: Revoke all active tokens FIRST before status change
		if s.sessionManager != nil {
			if err := s.sessionManager.RevokeAllUserTokens(txCtx, targetUserID); err != nil {
				// Log warning but don't fail - tokens will expire naturally
				slog.Warn("Failed to revoke tokens during deactivation",
					"error", err, "user_id", targetUserID)
			}
		}

		// AC1: Update user status and deactivation fields
		now := time.Now()
		user.Status = UserStatusInactive
		user.DeactivatedAt = &now
		user.DeactivatedBy = &adminID
		user.DeactivationReason = reason

		if err := s.repo.Update(txCtx, user); err != nil {
			return fmt.Errorf("failed to deactivate user: %w", err)
		}

		deactivatedUser = user
		return nil
	})

	if err != nil {
		return nil, err
	}

	return deactivatedUser, nil
}
