package services

import (
	"context"
	"time"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/dto"
)

// ReportService defines the interface for report business operations
// Story 5.1, Task 3.1: Service interface for financial reporting logic
type ReportService interface {
	// GenerateDailySalesSummary generates comprehensive daily sales summary report
	// Story 5.1, Task 3.1, AC1: Returns total sales, payment breakdown, top products, hourly sales
	// Task 3.3: RBAC validation (Owner role required)
	// Task 3.4: Branch filtering based on user role
	// Task 3.5: Caching with Redis for 5-minute TTL
	// Task 3.6: Performance requirement <10 seconds with context timeout
	GenerateDailySalesSummary(ctx context.Context, req *dto.DailySalesRequest) (*dto.DailySalesSummaryDTO, error)

	// GenerateDailySales generates daily sales summary report (legacy method for backwards compatibility)
	// Filters by date range and branch
	// Deprecated: Use GenerateDailySalesSummary with DTO instead
	GenerateDailySales(ctx context.Context, branchID uint, startDate, endDate time.Time) (*SalesReport, error)

	// GenerateProfitLossSummary generates comprehensive profit/loss summary report
	// Story 5.2, Task 3.1-3.7, AC1, AC2, AC3: Returns revenue, COGS, gross profit with breakdowns
	// Task 3.3: RBAC validation (Owner role required)
	// Task 3.4: Branch filtering based on user role
	// Task 3.5: Caching with Redis for 5-minute TTL
	// Task 3.6: Performance requirement <10 seconds with context timeout
	// Task 3.7: Calculate gross profit margin percentage
	GenerateProfitLossSummary(ctx context.Context, req *dto.ProfitLossRequest) (*dto.ProfitLossSummaryDTO, error)

	// GenerateProfitLoss generates profit and loss report
	// Calculates: Revenue - Cost of Goods Sold
	GenerateProfitLoss(ctx context.Context, branchID uint, startDate, endDate time.Time) (*ProfitLossReport, error)

	// ExportReport exports report in various formats
	// Stub for future story
	ExportReport(ctx context.Context, reportType string, format string) ([]byte, error)
}

// Constants for report-related errors
const (
	// ErrUnauthorizedReportAccess is returned when user lacks permission for reports
	// Story 5.1, Security Requirements: Only Owner and Admin can access financial reports
	ErrUnauthorizedReportAccess = "unauthorized: insufficient permissions to access financial reports"

	// ErrInvalidDateFormat is returned when date format is invalid
	// Story 5.1, AC1: Date must be in YYYY-MM-DD format
	ErrInvalidDateFormat = "invalid date format: use YYYY-MM-DD"

	// ErrReportGenerationTimeout is returned when report generation exceeds timeout
	// Story 5.1, Task 3.6: Performance requirement <10 seconds
	ErrReportGenerationTimeout = "report generation timeout: exceeded 10 seconds"
)

// SalesReport represents daily sales summary report
type SalesReport struct {
	BranchID        uint
	BranchName      string
	StartDate       time.Time
	EndDate         time.Time
	TotalSales      string
	TotalTransactions int64
	AverageTransactionValue string
	PaymentMethods  []PaymentMethodBreakdown
}

// ProfitLossReport represents profit and loss report
type ProfitLossReport struct {
	BranchID        uint
	BranchName      string
	StartDate       time.Time
	EndDate         time.Time
	Revenue         string
	CostOfGoodsSold string
	GrossProfit     string
	GrossMargin     float64
	Expenses        string
	NetProfit       string
	NetMargin       float64
}

// PaymentMethodBreakdown represents sales breakdown by payment method
type PaymentMethodBreakdown struct {
	PaymentMethod string
	Count         int64
	TotalAmount   string
	Percentage    float64
}
