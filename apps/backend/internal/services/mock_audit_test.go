package services

import (
	"context"
)

// MockAuditService is a mock implementation of AuditService for testing
type MockAuditService struct {
	LogLoginAttemptFunc           func(ctx context.Context, entry AuditLogEntry) error
	LogAuthorizationFailureFunc    func(ctx context.Context, entry AuditLogEntry) error
	LogUserCreationFunc            func(ctx context.Context, adminID uint, createdUserID uint, adminUsername string, createdUsername string, ipAddress string) error
	LogCount                      int // Track how many times logging was called
}

func (m *MockAuditService) LogLoginAttempt(ctx context.Context, entry AuditLogEntry) error {
	m.LogCount++
	if m.LogLoginAttemptFunc != nil {
		return m.LogLoginAttemptFunc(ctx, entry)
	}
	return nil
}

func (m *MockAuditService) LogAuthorizationFailure(ctx context.Context, entry AuditLogEntry) error {
	m.LogCount++
	if m.LogAuthorizationFailureFunc != nil {
		return m.LogAuthorizationFailureFunc(ctx, entry)
	}
	return nil
}

// LogUserCreation logs user creation actions (Story 1.7, AC7)
func (m *MockAuditService) LogUserCreation(ctx context.Context, adminID uint, createdUserID uint, adminUsername string, createdUsername string, ipAddress string) error {
	m.LogCount++
	if m.LogUserCreationFunc != nil {
		return m.LogUserCreationFunc(ctx, adminID, createdUserID, adminUsername, createdUsername, ipAddress)
	}
	return nil
}
