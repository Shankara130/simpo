package utils

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/models"
)

// Story 10.7: Implement Supplier Transaction Audit Trail

// formatIndonesianCurrency formats numbers in Indonesian currency style
// Formats numbers as Rp 1.000.000,00 (thousand separators, decimal comma)
func formatIndonesianCurrency(amount float64) string {
	// Format to string with 2 decimal places
	strValue := fmt.Sprintf("%.2f", amount)

	// Find decimal point
	dotIndex := len(strValue)
	for i, c := range strValue {
		if c == '.' {
			dotIndex = i
			break
		}
	}

	// Split into integer and fractional parts
	integerPart := strValue[:dotIndex]
	fractionalPart := strValue[dotIndex+1:]

	// Pad fractional part to 2 digits if needed
	for len(fractionalPart) < 2 {
		fractionalPart += "0"
	}

	// Add thousand separators to integer part
	formattedInteger := ""
	n := len(integerPart)
	for i := 0; i < n; i++ {
		if i > 0 && (n-i)%3 == 0 {
			formattedInteger += "."
		}
		formattedInteger += string(integerPart[i])
	}

	return "Rp " + formattedInteger + "," + fractionalPart
}

// formatIndonesianDate formats date in Indonesian locale (DD/MM/YYYY)
func formatIndonesianDate(t time.Time) string {
	return t.Format("02/01/2006")
}

// ExportSupplierAuditTrailToCSV exports audit trail data to CSV format
// Story 10.7, AC3: CSV export for compliance inspections
// AC: UTF-8 BOM for Excel compatibility with Indonesian text
func ExportSupplierAuditTrailToCSV(audits []models.SupplierAuditTrail, writer io.Writer) error {
	// Create CSV writer
	csvWriter := csv.NewWriter(writer)

	// Write UTF-8 BOM for Excel compatibility with Indonesian text
	// This ensures Excel correctly recognizes UTF-8 encoded Indonesian characters
	bom := []byte{0xEF, 0xBB, 0xBF}
	if _, err := writer.Write(bom); err != nil {
		return fmt.Errorf("failed to write UTF-8 BOM: %w", err)
	}

	// Write CSV headers
	headers := []string{
		"Timestamp",
		"User ID",
		"User Role",
		"Transaction Type",
		"Entity Type",
		"Entity ID",
		"Action Type",
		"Action Description",
		"Reason",
		"Transaction Amount",
		"Affected Items",
		"IP Address",
		"Branch ID",
	}

	if err := csvWriter.Write(headers); err != nil {
		return fmt.Errorf("failed to write CSV headers: %w", err)
	}

	// Write audit data rows
	for _, audit := range audits {
		// Format currency amount if present
		var amountStr string
		if audit.TransactionAmount != nil {
			amountStr = formatIndonesianCurrency(*audit.TransactionAmount)
		} else {
			amountStr = "-"
		}

		// Format date in Indonesian locale
		timestamp := formatIndonesianDate(audit.CreatedAt)

		record := []string{
			timestamp,
			strconv.Itoa(int(audit.UserID)),
			audit.UserRole,
			audit.TransactionType,
			audit.EntityType,
			strconv.Itoa(int(audit.EntityID)),
			audit.ActionType,
			audit.ActionDescription,
			func() string {
				if audit.Reason != "" {
					return audit.Reason
				}
				return "-"
			}(),
			amountStr,
			strconv.Itoa(audit.AffectedItemsCount),
			func() string {
				if audit.IPAddress != "" {
					return audit.IPAddress
				}
				return "-"
			}(),
			strconv.Itoa(int(audit.BranchID)),
		}

		if err := csvWriter.Write(record); err != nil {
			return fmt.Errorf("failed to write CSV record: %w", err)
		}
	}

	// Flush the CSV writer
	if err := csvWriter.Error(); err != nil {
		return fmt.Errorf("CSV writer error: %w", err)
	}

	csvWriter.Flush()
	return nil
}

// GenerateCSVFileName generates a filename for CSV export
// Story 10.7: File naming convention for exports
func GenerateCSVFileName(startDate, endDate time.Time) string {
	return fmt.Sprintf("supplier-audit-trail-%s-to-%s.csv",
		startDate.Format("2006-01-02"),
		endDate.Format("2006-01-02"))
}

// GeneratePDFFileName generates a filename for PDF export
// Story 10.7: File naming convention for PDF exports
func GeneratePDFFileName(startDate, endDate time.Time) string {
	return fmt.Sprintf("supplier-audit-trail-%s-to-%s.pdf",
		startDate.Format("2006-01-02"),
		endDate.Format("2006-01-02"))
}

// ValidateExportDateRange validates that the export date range is acceptable
// Story 10.7: Limit exports to maximum 1 year to prevent excessive data
func ValidateExportDateRange(startDate, endDate time.Time) error {
	// Ensure end date is after start date
	if endDate.Before(startDate) {
		return fmt.Errorf("end date must be after start date")
	}

	// Calculate duration
	duration := endDate.Sub(startDate)

	// Limit to 1 year (365 days)
	maxDuration := 365 * 24 * time.Hour
	if duration > maxDuration {
		return fmt.Errorf("export date range cannot exceed 1 year (maximum: 365 days)")
	}

	return nil
}

// CalculateAuditTrailSummary calculates summary statistics for audit trail export
// Story 10.7: Summary statistics for PDF export header
func CalculateAuditTrailSummary(audits []models.SupplierAuditTrail) map[string]interface{} {
	totalOperations := len(audits)
	totalAmount := 0.0
	userBreakdown := make(map[uint]int)
	transactionTypeBreakdown := make(map[string]int)

	for _, audit := range audits {
		// Count total operations
		totalOperations++

		// Sum transaction amounts
		if audit.TransactionAmount != nil {
			totalAmount += *audit.TransactionAmount
		}

		// User breakdown
		userBreakdown[audit.UserID]++

		// Transaction type breakdown
		transactionTypeBreakdown[audit.TransactionType]++
	}

	return map[string]interface{}{
		"total_operations":           totalOperations,
		"total_amount":              totalAmount,
		"user_breakdown":             userBreakdown,
		"transaction_type_breakdown": transactionTypeBreakdown,
	}
}
