package repositories

import (
	"context"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/dto"
)

// ReportRepository defines the interface for report data operations
// Story 5.1, Task 2: Repository for complex SQL aggregation queries for reports
type ReportRepository interface {
	// GetDailySalesSummary retrieves comprehensive daily sales summary
	// Story 5.1, Task 2.1, AC1: Returns total sales, payment breakdown, top products, hourly sales
	// Task 2.2: SQL query for total sales and transaction count with date and optional branch filter
	// Task 2.3: SQL query for payment method breakdown with GROUP BY
	// Task 2.4: SQL query for top 10 products with JOIN transaction_items and transactions
	// Task 2.5: SQL query for hourly sales with EXTRACT(HOUR FROM created_at) GROUP BY
	//
	// Parameters:
	//   - ctx: Context for cancellation and timeout
	//   - date: Report date (time component will be ignored)
	//   - branchID: Branch filter (0 means all branches aggregated)
	//
	// Returns complete DailySalesSummaryDTO or error
	GetDailySalesSummary(ctx context.Context, date string, branchID uint) (*dto.DailySalesSummaryDTO, error)

	// GetProfitLossSummary retrieves comprehensive profit/loss summary
	// Story 5.2, Task 2.1-2.6, AC1, AC2, AC3: Returns revenue, COGS, gross profit with breakdowns
	// Task 2.2: SQL query for total revenue from transactions table
	// Task 2.3: SQL query for COGS using transaction_items.cost_price
	// Task 2.4: Query for breakdown by product category
	// Task 2.5: Query for breakdown by branch location
	// Task 2.6: Query for breakdown by payment method
	//
	// Parameters:
	//   - ctx: Context for cancellation and timeout
	//   - startDate: Report start date (YYYY-MM-DD format)
	//   - endDate: Report end date (YYYY-MM-DD format)
	//   - branchID: Branch filter (0 means all branches aggregated)
	//   - breakdownBy: Breakdown type (category, branch, payment_method, or empty)
	//
	// Returns complete ProfitLossSummaryDTO or error
	GetProfitLossSummary(ctx context.Context, startDate, endDate string, branchID uint, breakdownBy string) (*dto.ProfitLossSummaryDTO, error)
}
