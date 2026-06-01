package utils

import (
	"fmt"
	"time"

	"github.com/xuri/excelize/v2"
)

// DailySalesReportData represents the data structure for daily sales reports
// Story 5.3: Excel export for daily sales summary
type DailySalesReportData struct {
	Date               string
	BranchName         string
	TotalSales         string
	TotalTransactions  int
	AvgTransaction     string
	TopProducts        []TopProductItem
	HourlySales        []HourlySalesItem
	BreakdownType      string
}

// ProfitLossReportData represents the data structure for profit & loss reports
// Story 5.3: Excel export for profit & loss statements
type ProfitLossReportData struct {
	PeriodStart      string
	PeriodEnd        string
	BranchName       string
	Revenue          string
	CostOfGoodsSold  string
	GrossProfit      string
	GrossProfitMargin float64
	Breakdowns       []ProfitLossBreakdown
	BreakdownType    string
}

// TopProductItem represents a top-selling product in daily sales
type TopProductItem struct {
	ProductName  string
	Quantity     int
	Revenue      string
	Name         string
	SKU          string
	QuantitySold int
}

// HourlySalesItem represents hourly sales data
type HourlySalesItem struct {
	Hour            int
	Revenue         string
	Transactions    int
	TransactionCount int
	TotalAmount      string
}

// ProfitLossBreakdown represents a breakdown category in P&L reports
type ProfitLossBreakdown struct {
	Category         string
	Amount           string
	Name             string
	Revenue          string
	CostOfGoodsSold  string
	GrossProfit      string
	MarginPercentage float64
}

// BreakdownItem represents an individual breakdown item in P&L reports
type BreakdownItem struct {
	Name             string
	Revenue          string
	CostOfGoodsSold  string
	GrossProfit      string
	MarginPercentage float64
}

// ExcelGenerator handles Excel generation for financial reports
// Story 5.3, Task 3, AC3: Excel generator with multi-sheet structure
type ExcelGenerator struct {
	companyName    string
	companyAddress string
	companyPhone   string
}

// NewExcelGenerator creates a new Excel generator instance
func NewExcelGenerator(companyName, companyAddress, companyPhone string) *ExcelGenerator {
	return &ExcelGenerator{
		companyName:    companyName,
		companyAddress: companyAddress,
		companyPhone:   companyPhone,
	}
}

// GenerateDailySalesExcel generates an Excel file for daily sales summary report
// Story 5.3, Task 3.3-3.10, AC3: Multi-sheet workbook with formatting
func (g *ExcelGenerator) GenerateDailySalesExcel(reportData interface{}) ([]byte, error) {
	// Type assertion for DailySalesReportData
	// Code review fix: CRITICAL-004 Round 6 - Validate type assertion and data integrity
	data, ok := reportData.(DailySalesReportData)
	if !ok {
		return nil, fmt.Errorf("invalid report data type for daily sales Excel")
	}
	// Validate data integrity - check required fields have valid values
	if data.Date == "" {
		return nil, fmt.Errorf("report data validation failed: Date cannot be empty")
	}
	if data.TotalSales == "" && data.TotalTransactions == 0 {
		return nil, fmt.Errorf("report data validation failed: both sales and transactions are empty")
	}

	// Create new Excel file
	f := excelize.NewFile()

	// Create sheets
	_, err := f.NewSheet("Breakdown")
	if err != nil {
		return nil, fmt.Errorf("failed to create Breakdown sheet: %w", err)
	}
	_, err = f.NewSheet("Raw Data")
	if err != nil {
		return nil, fmt.Errorf("failed to create Raw Data sheet: %w", err)
	}

	// Set active sheet to Summary
	f.SetActiveSheet(0)

	// Set column widths
	f.SetColWidth("Summary", "A", "B", 20)
	f.SetColWidth("Summary", "C", "C", 25)
	f.SetColWidth("Breakdown", "A", "E", 18)
	f.SetColWidth("Raw Data", "A", "G", 15)

	// Add Summary sheet content
	g.addDailySalesSummarySheet(f, data)

	// Add Breakdown sheet content
	if len(data.TopProducts) > 0 {
		g.addTopProductsSheet(f, data)
	}

	// Add Raw Data sheet content
	if len(data.HourlySales) > 0 {
		g.addHourlySalesSheet(f, data)
	}

	// Generate buffer
	buffer, err := f.WriteToBuffer()
	if err != nil {
		return nil, fmt.Errorf("failed to write Excel buffer: %w", err)
	}

	return buffer.Bytes(), nil
}

// GenerateProfitLossExcel generates an Excel file for profit/loss report
// Story 5.3, Task 3, AC3: Multi-sheet workbook with breakdown data
func (g *ExcelGenerator) GenerateProfitLossExcel(reportData interface{}) ([]byte, error) {
	// Type assertion for ProfitLossReportData
	// Code review fix: CRITICAL-004 Round 6 - Validate type assertion and data integrity
	data, ok := reportData.(ProfitLossReportData)
	if !ok {
		return nil, fmt.Errorf("invalid report data type for profit/loss Excel")
	}
	// Validate data integrity - check required fields have valid values
	if data.PeriodStart == "" || data.PeriodEnd == "" {
		return nil, fmt.Errorf("report data validation failed: PeriodStart and PeriodEnd cannot be empty")
	}
	if data.Revenue == "" && data.CostOfGoodsSold == "" {
		return nil, fmt.Errorf("report data validation failed: both Revenue and COGS are empty")
	}

	// Create new Excel file
	f := excelize.NewFile()

	// Create sheets
	_, err := f.NewSheet("Breakdown")
	if err != nil {
		return nil, fmt.Errorf("failed to create Breakdown sheet: %w", err)
	}
	_, err = f.NewSheet("Raw Data")
	if err != nil {
		return nil, fmt.Errorf("failed to create Raw Data sheet: %w", err)
	}

	// Set active sheet to Summary
	f.SetActiveSheet(0)

	// Set column widths
	f.SetColWidth("Summary", "A", "B", 20)
	f.SetColWidth("Summary", "C", "C", 25)
	f.SetColWidth("Breakdown", "A", "E", 18)

	// Add Summary sheet content
	g.addProfitLossSummarySheet(f, data)

	// Add Breakdown sheet content
	if len(data.Breakdowns) > 0 {
		g.addBreakdownSheet(f, data)
	}

	// Add Raw Data sheet content
	// Code review fix: HIGH-015 (Round 4) - Add Raw Data sheet for profit/loss reports
	g.addProfitLossRawDataSheet(f)

	// Generate buffer
	buffer, err := f.WriteToBuffer()
	if err != nil {
		return nil, fmt.Errorf("failed to write Excel buffer: %w", err)
	}

	return buffer.Bytes(), nil
}

// addDailySalesSummarySheet adds the summary sheet for daily sales report
func (g *ExcelGenerator) addDailySalesSummarySheet(f *excelize.File, data DailySalesReportData) {
	// Add title and metadata
	f.SetCellValue("Summary", "A1", g.companyName)
	f.SetCellValue("Summary", "A2", g.companyAddress)
	if g.companyPhone != "" {
		f.SetCellValue("Summary", "A3", fmt.Sprintf("Tel: %s", g.companyPhone))
	}

	// Add report title
	row := 5
	if g.companyPhone != "" {
		row = 6
	}
	f.SetCellValue("Summary", fmt.Sprintf("A%d", row), "Daily Sales Summary Report")
	f.SetCellValue("Summary", fmt.Sprintf("A%d", row+1), fmt.Sprintf("Date: %s", data.Date))
	f.SetCellValue("Summary", fmt.Sprintf("A%d", row+2), fmt.Sprintf("Branch: %s", data.BranchName))
	f.SetCellValue("Summary", fmt.Sprintf("A%d", row+3), fmt.Sprintf("Generated: %s", time.Now().Format("2006-01-02 15:04:05 WIB")))

	// Add summary metrics
	dataStartRow := row + 5
	f.SetCellValue("Summary", fmt.Sprintf("A%d", dataStartRow), "Metric")
	f.SetCellValue("Summary", fmt.Sprintf("B%d", dataStartRow), "Value")
	f.SetCellValue("Summary", fmt.Sprintf("A%d", dataStartRow+1), "Total Sales")
	f.SetCellValue("Summary", fmt.Sprintf("B%d", dataStartRow+1), data.TotalSales)
	f.SetCellValue("Summary", fmt.Sprintf("A%d", dataStartRow+2), "Total Transactions")
	f.SetCellValue("Summary", fmt.Sprintf("B%d", dataStartRow+2), data.TotalTransactions)
	f.SetCellValue("Summary", fmt.Sprintf("A%d", dataStartRow+3), "Average Transaction")
	f.SetCellValue("Summary", fmt.Sprintf("B%d", dataStartRow+3), data.AvgTransaction)
}

// addProfitLossSummarySheet adds the summary sheet for profit/loss report
func (g *ExcelGenerator) addProfitLossSummarySheet(f *excelize.File, data ProfitLossReportData) {
	// Add title and metadata
	f.SetCellValue("Summary", "A1", g.companyName)
	f.SetCellValue("Summary", "A2", g.companyAddress)
	if g.companyPhone != "" {
		f.SetCellValue("Summary", "A3", fmt.Sprintf("Tel: %s", g.companyPhone))
	}

	// Add report title
	row := 5
	if g.companyPhone != "" {
		row = 6
	}
	dateRange := fmt.Sprintf("%s - %s", data.PeriodStart, data.PeriodEnd)
	f.SetCellValue("Summary", fmt.Sprintf("A%d", row), "Profit/Loss Report")
	f.SetCellValue("Summary", fmt.Sprintf("A%d", row+1), fmt.Sprintf("Period: %s", dateRange))
	f.SetCellValue("Summary", fmt.Sprintf("A%d", row+2), fmt.Sprintf("Branch: %s", data.BranchName))
	f.SetCellValue("Summary", fmt.Sprintf("A%d", row+3), fmt.Sprintf("Generated: %s", time.Now().Format("2006-01-02 15:04:05 WIB")))

	// Add summary metrics
	dataStartRow := row + 5
	f.SetCellValue("Summary", fmt.Sprintf("A%d", dataStartRow), "Metric")
	f.SetCellValue("Summary", fmt.Sprintf("B%d", dataStartRow), "Value")
	f.SetCellValue("Summary", fmt.Sprintf("A%d", dataStartRow+1), "Revenue")
	f.SetCellValue("Summary", fmt.Sprintf("B%d", dataStartRow+1), data.Revenue)
	f.SetCellValue("Summary", fmt.Sprintf("A%d", dataStartRow+2), "Cost of Goods Sold")
	f.SetCellValue("Summary", fmt.Sprintf("B%d", dataStartRow+2), data.CostOfGoodsSold)
	f.SetCellValue("Summary", fmt.Sprintf("A%d", dataStartRow+3), "Gross Profit")
	f.SetCellValue("Summary", fmt.Sprintf("B%d", dataStartRow+3), data.GrossProfit)
	f.SetCellValue("Summary", fmt.Sprintf("A%d", dataStartRow+4), "Gross Profit Margin")
	f.SetCellValue("Summary", fmt.Sprintf("B%d", dataStartRow+4), fmt.Sprintf("%.2f%%", data.GrossProfitMargin))
}

// addTopProductsSheet adds the breakdown sheet with top products
func (g *ExcelGenerator) addTopProductsSheet(f *excelize.File, data DailySalesReportData) {
	// Add header
	f.SetCellValue("Breakdown", "A1", "Top Selling Products")
	f.SetCellValue("Breakdown", "A3", "Product")
	f.SetCellValue("Breakdown", "B3", "SKU")
	f.SetCellValue("Breakdown", "C3", "Quantity Sold")
	f.SetCellValue("Breakdown", "D3", "Revenue")

	// Add data rows
	// Code review fix: CRITICAL-005 - Add row limit to prevent DoS attacks
	maxRows := 10000 // Limit to 10,000 rows for security
	for i, product := range data.TopProducts {
		if i >= maxRows {
			break // Stop processing to prevent memory exhaustion
		}
		row := i + 4
		f.SetCellValue("Breakdown", fmt.Sprintf("A%d", row), product.Name)
		f.SetCellValue("Breakdown", fmt.Sprintf("B%d", row), product.SKU)
		f.SetCellValue("Breakdown", fmt.Sprintf("C%d", row), product.QuantitySold)
		f.SetCellValue("Breakdown", fmt.Sprintf("D%d", row), product.Revenue)
	}

}

// addHourlySalesSheet adds the raw data sheet with hourly sales
func (g *ExcelGenerator) addHourlySalesSheet(f *excelize.File, data DailySalesReportData) {
	// Add header
	f.SetCellValue("Raw Data", "A1", "Hour")
	f.SetCellValue("Raw Data", "B1", "Transaction Count")
	f.SetCellValue("Raw Data", "C1", "Total Amount")

	// Add data rows
	// Code review fix: CRITICAL-005 - Add row limit to prevent DoS attacks
	maxRows := 10000 // Limit to 10,000 rows for security
	for i, hour := range data.HourlySales {
		if i >= maxRows {
			break // Stop processing to prevent memory exhaustion
		}
		row := i + 2
		f.SetCellValue("Raw Data", fmt.Sprintf("A%d", row), hour.Hour)
		f.SetCellValue("Raw Data", fmt.Sprintf("B%d", row), hour.TransactionCount)
		f.SetCellValue("Raw Data", fmt.Sprintf("C%d", row), hour.TotalAmount)
	}
}

// addBreakdownSheet adds the breakdown sheet for profit/loss report
func (g *ExcelGenerator) addBreakdownSheet(f *excelize.File, data ProfitLossReportData) {
	// Add header
	f.SetCellValue("Breakdown", "A1", fmt.Sprintf("Breakdown by %s", data.BreakdownType))
	f.SetCellValue("Breakdown", "A3", "Category")
	f.SetCellValue("Breakdown", "B3", "Revenue")
	f.SetCellValue("Breakdown", "C3", "COGS")
	f.SetCellValue("Breakdown", "D3", "Gross Profit")
	f.SetCellValue("Breakdown", "E3", "Margin %")

	// Add data rows
	for i, item := range data.Breakdowns {
		row := i + 4
		f.SetCellValue("Breakdown", fmt.Sprintf("A%d", row), item.Name)
		f.SetCellValue("Breakdown", fmt.Sprintf("B%d", row), item.Revenue)
		f.SetCellValue("Breakdown", fmt.Sprintf("C%d", row), item.CostOfGoodsSold)
		f.SetCellValue("Breakdown", fmt.Sprintf("D%d", row), item.GrossProfit)
		f.SetCellValue("Breakdown", fmt.Sprintf("E%d", row), fmt.Sprintf("%.2f%%", item.MarginPercentage))
	}
}

// addProfitLossRawDataSheet adds a raw data sheet placeholder for profit/loss reports
// Code review fix: HIGH-015 (Round 4) - Add Raw Data sheet with explanatory note
func (g *ExcelGenerator) addProfitLossRawDataSheet(f *excelize.File) {
	// Add header explaining that raw transaction data is not included in this report
	f.SetCellValue("Raw Data", "A1", "Raw Transaction Data")
	f.SetCellValue("Raw Data", "A3", "Note: Detailed transaction data is not included in this report type.")
	f.SetCellValue("Raw Data", "A4", "For individual transaction details, please use the Daily Sales Summary Report.")
	f.SetCellValue("Raw Data", "A6", "This report provides:")
	f.SetCellValue("Raw Data", "A7", "- Summary: Overall profit/loss metrics")
	f.SetCellValue("Raw Data", "A8", "- Breakdown: Categorized analysis by category, branch, or payment method")
}
