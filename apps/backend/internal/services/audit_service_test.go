package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestAuditService_LogBlockedSaleAttempt tests logging blocked sale attempts
func TestAuditService_LogBlockedSaleAttempt(t *testing.T) {
	// Arrange
	service := NewAuditService()

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
	)

	// Assert
	assert.NoError(t, err, "Should log blocked sale attempt without error")
}
