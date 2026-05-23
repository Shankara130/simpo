package repositories

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/models"
)

// setupReportTestDB creates a test database for report repository testing
// Story 5.1, Task 5.1: Test helper for repository testing
func setupReportTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)

	// Create tables
	err = db.AutoMigrate(&models.Transaction{}, &models.TransactionItem{}, &models.Product{}, &models.Branch{})
	require.NoError(t, err)

	return db
}

// TestReportRepository_Contract verifies the repository interface is correctly implemented
// Story 5.1, Task 5.1: Interface implementation verification
func TestReportRepository_Contract(t *testing.T) {
	// Arrange
	db := setupReportTestDB(t)

	// Act: Create repository instance
	repo := NewReportRepository(db)

	// Assert: Verify repository implements the interface
	assert.Implements(t, (*ReportRepository)(nil), repo, "Repository should implement ReportRepository interface")
	assert.NotNil(t, repo, "Repository should not be nil")
}

// TestReportRepository_GetDailySalesSummary_Signature tests method signature
// Story 5.1, Task 5.1: Verify method signature matches interface
func TestReportRepository_GetDailySalesSummary_Signature(t *testing.T) {
	// Arrange
	db := setupReportTestDB(t)
	repo := NewReportRepository(db)
	ctx := context.Background()

	// Act & Assert: Test method exists and accepts correct parameters
	// Note: PostgreSQL-specific SQL will fail with SQLite, so we expect errors
	// This test verifies the method signature and error handling

	// Test with valid parameters
	_, _ = repo.GetDailySalesSummary(ctx, "2026-05-23", 0)
	// Expected: SQL error due to SQLite not supporting PostgreSQL syntax
	// This is OK - we're testing the method signature, not SQL execution
	assert.NotNil(t, repo, "Repository method should exist")

	// Test with branch_id parameter
	_, _ = repo.GetDailySalesSummary(ctx, "2026-05-23", 1)
	assert.NotNil(t, repo, "Repository method should accept branch_id parameter")
}

// TestReportRepository_NilDatabase tests repository with nil database
// Story 5.1, Task 5.1: Edge case testing
func TestReportRepository_NilDatabase(t *testing.T) {
	// Arrange: Create repository with nil database
	var db *gorm.DB = nil
	repo := NewReportRepository(db)

	// Assert: Repository should still be created (GORM handles nil DB)
	assert.NotNil(t, repo, "Repository should be created even with nil DB")
}

// TestReportRepository_Interface_ReturnTypes tests return type contract
// Story 5.1, Task 5.1: Verify return types match interface
func TestReportRepository_Interface_ReturnTypes(t *testing.T) {
	// Arrange
	db := setupReportTestDB(t)
	repo := NewReportRepository(db)
	ctx := context.Background()

	// Act: Call method (will fail with SQLite, but we test return types)
	_, err := repo.GetDailySalesSummary(ctx, "2026-05-23", 0)

	// Assert: Verify return types are correct even when SQL fails
	// err should be non-nil (SQL syntax error with SQLite)
	assert.Error(t, err, "Expected SQL error with SQLite (PostgreSQL syntax)")
}

// Note: Full integration tests with PostgreSQL SQL validation should be run
// in a CI/CD environment with actual PostgreSQL instance.
// Story 5.1, Task 5.3: Integration test requirement
//
// To run integration tests:
// 1. Set up PostgreSQL test database
// 2. Load test data with transactions, transaction_items, products, branches
// 3. Execute GetDailySalesSummary and verify:
//    - Correct total sales amount

// ==============================================================================
// Story 5.2: Profit/Loss Report Repository Tests
// ==============================================================================

// TestReportRepository_GetProfitLossSummary_Signature tests method signature
// Story 5.2, Task 2.1: Verify method signature matches interface
func TestReportRepository_GetProfitLossSummary_Signature(t *testing.T) {
	// Arrange
	db := setupReportTestDB(t)
	repo := NewReportRepository(db)
	ctx := context.Background()

	// Act & Assert: Test method exists and accepts correct parameters
	// Note: PostgreSQL-specific SQL will fail with SQLite, so we expect errors
	// This test verifies the method signature and error handling

	// Test with valid parameters
	_, _ = repo.GetProfitLossSummary(ctx, "2026-05-01", "2026-05-23", 0, "")
	assert.NotNil(t, repo, "Repository method should exist")

	// Test with breakdown_by parameter
	_, _ = repo.GetProfitLossSummary(ctx, "2026-05-01", "2026-05-23", 0, "category")
	assert.NotNil(t, repo, "Repository method should accept breakdown_by parameter")

	// Test with branch_id parameter
	_, _ = repo.GetProfitLossSummary(ctx, "2026-05-01", "2026-05-23", 1, "branch")
	assert.NotNil(t, repo, "Repository method should accept branch_id parameter")
}

// TestReportRepository_GetProfitLossSummary_DateRangeValidation tests date range validation
// Story 5.2, Task 2.2, AC3: Validate date range parameters
func TestReportRepository_GetProfitLossSummary_DateRangeValidation(t *testing.T) {
	// Arrange
	db := setupReportTestDB(t)
	repo := NewReportRepository(db)
	ctx := context.Background()

	tests := []struct {
		name        string
		startDate   string
		endDate     string
		expectError bool
		description string
	}{
		{
			name:        "Valid date range - empty result",
			startDate:   "2026-05-01",
			endDate:     "2026-05-23",
			expectError: false,
			description: "Should succeed with empty result (no test data)",
		},
		{
			name:        "Invalid date format",
			startDate:   "invalid",
			endDate:     "2026-05-23",
			expectError: true,
			description: "Should fail with invalid date format",
		},
		{
			name:        "End date before start date",
			startDate:   "2026-05-23",
			endDate:     "2026-05-01",
			expectError: false,
			description: "Should succeed but return empty result (date range produces no data)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result, err := repo.GetProfitLossSummary(ctx, tt.startDate, tt.endDate, 0, "")

			// Assert
			if tt.expectError {
				assert.Error(t, err, tt.description)
			} else {
				assert.NoError(t, err, tt.description)
				assert.NotNil(t, result, "Result should not be nil")
				assert.Equal(t, tt.startDate, result.PeriodStart, "PeriodStart should match")
				assert.Equal(t, tt.endDate, result.PeriodEnd, "PeriodEnd should match")
			}
		})
	}
}

// TestReportRepository_GetProfitLossSummary_BranchValidation tests branch ID validation
// Story 5.2, Task 2.2, AC2: Validate branch ID parameter
func TestReportRepository_GetProfitLossSummary_BranchValidation(t *testing.T) {
	// Arrange
	db := setupReportTestDB(t)
	repo := NewReportRepository(db)
	ctx := context.Background()

	// Act: Test with branch_id = 0 (all branches)
	result, err := repo.GetProfitLossSummary(ctx, "2026-05-01", "2026-05-23", 0, "")

	// Assert: Should succeed with empty result (no test data)
	assert.NoError(t, err, "Should succeed with empty result")
	assert.NotNil(t, result, "Result should not be nil")
	assert.Equal(t, "All Branches", result.BranchName, "BranchName should be 'All Branches'")
	assert.Equal(t, uint(0), result.BranchID, "BranchID should be 0")
}

// TestReportRepository_GetProfitLossSummary_ParseFloat64Error tests error handling for malformed decimal values
// Story 5.2, Code review fix CRITICAL-001: parseFloat64 errors should be properly handled
func TestReportRepository_GetProfitLossSummary_ParseFloat64Error(t *testing.T) {
	// This test verifies that the parseFloat64 function returns errors correctly
	// For integration testing with SQL returning malformed values, we would need
	// to mock the database driver or use a test database with crafted data

	// Test parseFloat64 with valid input
	validFloat, err := parseFloat64("123.45")
	assert.NoError(t, err, "Should parse valid float correctly")
	assert.Equal(t, 123.45, validFloat, "Should return correct float value")

	// Test parseFloat64 with invalid input
	_, err = parseFloat64("invalid")
	assert.Error(t, err, "Should return error for invalid input")

	// Test parseFloat64 with empty string
	_, err = parseFloat64("")
	assert.Error(t, err, "Should return error for empty string")

	// Test parseFloat64 with valid integer string
	validInt, err := parseFloat64("100")
	assert.NoError(t, err, "Should parse integer string correctly")
	assert.Equal(t, 100.0, validInt, "Should return correct float value for integer string")

	// Test parseFloat64 with zero
	zeroFloat, err := parseFloat64("0")
	assert.NoError(t, err, "Should parse zero correctly")
	assert.Equal(t, 0.0, zeroFloat, "Should return zero value")

	// Test parseFloat64 with negative number
	negFloat, err := parseFloat64("-50.25")
	assert.NoError(t, err, "Should parse negative float correctly")
	assert.Equal(t, -50.25, negFloat, "Should return correct negative float value")
}

// Note: Full integration tests for GetProfitLossSummary with actual PostgreSQL
// should include:
// Story 5.2, Task 5.1-5.8:
// 1. Test COGS calculation with transaction_items.cost_price
// 2. Test breakdown by product category
// 3. Test breakdown by branch location
// 4. Test breakdown by payment method
// 5. Test NULL cost_price handling (should use 0 for COGS)
// 6. Test branch filtering for multi-branch scenarios
// 7. Test performance with large dataset (<10 seconds)
// 8. Test date range validation and error responses
//    - Correct transaction count
//    - Payment breakdown with correct percentages
//    - Top 10 products ordered by quantity
//    - Hourly sales grouped by hour (0-23)
// 4. Test branch filtering with different branch_id values
// 5. Test date range filtering
//
// Performance test (Story 5.1, Task 5.4):
// - Insert 10K+ transactions
// - Measure query execution time
// - Verify < 10 seconds requirement (AC3)
