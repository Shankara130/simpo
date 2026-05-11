package user

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestService_RegisterUserForAdmin_SuccessfulCreation tests successful user creation for all roles
func TestService_RegisterUserForAdmin_SuccessfulCreation(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo)

	t.Run("create SYSTEM_ADMIN user", func(t *testing.T) {
		req := CreateUserRequest{
			Username: "adminuser",
			Password: "SecurePass123!",
			Email:    "admin@example.com",
			Role:     RoleSystemAdmin,
		}

		// Mock expectations
		mockRepo.On("CheckUsernameExists", context.Background(), "adminuser").Return(false, nil)
		mockRepo.On("CheckEmailExists", context.Background(), "admin@example.com").Return(false, nil)
		mockRepo.On("Create", context.Background(), mock.AnythingOfType("*user.User")).Return(nil)
		mockRepo.On("FindByID", context.Background(), mock.AnythingOfType("uint")).Return(&User{
			ID:       1,
			Username: "adminuser",
			Email:    "admin@example.com",
			Role:     RoleSystemAdmin,
			Status:   UserStatusActive,
		}, nil)

		result, err := service.RegisterUserForAdmin(context.Background(), req, 100)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "adminuser", result.Username)
		assert.Equal(t, "admin@example.com", result.Email)
		assert.Equal(t, RoleSystemAdmin, result.Role)
		assert.Equal(t, UserStatusActive, result.Status)
		assert.Nil(t, result.BranchID)

		mockRepo.AssertExpectations(t)
		mockRepo.ExpectedCalls = nil // Reset for next test
	})

	t.Run("create OWNER user", func(t *testing.T) {
		req := CreateUserRequest{
			Username: "owneruser",
			Password: "SecurePass123!",
			Email:    "owner@example.com",
			Role:     RoleOwner,
		}

		mockRepo.On("CheckUsernameExists", context.Background(), "owneruser").Return(false, nil)
		mockRepo.On("CheckEmailExists", context.Background(), "owner@example.com").Return(false, nil)
		mockRepo.On("Create", context.Background(), mock.AnythingOfType("*user.User")).Return(nil)
		mockRepo.On("FindByID", context.Background(), mock.AnythingOfType("uint")).Return(&User{
			ID:       2,
			Username: "owneruser",
			Email:    "owner@example.com",
			Role:     RoleOwner,
			Status:   UserStatusActive,
		}, nil)

		result, err := service.RegisterUserForAdmin(context.Background(), req, 100)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "owneruser", result.Username)
		assert.Equal(t, "owner@example.com", result.Email)
		assert.Equal(t, RoleOwner, result.Role)

		mockRepo.AssertExpectations(t)
		mockRepo.ExpectedCalls = nil
	})

	t.Run("create CASHIER user with branch_id", func(t *testing.T) {
		branchID := uint(5)
		req := CreateUserRequest{
			Username: "cashieruser",
			Password: "SecurePass123!",
			Email:    "cashier@example.com",
			Role:     RoleCashier,
			BranchID: &branchID,
		}

		mockRepo.On("CheckUsernameExists", context.Background(), "cashieruser").Return(false, nil)
		mockRepo.On("CheckEmailExists", context.Background(), "cashier@example.com").Return(false, nil)
		mockRepo.On("CheckBranchExists", context.Background(), uint(5)).Return(true, nil)
		mockRepo.On("Create", context.Background(), mock.AnythingOfType("*user.User")).Return(nil)
		mockRepo.On("FindByID", context.Background(), mock.AnythingOfType("uint")).Return(&User{
			ID:       3,
			Username: "cashieruser",
			Email:    "cashier@example.com",
			Role:     RoleCashier,
			BranchID: &branchID,
			Status:   UserStatusActive,
		}, nil)

		result, err := service.RegisterUserForAdmin(context.Background(), req, 100)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "cashieruser", result.Username)
		assert.Equal(t, RoleCashier, result.Role)
		assert.NotNil(t, result.BranchID)
		assert.Equal(t, uint(5), *result.BranchID)

		mockRepo.AssertExpectations(t)
		mockRepo.ExpectedCalls = nil
	})
}

// TestService_RegisterUserForAdmin_DuplicateUsername tests duplicate username prevention
func TestService_RegisterUserForAdmin_DuplicateUsername(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo)

	req := CreateUserRequest{
		Username: "existinguser",
		Password: "SecurePass123!",
		Email:    "new@example.com",
		Role:     RoleCashier,
	}

	mockRepo.On("CheckUsernameExists", context.Background(), "existinguser").Return(true, nil)

	result, err := service.RegisterUserForAdmin(context.Background(), req, 100)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, ErrUsernameExists, err)

	mockRepo.AssertExpectations(t)
}

// TestService_RegisterUserForAdmin_DuplicateEmail tests duplicate email prevention
func TestService_RegisterUserForAdmin_DuplicateEmail(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo)

	req := CreateUserRequest{
		Username: "newuser",
		Password: "SecurePass123!",
		Email:    "existing@example.com",
		Role:     RoleCashier,
	}

	mockRepo.On("CheckUsernameExists", context.Background(), "newuser").Return(false, nil)
	mockRepo.On("CheckEmailExists", context.Background(), "existing@example.com").Return(true, nil)

	result, err := service.RegisterUserForAdmin(context.Background(), req, 100)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, ErrEmailExists, err)

	mockRepo.AssertExpectations(t)
}

// TestService_RegisterUserForAdmin_InvalidRole tests invalid role validation
func TestService_RegisterUserForAdmin_InvalidRole(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo)

	req := CreateUserRequest{
		Username: "testuser",
		Password: "SecurePass123!",
		Email:    "test@example.com",
		Role:     "INVALID_ROLE",
	}

	result, err := service.RegisterUserForAdmin(context.Background(), req, 100)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, ErrInvalidRoleForCreate, err)

	// No repository calls should be made for invalid role
	mockRepo.AssertNotCalled(t, "CheckUsernameExists", mock.Anything, mock.Anything)
	mockRepo.AssertNotCalled(t, "CheckEmailExists", mock.Anything, mock.Anything)
}

// TestService_RegisterUserForAdmin_MissingBranchIDForCashier tests branch_id validation for CASHIER role
func TestService_RegisterUserForAdmin_MissingBranchIDForCashier(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo)

	req := CreateUserRequest{
		Username: "cashieruser",
		Password: "SecurePass123!",
		Email:    "cashier@example.com",
		Role:     RoleCashier,
		BranchID: nil, // Missing branch_id
	}

	// Set up mocks for username/email checks (they happen before branch_id validation)
	mockRepo.On("CheckUsernameExists", context.Background(), "cashieruser").Return(false, nil)
	mockRepo.On("CheckEmailExists", context.Background(), "cashier@example.com").Return(false, nil)

	result, err := service.RegisterUserForAdmin(context.Background(), req, 100)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, ErrBranchIDRequired, err)

	mockRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
	mockRepo.AssertExpectations(t)
}

// TestService_RegisterUserForAdmin_BranchNotFound tests validation for non-existent branch
func TestService_RegisterUserForAdmin_BranchNotFound(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo)

	branchID := uint(999)
	req := CreateUserRequest{
		Username: "cashieruser",
		Password: "SecurePass123!",
		Email:    "cashier@example.com",
		Role:     RoleCashier,
		BranchID: &branchID,
	}

	mockRepo.On("CheckUsernameExists", context.Background(), "cashieruser").Return(false, nil)
	mockRepo.On("CheckEmailExists", context.Background(), "cashier@example.com").Return(false, nil)
	mockRepo.On("CheckBranchExists", context.Background(), uint(999)).Return(false, nil)

	result, err := service.RegisterUserForAdmin(context.Background(), req, 100)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "branch with ID 999 does not exist")

	mockRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
	mockRepo.AssertExpectations(t)
}

// TestService_RegisterUserForAdmin_RepositoryErrors tests various repository error scenarios
func TestService_RegisterUserForAdmin_RepositoryErrors(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo)

	t.Run("error checking username existence", func(t *testing.T) {
		req := CreateUserRequest{
			Username: "testuser",
			Password: "SecurePass123!",
			Email:    "test@example.com",
			Role:     RoleCashier,
		}

		mockRepo.On("CheckUsernameExists", context.Background(), "testuser").Return(false, errors.New("database error"))

		result, err := service.RegisterUserForAdmin(context.Background(), req, 100)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to check username existence")

		mockRepo.AssertExpectations(t)
		mockRepo.ExpectedCalls = nil
	})

	t.Run("error checking email existence", func(t *testing.T) {
		req := CreateUserRequest{
			Username: "testuser",
			Password: "SecurePass123!",
			Email:    "test@example.com",
			Role:     RoleCashier,
		}

		mockRepo.On("CheckUsernameExists", context.Background(), "testuser").Return(false, nil)
		mockRepo.On("CheckEmailExists", context.Background(), "test@example.com").Return(false, errors.New("database error"))

		result, err := service.RegisterUserForAdmin(context.Background(), req, 100)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to check email existence")

		mockRepo.AssertExpectations(t)
		mockRepo.ExpectedCalls = nil
	})

	t.Run("error creating user", func(t *testing.T) {
		branchID := uint(1)
		req := CreateUserRequest{
			Username: "testuser",
			Password: "SecurePass123!",
			Email:    "test@example.com",
			Role:     RoleCashier,
			BranchID: &branchID,
		}

		mockRepo.On("CheckUsernameExists", context.Background(), "testuser").Return(false, nil)
		mockRepo.On("CheckEmailExists", context.Background(), "test@example.com").Return(false, nil)
			mockRepo.On("CheckBranchExists", context.Background(), uint(1)).Return(true, nil)
		mockRepo.On("Create", context.Background(), mock.AnythingOfType("*user.User")).Return(errors.New("failed to create user"))

		result, err := service.RegisterUserForAdmin(context.Background(), req, 100)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to create user")

		mockRepo.AssertExpectations(t)
		mockRepo.ExpectedCalls = nil
	})

	t.Run("error reloading user after creation", func(t *testing.T) {
		branchID := uint(1)
		req := CreateUserRequest{
			Username: "testuser",
			Password: "SecurePass123!",
			Email:    "test@example.com",
			Role:     RoleCashier,
			BranchID: &branchID,
		}

		mockRepo.On("CheckUsernameExists", context.Background(), "testuser").Return(false, nil)
		mockRepo.On("CheckEmailExists", context.Background(), "test@example.com").Return(false, nil)
		mockRepo.On("CheckBranchExists", context.Background(), uint(1)).Return(true, nil)
		mockRepo.On("Create", context.Background(), mock.AnythingOfType("*user.User")).Return(nil)
		mockRepo.On("FindByID", context.Background(), mock.AnythingOfType("uint")).Return(nil, errors.New("failed to reload user"))

		result, err := service.RegisterUserForAdmin(context.Background(), req, 100)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to reload user")

		mockRepo.AssertExpectations(t)
		mockRepo.ExpectedCalls = nil
	})

	t.Run("nil user returned after successful creation", func(t *testing.T) {
		branchID := uint(1)
		req := CreateUserRequest{
			Username: "testuser",
			Password: "SecurePass123!",
			Email:    "test@example.com",
			Role:     RoleCashier,
			BranchID: &branchID,
		}

		mockRepo.On("CheckUsernameExists", context.Background(), "testuser").Return(false, nil)
		mockRepo.On("CheckEmailExists", context.Background(), "test@example.com").Return(false, nil)
		mockRepo.On("CheckBranchExists", context.Background(), uint(1)).Return(true, nil)
		mockRepo.On("Create", context.Background(), mock.AnythingOfType("*user.User")).Return(nil)
		mockRepo.On("FindByID", context.Background(), mock.AnythingOfType("uint")).Return(nil, nil) // nil user, nil error

		result, err := service.RegisterUserForAdmin(context.Background(), req, 100)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to reload user")

		mockRepo.AssertExpectations(t)
		mockRepo.ExpectedCalls = nil
	})
}

// TestService_RegisterUserForAdmin_PasswordHashing tests that password is properly hashed
func TestService_RegisterUserForAdmin_PasswordHashing(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo)

	req := CreateUserRequest{
		Username: "testuser",
		Password: "plainPassword123",
		Email:    "test@example.com",
		Role:     RoleOwner,
	}

	var capturedUser *User
	mockRepo.On("CheckUsernameExists", context.Background(), "testuser").Return(false, nil)
	mockRepo.On("CheckEmailExists", context.Background(), "test@example.com").Return(false, nil)
	mockRepo.On("Create", context.Background(), mock.MatchedBy(func(u *User) bool {
		capturedUser = u
		return u.Username == "testuser" && u.Email == "test@example.com"
	})).Return(nil)
	mockRepo.On("FindByID", context.Background(), mock.AnythingOfType("uint")).Return(&User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
		Role:     RoleOwner,
		Status:   UserStatusActive,
	}, nil)

	result, err := service.RegisterUserForAdmin(context.Background(), req, 100)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, capturedUser)
	assert.NotEqual(t, "plainPassword123", capturedUser.PasswordHash)
	assert.NotEmpty(t, capturedUser.PasswordHash)
	assert.Greater(t, len(capturedUser.PasswordHash), 50) // bcrypt hashes are longer

	// Verify the hash is a valid bcrypt hash
	err = verifyPassword(capturedUser.PasswordHash, "plainPassword123")
	assert.NoError(t, err, "Hashed password should verify correctly")

	mockRepo.AssertExpectations(t)
}

// TestService_RegisterUserForAdmin_UserFields tests that user fields are set correctly
func TestService_RegisterUserForAdmin_UserFields(t *testing.T) {
	mockRepo := new(MockRepository)
	service := NewService(mockRepo)

	branchID := uint(10)
	req := CreateUserRequest{
		Username: "fieldstest",
		Password: "SecurePass123!",
		Email:    "fields@example.com",
		Role:     RoleCashier,
		BranchID: &branchID,
	}

	var capturedUser *User
	mockRepo.On("CheckUsernameExists", context.Background(), "fieldstest").Return(false, nil)
	mockRepo.On("CheckEmailExists", context.Background(), "fields@example.com").Return(false, nil)
	mockRepo.On("CheckBranchExists", context.Background(), uint(10)).Return(true, nil)
	mockRepo.On("Create", context.Background(), mock.MatchedBy(func(u *User) bool {
		capturedUser = u
		return true
	})).Return(nil)
	mockRepo.On("FindByID", context.Background(), mock.AnythingOfType("uint")).Return(&User{
		ID:       1,
		Username: "fieldstest",
		Email:    "fields@example.com",
		Role:     RoleCashier,
		BranchID: &branchID,
		Status:   UserStatusActive,
	}, nil)

	result, err := service.RegisterUserForAdmin(context.Background(), req, 100)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, capturedUser)
	assert.Equal(t, "fieldstest", capturedUser.Username)
	assert.Equal(t, "fields@example.com", capturedUser.Email)
	assert.Equal(t, RoleCashier, capturedUser.Role)
	assert.Equal(t, UserStatusActive, capturedUser.Status)
	assert.NotNil(t, capturedUser.BranchID)
	assert.Equal(t, uint(10), *capturedUser.BranchID)
	assert.NotEmpty(t, capturedUser.PasswordHash)

	mockRepo.AssertExpectations(t)
}
