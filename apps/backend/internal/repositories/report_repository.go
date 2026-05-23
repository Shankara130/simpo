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
}
