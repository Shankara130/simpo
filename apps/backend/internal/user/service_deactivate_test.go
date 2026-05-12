package user

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

// TestService_DeactivateUser_SelfDeactivationPrevented tests that admin cannot deactivate themselves (Story 1.10, AC4)
func TestService_DeactivateUser_SelfDeactivationPrevented(t *testing.T) {
	mockRepo := new(MockRepository)
	mockSessionMgr := new(mockSessionManager)

	svc := &service{
		repo:           mockRepo,
		sessionManager: mockSessionMgr,
	}

	// Execute - try to deactivate self
	user, err := svc.DeactivateUser(context.Background(), 1, 1, "Test")

	// Verify
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrCannotDeactivateSelf))
	assert.Nil(t, user)
}

// TestService_AuthenticateUser_InactiveUserRejected tests that inactive users cannot login (Story 1.10, AC1)
func TestService_AuthenticateUser_InactiveUserRejected(t *testing.T) {
	mockRepo := new(MockRepository)

	// Setup mock - inactive user
	mockRepo.On("FindByEmail", mock.Anything, "inactive@simpo.pharmacy").Return(&User{
		ID:           10,
		Username:     "inactiveuser",
		Email:        "inactive@simpo.pharmacy",
		PasswordHash: "$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewdBPj9SjKEQ7mW",
		Status:       UserStatusInactive,
	}, nil).Once()

	svc := &service{
		repo: mockRepo,
	}

	// Execute
	request := LoginRequest{
		Email:    "inactive@simpo.pharmacy",
		Password: "password123",
	}

	user, err := svc.AuthenticateUser(context.Background(), request)

	// Verify
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrAccountDeactivated))
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "deactivated")

	mockRepo.AssertExpectations(t)
}

// TestService_AuthenticateUser_ActiveUserCanLogin tests that active users can login (Story 1.10)
func TestService_AuthenticateUser_ActiveUserCanLogin(t *testing.T) {
	mockRepo := new(MockRepository)

	// Generate valid bcrypt hash for "password123"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), 12)
	assert.NoError(t, err)

	// Setup mock - active user
	mockRepo.On("FindByEmail", mock.Anything, "active@simpo.pharmacy").Return(&User{
		ID:           11,
		Username:     "activeuser",
		Email:        "active@simpo.pharmacy",
		PasswordHash: string(hashedPassword),
		Status:       UserStatusActive,
	}, nil).Once()

	svc := &service{
		repo: mockRepo,
	}

	// Execute
	request := LoginRequest{
		Email:    "active@simpo.pharmacy",
		Password: "password123",
	}

	user, err := svc.AuthenticateUser(context.Background(), request)

	// Verify
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, uint(11), user.ID)
	assert.Equal(t, "activeuser", user.Username)

	mockRepo.AssertExpectations(t)
}
