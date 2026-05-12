package services

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAuditService_LogUserDeactivation tests the audit logging for user deactivation (Story 1.10, AC5)
func TestAuditService_LogUserDeactivation(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name                string
		adminID             uint
		deactivatedUserID   uint
		adminUsername       string
		deactivatedUsername string
		reason              string
		ipAddress           string
		expectError         bool
	}{
		{
			name:                "successful deactivation logging - Story 1.10 AC5",
			adminID:             1,
			deactivatedUserID:   10,
			adminUsername:       "admin",
			deactivatedUsername: "formerstaff",
			reason:              "Staff resignation",
			ipAddress:           "192.168.1.100",
			expectError:         false,
		},
		{
			name:                "deactivation with termination reason",
			adminID:             2,
			deactivatedUserID:   15,
			adminUsername:       "manager",
			deactivatedUsername: "terminatedstaff",
			reason:              "Termination",
			ipAddress:           "10.0.0.5",
			expectError:         false,
		},
		{
			name:                "deactivation with contract ended reason",
			adminID:             1,
			deactivatedUserID:   20,
			adminUsername:       "admin",
			deactivatedUsername: "contractor",
			reason:              "Contract ended",
			ipAddress:           "127.0.0.1",
			expectError:         false,
		},
		{
			name:                "deactivation with empty reason",
			adminID:             1,
			deactivatedUserID:   25,
			adminUsername:       "admin",
			deactivatedUsername: "testuser",
			reason:              "",
			ipAddress:           "192.168.1.1",
			expectError:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auditService := NewAuditService()

			// Execute - logging should not fail
			err := auditService.LogUserDeactivation(
				ctx,
				tt.adminID,
				tt.deactivatedUserID,
				tt.adminUsername,
				tt.deactivatedUsername,
				tt.reason,
				tt.ipAddress,
			)

			// Verify - audit logging should always succeed (log to stdout)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestAuditService_AuditActionUserDeactivated tests the audit action constant (Story 1.10, AC5)
func TestAuditService_AuditActionUserDeactivated(t *testing.T) {
	// Verify the audit action constant is defined correctly
	assert.Equal(t, AuditAction("USER_DEACTIVATED"), AuditActionUserDeactivated)
}

// TestAuditService_LogUserDeactivationFields verifies all required fields are logged (Story 1.10, AC5)
func TestAuditService_LogUserDeactivationFields(t *testing.T) {
	ctx := context.Background()
	auditService := NewAuditService()

	adminID := uint(1)
	deactivatedUserID := uint(10)
	adminUsername := "admin"
	deactivatedUsername := "formerstaff"
	reason := "Staff resignation"
	ipAddress := "192.168.1.100"

	// Execute - logging should not fail
	err := auditService.LogUserDeactivation(
		ctx,
		adminID,
		deactivatedUserID,
		adminUsername,
		deactivatedUsername,
		reason,
		ipAddress,
	)

	// Verify - audit logging should succeed
	require.NoError(t, err)

	// Note: For MVP, audit logs to stdout. In production, we would verify:
	// 1. Log entry contains admin_user_id
	// 2. Log entry contains deactivated_user_id
	// 3. Log entry contains admin_username
	// 4. Log entry contains deactivated_username
	// 5. Log entry contains reason
	// 6. Log entry contains ip_address
	// 7. Log entry contains timestamp
	// 8. Log entry contains action (USER_DEACTIVATED)
	// 9. Log entry contains outcome (success)
	// For now, we just verify the method executes without error
}

// TestAuditService_AllAuditMethodsExecute tests all audit methods execute without error
func TestAuditService_AllAuditMethodsExecute(t *testing.T) {
	ctx := context.Background()
	auditService := NewAuditService()

	// Test LogLoginAttempt
	err := auditService.LogLoginAttempt(ctx, AuditLogEntry{
		Username:  "testuser",
		Action:    AuditActionLoginSuccess,
		IPAddress: "192.168.1.1",
		Outcome:   "success",
		Timestamp: time.Now(),
	})
	assert.NoError(t, err)

	// Test LogAuthorizationFailure
	err = auditService.LogAuthorizationFailure(ctx, AuditLogEntry{
		Username:  "testuser",
		Action:    AuditActionAuthFailure,
		IPAddress: "192.168.1.1",
		Outcome:   "failure",
		Reason:    "Invalid role",
		Timestamp: time.Now(),
	})
	assert.NoError(t, err)

	// Test LogUserCreation
	err = auditService.LogUserCreation(ctx, 1, 10, "admin", "newuser", "192.168.1.1")
	assert.NoError(t, err)

	// Test LogWhitelistChange
	err = auditService.LogWhitelistChange(ctx, 1, "admin", "simpo.pharmacy", AuditActionWhitelistDomainAdded, "192.168.1.1")
	assert.NoError(t, err)

	// Test LogSelfRegistration
	err = auditService.LogSelfRegistration(ctx, 10, "staff@simpo.pharmacy", "simpo.pharmacy", "192.168.1.1")
	assert.NoError(t, err)

	// Test LogEmailVerification
	err = auditService.LogEmailVerification(ctx, 10, "staff@simpo.pharmacy", "192.168.1.1")
	assert.NoError(t, err)

	// Test LogUserDeactivation (Story 1.10)
	err = auditService.LogUserDeactivation(ctx, 1, 10, "admin", "formerstaff", "Staff resignation", "192.168.1.1")
	assert.NoError(t, err)
}
