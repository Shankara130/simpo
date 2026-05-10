package services

import (
	"context"
)

// MockAuditService is a mock implementation of AuditService for testing
type MockAuditService struct {
	LogLoginAttemptFunc           func(ctx context.Context, entry AuditLogEntry) error
	LogAuthorizationFailureFunc    func(ctx context.Context, entry AuditLogEntry) error
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
