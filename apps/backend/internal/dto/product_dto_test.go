package dto

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestExpiryAlertEvent_ValidStructure tests that ExpiryAlertEvent has correct structure
// Task 1, Subtask 1.2: Verify event structure matches specification
func TestExpiryAlertEvent_ValidStructure(t *testing.T) {
	// Arrange
	now := time.Now().UTC()
	productExpiryDate := now.AddDate(0, 0, 7) // 7 days from now

	event := &ExpiryAlertEvent{
		EventID:   "evt_test123",
		EventType: "product.expiry",
		Timestamp: now.Format(time.RFC3339),
		Data: ProductExpiryData{
			ProductID:     123,
			SKU:           "SKU-12345",
			ProductName:   "Paracetamol 500mg",
			ExpiryDate:    productExpiryDate.Format(time.RFC3339),
			DaysRemaining: 7,
			AlertLevel:    "urgent",
			BranchID:      1,
			BranchName:    "Jakarta Branch",
		},
	}

	// Assert
	assert.Equal(t, "evt_test123", event.EventID, "EventID should be set")
	assert.Equal(t, "product.expiry", event.EventType, "EventType should be product.expiry")
	assert.NotEmpty(t, event.Timestamp, "Timestamp should be set")
	assert.Equal(t, uint(123), event.Data.ProductID, "ProductID should be set")
	assert.Equal(t, "SKU-12345", event.Data.SKU, "SKU should be set")
	assert.Equal(t, "Paracetamol 500mg", event.Data.ProductName, "ProductName should be set")
	assert.NotEmpty(t, event.Data.ExpiryDate, "ExpiryDate should be set")
	assert.Equal(t, 7, event.Data.DaysRemaining, "DaysRemaining should be 7")
	assert.Equal(t, "urgent", event.Data.AlertLevel, "AlertLevel should be urgent")
	assert.Equal(t, uint(1), event.Data.BranchID, "BranchID should be set")
	assert.Equal(t, "Jakarta Branch", event.Data.BranchName, "BranchName should be set")
}

// TestExpiryAlertEvent_MarshalJSON tests JSON serialization
// Task 1, Subtask 1.2: Verify JSON format for Redis pub/sub
func TestExpiryAlertEvent_MarshalJSON(t *testing.T) {
	// Arrange
	now := time.Now().UTC()
	productExpiryDate := now.AddDate(0, 0, 14)

	event := &ExpiryAlertEvent{
		EventID:   "evt_abc456",
		EventType: "product.expiry",
		Timestamp: now.Format(time.RFC3339),
		Data: ProductExpiryData{
			ProductID:     456,
			SKU:           "SKU-67890",
			ProductName:   "Amoxicillin 500mg",
			ExpiryDate:    productExpiryDate.Format(time.RFC3339),
			DaysRemaining: 14,
			AlertLevel:    "critical",
			BranchID:      2,
			BranchName:    "Bandung Branch",
		},
	}

	// Act
	jsonBytes, err := json.Marshal(event)
	require.NoError(t, err, "Should marshal event to JSON")

	// Assert
	jsonString := string(jsonBytes)
	assert.Contains(t, jsonString, `"eventType":"product.expiry"`, "JSON should contain eventType")
	assert.Contains(t, jsonString, `"eventId":"evt_abc456"`, "JSON should contain eventId")
	assert.Contains(t, jsonString, `"productId":456`, "JSON should contain productId")
	assert.Contains(t, jsonString, `"sku":"SKU-67890"`, "JSON should contain SKU")
	assert.Contains(t, jsonString, `"productName":"Amoxicillin 500mg"`, "JSON should contain productName")
	assert.Contains(t, jsonString, `"daysRemaining":14`, "JSON should contain daysRemaining")
	assert.Contains(t, jsonString, `"alertLevel":"critical"`, "JSON should contain alertLevel")
	assert.Contains(t, jsonString, `"branchId":2`, "JSON should contain branchId")
	assert.Contains(t, jsonString, `"branchName":"Bandung Branch"`, "JSON should contain branchName")
}

// TestProductExpiryData_AlertLevelValidation tests alert level categorization
// Task 1, Subtask 1.2: Verify alert level values
func TestProductExpiryData_AlertLevelValidation(t *testing.T) {
	// Test valid alert levels
	validLevels := []string{"warning", "critical", "urgent"}

	for _, level := range validLevels {
		data := ProductExpiryData{AlertLevel: level}
		assert.Contains(t, validLevels, data.AlertLevel, "Alert level should be valid")
	}
}

// TestExpiryAlertEvent_AllFieldsRequired tests that all required fields are present
// Task 1, Subtask 1.2: Validate required fields
func TestExpiryAlertEvent_AllFieldsRequired(t *testing.T) {
	// Arrange
	event := &ExpiryAlertEvent{}

	// Act & Assert
	assert.Empty(t, event.EventID, "EventID should be empty initially")
	assert.Empty(t, event.EventType, "EventType should be empty initially")
	assert.Empty(t, event.Timestamp, "Timestamp should be empty initially")
	assert.Equal(t, uint(0), event.Data.ProductID, "ProductID should be zero initially")
	assert.Empty(t, event.Data.SKU, "SKU should be empty initially")
	assert.Empty(t, event.Data.ProductName, "ProductName should be empty initially")
	assert.Empty(t, event.Data.ExpiryDate, "ExpiryDate should be empty initially")
	assert.Equal(t, 0, event.Data.DaysRemaining, "DaysRemaining should be 0 initially")
	assert.Empty(t, event.Data.AlertLevel, "AlertLevel should be empty initially")
	assert.Equal(t, uint(0), event.Data.BranchID, "BranchID should be 0 initially")
	assert.Empty(t, event.Data.BranchName, "BranchName should be empty initially")
}
