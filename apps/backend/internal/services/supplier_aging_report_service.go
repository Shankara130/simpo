package services

import (
	"context"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/dto"
)

// SupplierAgingReportService defines the interface for supplier aging report business logic
// Story 10.6: Service interface for generating supplier aging reports with bucket breakdown
type SupplierAgingReportService interface {
	// GenerateAgingReport generates a comprehensive supplier aging report
	// Story 10.6, AC1: Calculates aging buckets, outstanding balances, and supplier summaries
	// Parameters:
	//   - ctx: Request context for cancellation and timeout
	//   - request: Aging report parameters (asOfDate, supplierID, branchID, pagination)
	// Returns:
	//   - response: Complete aging report with supplier summaries and grand totals
	//   - error: Error if report generation fails (invalid date, database error, etc.)
	GenerateAgingReport(ctx context.Context, request *dto.SupplierAgingReportRequest) (*dto.SupplierAgingReportResponse, error)

	// ExportAgingReportPDF generates a PDF export of the aging report
	// Story 10.6, AC1: Creates formatted PDF with supplier summaries and invoice details
	// Parameters:
	//   - ctx: Request context for cancellation and timeout
	//   - request: Aging report parameters (asOfDate, supplierID, branchID)
	// Returns:
	//   - pdfBytes: PDF file contents
	//   - filename: Suggested filename for download (e.g., "supplier-aging-report-2026-05-31.pdf")
	//   - error: Error if PDF generation fails
	ExportAgingReportPDF(ctx context.Context, request *dto.SupplierAgingReportRequest) (pdfBytes []byte, filename string, err error)

	// ExportAgingReportExcel generates an Excel export of the aging report
	// Story 10.6, AC1: Creates Excel workbook with supplier summary and invoice details sheets
	// Parameters:
	//   - ctx: Request context for cancellation and timeout
	//   - request: Aging report parameters (asOfDate, supplierID, branchID)
	// Returns:
	//   - excelBytes: Excel file contents
	//   - filename: Suggested filename for download (e.g., "supplier-aging-report-2026-05-31.xlsx")
	//   - error: Error if Excel generation fails
	ExportAgingReportExcel(ctx context.Context, request *dto.SupplierAgingReportRequest) (excelBytes []byte, filename string, err error)
}

// SupplierAgingReportFilter defines filtering options for aging report queries
// Story 10.6, AC1: Filter struct for database queries with pagination support
type SupplierAgingReportFilter struct {
	// AsOfDate is the snapshot date for aging calculations
	AsOfDate string

	// SupplierID filters by specific supplier (nil for all suppliers)
	SupplierID *uint

	// BranchID filters by branch (nil for all branches, subject to access control)
	BranchID *uint

	// Page is the page number for supplier list pagination (1-indexed)
	Page int

	// Limit is the number of suppliers per page
	Limit int

	// IncludeDetails controls whether to include invoice-level details
	IncludeDetails bool
}
