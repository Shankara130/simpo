package utils

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestPDFGenerator_NewPDFGenerator tests the constructor
// Story 5.3, Task 2.2: PDF generator initialization with company details
func TestPDFGenerator_NewPDFGenerator(t *testing.T) {
	generator := NewPDFGenerator("Apotek Sehat", "Jl. Kesehatan No. 123", "021-12345678")

	assert.NotNil(t, generator, "PDFGenerator should not be nil")
	assert.Equal(t, "Apotek Sehat", generator.companyName, "Company name should match")
	assert.Equal(t, "Jl. Kesehatan No. 123", generator.companyAddress, "Company address should match")
	assert.Equal(t, "021-12345678", generator.companyPhone, "Company phone should match")
}

// TestPDFGenerator_GenerateDailySalesPDF_Basic tests basic daily sales PDF generation
// Story 5.3, Task 2.3-2.4, AC2: PDF with header, title, and summary
func TestPDFGenerator_GenerateDailySalesPDF_Basic(t *testing.T) {
	generator := NewPDFGenerator("Apotek Sehat", "Jl. Kesehatan No. 123", "021-12345678")

	data := DailySalesReportData{
		Date:              "2026-05-24",
		BranchName:        "Jakarta Pusat",
		TotalSales:        "Rp 15.500.000",
		TotalTransactions: 150,
		AvgTransaction:    "Rp 103.333",
		TopProducts:       []TopProductItem{},
		HourlySales:       []HourlySalesItem{},
	}

	pdfBytes, err := generator.GenerateDailySalesPDF(data)

	assert.NoError(t, err, "PDF generation should succeed")
	assert.NotEmpty(t, pdfBytes, "PDF bytes should not be empty")
	assert.True(t, bytes.HasPrefix(pdfBytes, []byte("%PDF-")), "Output should be a valid PDF file")
}

// TestPDFGenerator_GenerateDailySalesPDF_WithTopProducts tests PDF with top products
// Story 5.3, Task 2.6, AC2: Table with product details
func TestPDFGenerator_GenerateDailySalesPDF_WithTopProducts(t *testing.T) {
	generator := NewPDFGenerator("Apotek Sehat", "Jl. Kesehatan No. 123", "021-12345678")

	data := DailySalesReportData{
		Date:              "2026-05-24",
		BranchName:        "Jakarta Pusat",
		TotalSales:        "Rp 15.500.000",
		TotalTransactions: 150,
		AvgTransaction:    "Rp 103.333",
		TopProducts: []TopProductItem{
			{
				Name:         "Paracetamol 500mg",
				SKU:          "PARA-500",
				QuantitySold: 50,
				Revenue:      "Rp 250.000",
			},
			{
				Name:         "Amoxicillin 500mg",
				SKU:          "AMOX-500",
				QuantitySold: 30,
				Revenue:      "Rp 300.000",
			},
		},
		HourlySales: []HourlySalesItem{},
	}

	pdfBytes, err := generator.GenerateDailySalesPDF(data)

	assert.NoError(t, err, "PDF generation should succeed")
	assert.NotEmpty(t, pdfBytes, "PDF bytes should not be empty")
	assert.True(t, bytes.HasPrefix(pdfBytes, []byte("%PDF-")), "Output should be a valid PDF file")
}

// TestPDFGenerator_GenerateDailySalesPDF_AllBranches tests PDF for all branches
// Story 5.3, Task 2.4, AC2: Branch name handling
func TestPDFGenerator_GenerateDailySalesPDF_AllBranches(t *testing.T) {
	generator := NewPDFGenerator("Apotek Sehat", "Jl. Kesehatan No. 123", "021-12345678")

	data := DailySalesReportData{
		Date:              "2026-05-24",
		BranchName:        "All Branches",
		TotalSales:        "Rp 45.000.000",
		TotalTransactions: 450,
		AvgTransaction:    "Rp 100.000",
		TopProducts:       []TopProductItem{},
		HourlySales:       []HourlySalesItem{},
	}

	pdfBytes, err := generator.GenerateDailySalesPDF(data)

	assert.NoError(t, err, "PDF generation should succeed")
	assert.NotEmpty(t, pdfBytes, "PDF bytes should not be empty")
}

// TestPDFGenerator_GenerateProfitLossPDF_Basic tests basic profit/loss PDF generation
// Story 5.3, Task 2.4, AC2: Summary with revenue, COGS, gross profit, margin
func TestPDFGenerator_GenerateProfitLossPDF_Basic(t *testing.T) {
	generator := NewPDFGenerator("Apotek Sehat", "Jl. Kesehatan No. 123", "021-12345678")

	data := ProfitLossReportData{
		PeriodStart:       "2026-05-01",
		PeriodEnd:         "2026-05-24",
		BranchName:        "Jakarta Pusat",
		Revenue:           "Rp 450.000.000",
		CostOfGoodsSold:   "Rp 270.000.000",
		GrossProfit:       "Rp 180.000.000",
		GrossProfitMargin: 40.0,
		BreakdownType:     "",
		Breakdowns:        []BreakdownItem{},
	}

	pdfBytes, err := generator.GenerateProfitLossPDF(data)

	assert.NoError(t, err, "PDF generation should succeed")
	assert.NotEmpty(t, pdfBytes, "PDF bytes should not be empty")
	assert.True(t, bytes.HasPrefix(pdfBytes, []byte("%PDF-")), "Output should be a valid PDF file")
}

// TestPDFGenerator_GenerateProfitLossPDF_WithBreakdown tests PDF with category breakdown
// Story 5.3, Task 2.4, AC2: Breakdown by category, branch, or payment method
func TestPDFGenerator_GenerateProfitLossPDF_WithBreakdown(t *testing.T) {
	generator := NewPDFGenerator("Apotek Sehat", "Jl. Kesehatan No. 123", "021-12345678")

	data := ProfitLossReportData{
		PeriodStart:       "2026-05-01",
		PeriodEnd:         "2026-05-24",
		BranchName:        "Jakarta Pusat",
		Revenue:           "Rp 450.000.000",
		CostOfGoodsSold:   "Rp 270.000.000",
		GrossProfit:       "Rp 180.000.000",
		GrossProfitMargin: 40.0,
		BreakdownType:     "category",
		Breakdowns: []BreakdownItem{
			{
				Name:             "Obat Resep",
				Revenue:          "Rp 300.000.000",
				CostOfGoodsSold:  "Rp 180.000.000",
				GrossProfit:      "Rp 120.000.000",
				MarginPercentage: 40.0,
			},
			{
				Name:             "Obat Bebas",
				Revenue:          "Rp 100.000.000",
				CostOfGoodsSold:  "Rp 60.000.000",
				GrossProfit:      "Rp 40.000.000",
				MarginPercentage: 40.0,
			},
			{
				Name:             "Alat Kesehatan",
				Revenue:          "Rp 50.000.000",
				CostOfGoodsSold:  "Rp 30.000.000",
				GrossProfit:      "Rp 20.000.000",
				MarginPercentage: 40.0,
			},
		},
	}

	pdfBytes, err := generator.GenerateProfitLossPDF(data)

	assert.NoError(t, err, "PDF generation should succeed")
	assert.NotEmpty(t, pdfBytes, "PDF bytes should not be empty")
	assert.True(t, bytes.HasPrefix(pdfBytes, []byte("%PDF-")), "Output should be a valid PDF file")
}

// TestPDFGenerator_GenerateDailySalesPDF_InvalidType tests type assertion error handling
// Story 5.3, Task 2.3, AC2: Error handling for invalid data types
func TestPDFGenerator_GenerateDailySalesPDF_InvalidType(t *testing.T) {
	generator := NewPDFGenerator("Apotek Sehat", "Jl. Kesehatan No. 123", "021-12345678")

	// Pass wrong data type
	pdfBytes, err := generator.GenerateDailySalesPDF("invalid data")

	assert.Error(t, err, "Should return error for invalid data type")
	assert.Nil(t, pdfBytes, "PDF bytes should be nil on error")
	assert.Contains(t, err.Error(), "invalid report data type", "Error should mention invalid type")
}

// TestPDFGenerator_GenerateProfitLossPDF_InvalidType tests type assertion error handling
func TestPDFGenerator_GenerateProfitLossPDF_InvalidType(t *testing.T) {
	generator := NewPDFGenerator("Apotek Sehat", "Jl. Kesehatan No. 123", "021-12345678")

	// Pass wrong data type
	pdfBytes, err := generator.GenerateProfitLossPDF(12345)

	assert.Error(t, err, "Should return error for invalid data type")
	assert.Nil(t, pdfBytes, "PDF bytes should be nil on error")
	assert.Contains(t, err.Error(), "invalid report data type", "Error should mention invalid type")
}

// TestPDFGenerator_NoPhone tests PDF generation without phone number
// Story 5.3, Task 2.3, AC2: Optional phone field
func TestPDFGenerator_NoPhone(t *testing.T) {
	generator := NewPDFGenerator("Apotek Sehat", "Jl. Kesehatan No. 123", "")

	data := DailySalesReportData{
		Date:              "2026-05-24",
		BranchName:        "Jakarta Pusat",
		TotalSales:        "Rp 15.500.000",
		TotalTransactions: 150,
		AvgTransaction:    "Rp 103.333",
		TopProducts:       []TopProductItem{},
		HourlySales:       []HourlySalesItem{},
	}

	pdfBytes, err := generator.GenerateDailySalesPDF(data)

	assert.NoError(t, err, "PDF generation should succeed without phone number")
	assert.NotEmpty(t, pdfBytes, "PDF bytes should not be empty")
}

// TestPDFGenerator_UTF8Support tests UTF-8 character support for Indonesian
// Story 5.3, Task 2.9, AC2: UTF-8 support for Indonesian characters
func TestPDFGenerator_UTF8Support(t *testing.T) {
	generator := NewPDFGenerator("Apotek Sehat", "Jl. Kesehatan No. 1, Jakarta", "021-12345678")

	data := DailySalesReportData{
		Date:              "2026-05-24",
		BranchName:        "Cabang Jakarta Selatan",
		TotalSales:        "Rp 15.500.000",
		TotalTransactions: 150,
		AvgTransaction:    "Rp 103.333",
		TopProducts: []TopProductItem{
			{
				Name:         "Paracetamol 500mg",
				SKU:          "PARA-500",
				QuantitySold: 50,
				Revenue:      "Rp 250.000",
			},
		},
		HourlySales: []HourlySalesItem{},
	}

	pdfBytes, err := generator.GenerateDailySalesPDF(data)

	assert.NoError(t, err, "PDF generation should succeed with Indonesian characters")
	assert.NotEmpty(t, pdfBytes, "PDF bytes should not be empty")
	assert.True(t, bytes.HasPrefix(pdfBytes, []byte("%PDF-")), "Output should be a valid PDF file")
}
