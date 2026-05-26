package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestAuditService_LogBlockedSaleAttempt tests logging blocked sale attempts
func TestAuditService_LogBlockedSaleAttempt(t *testing.T) {
	// Arrange
	service := NewAuditService(nil) // Pass nil for repository in tests (uses stdout fallback)

	// Act
	err := service.LogBlockedSaleAttempt(
		context.Background(),
		1,                    // userID
		"testuser",           // username
		123,                  // productID
		"EXP001",             // productSKU
		"Expired Medicine",   // productName
		"2024-01-01",         // expiryDate
		"Product expired and cannot be sold", // reason
		"127.0.0.1",          // ipAddress (Story 5.4, Task 4.5)
	)

	// Assert
	assert.NoError(t, err, "Should log blocked sale attempt without error")
}
