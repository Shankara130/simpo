package dto

import "time"

// SupplierAgingReport DTOs for API request/response
// Story 10.6: Data transfer objects for supplier aging report endpoints

// SupplierAgingReportRequest represents the request payload for generating a supplier aging report
// Story 10.6, AC1: Validation tags for report generation parameters
type SupplierAgingReportRequest struct {
	// AsOfDate is the snapshot date for aging calculations (required, ISO 8601 format)
	// Used to calculate days overdue: asOfDate - invoiceDueDate
	// Example: "2026-05-31"
	AsOfDate string `json:"asOfDate" binding:"required" example:"2026-05-31"`

	// SupplierID filters by specific supplier (optional)
	// When provided, only show aging for this supplier
	// Example: 1
	SupplierID *uint `json:"supplierId,omitempty" example:"1"`

	// BranchID filters by branch (optional)
	// Required for branch access control validation
	// Example: 1
	BranchID *uint `json:"branchId,omitempty" example:"1"`
}

// SupplierAgingReportResponse represents the response payload for supplier aging report
// Story 10.6, AC1: Complete aging report with supplier summaries and totals
type SupplierAgingReportResponse struct {
	// AsOfDate is the snapshot date used for aging calculations
	// Example: "2026-05-31"
	AsOfDate string `json:"asOfDate" example:"2026-05-31"`

	// ReportGeneratedAt is the timestamp when the report was generated
	// Example: "2026-05-31T10:00:00Z"
	ReportGeneratedAt string `json:"reportGeneratedAt" example:"2026-05-31T10:00:00Z"`

	// Currency is the currency code for all amounts
	// Example: "IDR"
	Currency string `json:"currency" example:"IDR"`

	// Suppliers contains aging summaries for each supplier
	Suppliers []SupplierAgingSummary `json:"suppliers"`

	// GrandTotals contains aggregated totals across all suppliers
	GrandTotals AgingGrandTotals `json:"grandTotals"`

	// Pagination contains pagination metadata
	Pagination PaginationResponse `json:"pagination,omitempty"`
}

// SupplierAgingSummary represents aging information for a single supplier
// Story 10.6, AC1: Supplier-level aging breakdown with buckets
type SupplierAgingSummary struct {
	// SupplierID is the unique identifier for the supplier
	// Example: 1
	SupplierID uint `json:"supplierId" example:"1"`

	// SupplierName is the name of the supplier
	// Example: "PT. Pharmasi Jaya"
	SupplierName string `json:"supplierName,omitempty" example:"PT. Pharmasi Jaya"`

	// ContactPerson is the primary contact at the supplier
	// Example: "John Doe"
	ContactPerson string `json:"contactPerson,omitempty" example:"John Doe"`

	// Phone is the supplier's phone number
	// Example: "+62-21-555-1234"
	Phone string `json:"phone,omitempty" example:"+62-21-555-1234"`

	// Email is the supplier's email address
	// Example: "orders@pharmasi-jaya.co.id"
	Email string `json:"email,omitempty" example:"orders@pharmasi-jaya.co.id"`

	// Address is the supplier's address
	// Example: "Jl. Industri No. 123, Jakarta"
	Address string `json:"address,omitempty" example:"Jl. Industri No. 123, Jakarta"`

	// AgingBuckets contains the aging breakdown for this supplier
	AgingBuckets AgingBucket `json:"agingBuckets"`

	// TotalOutstanding is the sum of all aging buckets for this supplier
	// Example: 15000000.00
	TotalOutstanding float64 `json:"totalOutstanding" example:"15000000.00"`

	// InvoiceCount is the total number of outstanding invoices for this supplier
	// Example: 5
	InvoiceCount int `json:"invoiceCount" example:"5"`

	// Invoices contains detailed aging breakdown for each invoice
	// Example: null (omitted for summary view) or populated for detail view
	Invoices []InvoiceAgingDetail `json:"invoices,omitempty"`
}

// AgingBucket represents the standard aging period breakdown
// Story 10.6, AC1: Four aging buckets (0-30, 31-60, 61-90, 90+ days)
type AgingBucket struct {
	// Current represents invoices 0-30 days overdue (or not yet due)
	// Example: 5000000.00
	Current float64 `json:"current" example:"5000000.00"`

	// CurrentCount is the number of invoices in the current bucket
	// Example: 2
	CurrentCount int `json:"currentCount" example:"2"`

	// Days31to60 represents invoices 31-60 days overdue
	// Example: 3000000.00
	Days31to60 float64 `json:"days31to60" example:"3000000.00"`

	// Days31to60Count is the number of invoices 31-60 days overdue
	// Example: 1
	Days31to60Count int `json:"days31to60Count" example:"1"`

	// Days61to90 represents invoices 61-90 days overdue
	// Example: 4000000.00
	Days61to90 float64 `json:"days61to90" example:"4000000.00"`

	// Days61to90Count is the number of invoices 61-90 days overdue
	// Example: 1
	Days61to90Count int `json:"days61to90Count" example:"1"`

	// DaysOver90 represents invoices over 90 days overdue (critical)
	// Example: 3000000.00
	DaysOver90 float64 `json:"daysOver90" example:"3000000.00"`

	// DaysOver90Count is the number of invoices over 90 days overdue
	// Example: 1
	DaysOver90Count int `json:"daysOver90Count" example:"1"`
}

// InvoiceAgingDetail represents aging information for a single invoice
// Story 10.6, AC1: Individual invoice aging breakdown for detail views
type InvoiceAgingDetail struct {
	// InvoiceID is the unique identifier for the invoice
	// Example: 1
	InvoiceID uint `json:"invoiceId" example:"1"`

	// InvoiceNumber is the invoice number from the supplier
	// Example: "INV-2026-001"
	InvoiceNumber string `json:"invoiceNumber" example:"INV-2026-001"`

	// InvoiceDate is the date of the invoice
	// Example: "2026-04-15"
	InvoiceDate string `json:"invoiceDate" example:"2026-04-15"`

	// DueDate is the date when the invoice was due (invoice_date + payment_terms)
	// Example: "2026-05-15"
	DueDate string `json:"dueDate" example:"2026-05-15"`

	// TotalAmount is the original invoice total
	// Example: 5000000.00
	TotalAmount float64 `json:"totalAmount" example:"5000000.00"`

	// PaidAmount is the sum of all payments made for this invoice
	// Example: 2000000.00
	PaidAmount float64 `json:"paidAmount" example:"2000000.00"`

	// OutstandingBalance is the remaining unpaid amount
	// Calculated as: totalAmount - paidAmount
	// Example: 3000000.00
	OutstandingBalance float64 `json:"outstandingBalance" example:"3000000.00"`

	// DaysOverdue is the number of days past the due date (as of the report date)
	// Can be negative for future-due invoices
	// Example: 46
	DaysOverdue int `json:"daysOverdue" example:"46"`

	// AgingBucket is the aging bucket this invoice falls into
	// Values: "current", "31-60", "61-90", "90+"
	// Example: "31-60"
	AgingBucket string `json:"agingBucket" example:"31-60"`

	// PaymentStatus is the current payment status
	// Values: "unpaid", "partial", "paid"
	// Example: "partial"
	PaymentStatus string `json:"paymentStatus" example:"partial"`
}

// AgingGrandTotals represents aggregated totals across all suppliers
// Story 10.6, AC1: Grand totals for the entire report
type AgingGrandTotals struct {
	// Current represents total outstanding 0-30 days overdue across all suppliers
	// Example: 50000000.00
	Current float64 `json:"current" example:"50000000.00"`

	// Days31to60 represents total outstanding 31-60 days overdue across all suppliers
	// Example: 15000000.00
	Days31to60 float64 `json:"days31to60" example:"15000000.00"`

	// Days61to90 represents total outstanding 61-90 days overdue across all suppliers
	// Example: 10000000.00
	Days61to90 float64 `json:"days61to90" example:"10000000.00"`

	// DaysOver90 represents total outstanding over 90 days overdue across all suppliers
	// Example: 5000000.00
	DaysOver90 float64 `json:"daysOver90" example:"5000000.00"`

	// TotalOutstanding is the sum of all aging buckets across all suppliers
	// Example: 80000000.00
	TotalOutstanding float64 `json:"totalOutstanding" example:"80000000.00"`

	// TotalInvoices is the total number of outstanding invoices across all suppliers
	// Example: 25
	TotalInvoices int `json:"totalInvoices" example:"25"`

	// TotalSuppliers is the total number of suppliers with outstanding balances
	// Example: 8
	TotalSuppliers int `json:"totalSuppliers" example:"8"`
}

// SupplierAgingReportListFilter represents the query parameters for aging report requests
// Story 10.6, AC1: Filter parameters for aging report endpoint
type SupplierAgingReportListFilter struct {
	// AsOfDate is the snapshot date for aging calculations (required)
	// Example: "2026-05-31"
	AsOfDate string `form:"asOfDate" binding:"required" example:"2026-05-31"`

	// SupplierID filters by specific supplier (optional)
	// Example: 1
	SupplierID *uint `form:"supplierId" example:"1"`

	// BranchID filters by branch (optional, for access control)
	// Example: 1
	BranchID *uint `form:"branchId" example:"1"`

	// Page is the page number for supplier list pagination (default 1)
	// Example: 1
	Page int `form:"page,default=1" example:"1"`

	// Limit is the number of suppliers per page (default 20, max 100)
	// Example: 20
	Limit int `form:"limit,default=20" example:"20"`

	// IncludeDetails controls whether to include invoice-level details (default false)
	// When true, includes InvoiceAgingDetail for each supplier
	// Example: false
	IncludeDetails bool `form:"includeDetails,default=false" example:"false"`
}

// Helper function to parse date string for aging calculations
func ParseAgingDate(dateStr string) (time.Time, error) {
	return time.Parse("2006-01-02", dateStr)
}

// Helper function to calculate days between two dates
func DaysBetween(start, end time.Time) int {
	return int(end.Sub(start).Hours() / 24)
}

// Helper function to categorize days overdue into aging bucket
func CategorizeIntoBucket(daysOverdue int) string {
	if daysOverdue <= 30 {
		return "current"
	} else if daysOverdue <= 60 {
		return "31-60"
	} else if daysOverdue <= 90 {
		return "61-90"
	}
	return "90+"
}
