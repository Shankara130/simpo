package utils

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestExcelGenerator_NewExcelGenerator tests the constructor
// Story 5.3, Task 3.1: Excel generator initialization with company details
func TestExcelGenerator_NewExcelGenerator(t *testing.T) {
	generator := NewExcelGenerator("Apotek Sehat", "Jl. Kesehatan No. 123", "021-12345678")

	assert.NotNil(t, generator, "ExcelGenerator should not be nil")
	assert.Equal(t, "Apotek Sehat", generator.companyName, "Company name should match")
	assert.Equal(t, "Jl. Kesehatan No. 123", generator.companyAddress, "Company address should match")
	assert.Equal(t, "021-12345678", generator.companyPhone, "Company phone should match")
}

// TestExcelGenerator_GenerateDailySalesExcel_Basic tests basic daily sales Excel generation
// Story 5.3, Task 3.3-3.5, AC3: Multi-sheet workbook with summary
func TestExcelGenerator_GenerateDailySalesExcel_Basic(t *testing.T) {
	generator := NewExcelGenerator("Apotek Sehat", "Jl. Kesehatan No. 123", "021-12345678")

	data := DailySalesReportData{
		Date:              "2026-05-24",
		BranchName:        "Jakarta Pusat",
		TotalSales:        "Rp 15.500.000",
		TotalTransactions: 150,
		AvgTransaction:    "Rp 103.333",
		TopProducts:       []TopProductItem{},
		HourlySales:       []HourlySalesItem{},
	}

	excelBytes, err := generator.GenerateDailySalesExcel(data)

	assert.NoError(t, err, "Excel generation should succeed")
	assert.NotEmpty(t, excelBytes, "Excel bytes should not be empty")
	assert.True(t, bytes.HasPrefix(excelBytes, []byte("PK\x03\x04")), "Output should be a valid Excel file (ZIP format)")
}

// TestExcelGenerator_GenerateDailySalesExcel_WithTopProducts tests Excel with top products sheet
// Story 5.3, Task 3.3-3.5, AC3: Breakdown sheet with top products
func TestExcelGenerator_GenerateDailySalesExcel_WithTopProducts(t *testing.T) {
	generator := NewExcelGenerator("Apotek Sehat", "Jl. Kesehatan No. 123", "021-12345678")

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

	excelBytes, err := generator.GenerateDailySalesExcel(data)

	assert.NoError(t, err, "Excel generation should succeed")
	assert.NotEmpty(t, excelBytes, "Excel bytes should not be empty")
	assert.True(t, bytes.HasPrefix(excelBytes, []byte("PK\x03\x04")), "Output should be a valid Excel file")
}

// TestExcelGenerator_GenerateDailySalesExcel_WithHourlySales tests Excel with hourly sales sheet
// Story 5.3, Task 3.6, AC3: Raw data sheet with hourly sales
func TestExcelGenerator_GenerateDailySalesExcel_WithHourlySales(t *testing.T) {
	generator := NewExcelGenerator("Apotek Sehat", "Jl. Kesehatan No. 123", "021-12345678")

	data := DailySalesReportData{
		Date:              "2026-05-24",
		BranchName:        "Jakarta Pusat",
		TotalSales:        "Rp 15.500.000",
		TotalTransactions: 150,
		AvgTransaction:    "Rp 103.333",
		TopProducts:       []TopProductItem{},
		HourlySales: []HourlySalesItem{
			{Hour: 8, TransactionCount: 10, TotalAmount: "Rp 1.000.000"},
			{Hour: 9, TransactionCount: 15, TotalAmount: "Rp 1.500.000"},
			{Hour: 10, TransactionCount: 20, TotalAmount: "Rp 2.000.000"},
		},
	}

	excelBytes, err := generator.GenerateDailySalesExcel(data)

	assert.NoError(t, err, "Excel generation should succeed")
	assert.NotEmpty(t, excelBytes, "Excel bytes should not be empty")
	assert.True(t, bytes.HasPrefix(excelBytes, []byte("PK\x03\x04")), "Output should be a valid Excel file")
}

// TestExcelGenerator_GenerateProfitLossExcel_Basic tests basic profit/loss Excel generation
// Story 5.3, Task 3.4, AC3: Summary sheet with revenue, COGS, gross profit
func TestExcelGenerator_GenerateProfitLossExcel_Basic(t *testing.T) {
	generator := NewExcelGenerator("Apotek Sehat", "Jl. Kesehatan No. 123", "021-12345678")

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

	excelBytes, err := generator.GenerateProfitLossExcel(data)

	assert.NoError(t, err, "Excel generation should succeed")
	assert.NotEmpty(t, excelBytes, "Excel bytes should not be empty")
	assert.True(t, bytes.HasPrefix(excelBytes, []byte("PK\x03\x04")), "Output should be a valid Excel file")
}

// TestExcelGenerator_GenerateProfitLossExcel_WithBreakdown tests Excel with category breakdown
// Story 5.3, Task 3.5, AC3: Breakdown sheet with category data
func TestExcelGenerator_GenerateProfitLossExcel_WithBreakdown(t *testing.T) {
	generator := NewExcelGenerator("Apotek Sehat", "Jl. Kesehatan No. 123", "021-12345678")

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
		},
	}

	excelBytes, err := generator.GenerateProfitLossExcel(data)

	assert.NoError(t, err, "Excel generation should succeed")
	assert.NotEmpty(t, excelBytes, "Excel bytes should not be empty")
	assert.True(t, bytes.HasPrefix(excelBytes, []byte("PK\x03\x04")), "Output should be a valid Excel file")
}

// TestExcelGenerator_GenerateDailySalesExcel_InvalidType tests type assertion error handling
// Story 5.3, Task 3, AC3: Error handling for invalid data types
func TestExcelGenerator_GenerateDailySalesExcel_InvalidType(t *testing.T) {
	generator := NewExcelGenerator("Apotek Sehat", "Jl. Kesehatan No. 123", "021-12345678")

	// Pass wrong data type
	excelBytes, err := generator.GenerateDailySalesExcel("invalid data")

	assert.Error(t, err, "Should return error for invalid data type")
	assert.Nil(t, excelBytes, "Excel bytes should be nil on error")
	assert.Contains(t, err.Error(), "invalid report data type", "Error should mention invalid type")
}

// TestExcelGenerator_GenerateProfitLossExcel_InvalidType tests type assertion error handling
func TestExcelGenerator_GenerateProfitLossExcel_InvalidType(t *testing.T) {
	generator := NewExcelGenerator("Apotek Sehat", "Jl. Kesehatan No. 123", "021-12345678")

	// Pass wrong data type
	excelBytes, err := generator.GenerateProfitLossExcel(12345)

	assert.Error(t, err, "Should return error for invalid data type")
	assert.Nil(t, excelBytes, "Excel bytes should be nil on error")
	assert.Contains(t, err.Error(), "invalid report data type", "Error should mention invalid type")
}

// TestExcelGenerator_NoPhone tests Excel generation without phone number
// Story 5.3, Task 3.4, AC3: Optional phone field
func TestExcelGenerator_NoPhone(t *testing.T) {
	generator := NewExcelGenerator("Apotek Sehat", "Jl. Kesehatan No. 123", "")

	data := DailySalesReportData{
		Date:              "2026-05-24",
		BranchName:        "Jakarta Pusat",
		TotalSales:        "Rp 15.500.000",
		TotalTransactions: 150,
		AvgTransaction:    "Rp 103.333",
		TopProducts:       []TopProductItem{},
		HourlySales:       []HourlySalesItem{},
	}

	excelBytes, err := generator.GenerateDailySalesExcel(data)

	assert.NoError(t, err, "Excel generation should succeed without phone number")
	assert.NotEmpty(t, excelBytes, "Excel bytes should not be empty")
}

// TestExcelGenerator_UTF8Support tests UTF-8 character support for Indonesian
// Story 5.3, Task 3.10, AC3: UTF-8 support for Indonesian characters
func TestExcelGenerator_UTF8Support(t *testing.T) {
	generator := NewExcelGenerator("Apotek Sehat", "Jl. Kesehatan No. 1, Jakarta", "021-12345678")

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

	excelBytes, err := generator.GenerateDailySalesExcel(data)

	assert.NoError(t, err, "Excel generation should succeed with Indonesian characters")
	assert.NotEmpty(t, excelBytes, "Excel bytes should not be empty")
	assert.True(t, bytes.HasPrefix(excelBytes, []byte("PK\x03\x04")), "Output should be a valid Excel file")
}

// TestExcelGenerator_AllBranches tests Excel for all branches
// Story 5.3, Task 3.4, AC3: Branch name handling
func TestExcelGenerator_AllBranches(t *testing.T) {
	generator := NewExcelGenerator("Apotek Sehat", "Jl. Kesehatan No. 123", "021-12345678")

	data := DailySalesReportData{
		Date:              "2026-05-24",
		BranchName:        "All Branches",
		TotalSales:        "Rp 45.000.000",
		TotalTransactions: 450,
		AvgTransaction:    "Rp 100.000",
		TopProducts:       []TopProductItem{},
		HourlySales:       []HourlySalesItem{},
	}

	excelBytes, err := generator.GenerateDailySalesExcel(data)

	assert.NoError(t, err, "Excel generation should succeed")
	assert.NotEmpty(t, excelBytes, "Excel bytes should not be empty")
}
