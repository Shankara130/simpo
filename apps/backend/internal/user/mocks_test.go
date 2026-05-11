package user

import (
	"context"

	"github.com/stretchr/testify/mock"
)

// MockService is a mock implementation of the user service for testing handlers
type MockService struct {
	mock.Mock
}

func (m *MockService) RegisterUser(ctx context.Context, req RegisterRequest) (*User, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}

// RegisterUserForAdmin registers a new user with admin-specified role (Story 1.7)
func (m *MockService) RegisterUserForAdmin(ctx context.Context, req CreateUserRequest, adminID uint) (*User, error) {
	args := m.Called(ctx, req, adminID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}

// MockAuditLogger is a mock implementation of AuditLogger for testing
type MockAuditLogger struct {
	mock.Mock
}

func (m *MockAuditLogger) LogUserCreation(ctx context.Context, adminID uint, createdUserID uint, adminUsername string, createdUsername string, ipAddress string) error {
	args := m.Called(ctx, adminID, createdUserID, adminUsername, createdUsername, ipAddress)
	return args.Error(0)
}

func (m *MockService) AuthenticateUser(ctx context.Context, req LoginRequest) (*User, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockService) GetUserByID(ctx context.Context, id uint) (*User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockService) UpdateUser(ctx context.Context, id uint, req UpdateUserRequest) (*User, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockService) DeleteUser(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockService) ListUsers(ctx context.Context, filters UserFilterParams, page, perPage int) ([]User, int64, error) {
	args := m.Called(ctx, filters, page, perPage)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]User), args.Get(1).(int64), args.Error(2)
}

func (m *MockService) PromoteToAdmin(ctx context.Context, userID uint) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

// MockRepository is a mock implementation of the user repository for testing services
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) Create(ctx context.Context, user *User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockRepository) FindByEmail(ctx context.Context, email string) (*User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockRepository) FindByUsername(ctx context.Context, username string) (*User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockRepository) FindByID(ctx context.Context, id uint) (*User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockRepository) Update(ctx context.Context, user *User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockRepository) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRepository) ListAllUsers(ctx context.Context, filters UserFilterParams, page, perPage int) ([]User, int64, error) {
	args := m.Called(ctx, filters, page, perPage)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]User), args.Get(1).(int64), args.Error(2)
}

func (m *MockRepository) AssignRole(ctx context.Context, userID uint, roleName string) error {
	args := m.Called(ctx, userID, roleName)
	return args.Error(0)
}

func (m *MockRepository) RemoveRole(ctx context.Context, userID uint, roleName string) error {
	args := m.Called(ctx, userID, roleName)
	return args.Error(0)
}

func (m *MockRepository) FindRoleByName(ctx context.Context, roleName string) (*Role, error) {
	args := m.Called(ctx, roleName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Role), args.Error(1)
}

func (m *MockRepository) GetUserRoles(ctx context.Context, userID uint) ([]Role, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]Role), args.Error(1)
}

func (m *MockRepository) Transaction(ctx context.Context, fn func(context.Context) error) error {
	// Execute the transaction function directly for testing
	return fn(ctx)
}

// Story 1.7: CheckUsernameExists checks if a username already exists
func (m *MockRepository) CheckUsernameExists(ctx context.Context, username string) (bool, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return false, args.Error(1)
	}
	return args.Get(0).(bool), args.Error(1)
}

// Story 1.7: CheckEmailExists checks if an email already exists
func (m *MockRepository) CheckEmailExists(ctx context.Context, email string) (bool, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return false, args.Error(1)
	}
	return args.Get(0).(bool), args.Error(1)
}

// Story 1.7: CheckBranchExists checks if a branch ID exists in the branches table
func (m *MockRepository) CheckBranchExists(ctx context.Context, branchID uint) (bool, error) {
	args := m.Called(ctx, branchID)
	if args.Get(0) == nil {
		return false, args.Error(1)
	}
	return args.Get(0).(bool), args.Error(1)
}
