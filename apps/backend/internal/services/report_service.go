package services

import (
	"context"
	"time"
)

// ReportService defines the interface for report business operations
// AC1: Service interface for report domain with clear business method signatures
type ReportService interface {
	// GenerateDailySales generates daily sales summary report
	// Filters by date range and branch
	GenerateDailySales(ctx context.Context, branchID uint, startDate, endDate time.Time) (*SalesReport, error)

	// GenerateProfitLoss generates profit and loss report
	// Calculates: Revenue - Cost of Goods Sold
	GenerateProfitLoss(ctx context.Context, branchID uint, startDate, endDate time.Time) (*ProfitLossReport, error)

	// ExportReport exports report in various formats
	// Stub for future story
	ExportReport(ctx context.Context, reportType string, format string) ([]byte, error)
}

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
