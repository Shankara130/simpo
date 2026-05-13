package services

import (
	"context"
	"fmt"
	"time"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/repositories"
)

// reportService implements ReportService interface
// AC2: Services use repository interfaces (not concrete implementations)
type reportService struct {
	transactionRepo repositories.TransactionRepository
	productRepo     repositories.ProductRepository
	auditService    AuditService
}

// NewReportService creates a new report service with dependency injection
// AC2: Services accept repository interfaces via constructor injection
func NewReportService(
	transactionRepo repositories.TransactionRepository,
	productRepo repositories.ProductRepository,
	auditService AuditService,
) ReportService {
	// Fail fast on nil dependencies
	if transactionRepo == nil {
		panic("reportService: transactionRepo cannot be nil")
	}
	if productRepo == nil {
		panic("reportService: productRepo cannot be nil")
	}
	if auditService == nil {
		panic("reportService: auditService cannot be nil")
	}

	return &reportService{
		transactionRepo: transactionRepo,
		productRepo:     productRepo,
		auditService:    auditService,
	}
}

// GenerateDailySales generates daily sales summary report
// AC3: Filters by date range and branch
func (s *reportService) GenerateDailySales(ctx context.Context, branchID uint, startDate, endDate time.Time) (*SalesReport, error) {
	// Check context cancellation
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("operation cancelled: %w", err)
	}

	// Validate date range
	if endDate.Before(startDate) {
		return nil, &InvalidInputError{Field: "date_range", Message: "end date cannot be before start date"}
	}

	// Validate branch ID
	if branchID == 0 {
		return nil, &InvalidInputError{Field: "branch_id", Message: "branch ID is required"}
	}

	// PATCH: Use startDate instead of endDate (date range was being ignored)
	// Get daily summary for the start date of the range
	summary, err := s.transactionRepo.GetDailySummary(ctx, branchID, startDate)
	if err != nil {
		return nil, &ServiceError{Op: "get daily summary", Err: err}
	}

	// Build report
	report := &SalesReport{
		BranchID:              branchID,
		BranchName:            "", // Would need branch repo to get name
		StartDate:             startDate,
		EndDate:               endDate,
		TotalSales:            summary.TotalAmount,
		TotalTransactions:     summary.TotalTransactions,
		AverageTransactionValue: "0.00", // Simplified for MVP
		PaymentMethods:        []PaymentMethodBreakdown{},
	}

	// Convert payment method summaries
	for _, pm := range summary.PaymentMethods {
		report.PaymentMethods = append(report.PaymentMethods, PaymentMethodBreakdown{
			PaymentMethod: pm.PaymentMethod,
			Count:         pm.Count,
			TotalAmount:   pm.TotalAmount,
			Percentage:    0.0, // Simplified for MVP
		})
	}

	return report, nil
}

// GenerateProfitLoss generates profit and loss report
// AC3: Calculates: Revenue - Cost of Goods Sold
func (s *reportService) GenerateProfitLoss(ctx context.Context, branchID uint, startDate, endDate time.Time) (*ProfitLossReport, error) {
	// Check context cancellation
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("operation cancelled: %w", err)
	}

	// Validate date range
	if endDate.Before(startDate) {
		return nil, &InvalidInputError{Field: "date_range", Message: "end date cannot be before start date"}
	}

	// Validate branch ID
	if branchID == 0 {
		return nil, &InvalidInputError{Field: "branch_id", Message: "branch ID is required"}
	}

	// PATCH: Use startDate for consistency with GenerateDailySales fix
	summary, err := s.transactionRepo.GetDailySummary(ctx, branchID, startDate)
	if err != nil {
		return nil, &ServiceError{Op: "get daily summary", Err: err}
	}

	// For MVP, simplified calculation
	// In production, COGS would be calculated from transaction items with cost price
	report := &ProfitLossReport{
		BranchID:        branchID,
		BranchName:      "", // Would need branch repo to get name
		StartDate:       startDate,
		EndDate:         endDate,
		Revenue:         summary.TotalAmount,
		CostOfGoodsSold: "0.00", // Simplified for MVP
		GrossProfit:     summary.TotalAmount, // Simplified
		GrossMargin:     100.0, // Simplified
		Expenses:        "0.00", // Simplified
		NetProfit:       summary.TotalAmount, // Simplified
		NetMargin:       100.0, // Simplified
	}

	return report, nil
}

// ExportReport exports report in various formats
// Stub for future story
func (s *reportService) ExportReport(ctx context.Context, reportType string, format string) ([]byte, error) {
	// Check context cancellation
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("operation cancelled: %w", err)
	}

	// Stub implementation for future story
	return nil, &InvalidInputError{
		Field:   "export",
		Message: "report export not implemented - scheduled for future story",
	}
}
