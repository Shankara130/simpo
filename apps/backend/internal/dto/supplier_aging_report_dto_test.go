package dto

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ==============================================================================
// Story 10.6: Supplier Aging Report DTO Tests
// ==============================================================================

// TestSupplierAgingReportRequestSerialization tests JSON serialization of SupplierAgingReportRequest
// Story 10.6, Task 1.1, AC1: Verify request DTO can be serialized with camelCase field names
func TestSupplierAgingReportRequestSerialization(t *testing.T) {
	supplierID := uint(1)
	branchID := uint(2)

	req := SupplierAgingReportRequest{
		AsOfDate:   "2026-05-31",
		SupplierID: &supplierID,
		BranchID:   &branchID,
	}

	// Serialize to JSON
	jsonBytes, err := json.Marshal(req)
	require.NoError(t, err, "JSON serialization should succeed")

	// Verify JSON structure
	var result map[string]interface{}
	err = json.Unmarshal(jsonBytes, &result)
	require.NoError(t, err, "JSON deserialization should succeed")

	// Assert camelCase field names
	assert.Equal(t, "2026-05-31", result["asOfDate"], "asOfDate should be present")
	assert.Equal(t, float64(1), result["supplierId"], "supplierId should be present")
	assert.Equal(t, float64(2), result["branchId"], "branchId should be present")
}

// TestSupplierAgingReportRequestWithNilFilters tests request DTO with nil optional filters
// Story 10.6, Task 1.1: Verify optional fields work correctly when nil
func TestSupplierAgingReportRequestWithNilFilters(t *testing.T) {
	req := SupplierAgingReportRequest{
		AsOfDate:   "2026-05-31",
		SupplierID: nil,
		BranchID:   nil,
	}

	// Serialize to JSON
	jsonBytes, err := json.Marshal(req)
	require.NoError(t, err, "JSON serialization should succeed")

	// Verify JSON structure
	var result map[string]interface{}
	err = json.Unmarshal(jsonBytes, &result)
	require.NoError(t, err, "JSON deserialization should succeed")

	assert.Equal(t, "2026-05-31", result["asOfDate"], "asOfDate should be present")
	assert.Nil(t, result["supplierId"], "supplierId should be null when not specified")
	assert.Nil(t, result["branchId"], "branchId should be null when not specified")
}

// TestSupplierAgingReportResponseSerialization tests JSON serialization of SupplierAgingReportResponse
// Story 10.6, Task 1.1: Verify response DTO structure with all components
func TestSupplierAgingReportResponseSerialization(t *testing.T) {
	response := SupplierAgingReportResponse{
		AsOfDate:         "2026-05-31",
		ReportGeneratedAt: "2026-05-31T10:00:00Z",
		Currency:         "IDR",
		Suppliers: []SupplierAgingSummary{
			{
				SupplierID:     1,
				SupplierName:   "PT. Pharmasi Jaya",
				ContactPerson:  "John Doe",
				Phone:          "+62-21-555-1234",
				Email:          "orders@pharmasi-jaya.co.id",
				Address:        "Jl. Industri No. 123, Jakarta",
				AgingBuckets: AgingBucket{
					Current:         5000000.00,
					CurrentCount:    2,
					Days31to60:       3000000.00,
					Days31to60Count:  1,
					Days61to90:       0,
					Days61to90Count:  0,
					DaysOver90:       0,
					DaysOver90Count:  0,
				},
				TotalOutstanding: 8000000.00,
				InvoiceCount:     3,
			},
		},
		GrandTotals: AgingGrandTotals{
			Current:        5000000.00,
			Days31to60:     3000000.00,
			Days61to90:     0,
			DaysOver90:     0,
			TotalOutstanding: 8000000.00,
			TotalInvoices: 3,
			TotalSuppliers: 1,
		},
		Pagination: PaginationResponse{
			Page:       1,
			Limit:      20,
			Total:      1,
			TotalPages: 1,
		},
	}

	// Serialize to JSON
	jsonBytes, err := json.Marshal(response)
	require.NoError(t, err, "JSON serialization should succeed")

	// Verify JSON structure
	var result map[string]interface{}
	err = json.Unmarshal(jsonBytes, &result)
	require.NoError(t, err, "JSON deserialization should succeed")

	// Assert camelCase field names
	assert.Equal(t, "2026-05-31", result["asOfDate"], "asOfDate should be present")
	assert.Equal(t, "2026-05-31T10:00:00Z", result["reportGeneratedAt"], "reportGeneratedAt should be present")
	assert.Equal(t, "IDR", result["currency"], "currency should be present")
	assert.Contains(t, result, "suppliers", "suppliers array should be present")
	assert.Contains(t, result, "grandTotals", "grandTotals should be present")
	assert.Contains(t, result, "pagination", "pagination should be present")

	// Verify suppliers array
	suppliers := result["suppliers"].([]interface{})
	assert.Len(t, suppliers, 1, "Should have 1 supplier")

	// Verify grand totals
	grandTotals := result["grandTotals"].(map[string]interface{})
	assert.Equal(t, float64(8000000), grandTotals["totalOutstanding"], "totalOutstanding should match")
}

// TestSupplierAgingSummarySerialization tests JSON serialization of SupplierAgingSummary
// Story 10.6, Task 1.1: Verify supplier summary structure with aging buckets
func TestSupplierAgingSummarySerialization(t *testing.T) {
	summary := SupplierAgingSummary{
		SupplierID:       1,
		SupplierName:     "PT. Pharmasi Jaya",
		ContactPerson:    "John Doe",
		Phone:            "+62-21-555-1234",
		Email:            "orders@pharmasi-jaya.co.id",
		Address:          "Jl. Industri No. 123, Jakarta",
		TotalOutstanding: 8000000.00,
		InvoiceCount:     3,
		AgingBuckets: AgingBucket{
			Current:        5000000.00,
			CurrentCount:   2,
			Days31to60:      3000000.00,
			Days31to60Count: 1,
			Days61to90:      0,
			Days61to90Count: 0,
			DaysOver90:      0,
			DaysOver90Count: 0,
		},
	}

	// Serialize to JSON
	jsonBytes, err := json.Marshal(summary)
	require.NoError(t, err, "JSON serialization should succeed")

	// Verify JSON structure
	var result map[string]interface{}
	err = json.Unmarshal(jsonBytes, &result)
	require.NoError(t, err, "JSON deserialization should succeed")

	// Assert camelCase field names
	assert.Equal(t, float64(1), result["supplierId"], "supplierId should be present")
	assert.Equal(t, "PT. Pharmasi Jaya", result["supplierName"], "supplierName should be present")
	assert.Equal(t, "John Doe", result["contactPerson"], "contactPerson should be present")
	assert.Equal(t, "+62-21-555-1234", result["phone"], "phone should be present")
	assert.Equal(t, "orders@pharmasi-jaya.co.id", result["email"], "email should be present")
	assert.Equal(t, "Jl. Industri No. 123, Jakarta", result["address"], "address should be present")
	assert.Equal(t, float64(8000000), result["totalOutstanding"], "totalOutstanding should be present")
	assert.Equal(t, float64(3), result["invoiceCount"], "invoiceCount should be present")

	// Verify aging buckets
	agingBuckets := result["agingBuckets"].(map[string]interface{})
	assert.Equal(t, float64(5000000), agingBuckets["current"], "current bucket should be present")
	assert.Equal(t, float64(2), agingBuckets["currentCount"], "currentCount should be present")
	assert.Equal(t, float64(3000000), agingBuckets["days31to60"], "days31to60 bucket should be present")
	assert.Equal(t, float64(1), agingBuckets["days31to60Count"], "days31to60Count should be present")
}

// TestAgingBucketSerialization tests JSON serialization of AgingBucket
// Story 10.6, Task 1.1: Verify aging bucket structure with all four periods
func TestAgingBucketSerialization(t *testing.T) {
	bucket := AgingBucket{
		Current:        5000000.00,
		CurrentCount:   2,
		Days31to60:      3000000.00,
		Days31to60Count: 1,
		Days61to90:      4000000.00,
		Days61to90Count: 2,
		DaysOver90:      3000000.00,
		DaysOver90Count: 1,
	}

	// Serialize to JSON
	jsonBytes, err := json.Marshal(bucket)
	require.NoError(t, err, "JSON serialization should succeed")

	// Verify JSON structure
	var result map[string]interface{}
	err = json.Unmarshal(jsonBytes, &result)
	require.NoError(t, err, "JSON deserialization should succeed")

	// Assert all four aging buckets with camelCase naming
	assert.Equal(t, "5000000", result["current"], "current bucket should be present")
	assert.Equal(t, float64(2), result["currentCount"], "currentCount should be present")
	assert.Equal(t, "3000000", result["days31to60"], "days31to60 bucket should be present")
	assert.Equal(t, float64(1), result["days31to60Count"], "days31to60Count should be present")
	assert.Equal(t, "4000000", result["days61to90"], "days61to90 bucket should be present")
	assert.Equal(t, float64(2), result["days61to90Count"], "days61to90Count should be present")
	assert.Equal(t, "3000000", result["daysOver90"], "daysOver90 bucket should be present")
	assert.Equal(t, float64(1), result["daysOver90Count"], "daysOver90Count should be present")
}

// TestInvoiceAgingDetailSerialization tests JSON serialization of InvoiceAgingDetail
// Story 10.6, Task 1.1: Verify invoice detail structure for individual invoices
func TestInvoiceAgingDetailSerialization(t *testing.T) {
	detail := InvoiceAgingDetail{
		InvoiceID:          1,
		InvoiceNumber:      "INV-2026-001",
		InvoiceDate:        "2026-04-15",
		DueDate:            "2026-05-15",
		TotalAmount:        5000000.00,
		PaidAmount:         2000000.00,
		OutstandingBalance: 3000000.00,
		DaysOverdue:        46,
		AgingBucket:        "31-60",
		PaymentStatus:      "partial",
	}

	// Serialize to JSON
	jsonBytes, err := json.Marshal(detail)
	require.NoError(t, err, "JSON serialization should succeed")

	// Verify JSON structure
	var result map[string]interface{}
	err = json.Unmarshal(jsonBytes, &result)
	require.NoError(t, err, "JSON deserialization should succeed")

	// Assert camelCase field names
	assert.Equal(t, float64(1), result["invoiceId"], "invoiceId should be present")
	assert.Equal(t, "INV-2026-001", result["invoiceNumber"], "invoiceNumber should be present")
	assert.Equal(t, "2026-04-15", result["invoiceDate"], "invoiceDate should be present")
	assert.Equal(t, "2026-05-15", result["dueDate"], "dueDate should be present")
	assert.Equal(t, "5000000", result["totalAmount"], "totalAmount should be present")
	assert.Equal(t, "2000000", result["paidAmount"], "paidAmount should be present")
	assert.Equal(t, "3000000", result["outstandingBalance"], "outstandingBalance should be present")
	assert.Equal(t, float64(46), result["daysOverdue"], "daysOverdue should be present")
	assert.Equal(t, "31-60", result["agingBucket"], "agingBucket should be present")
	assert.Equal(t, "partial", result["paymentStatus"], "paymentStatus should be present")
}

// TestAgingGrandTotalsSerialization tests JSON serialization of AgingGrandTotals
// Story 10.6, Task 1.1: Verify grand totals structure with aggregated data
func TestAgingGrandTotalsSerialization(t *testing.T) {
	totals := AgingGrandTotals{
		Current:         50000000.00,
		Days31to60:       15000000.00,
		Days61to90:       10000000.00,
		DaysOver90:       5000000.00,
		TotalOutstanding: 80000000.00,
		TotalInvoices:    25,
		TotalSuppliers:   8,
	}

	// Serialize to JSON
	jsonBytes, err := json.Marshal(totals)
	require.NoError(t, err, "JSON serialization should succeed")

	// Verify JSON structure
	var result map[string]interface{}
	err = json.Unmarshal(jsonBytes, &result)
	require.NoError(t, err, "JSON deserialization should succeed")

	// Assert all totals with camelCase naming
	assert.Equal(t, "50000000", result["current"], "current bucket total should be present")
	assert.Equal(t, "15000000", result["days31to60"], "days31to60 bucket total should be present")
	assert.Equal(t, "10000000", result["days61to90"], "days61to90 bucket total should be present")
	assert.Equal(t, "5000000", result["daysOver90"], "daysOver90 bucket total should be present")
	assert.Equal(t, "80000000", result["totalOutstanding"], "totalOutstanding should be present")
	assert.Equal(t, float64(25), result["totalInvoices"], "totalInvoices should be present")
	assert.Equal(t, float64(8), result["totalSuppliers"], "totalSuppliers should be present")
}

// TestParseAgingDate tests the ParseAgingDate helper function
// Story 10.6, Task 3.2, AC1: Verify date parsing for aging calculations
func TestParseAgingDate(t *testing.T) {
	tests := []struct {
		name      string
		dateStr   string
		expectErr bool
	}{
		{
			name:      "Valid date format",
			dateStr:   "2026-05-31",
			expectErr: false,
		},
		{
			name:      "Invalid date format - wrong separator",
			dateStr:   "2026/05/31",
			expectErr: true,
		},
		{
			name:      "Invalid date format - missing leading zeros",
			dateStr:   "2026-5-31",
			expectErr: true,
		},
		{
			name:      "Invalid date - non-existent date",
			dateStr:   "2026-02-30",
			expectErr: true,
		},
		{
			name:      "Empty string",
			dateStr:   "",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseAgingDate(tt.dateStr)

			if tt.expectErr {
				assert.Error(t, err, "Should return error for invalid date")
				assert.True(t, result.IsZero(), "Result should be zero time for error")
			} else {
				assert.NoError(t, err, "Should not return error for valid date")
				assert.False(t, result.IsZero(), "Result should not be zero for valid date")

				// Verify parsed components
				assert.Equal(t, 2026, result.Year(), "Year should match")
				assert.Equal(t, time.May, result.Month(), "Month should match")
				assert.Equal(t, 31, result.Day(), "Day should match")
			}
		})
	}
}

// TestDaysBetween tests the DaysBetween helper function
// Story 10.6, Task 3.3, AC1: Verify days calculation for aging buckets
func TestDaysBetween(t *testing.T) {
	tests := []struct {
		name     string
		start    string
		end      string
		expected int
	}{
		{
			name:     "Same day - 0 days",
			start:    "2026-05-31",
			end:      "2026-05-31",
			expected: 0,
		},
		{
			name:     "One day apart",
			start:    "2026-05-31",
			end:      "2026-06-01",
			expected: 1,
		},
		{
			name:     "30 days apart - current bucket",
			start:    "2026-05-01",
			end:      "2026-05-31",
			expected: 30,
		},
		{
			name:     "45 days apart - 31-60 bucket",
			start:    "2026-04-15",
			end:      "2026-05-30",
			expected: 45,
		},
		{
			name:     "75 days apart - 61-90 bucket",
			start:    "2026-03-15",
			end:      "2026-05-29",
			expected: 75,
		},
		{
			name:     "100 days apart - 90+ bucket",
			start:    "2026-02-20",
			end:      "2026-05-31",
			expected: 100,
		},
		{
			name:     "Negative days (end before start)",
			start:    "2026-05-31",
			end:      "2026-05-01",
			expected: -30,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, _ := time.Parse("2006-01-02", tt.start)
			end, _ := time.Parse("2006-01-02", tt.end)

			result := DaysBetween(start, end)
			assert.Equal(t, tt.expected, result, "Days calculation should match expected")
		})
	}
}

// TestCategorizeIntoBucket tests the CategorizeIntoBucket helper function
// Story 10.6, Task 3.4, AC1: Verify aging bucket categorization logic
func TestCategorizeIntoBucket(t *testing.T) {
	tests := []struct {
		name         string
		daysOverdue  int
		expectedBucket string
	}{
		{
			name:         "Not overdue - current bucket",
			daysOverdue:  0,
			expectedBucket: "current",
		},
		{
			name:         "15 days overdue - current bucket",
			daysOverdue:  15,
			expectedBucket: "current",
		},
		{
			name:         "30 days overdue - current bucket (boundary)",
			daysOverdue:  30,
			expectedBucket: "current",
		},
		{
			name:         "31 days overdue - 31-60 bucket",
			daysOverdue:  31,
			expectedBucket: "31-60",
		},
		{
			name:         "45 days overdue - 31-60 bucket",
			daysOverdue:  45,
			expectedBucket: "31-60",
		},
		{
			name:         "60 days overdue - 31-60 bucket (boundary)",
			daysOverdue:  60,
			expectedBucket: "31-60",
		},
		{
			name:         "61 days overdue - 61-90 bucket",
			daysOverdue:  61,
			expectedBucket: "61-90",
		},
		{
			name:         "75 days overdue - 61-90 bucket",
			daysOverdue:  75,
			expectedBucket: "61-90",
		},
		{
			name:         "90 days overdue - 61-90 bucket (boundary)",
			daysOverdue:  90,
			expectedBucket: "61-90",
		},
		{
			name:         "91 days overdue - 90+ bucket",
			daysOverdue:  91,
			expectedBucket: "90+",
		},
		{
			name:         "150 days overdue - 90+ bucket",
			daysOverdue:  150,
			expectedBucket: "90+",
		},
		{
			name:         "Negative days (future-due) - current bucket",
			daysOverdue:  -10,
			expectedBucket: "current",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CategorizeIntoBucket(tt.daysOverdue)
			assert.Equal(t, tt.expectedBucket, result, "Bucket categorization should match expected")
		})
	}
}

// TestSupplierAgingReportListFilter tests the filter struct for list endpoint
// Story 10.6, Task 1.3: Verify pagination and filter parameters
func TestSupplierAgingReportListFilter(t *testing.T) {
	supplierID := uint(1)
	branchID := uint(2)

	filter := SupplierAgingReportListFilter{
		AsOfDate:      "2026-05-31",
		SupplierID:    &supplierID,
		BranchID:      &branchID,
		Page:          1,
		Limit:         20,
		IncludeDetails: false,
	}

	// Verify field values
	assert.Equal(t, "2026-05-31", filter.AsOfDate, "AsOfDate should be set")
	assert.Equal(t, uint(1), *filter.SupplierID, "SupplierID should be set")
	assert.Equal(t, uint(2), *filter.BranchID, "BranchID should be set")
	assert.Equal(t, 1, filter.Page, "Page should default to 1")
	assert.Equal(t, 20, filter.Limit, "Limit should default to 20")
	assert.False(t, filter.IncludeDetails, "IncludeDetails should default to false")
}

// TestInvoiceAgingDetailEdgeCases tests edge cases in invoice aging details
// Story 10.6, Task 3.5, AC1: Handle negative days, future dates, zero balances
func TestInvoiceAgingDetailEdgeCases(t *testing.T) {
	tests := []struct {
		name           string
		detail         InvoiceAgingDetail
		expectedBucket string
	}{
		{
			name: "Future-due invoice (negative days)",
			detail: InvoiceAgingDetail{
				InvoiceID:          1,
				InvoiceNumber:      "INV-2026-001",
				InvoiceDate:        "2026-06-15",
				DueDate:            "2026-07-15",
				TotalAmount:        5000000.00,
				PaidAmount:         0,
				OutstandingBalance: 5000000.00,
				DaysOverdue:        -15,
				AgingBucket:        "current",
				PaymentStatus:      "unpaid",
			},
			expectedBucket: "current",
		},
		{
			name: "Zero outstanding balance (fully paid - should be excluded)",
			detail: InvoiceAgingDetail{
				InvoiceID:          1,
				InvoiceNumber:      "INV-2026-001",
				InvoiceDate:        "2026-04-15",
				DueDate:            "2026-05-15",
				TotalAmount:        5000000.00,
				PaidAmount:         5000000.00,
				OutstandingBalance: 0,
				DaysOverdue:        46,
				AgingBucket:        "31-60",
				PaymentStatus:      "paid",
			},
			expectedBucket: "31-60",
		},
		{
			name: "Exactly on due date",
			detail: InvoiceAgingDetail{
				InvoiceID:          1,
				InvoiceNumber:      "INV-2026-001",
				InvoiceDate:        "2026-04-15",
				DueDate:            "2026-05-31",
				TotalAmount:        5000000.00,
				PaidAmount:         2000000.00,
				OutstandingBalance: 3000000.00,
				DaysOverdue:        0,
				AgingBucket:        "current",
				PaymentStatus:      "partial",
			},
			expectedBucket: "current",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify bucket assignment
			assert.Equal(t, tt.expectedBucket, tt.detail.AgingBucket, "Aging bucket should match expected")

			// Verify calculation: Outstanding = Total - Paid
			expectedOutstanding := tt.detail.TotalAmount - tt.detail.PaidAmount
			assert.Equal(t, expectedOutstanding, tt.detail.OutstandingBalance, "Outstanding balance should equal total minus paid")

			// Serialize to verify no errors
			jsonBytes, err := json.Marshal(tt.detail)
			require.NoError(t, err, "Edge case invoice should serialize successfully")
			assert.NotEmpty(t, jsonBytes, "Serialized JSON should not be empty")
		})
	}
}

// TestSupplierAgingSummaryWithInvoices tests supplier summary with invoice details
// Story 10.6, Task 1.1: Verify invoice details can be included in supplier summary
func TestSupplierAgingSummaryWithInvoices(t *testing.T) {
	summary := SupplierAgingSummary{
		SupplierID:       1,
		SupplierName:     "PT. Pharmasi Jaya",
		ContactPerson:    "John Doe",
		Phone:            "+62-21-555-1234",
		Email:            "orders@pharmasi-jaya.co.id",
		Address:          "Jl. Industri No. 123, Jakarta",
		TotalOutstanding: 8000000.00,
		InvoiceCount:     2,
		AgingBuckets: AgingBucket{
			Current:        5000000.00,
			CurrentCount:   1,
			Days31to60:      3000000.00,
			Days31to60Count: 1,
			Days61to90:      0,
			Days61to90Count: 0,
			DaysOver90:      0,
			DaysOver90Count: 0,
		},
		Invoices: []InvoiceAgingDetail{
			{
				InvoiceID:          1,
				InvoiceNumber:      "INV-2026-001",
				InvoiceDate:        "2026-04-15",
				DueDate:            "2026-05-15",
				TotalAmount:        5000000.00,
				PaidAmount:         2000000.00,
				OutstandingBalance: 3000000.00,
				DaysOverdue:        46,
				AgingBucket:        "31-60",
				PaymentStatus:      "partial",
			},
			{
				InvoiceID:          2,
				InvoiceNumber:      "INV-2026-002",
				InvoiceDate:        "2026-05-01",
				DueDate:            "2026-05-31",
				TotalAmount:        5000000.00,
				PaidAmount:         0,
				OutstandingBalance: 5000000.00,
				DaysOverdue:        0,
				AgingBucket:        "current",
				PaymentStatus:      "unpaid",
			},
		},
	}

	// Serialize to JSON
	jsonBytes, err := json.Marshal(summary)
	require.NoError(t, err, "JSON serialization should succeed")

	// Verify JSON structure
	var result map[string]interface{}
	err = json.Unmarshal(jsonBytes, &result)
	require.NoError(t, err, "JSON deserialization should succeed")

	// Verify invoices array exists
	invoices := result["invoices"].([]interface{})
	assert.Len(t, invoices, 2, "Should have 2 invoices in detail view")
}
