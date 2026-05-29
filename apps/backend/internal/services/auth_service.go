package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/config"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/dto"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/models"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/user"
)

// Auth errors
var (
	ErrUserNotFound    = errors.New("user not found")
	ErrInvalidPassword = errors.New("invalid password")
	ErrUserInactive    = errors.New("user account is inactive")
	ErrEmptyUsername   = errors.New("username cannot be empty")
	ErrEmptyPassword   = errors.New("password cannot be empty")
)

// AuthInterface defines authentication service interface for handlers
type AuthInterface interface {
	Login(ctx context.Context, username, password, ipAddress string) (*dto.LoginResponse, error)
}

// UserFinder defines interface for finding users by username
type UserFinder interface {
	FindByUsername(ctx context.Context, username string) (*user.User, error)
}

// LoginResult represents successful login result (Story 1.5, AC4)
type LoginResult struct {
	User      *user.User `json:"user"`
	Token     string     `json:"token"`
	ExpiresIn int64      `json:"expires_in"`
}

// JWTClaims represents JWT token claims (Story 1.5, AC2)
type JWTClaims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	BranchID *uint  `json:"branch_id,omitempty"`
	jwt.RegisteredClaims
}

// AuthService provides authentication functionality (Story 1.5, Task 2)
type AuthService struct {
	jwtSecret      string
	accessTokenTTL time.Duration
	userRepo       UserFinder
	auditService   AuditService // Story 1.5, AC7: audit logging
}

// NewAuthService creates a new authentication service (Story 1.5, Task 2)
func NewAuthService(cfg *config.JWTConfig, userRepo UserFinder, auditService AuditService) *AuthService {
	if cfg == nil {
		panic("authService: config cannot be nil") // Prevent nil pointer dereference
	}
	if userRepo == nil {
		panic("authService: userRepo cannot be nil") // Prevent nil pointer dereference
	}
	if auditService == nil {
		panic("authService: auditService cannot be nil") // Prevent nil pointer dereference
	}

	jwtSecret := cfg.Secret
	if jwtSecret == "" {
		jwtSecret = "default-secret-change-in-production"
	}

	accessTokenTTL := cfg.AccessTokenTTL
	if accessTokenTTL == 0 {
		if cfg.TTLHours > 0 {
			accessTokenTTL = time.Duration(cfg.TTLHours) * time.Hour
		} else {
			accessTokenTTL = 8 * time.Hour // Story 1.5, NFR-SEC-002: 8 hours
		}
	}

	return &AuthService{
		jwtSecret:      jwtSecret,
		accessTokenTTL: accessTokenTTL,
		userRepo:       userRepo,
		auditService:   auditService,
	}
}

// Login authenticates user with username and password (Story 1.5, AC1, AC3, AC6, AC7)
func (s *AuthService) Login(ctx context.Context, username, password, ipAddress string) (*dto.LoginResponse, error) {
	// Validate input (Story 1.5, AC5)
	if username == "" {
		_ = s.auditService.LogLoginAttempt(ctx, AuditLogEntry{
			Username:  username,
			Action:    models.AuditActionLoginFailure,
			IPAddress: ipAddress,
			Outcome:   "EMPTY_USERNAME",
			Timestamp: time.Now(),
		})
		return nil, ErrEmptyUsername
	}
	if password == "" {
		_ = s.auditService.LogLoginAttempt(ctx, AuditLogEntry{
			Username:  username,
			Action:    models.AuditActionLoginFailure,
			IPAddress: ipAddress,
			Outcome:   "EMPTY_PASSWORD",
			Timestamp: time.Now(),
		})
		return nil, ErrEmptyPassword
	}

	// Find user by username
	foundUser, err := s.userRepo.FindByUsername(ctx, username)
	if err != nil {
		// Log failed login attempt - user not found (Story 1.5, AC7)
		_ = s.auditService.LogLoginAttempt(ctx, AuditLogEntry{
			Username:  username,
			Action:    models.AuditActionLoginFailure,
			IPAddress: ipAddress,
			Outcome:   "USER_NOT_FOUND",
			Reason:    "username not found in database",
			Timestamp: time.Now(),
		})
		if errors.Is(err, ErrUserNotFound) || errors.Is(err, context.Canceled) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	// Check user status (Story 1.5, AC6)
	if foundUser.Status != user.UserStatusActive {
		// Log failed login attempt - inactive user (Story 1.5, AC7)
		_ = s.auditService.LogLoginAttempt(ctx, AuditLogEntry{
			UserID:    &foundUser.ID,
			Username:  username,
			Action:    models.AuditActionLoginFailure,
			IPAddress: ipAddress,
			Outcome:   "USER_INACTIVE",
			Reason:    "user account is inactive",
			Timestamp: time.Now(),
		})
		return nil, ErrUserInactive
	}

	// Compare password using bcrypt (Story 1.5, AC3, cost factor 12)
	// Guard against nil/empty password hash (prevents runtime panic)
	if foundUser.PasswordHash == "" {
		_ = s.auditService.LogLoginAttempt(ctx, AuditLogEntry{
			UserID:    &foundUser.ID,
			Username:  username,
			Action:    models.AuditActionLoginFailure,
			IPAddress: ipAddress,
			Outcome:   "INVALID_USER_STATE",
			Reason:    "password hash is missing",
			Timestamp: time.Now(),
		})
		return nil, fmt.Errorf("invalid user state: missing password hash")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(foundUser.PasswordHash), []byte(password)); err != nil {
		// Log failed login attempt - invalid password (Story 1.5, AC7)
		_ = s.auditService.LogLoginAttempt(ctx, AuditLogEntry{
			UserID:    &foundUser.ID,
			Username:  username,
			Action:    models.AuditActionLoginFailure,
			IPAddress: ipAddress,
			Outcome:   "INVALID_PASSWORD",
			Reason:    "password does not match",
			Timestamp: time.Now(),
		})
		return nil, ErrInvalidPassword
	}

	// Generate JWT token (Story 1.5, AC2)
	token, err := s.generateToken(foundUser)
	if err != nil {
		// Log token generation failure (Story 1.5, AC7)
		_ = s.auditService.LogLoginAttempt(ctx, AuditLogEntry{
			UserID:    &foundUser.ID,
			Username:  username,
			Action:    models.AuditActionLoginFailure,
			IPAddress: ipAddress,
			Outcome:   "TOKEN_GENERATION_FAILED",
			Reason:    err.Error(),
			Timestamp: time.Now(),
		})
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Log successful login attempt (Story 1.5, AC7)
	_ = s.auditService.LogLoginAttempt(ctx, AuditLogEntry{
		UserID:    &foundUser.ID,
		Username:  username,
		Action:    models.AuditActionLoginSuccess,
		IPAddress: ipAddress,
		Outcome:   "SUCCESS",
		Timestamp: time.Now(),
	})

	return &dto.LoginResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   int(s.accessTokenTTL.Seconds()),
		User: dto.UserInfo{
			ID:       foundUser.ID,
			Username: foundUser.Username,
			Email:    foundUser.Email,
			Role:     foundUser.Role,
			BranchID: foundUser.BranchID,
		},
	}, nil
}

// generateToken generates JWT token for user (Story 1.5, AC2)
// Story 1.8, AC1: Token includes unique ID (jti claim) for session tracking
func (s *AuthService) generateToken(u *user.User) (string, error) {
	now := time.Now()
	expirationTime := now.Add(s.accessTokenTTL)

	// Generate unique token ID for session tracking (Story 1.8, Task 1)
	tokenID := uuid.New().String()

	claims := &JWTClaims{
		UserID:   u.ID,
		Username: u.Username,
		Email:    u.Email,
		Role:     u.Role,
		BranchID: u.BranchID,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        tokenID, // Story 1.8: JWT ID for tracking individual tokens
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    "simpo-api",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}
