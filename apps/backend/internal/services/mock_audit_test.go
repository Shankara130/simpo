package services

import (
	"context"
)

// MockAuditService is a mock implementation of AuditService for testing
type MockAuditService struct {
	LogLoginAttemptFunc           func(ctx context.Context, entry AuditLogEntry) error
	LogAuthorizationFailureFunc    func(ctx context.Context, entry AuditLogEntry) error
	LogUserCreationFunc            func(ctx context.Context, adminID uint, createdUserID uint, adminUsername string, createdUsername string, ipAddress string) error
	LogWhitelistChangeFunc        func(ctx context.Context, adminID uint, adminUsername string, domain string, action AuditAction, ipAddress string) error
	LogSelfRegistrationFunc       func(ctx context.Context, userID uint, email string, domain string, ipAddress string) error
	LogEmailVerificationFunc       func(ctx context.Context, userID uint, email string, ipAddress string) error
	LogUserDeactivationFunc       func(ctx context.Context, adminID uint, deactivatedUserID uint, adminUsername string, deactivatedUsername string, reason string, ipAddress string) error
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

// LogWhitelistChange logs whitelist domain management actions (Story 1.9, AC8)
func (m *MockAuditService) LogWhitelistChange(ctx context.Context, adminID uint, adminUsername string, domain string, action AuditAction, ipAddress string) error {
	m.LogCount++
	if m.LogWhitelistChangeFunc != nil {
		return m.LogWhitelistChangeFunc(ctx, adminID, adminUsername, domain, action, ipAddress)
	}
	return nil
}

// LogSelfRegistration logs staff self-registration actions (Story 1.9, AC8)
func (m *MockAuditService) LogSelfRegistration(ctx context.Context, userID uint, email string, domain string, ipAddress string) error {
	m.LogCount++
	if m.LogSelfRegistrationFunc != nil {
		return m.LogSelfRegistrationFunc(ctx, userID, email, domain, ipAddress)
	}
	return nil
}

// LogEmailVerification logs email verification actions (Story 1.9, AC8)
func (m *MockAuditService) LogEmailVerification(ctx context.Context, userID uint, email string, ipAddress string) error {
	m.LogCount++
	if m.LogEmailVerificationFunc != nil {
		return m.LogEmailVerificationFunc(ctx, userID, email, ipAddress)
	}
	return nil
}

// LogUserDeactivation logs user deactivation actions (Story 1.10, AC5)
func (m *MockAuditService) LogUserDeactivation(ctx context.Context, adminID uint, deactivatedUserID uint, adminUsername string, deactivatedUsername string, reason string, ipAddress string) error {
	m.LogCount++
	if m.LogUserDeactivationFunc != nil {
		return m.LogUserDeactivationFunc(ctx, adminID, deactivatedUserID, adminUsername, deactivatedUsername, reason, ipAddress)
	}
	return nil
}
