package user

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupServiceTestDB creates a test environment with in-memory database
func setupServiceTestDB(t *testing.T) (*gorm.DB, Service, Repository, *mockWhitelistRepo, *mockVerificationRepo) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Migrate tables
	err = db.AutoMigrate(&User{}, &EmailVerificationToken{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	repo := NewRepository(db)
	mockWhitelist := &mockWhitelistRepo{}
	mockVerification := &mockVerificationRepo{}

	service := NewService(repo)
	service.SetWhitelistRepo(mockWhitelist)
	service.SetVerificationRepo(mockVerification)

	return db, service, repo, mockWhitelist, mockVerification
}

// mockWhitelistRepo is a mock for WhitelistRepository
type mockWhitelistRepo struct {
 findByDomainFunc func(ctx context.Context, domain string) (*WhitelistEntry, error)
}

func (m *mockWhitelistRepo) FindByDomain(ctx context.Context, domain string) (*WhitelistEntry, error) {
	if m.findByDomainFunc != nil {
		return m.findByDomainFunc(ctx, domain)
	}
	return nil, errors.New("domain not whitelisted")
}

// mockVerificationRepo is a mock for VerificationRepository
type mockVerificationRepo struct {
	createTokenFunc  func(ctx context.Context, token, email string, expiresAt time.Time) error
	findByTokenFunc  func(ctx context.Context, token string) (*EmailVerificationToken, error)
	deleteTokenFunc  func(ctx context.Context, token string) error
}

func (m *mockVerificationRepo) CreateToken(ctx context.Context, token, email string, expiresAt time.Time) error {
	if m.createTokenFunc != nil {
		return m.createTokenFunc(ctx, token, email, expiresAt)
	}
	return nil
}

func (m *mockVerificationRepo) FindByToken(ctx context.Context, token string) (*EmailVerificationToken, error) {
	if m.findByTokenFunc != nil {
		return m.findByTokenFunc(ctx, token)
	}
	return nil, ErrVerificationTokenNotFound
}

func (m *mockVerificationRepo) DeleteToken(ctx context.Context, token string) error {
	if m.deleteTokenFunc != nil {
		return m.deleteTokenFunc(ctx, token)
	}
	return nil
}

// TestRegisterStaffSuccess verifies successful staff registration
func TestRegisterStaffSuccess(t *testing.T) {
	_, service, _, mockWhitelist, mockVerification := setupServiceTestDB(t)
	ctx := context.Background()

	// Mock whitelist to return CASHIER role
	mockWhitelist.findByDomainFunc = func(ctx context.Context, domain string) (*WhitelistEntry, error) {
		return &WhitelistEntry{
			Domain:      "simpo.pharmacy",
			DefaultRole: RoleCashier,
		}, nil
	}

	// Mock verification token creation
	mockVerification.createTokenFunc = func(ctx context.Context, token, email string, expiresAt time.Time) error {
		return nil
	}

	req := StaffRegistrationRequest{
		Username: "newstaff",
		Password: "SecurePass123!",
		Email:    "newstaff@simpo.pharmacy",
		FullName: "New Staff Member",
	}

	user, token, err := service.RegisterStaff(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.NotEmpty(t, token)
	assert.Equal(t, "newstaff", user.Username)
	assert.Equal(t, "newstaff@simpo.pharmacy", user.Email)
	assert.Equal(t, RoleCashier, user.Role)
	assert.Equal(t, UserStatusPending, user.Status)
}

// TestRegisterStaffDomainNotWhitelisted verifies rejection for non-whitelisted domains
func TestRegisterStaffDomainNotWhitelisted(t *testing.T) {
	_, service, _, mockWhitelist, _ := setupServiceTestDB(t)
	ctx := context.Background()

	// Mock whitelist to return error (domain not found)
	mockWhitelist.findByDomainFunc = func(ctx context.Context, domain string) (*WhitelistEntry, error) {
		return nil, errors.New("domain not whitelisted")
	}

	req := StaffRegistrationRequest{
		Username: "newstaff",
		Password: "SecurePass123!",
		Email:    "newstaff@external.com",
		FullName: "New Staff Member",
	}

	_, _, err := service.RegisterStaff(ctx, req)
	assert.Error(t, err)
	assert.Equal(t, ErrDomainNotWhitelisted, err)
}

// TestRegisterStaffDuplicateUsername verifies duplicate username rejection
func TestRegisterStaffDuplicateUsername(t *testing.T) {
	_, service, _, mockWhitelist, mockVerification := setupServiceTestDB(t)
	ctx := context.Background()

	// Mock whitelist to return CASHIER role
	mockWhitelist.findByDomainFunc = func(ctx context.Context, domain string) (*WhitelistEntry, error) {
		return &WhitelistEntry{
			Domain:      "simpo.pharmacy",
			DefaultRole: RoleCashier,
		}, nil
	}

	// Mock verification token creation
	mockVerification.createTokenFunc = func(ctx context.Context, token, email string, expiresAt time.Time) error {
		return nil
	}

	// Create first user
	req1 := StaffRegistrationRequest{
		Username: "newstaff",
		Password: "SecurePass123!",
		Email:    "newstaff@simpo.pharmacy",
		FullName: "New Staff Member",
	}

	_, _, err := service.RegisterStaff(ctx, req1)
	assert.NoError(t, err)

	// Try to create second user with same username but different email
	req2 := StaffRegistrationRequest{
		Username: "newstaff", // Duplicate username
		Password: "SecurePass456!",
		Email:    "another@simpo.pharmacy",
		FullName: "Another Staff",
	}

	_, _, err = service.RegisterStaff(ctx, req2)
	assert.Error(t, err)
	assert.Equal(t, ErrUsernameExists, err)
}

// TestRegisterStaffDuplicateEmail verifies duplicate email rejection
func TestRegisterStaffDuplicateEmail(t *testing.T) {
	_, service, _, mockWhitelist, mockVerification := setupServiceTestDB(t)
	ctx := context.Background()

	// Mock whitelist to return CASHIER role
	mockWhitelist.findByDomainFunc = func(ctx context.Context, domain string) (*WhitelistEntry, error) {
		return &WhitelistEntry{
			Domain:      "simpo.pharmacy",
			DefaultRole: RoleCashier,
		}, nil
	}

	// Mock verification token creation
	mockVerification.createTokenFunc = func(ctx context.Context, token, email string, expiresAt time.Time) error {
		return nil
	}

	// Create first user
	req1 := StaffRegistrationRequest{
		Username: "newstaff1",
		Password:  "SecurePass123!",
		Email:     "newstaff@simpo.pharmacy",
		FullName:  "New Staff Member",
	}

	_, _, err := service.RegisterStaff(ctx, req1)
	assert.NoError(t, err)

	// Try to create second user with same email but different username
	req2 := StaffRegistrationRequest{
		Username: "newstaff2", // Different username
		Password: "SecurePass456!",
		Email:     "newstaff@simpo.pharmacy", // Duplicate email
		FullName:  "Another Staff",
	}

	_, _, err = service.RegisterStaff(ctx, req2)
	assert.Error(t, err)
	assert.Equal(t, ErrEmailExists, err)
}

// TestVerifyEmailSuccess verifies successful email verification
func TestVerifyEmailSuccess(t *testing.T) {
	_, service, repo, _, mockVerification := setupServiceTestDB(t)
	ctx := context.Background()

	// Create a PENDING user
	user := &User{
		Username:     "newstaff",
		Name:         "New Staff Member",
		Email:        "newstaff@simpo.pharmacy",
		PasswordHash: "hashedpassword",
		Role:         RoleCashier,
		Status:       UserStatusPending,
	}
	err := repo.Create(ctx, user)
	assert.NoError(t, err)

	// Mock verification token lookup
	mockVerification.findByTokenFunc = func(ctx context.Context, token string) (*EmailVerificationToken, error) {
		return &EmailVerificationToken{
			Token:     token,
			Email:     "newstaff@simpo.pharmacy",
			ExpiresAt: time.Now().Add(1 * time.Hour),
		}, nil
	}

	// Verify email
	verifiedUser, err := service.VerifyEmail(ctx, "test-token-123")
	assert.NoError(t, err)
	assert.NotNil(t, verifiedUser)
	assert.Equal(t, UserStatusActive, verifiedUser.Status)
}

// TestVerifyEmailInvalidToken verifies rejection for invalid tokens
func TestVerifyEmailInvalidToken(t *testing.T) {
	_, service, _, _, mockVerification := setupServiceTestDB(t)
	ctx := context.Background()

	// Mock verification token to return not found
	mockVerification.findByTokenFunc = func(ctx context.Context, token string) (*EmailVerificationToken, error) {
		return nil, ErrVerificationTokenNotFound
	}

	_, err := service.VerifyEmail(ctx, "invalid-token")
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidVerificationToken, err)
}

// TestVerifyEmailExpiredToken verifies rejection for expired tokens
func TestVerifyEmailExpiredToken(t *testing.T) {
	_, service, _, _, mockVerification := setupServiceTestDB(t)
	ctx := context.Background()

	// Mock verification token to return expired token
	mockVerification.findByTokenFunc = func(ctx context.Context, token string) (*EmailVerificationToken, error) {
		return &EmailVerificationToken{
			Token:     token,
			Email:     "newstaff@simpo.pharmacy",
			ExpiresAt: time.Now().Add(-1 * time.Hour), // Expired
		}, nil
	}

	_, err := service.VerifyEmail(ctx, "expired-token")
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidVerificationToken, err)
}

// TestExtractDomain verifies extractDomain helper function
func TestExtractDomain(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		expected string
	}{
		{"Valid email", "user@example.com", "example.com"},
		{"Valid email with subdomain", "user@mail.example.com", "mail.example.com"},
		{"Valid email with plus", "user+tag@example.com", "example.com"},
		{"Invalid email - no at", "invalidemail", ""},
		{"Invalid email - multiple at", "user@@example.com", ""}, // Returns "" because len(parts) != 2
		{"Empty email", "", ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := extractDomain(tc.email)
			assert.Equal(t, tc.expected, result)
		})
	}
}
