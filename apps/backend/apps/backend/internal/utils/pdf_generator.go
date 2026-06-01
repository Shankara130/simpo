package utils

import (
	"fmt"
	"time"

	"github.com/phpdave11/gofpdf"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/models"
)

// Story 10.7: Implement Supplier Transaction Audit Trail - PDF Export

// GenerateAuditTrailPDF generates a professional PDF report for supplier audit trail
// AC: PDF export for compliance inspections with proper formatting and Indonesian locale
func GenerateAuditTrailPDF(audits []models.SupplierAuditTrail, startDate, endDate time.Time, pharmacyName string) (*gofpdf.Fpdf, error) {
	pdf := gofpdf.Init("P", "mm", "A4", "")
	pdf.AddPage()

	// Set up fonts - support for Indonesian characters
	pdf.SetFont("Arial", "", 10)

	// Add header section
	addPDFHeader(pdf, pharmacyName, startDate, endDate)

	// Add summary statistics
	summary := CalculateAuditTrailSummary(audits)
	addPDFSummary(pdf, summary)

	// Add audit trail table
	addPDFAuditTable(pdf, audits)

	return pdf, nil
}

// addPDFHeader adds the report header with pharmacy name, date range, and title
func addPDFHeader(pdf *gofpdf.Fpdf, pharmacyName string, startDate, endDate time.Time) {
	// Set title
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(0, 10, "Laporan Jejak Audit Supplier")
	pdf.Ln(12)

	// Set pharmacy name
	pdf.SetFont("Arial", "", 12)
	if pharmacyName == "" {
		pharmacyName = "Simpo Pharmacy"
	}
	pdf.Cell(0, 8, pharmacyName)
	pdf.Ln(8)

	// Set date range
	dateRange := fmt.Sprintf("Periode: %s s/d %s",
		startDate.Format("02/01/2006"),
		endDate.Format("02/01/2006"))
	pdf.Cell(0, 8, dateRange)
	pdf.Ln(8)

	// Add generation timestamp
	pdf.SetFont("Arial", "I", 9)
	pdf.Cell(0, 6, fmt.Sprintf("Dicetak: %s", time.Now().Format("02/01/2006 15:04")))
	pdf.Ln(10)

	// Add separator line
	pdf.SetDrawColor(200, 200, 200)
	pdf.Line(10, pdf.GetY(), 200, pdf.GetY())
	pdf.Ln(5)
}

// addPDFSummary adds summary statistics section
func addPDFSummary(pdf *gofpdf.Fpdf, summary map[string]interface{}) {
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(0, 8, "Ringkasan Statistik")
	pdf.Ln(8)

	pdf.SetFont("Arial", "", 10)

	// Total operations
	totalOps := summary["total_operations"].(int)
	pdf.Cell(0, 6, fmt.Sprintf("Total Operasi: %d", totalOps))
	pdf.Ln(6)

	// Total amount
	totalAmount := summary["total_amount"].(float64)
	pdf.Cell(0, 6, fmt.Sprintf("Total Transaksi: %s", formatIndonesianCurrency(totalAmount)))
	pdf.Ln(6)

	// User breakdown
	userBreakdown := summary["user_breakdown"].(map[uint]int)
	if len(userBreakdown) > 0 {
		pdf.Cell(0, 6, fmt.Sprintf("Jumlah Pengguna: %d", len(userBreakdown)))
		pdf.Ln(6)
	}

	// Transaction type breakdown
	typeBreakdown := summary["transaction_type_breakdown"].(map[string]int)
	if len(typeBreakdown) > 0 {
		pdf.Cell(0, 6, fmt.Sprintf("Jenis Transaksi: %d", len(typeBreakdown)))
		pdf.Ln(6)
	}

	pdf.Ln(3)
	pdf.SetDrawColor(200, 200, 200)
	pdf.Line(10, pdf.GetY(), 200, pdf.GetY())
	pdf.Ln(5)
}

// addPDFAuditTable adds the audit trail data table
func addPDFAuditTable(pdf *gofpdf.Fpdf, audits []models.SupplierAuditTrail) {
	pdf.SetFont("Arial", "B", 9)

	// Table headers
	headers := []string{"Tanggal", "Pengguna", "Jenis", "Entitas", "Aksi", "Deskripsi", "Jumlah"}
	colWidths := []float64{20, 25, 25, 15, 12, 50, 20}

	// Print header row
	x := 10.0
	for i, header := range headers {
		pdf.SetXY(x, pdf.GetY())
		pdf.CellFormat(colWidths[i], 7, header, "1", 0, "C", true, 0, "")
		x += colWidths[i]
	}
	pdf.Ln(7)

	// Print data rows
	pdf.SetFont("Arial", "", 8)

	for _, audit := range audits {
		// Check if we need a new page
		if pdf.GetY() > 270 {
			pdf.AddPage()
			// Reprint headers on new page
			pdf.SetFont("Arial", "B", 9)
			x = 10.0
			for i, header := range headers {
				pdf.SetXY(x, pdf.GetY())
				pdf.CellFormat(colWidths[i], 7, header, "1", 0, "C", true, 0, "")
				x += colWidths[i]
			}
			pdf.Ln(7)
			pdf.SetFont("Arial", "", 8)
		}

		x = 10.0

		// Format cells
		cells := []string{
			formatIndonesianDate(audit.CreatedAt),
			fmt.Sprintf("User %d", audit.UserID),
			audit.TransactionType,
			audit.EntityType,
			audit.ActionType,
			truncateString(audit.ActionDescription, 40),
			formatCurrencyIfExists(audit.TransactionAmount),
		}

		for i, cell := range cells {
			pdf.SetXY(x, pdf.GetY())
			pdf.CellFormat(colWidths[i], 6, cell, "LR", 0, "L", false, 0, "")
			x += colWidths[i]
		}
		pdf.Ln(6)
	}

	// Close table
	pdf.SetDrawColor(0, 0, 0)
	x = 10.0
	for _, width := range colWidths {
		pdf.Line(x, pdf.GetY(), x+width, pdf.GetY())
		x += width
	}
}

// formatCurrencyIfExists formats currency amount if present, returns "-" otherwise
func formatCurrencyIfExists(amount *float64) string {
	if amount == nil {
		return "-"
	}
	return formatIndonesianCurrency(*amount)
}

// truncateString truncates a string to max length and adds "..." if needed
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
