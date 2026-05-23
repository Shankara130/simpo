package dto

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDailySalesSummaryDTOSerialization tests JSON serialization of DailySalesSummaryDTO
// Story 5.1, Task 1.1, AC1: Verify DTO can be serialized to JSON with camelCase field names
func TestDailySalesSummaryDTOSerialization(t *testing.T) {
	// Create a complete DailySalesSummaryDTO with all fields populated
	dto := DailySalesSummaryDTO{
		Date:              "2026-05-23",
		BranchID:          2,
		BranchName:        "Apotek Sehat - Jakarta Pusat",
		TotalSales:        "15000000.00",
		TotalTransactions: 45,
		PaymentBreakdown: []PaymentBreakdown{
			{
				PaymentMethod:   "CASH",
				Amount:          "8000000.00",
				TransactionCount: 25,
				Percentage:      53.33,
			},
			{
				PaymentMethod:   "TRANSFER",
				Amount:          "5000000.00",
				TransactionCount: 15,
				Percentage:      33.33,
			},
			{
				PaymentMethod:   "E-WALLET",
				Amount:          "2000000.00",
				TransactionCount: 5,
				Percentage:      13.34,
			},
		},
		TopProducts: []TopProduct{
			{
				ProductID:    123,
				SKU:          "SKU-001",
				Name:         "Paracetamol 500mg",
				QuantitySold: 20,
				Revenue:      "1500000.00",
			},
		},
		HourlySales: []HourlySales{
			{
				Hour:            8,
				TransactionCount: 5,
				TotalAmount:     "1500000.00",
			},
			{
				Hour:            9,
				TransactionCount: 10,
				TotalAmount:     "3000000.00",
			},
		},
		GeneratedAt: time.Date(2026, 5, 23, 15, 30, 0, 0, time.UTC),
	}

	// Serialize to JSON
	jsonBytes, err := json.Marshal(dto)
	require.NoError(t, err, "JSON serialization should succeed")

	// Verify JSON structure
	var result map[string]interface{}
	err = json.Unmarshal(jsonBytes, &result)
	require.NoError(t, err, "JSON deserialization should succeed")

	// Assert camelCase field names (Task 1.5 requirement)
	assert.Contains(t, result, "date", "Should have 'date' field (camelCase)")
	assert.Contains(t, result, "branchId", "Should have 'branchId' field (camelCase)")
	assert.Contains(t, result, "branchName", "Should have 'branchName' field (camelCase)")
	assert.Contains(t, result, "totalSales", "Should have 'totalSales' field (camelCase)")
	assert.Contains(t, result, "totalTransactions", "Should have 'totalTransactions' field (camelCase)")
	assert.Contains(t, result, "paymentBreakdown", "Should have 'paymentBreakdown' field (camelCase)")
	assert.Contains(t, result, "topProducts", "Should have 'topProducts' field (camelCase)")
	assert.Contains(t, result, "hourlySales", "Should have 'hourlySales' field (camelCase)")
	assert.Contains(t, result, "generatedAt", "Should have 'generatedAt' field (camelCase)")

	// Assert no snake_case fields
	assert.NotContains(t, result, "total_sales", "Should NOT have snake_case 'total_sales'")
	assert.NotContains(t, result, "branch_id", "Should NOT have snake_case 'branch_id'")
}

// TestPaymentBreakdownStructure tests PaymentBreakdown struct has all required fields
// Story 5.1, Task 1.2, AC1: Payment method breakdown with percentage
func TestPaymentBreakdownStructure(t *testing.T) {
	breakdown := PaymentBreakdown{
		PaymentMethod:    "CASH",
		Amount:           "8000000.00",
		TransactionCount: 25,
		Percentage:       53.33,
	}

	// Serialize to JSON and verify field structure
	jsonBytes, err := json.Marshal(breakdown)
	require.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal(jsonBytes, &result)
	require.NoError(t, err)

	// Assert required fields exist with camelCase naming
	assert.Equal(t, "CASH", result["paymentMethod"], "paymentMethod should be CASH")
	assert.Equal(t, "8000000.00", result["amount"], "amount should be '8000000.00'")
	assert.Equal(t, float64(25), result["transactionCount"], "transactionCount should be 25")
	assert.Equal(t, 53.33, result["percentage"], "percentage should be 53.33")
}

// TestTopProductStructure tests TopProduct struct has all required fields
// Story 5.1, Task 1.3, AC1: Top 10 products by quantity and revenue
func TestTopProductStructure(t *testing.T) {
	product := TopProduct{
		ProductID:    123,
		SKU:          "SKU-001",
		Name:         "Paracetamol 500mg",
		QuantitySold: 20,
		Revenue:      "1500000.00",
	}

	// Serialize to JSON and verify field structure
	jsonBytes, err := json.Marshal(product)
	require.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal(jsonBytes, &result)
	require.NoError(t, err)

	// Assert required fields exist with camelCase naming
	assert.Equal(t, float64(123), result["productId"], "productId should be 123")
	assert.Equal(t, "SKU-001", result["sku"], "sku should be 'SKU-001'")
	assert.Equal(t, "Paracetamol 500mg", result["name"], "name should be 'Paracetamol 500mg'")
	assert.Equal(t, float64(20), result["quantitySold"], "quantitySold should be 20")
	assert.Equal(t, "1500000.00", result["revenue"], "revenue should be '1500000.00'")
}

// TestHourlySalesStructure tests HourlySales struct has all required fields
// Story 5.1, Task 1.4, AC1: Sales by hour for operational insights
func TestHourlySalesStructure(t *testing.T) {
	hourly := HourlySales{
		Hour:            8,
		TransactionCount: 5,
		TotalAmount:     "1500000.00",
	}

	// Serialize to JSON and verify field structure
	jsonBytes, err := json.Marshal(hourly)
	require.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal(jsonBytes, &result)
	require.NoError(t, err)

	// Assert required fields exist with camelCase naming
	assert.Equal(t, float64(8), result["hour"], "hour should be 8")
	assert.Equal(t, float64(5), result["transactionCount"], "transactionCount should be 5")
	assert.Equal(t, "1500000.00", result["totalAmount"], "totalAmount should be '1500000.00'")
}

// TestDailySalesRequestStructure tests DailySalesRequest struct has required fields
// Story 5.1, Task 3.2, AC1, AC2: Date and optional branch_id parameters
func TestDailySalesRequestStructure(t *testing.T) {
	// Test with all fields
	reqWithBranch := DailySalesRequest{
		Date:     "2026-05-23",
		BranchID: uintPtr(2),
	}

	jsonBytes, err := json.Marshal(reqWithBranch)
	require.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal(jsonBytes, &result)
	require.NoError(t, err)

	assert.Equal(t, "2026-05-23", result["date"], "date should be '2026-05-23'")
	assert.Equal(t, float64(2), result["branchId"], "branchId should be 2")

	// Test with nil branch_id (all branches)
	reqAllBranches := DailySalesRequest{
		Date:     "2026-05-23",
		BranchID: nil,
	}

	jsonBytes, err = json.Marshal(reqAllBranches)
	require.NoError(t, err)

	err = json.Unmarshal(jsonBytes, &result)
	require.NoError(t, err)

	assert.Equal(t, "2026-05-23", result["date"], "date should be '2026-05-23'")
	assert.Nil(t, result["branchId"], "branchId should be null when requesting all branches")
}

// TestDailySalesSummaryDTOEmptyFields tests DTO handles empty/zero values correctly
// Story 5.1, Task 1: Verify DTO doesn't crash with empty data
func TestDailySalesSummaryDTOEmptyFields(t *testing.T) {
	dto := DailySalesSummaryDTO{
		// Only required fields
		Date:              "2026-05-23",
		TotalSales:        "0.00",
		TotalTransactions: 0,
		GeneratedAt:       time.Now(),
	}

	// Should serialize without errors
	jsonBytes, err := json.Marshal(dto)
	require.NoError(t, err, "Empty DTO should serialize successfully")
	assert.NotEmpty(t, jsonBytes, "Serialized JSON should not be empty")
}

// uintPtr is a helper function to create a pointer to uint
func uintPtr(i uint) *uint {
	return &i
}
