package services

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/repositories"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/user"
)

// MockFullUserRepository is a complete mock implementation of UserRepository for testing UserService
type MockFullUserRepository struct {
	mock.Mock
}

func (m *MockFullUserRepository) Create(ctx context.Context, u *user.User) error {
	args := m.Called(ctx, u)
	return args.Error(0)
}

func (m *MockFullUserRepository) GetByID(ctx context.Context, id uint) (*user.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *MockFullUserRepository) GetByUsername(ctx context.Context, username string) (*user.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *MockFullUserRepository) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *MockFullUserRepository) Update(ctx context.Context, u *user.User) error {
	args := m.Called(ctx, u)
	return args.Error(0)
}

func (m *MockFullUserRepository) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockFullUserRepository) Deactivate(ctx context.Context, id uint, reason string, deactivatedBy uint) error {
	args := m.Called(ctx, id, reason, deactivatedBy)
	return args.Error(0)
}

func (m *MockFullUserRepository) List(ctx context.Context, filter *repositories.UserFilter) ([]*user.User, int64, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*user.User), args.Get(1).(int64), args.Error(2)
}

func (m *MockFullUserRepository) GetByBranch(ctx context.Context, branchID uint) ([]*user.User, error) {
	args := m.Called(ctx, branchID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*user.User), args.Error(1)
}

func (m *MockFullUserRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	args := m.Called(ctx, username)
	return args.Bool(0), args.Error(1)
}

func (m *MockFullUserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	args := m.Called(ctx, email)
	return args.Bool(0), args.Error(1)
}

// Test NewUserService with nil dependencies
func TestNewUserService_PanicOnNilDependencies(t *testing.T) {
	assert.Panics(t, func() {
		NewUserService(nil, &MockAuditService{})
	}, "Should panic when userRepo is nil")

	assert.Panics(t, func() {
		NewUserService(&MockFullUserRepository{}, nil)
	}, "Should panic when auditService is nil")
}

// Test CreateUser
func TestUserService_CreateUser_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockFullUserRepository)
	mockAudit := new(MockAuditService)
	service := NewUserService(mockRepo, mockAudit)

	u := &user.User{
		Username:     "testuser",
		Email:        "test@example.com",
		Role:         user.RoleCashier,
		PasswordHash: "hashedpassword",
		Status:       user.UserStatusActive,
	}

	// Mock expectations
	mockRepo.On("ExistsByUsername", mock.Anything, "testuser").Return(false, nil)
	mockRepo.On("ExistsByEmail", mock.Anything, "test@example.com").Return(false, nil)
	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*user.User")).Return(nil)
	// MockAuditService uses function callbacks, not mock.On
	mockAudit.LogUserCreationFunc = func(ctx context.Context, adminID uint, createdUserID uint, adminUsername string, createdUsername string, ipAddress string) error {
		return nil
	}

	// Act
	err := service.CreateUser(context.Background(), u, "")

	// Assert
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestUserService_CreateUser_EmptyUsername(t *testing.T) {
	// Arrange
	mockRepo := new(MockFullUserRepository)
	mockAudit := new(MockAuditService)
	service := NewUserService(mockRepo, mockAudit)

	u := &user.User{
		Email:        "test@example.com",
		Role:         user.RoleCashier,
		PasswordHash: "hashedpassword",
	}

	// Act
	err := service.CreateUser(context.Background(), u, "")

	// Assert
	assert.Error(t, err)
	var invErr *InvalidInputError
	assert.True(t, errors.As(err, &invErr))
	assert.Equal(t, "username", invErr.Field)
}

func TestUserService_CreateUser_DuplicateUsername(t *testing.T) {
	// Arrange
	mockRepo := new(MockFullUserRepository)
	mockAudit := new(MockAuditService)
	service := NewUserService(mockRepo, mockAudit)

	u := &user.User{
		Username:     "existinguser",
		Email:        "test@example.com",
		Role:         user.RoleCashier,
		PasswordHash: "hashedpassword",
	}

	// Mock expectations
	mockRepo.On("ExistsByUsername", mock.Anything, "existinguser").Return(true, nil)

	// Act
	err := service.CreateUser(context.Background(), u, "")

	// Assert
	assert.Error(t, err)
	var dupErr *DuplicateUsernameError
	assert.True(t, errors.As(err, &dupErr))
	assert.Equal(t, "existinguser", dupErr.Username)
}

func TestUserService_CreateUser_DuplicateEmail(t *testing.T) {
	// Arrange
	mockRepo := new(MockFullUserRepository)
	mockAudit := new(MockAuditService)
	service := NewUserService(mockRepo, mockAudit)

	u := &user.User{
		Username:     "testuser",
		Email:        "existing@example.com",
		Role:         user.RoleCashier,
		PasswordHash: "hashedpassword",
	}

	// Mock expectations
	mockRepo.On("ExistsByUsername", mock.Anything, "testuser").Return(false, nil)
	mockRepo.On("ExistsByEmail", mock.Anything, "existing@example.com").Return(true, nil)

	// Act
	err := service.CreateUser(context.Background(), u, "")

	// Assert
	assert.Error(t, err)
	var invErr *InvalidInputError
	assert.True(t, errors.As(err, &invErr))
	assert.Equal(t, "email", invErr.Field)
}

// Test UpdateUser
func TestUserService_UpdateUser_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockFullUserRepository)
	mockAudit := new(MockAuditService)
	service := NewUserService(mockRepo, mockAudit)

	existing := &user.User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
		Role:     user.RoleCashier,
	}

	updated := &user.User{
		Email:  "newemail@example.com",
		Status: user.UserStatusInactive,
	}

	// Mock expectations
	mockRepo.On("GetByID", mock.Anything, uint(1)).Return(existing, nil)
	// PATCH: Added for email uniqueness check
	mockRepo.On("GetByEmail", mock.Anything, "newemail@example.com").Return((*user.User)(nil), gorm.ErrRecordNotFound)
	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*user.User")).Return(nil)

	// Act
	err := service.UpdateUser(context.Background(), 1, updated, 2, "")

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "testuser", updated.Username) // Username preserved
	mockRepo.AssertExpectations(t)
}

func TestUserService_UpdateUser_CannotChangeOwnRole(t *testing.T) {
	// Arrange
	mockRepo := new(MockFullUserRepository)
	mockAudit := new(MockAuditService)
	service := NewUserService(mockRepo, mockAudit)

	existing := &user.User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
		Role:     user.RoleCashier,
	}

	updated := &user.User{
		Role: user.RoleOwner, // Attempting to change own role
	}

	// Mock expectations
	mockRepo.On("GetByID", mock.Anything, uint(1)).Return(existing, nil)

	// Act
	err := service.UpdateUser(context.Background(), 1, updated, 1, "") // Same user ID

	// Assert
	assert.Error(t, err)
	var authErr *UnauthorizedError
	assert.True(t, errors.As(err, &authErr))
	assert.Equal(t, "update own role", authErr.Action)
}

// Test DeactivateUser
func TestUserService_DeactivateUser_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockFullUserRepository)
	mockAudit := new(MockAuditService)
	service := NewUserService(mockRepo, mockAudit)

	targetUser := &user.User{
		ID:       2,
		Username: "targetuser",
		Email:    "target@example.com",
	}

	adminUser := &user.User{
		ID:       1,
		Username: "adminuser",
	}

	// Mock expectations
	mockRepo.On("GetByID", mock.Anything, uint(2)).Return(targetUser, nil)
	mockRepo.On("GetByID", mock.Anything, uint(1)).Return(adminUser, nil)
	mockRepo.On("Deactivate", mock.Anything, uint(2), "Test reason", uint(1)).Return(nil)
	// MockAuditService uses function callbacks, not mock.On
	mockAudit.LogUserDeactivationFunc = func(ctx context.Context, adminID uint, deactivatedUserID uint, adminUsername string, deactivatedUsername string, reason string, ipAddress string) error {
		return nil
	}

	// Act
	err := service.DeactivateUser(context.Background(), 2, "Test reason", 1, "")

	// Assert
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestUserService_DeactivateUser_CannotDeactivateSelf(t *testing.T) {
	// Arrange
	mockRepo := new(MockFullUserRepository)
	mockAudit := new(MockAuditService)
	service := NewUserService(mockRepo, mockAudit)

	targetUser := &user.User{
		ID:       1,
		Username: "testuser",
	}

	// Mock expectations
	mockRepo.On("GetByID", mock.Anything, uint(1)).Return(targetUser, nil)

	// Act
	err := service.DeactivateUser(context.Background(), 1, "Test reason", 1, "") // Same user ID

	// Assert
	assert.Error(t, err)
	var authErr *UnauthorizedError
	assert.True(t, errors.As(err, &authErr))
	assert.Equal(t, "deactivate self", authErr.Action)
}

// Test GetUserByID
func TestUserService_GetUserByID_ZeroID(t *testing.T) {
	// Arrange
	mockRepo := new(MockFullUserRepository)
	mockAudit := new(MockAuditService)
	service := NewUserService(mockRepo, mockAudit)

	// Act
	_, err := service.GetUserByID(context.Background(), 0)

	// Assert
	assert.Error(t, err)
	var invErr *InvalidInputError
	assert.True(t, errors.As(err, &invErr))
	assert.Equal(t, "id", invErr.Field)
}

// Test ListUsers
func TestUserService_ListUsers_SanitizesWildcardCharacters(t *testing.T) {
	// Arrange
	mockRepo := new(MockFullUserRepository)
	mockAudit := new(MockAuditService)
	service := NewUserService(mockRepo, mockAudit)

	filter := &UserFilter{
		SearchQuery: "test%_wildcards",
		Page:        1,
		Limit:       20,
	}

	// Mock expectations - expect sanitized input
	mockRepo.On("List", mock.Anything, mock.MatchedBy(func(f *repositories.UserFilter) bool {
		return f.SearchQuery == "testwildcards" // Wildcards removed
	})).Return([]*user.User{}, int64(0), nil)

	// Act
	_, _, err := service.ListUsers(context.Background(), filter)

	// Assert
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestUserService_CreateUser_ContextCanceled(t *testing.T) {
	// Arrange
	mockRepo := new(MockFullUserRepository)
	mockAudit := new(MockAuditService)
	service := NewUserService(mockRepo, mockAudit)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	u := &user.User{
		Username:     "testuser",
		Email:        "test@example.com",
		Role:         user.RoleCashier,
		PasswordHash: "hashedpassword",
	}

	// Act
	err := service.CreateUser(ctx, u, "")

	// Assert
	assert.Error(t, err)
}
