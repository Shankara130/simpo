package services

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/dto"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/models"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/repositories"
)

// supplierAgingReportServiceImpl implements SupplierAgingReportService interface
// Story 10.6: Service layer with aging calculation logic and report generation
type supplierAgingReportServiceImpl struct {
	purchaseInvoiceRepo repositories.PurchaseInvoiceRepository
	supplierPaymentRepo  repositories.SupplierPaymentRepository
	supplierRepo         repositories.SupplierRepository
	auditService         AuditService
}

// NewSupplierAgingReportService creates a new supplier aging report service
// Story 10.6: Factory function with dependency injection
func NewSupplierAgingReportService(
	purchaseInvoiceRepo repositories.PurchaseInvoiceRepository,
	supplierPaymentRepo repositories.SupplierPaymentRepository,
	supplierRepo repositories.SupplierRepository,
	auditService AuditService,
) SupplierAgingReportService {
	return &supplierAgingReportServiceImpl{
		purchaseInvoiceRepo: purchaseInvoiceRepo,
		supplierPaymentRepo:  supplierPaymentRepo,
		supplierRepo:         supplierRepo,
		auditService:         auditService,
	}
}

// GenerateAgingReport generates a comprehensive supplier aging report
// Story 10.6, AC1: Calculates aging buckets and outstanding balances
func (s *supplierAgingReportServiceImpl) GenerateAgingReport(ctx context.Context, request *dto.SupplierAgingReportRequest) (*dto.SupplierAgingReportResponse, error) {
	// Validate request
	if request == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}
	if request.AsOfDate == "" {
		return nil, fmt.Errorf("asOfDate is required")
	}

	// Parse as-of date
	asOfDate, err := dto.ParseAgingDate(request.AsOfDate)
	if err != nil {
		return nil, fmt.Errorf("invalid asOfDate format: %w", err)
	}

	// Log report generation
	slog.InfoContext(ctx, "Generating supplier aging report",
		"asOfDate", request.AsOfDate,
		"supplierId", request.SupplierID,
		"branchId", request.BranchID,
	)

	// Query unpaid and partially paid invoices with outstanding balances
	invoices, err := s.getOutstandingInvoices(ctx, request.SupplierID, request.BranchID)
	if err != nil {
		return nil, fmt.Errorf("failed to query outstanding invoices: %w", err)
	}

	// Build supplier aging summaries
	suppliers, err := s.buildSupplierSummaries(ctx, invoices, asOfDate)
	if err != nil {
		return nil, fmt.Errorf("failed to build supplier summaries: %w", err)
	}

	// Calculate grand totals
	grandTotals := s.calculateGrandTotals(suppliers)

	// Build response
	response := &dto.SupplierAgingReportResponse{
		AsOfDate:         request.AsOfDate,
		ReportGeneratedAt: time.Now().UTC().Format(time.RFC3339),
		Currency:         "IDR",
		Suppliers:        suppliers,
		GrandTotals:      grandTotals,
	}

	// Log completion
	slog.InfoContext(ctx, "Supplier aging report generated successfully",
		"asOfDate", request.AsOfDate,
		"supplierCount", len(suppliers),
		"totalOutstanding", grandTotals.TotalOutstanding,
	)

	return response, nil
}

// getOutstandingInvoices retrieves unpaid and partially paid invoices with payment details
// Story 10.6, Task 2.5: Query invoices with payment history for aging calculations
func (s *supplierAgingReportServiceImpl) getOutstandingInvoices(ctx context.Context, supplierID, branchID *uint) ([]models.PurchaseInvoice, error) {
	// Query invoices with payment_status = 'unpaid' OR 'partial'
	// This should filter out fully paid invoices
	unpaidStatus := "unpaid"
	invoices, _, err := s.purchaseInvoiceRepo.List(ctx, &repositories.PurchaseInvoiceFilter{
		SupplierID:    supplierID,
		StartDate:     nil, // No date filter for aging reports
		EndDate:       nil,
		PaymentStatus:  &unpaidStatus,
		Page:           1,
		Limit:          1000, // Get all for report
		SortBy:         "invoice_date",
		SortOrder:      "asc",
	})
	if err != nil {
		return nil, err
	}

	// Also query partial payment invoices
	partialStatus := "partial"
	partialInvoices, _, err := s.purchaseInvoiceRepo.List(ctx, &repositories.PurchaseInvoiceFilter{
		SupplierID:    supplierID,
		StartDate:     nil,
		EndDate:       nil,
		PaymentStatus:  &partialStatus,
		Page:           1,
		Limit:          1000,
		SortBy:         "invoice_date",
		SortOrder:      "asc",
	})
	if err != nil {
		return nil, err
	}

	// Combine results
	var result []models.PurchaseInvoice
	for _, inv := range invoices {
		result = append(result, *inv)
	}
	for _, inv := range partialInvoices {
		result = append(result, *inv)
	}
	return result, nil
}

// buildSupplierSummaries creates aging summaries grouped by supplier
// Story 10.6, Task 2.4: Calculate aging buckets and outstanding amounts per supplier
func (s *supplierAgingReportServiceImpl) buildSupplierSummaries(ctx context.Context, invoices []models.PurchaseInvoice, asOfDate time.Time) ([]dto.SupplierAgingSummary, error) {
	// Group invoices by supplier
	supplierMap := make(map[uint][]models.PurchaseInvoice)
	for _, invoice := range invoices {
		supplierMap[invoice.SupplierID] = append(supplierMap[invoice.SupplierID], invoice)
	}

	// Build summaries for each supplier
	var summaries []dto.SupplierAgingSummary
	for supplierID, supplierInvoices := range supplierMap {
		// Load supplier details
		supplier, err := s.supplierRepo.GetByID(ctx, supplierID)
		if err != nil {
			slog.WarnContext(ctx, "Failed to load supplier details",
				"supplierId", supplierID,
				"error", err,
			)
			continue
		}

		// Calculate aging buckets for this supplier
		agingBuckets, invoices := s.calculateAgingBuckets(ctx, supplierInvoices, asOfDate)
		totalOutstanding := s.calculateTotalOutstanding(agingBuckets)

		summary := dto.SupplierAgingSummary{
			SupplierID:       supplier.ID,
			SupplierName:     supplier.Name,
			ContactPerson:    supplier.ContactPerson,
			Phone:            supplier.Phone,
			Email:            supplier.Email,
			Address:          supplier.Address,
			AgingBuckets:     agingBuckets,
			TotalOutstanding: totalOutstanding,
			InvoiceCount:     len(invoices),
		}

		summaries = append(summaries, summary)
	}

	return summaries, nil
}

// calculateAgingBuckets computes aging bucket breakdown for a set of invoices
// Story 10.6, Task 2.4 & Task 3: Categorize invoices into aging buckets with counts
func (s *supplierAgingReportServiceImpl) calculateAgingBuckets(ctx context.Context, invoices []models.PurchaseInvoice, asOfDate time.Time) (dto.AgingBucket, []dto.InvoiceAgingDetail) {
	bucket := dto.AgingBucket{}
	var invoiceDetails []dto.InvoiceAgingDetail

	for _, invoice := range invoices {
		// Calculate outstanding balance: totalAmount - SUM(paymentAmount)
		outstandingBalance := invoice.TotalAmount - s.getPaidAmount(ctx, invoice)

		// Skip fully paid invoices (shouldn't happen due to filtering, but safety check)
		if outstandingBalance <= 0 {
			continue
		}

		// Calculate days overdue: asOfDate - invoiceDueDate
		// Default due date is invoice date + 30 days (standard payment terms)
		invoiceDueDate := invoice.InvoiceDate.AddDate(0, 0, 30)
		daysOverdue := dto.DaysBetween(invoiceDueDate, asOfDate)

		// Categorize into aging bucket
		agingCategory := dto.CategorizeIntoBucket(daysOverdue)

		// Add to appropriate bucket
		switch agingCategory {
		case "current":
			bucket.Current += outstandingBalance
			bucket.CurrentCount++
		case "31-60":
			bucket.Days31to60 += outstandingBalance
			bucket.Days31to60Count++
		case "61-90":
			bucket.Days61to90 += outstandingBalance
			bucket.Days61to90Count++
		case "90+":
			bucket.DaysOver90 += outstandingBalance
			bucket.DaysOver90Count++
		}

		// Create invoice detail
		invoiceDetail := dto.InvoiceAgingDetail{
			InvoiceID:          invoice.ID,
			InvoiceNumber:      invoice.InvoiceNumber,
			InvoiceDate:        invoice.InvoiceDate.Format("2006-01-02"),
			DueDate:            invoiceDueDate.Format("2006-01-02"),
			TotalAmount:        invoice.TotalAmount,
			PaidAmount:         s.getPaidAmount(ctx, invoice),
			OutstandingBalance: outstandingBalance,
			DaysOverdue:        daysOverdue,
			AgingBucket:        agingCategory,
			PaymentStatus:      invoice.PaymentStatus,
		}
		invoiceDetails = append(invoiceDetails, invoiceDetail)
	}

	return bucket, invoiceDetails
}

// getPaidAmount calculates the sum of all payments for an invoice
// Story 10.6, Task 3.1: Calculate outstanding balance by summing payments
func (s *supplierAgingReportServiceImpl) getPaidAmount(ctx context.Context, invoice models.PurchaseInvoice) float64 {
	// Use supplier payment repository to get total paid amount
	paidAmount, err := s.supplierPaymentRepo.GetTotalPaidByInvoice(ctx, invoice.ID)
	if err != nil {
		// If error, assume no payments (default to 0)
		return 0
	}
	return paidAmount
}

// calculateTotalOutstanding sums all aging buckets for total outstanding
// Story 10.6, Task 3.6: Calculate total outstanding per supplier
func (s *supplierAgingReportServiceImpl) calculateTotalOutstanding(bucket dto.AgingBucket) float64 {
	return bucket.Current + bucket.Days31to60 + bucket.Days61to90 + bucket.DaysOver90
}

// calculateGrandTotals aggregates totals across all suppliers
// Story 10.6, Task 3.7: Calculate grand totals for the entire report
func (s *supplierAgingReportServiceImpl) calculateGrandTotals(summaries []dto.SupplierAgingSummary) dto.AgingGrandTotals {
	totals := dto.AgingGrandTotals{}

	for _, summary := range summaries {
		totals.Current += summary.AgingBuckets.Current
		totals.Days31to60 += summary.AgingBuckets.Days31to60
		totals.Days61to90 += summary.AgingBuckets.Days61to90
		totals.DaysOver90 += summary.AgingBuckets.DaysOver90
		totals.TotalOutstanding += summary.TotalOutstanding
		totals.TotalInvoices += summary.InvoiceCount
	}
	totals.TotalSuppliers = len(summaries)

	return totals
}

// ExportAgingReportPDF generates a PDF export of the aging report
// Story 10.6, Task 6: PDF generation with professional layout
func (s *supplierAgingReportServiceImpl) ExportAgingReportPDF(ctx context.Context, request *dto.SupplierAgingReportRequest) ([]byte, string, error) {
	// Generate the report first
	_, err := s.GenerateAgingReport(ctx, request)
	if err != nil {
		return nil, "", err
	}

	// TODO: Implement PDF generation using gofpdf library
	// Story 10.6, Task 6.1-6.6: PDF generation with formatting

	// Placeholder: Create simple text-based PDF (will be implemented in Task 6)
	slog.InfoContext(ctx, "PDF export not yet implemented - returning placeholder",
		"asOfDate", request.AsOfDate,
	)

	filename := fmt.Sprintf("supplier-aging-report-%s.pdf", request.AsOfDate)
	return []byte("PDF placeholder"), filename, nil
}

// ExportAgingReportExcel generates an Excel export of the aging report
// Story 10.6, Task 7: Excel generation with multiple sheets
func (s *supplierAgingReportServiceImpl) ExportAgingReportExcel(ctx context.Context, request *dto.SupplierAgingReportRequest) ([]byte, string, error) {
	// Generate the report first
	_, err := s.GenerateAgingReport(ctx, request)
	if err != nil {
		return nil, "", err
	}

	// TODO: Implement Excel generation using excelize library
	// Story 10.6, Task 7.1-7.6: Excel generation with multiple sheets

	// Placeholder: Create simple CSV-based Excel (will be implemented in Task 7)
	slog.InfoContext(ctx, "Excel export not yet implemented - returning placeholder",
		"asOfDate", request.AsOfDate,
	)

	filename := fmt.Sprintf("supplier-aging-report-%s.xlsx", request.AsOfDate)
	return []byte("Excel placeholder"), filename, nil
}