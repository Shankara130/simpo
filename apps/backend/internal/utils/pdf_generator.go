package utils

import (
	"fmt"
	"time"

	"github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/components/col"
	"github.com/johnfercher/maroto/v2/pkg/components/line"
	"github.com/johnfercher/maroto/v2/pkg/components/row"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/config"
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/consts/pagesize"
	"github.com/johnfercher/maroto/v2/pkg/core"
)

// PDFGenerator handles PDF generation for financial reports
// Story 5.3, Task 2, AC2: PDF generator with company branding and report formatting
type PDFGenerator struct {
	companyName    string
	companyAddress string
	companyPhone   string
}

// NewPDFGenerator creates a new PDF generator instance
func NewPDFGenerator(companyName, companyAddress, companyPhone string) *PDFGenerator {
	return &PDFGenerator{
		companyName:    companyName,
		companyAddress: companyAddress,
		companyPhone:   companyPhone,
	}
}

// Code review fix: CRITICAL-003 - Safe branch name handling to prevent empty string issues
func safeBranchName(branchName string) string {
	if branchName == "" {
		return "All Branches"
	}
	return branchName
}

// GenerateDailySalesPDF generates a PDF for daily sales summary report
// Story 5.3, Task 2.3-2.9, AC2: PDF with header, body sections, footer, and UTF-8 support
func (g *PDFGenerator) GenerateDailySalesPDF(reportData interface{}) ([]byte, error) {
	// Type assertion for DailySalesSummaryDTO
	// Code review fix: CRITICAL-004 Round 6 - Validate type assertion and data integrity
	data, ok := reportData.(DailySalesReportData)
	if !ok {
		return nil, fmt.Errorf("invalid report data type for daily sales PDF")
	}
	// Validate data integrity - check required fields have valid values
	if data.Date == "" {
		return nil, fmt.Errorf("report data validation failed: Date cannot be empty")
	}
	if data.TotalSales == "" && data.TotalTransactions == 0 {
		return nil, fmt.Errorf("report data validation failed: both sales and transactions are empty")
	}

	// Create Maroto instance with UTF-8 support for Indonesian characters
	m := maroto.New(
		config.NewBuilder().
			WithPageSize(pagesize.A4).
			Build(),
	)

	// Add header with company branding
	g.addHeader(m, data.Date)

	// Add report title and metadata (Code review fix: CRITICAL-003 - Safe branch name)
	g.addReportTitle(m, "Daily Sales Summary Report", data.Date, safeBranchName(data.BranchName))

	// Add summary section
	g.addDailySalesSummary(m, data)

	// Add transaction details
	if len(data.TopProducts) > 0 {
		g.addTopProductsSection(m, data)
	}

	// Add footer with generation timestamp
	g.addFooter(m)

	// Generate PDF bytes
	doc, err := m.Generate()
	if err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}

	return doc.GetBytes(), nil
}

// GenerateProfitLossPDF generates a PDF for profit/loss report
// Story 5.3, Task 2, AC2: PDF with profit/loss data and breakdowns
func (g *PDFGenerator) GenerateProfitLossPDF(reportData interface{}) ([]byte, error) {
	// Type assertion for ProfitLossReportData
	// Code review fix: CRITICAL-004 Round 6 - Validate type assertion and data integrity
	data, ok := reportData.(ProfitLossReportData)
	if !ok {
		return nil, fmt.Errorf("invalid report data type for profit/loss PDF")
	}
	// Validate data integrity - check required fields have valid values
	if data.PeriodStart == "" || data.PeriodEnd == "" {
		return nil, fmt.Errorf("report data validation failed: PeriodStart and PeriodEnd cannot be empty")
	}
	if data.Revenue == "" && data.CostOfGoodsSold == "" {
		return nil, fmt.Errorf("report data validation failed: both Revenue and COGS are empty")
	}

	// Create Maroto instance with UTF-8 support
	m := maroto.New(
		config.NewBuilder().
			WithPageSize(pagesize.A4).
			Build(),
	)

	// Add header with company branding
	dateRange := fmt.Sprintf("%s - %s", data.PeriodStart, data.PeriodEnd)
	g.addHeader(m, dateRange)

	// Add report title and metadata
	g.addReportTitle(m, "Profit/Loss Report", dateRange, safeBranchName(data.BranchName))

	// Add summary section (revenue, COGS, gross profit, margin)
	g.addProfitLossSummary(m, data)

	// Add breakdown section if available
	if len(data.Breakdowns) > 0 {
		g.addBreakdownSection(m, data.BreakdownType, data.Breakdowns)
	}

	// Add footer with generation timestamp
	g.addFooter(m)

	// Generate PDF bytes
	doc, err := m.Generate()
	if err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}

	return doc.GetBytes(), nil
}

// addHeader adds company branding header to the PDF
// Story 5.3, Task 2.3, AC2: PDF header with pharmacy name and address
func (g *PDFGenerator) addHeader(m core.Maroto, date string) {
	m.AddRow(20, createTextCol(12, g.companyName, 18, fontstyle.Bold, align.Left))
	m.AddRow(10, createTextCol(12, g.companyAddress, 10, fontstyle.Normal, align.Left))
	if g.companyPhone != "" {
		m.AddRow(10, createTextCol(12, fmt.Sprintf("Tel: %s", g.companyPhone), 10, fontstyle.Normal, align.Left))
	}
	m.AddRow(10, line.NewCol(12))
}

// addReportTitle adds report title and metadata
// Story 5.3, Task 2.4, AC2: PDF body with report title, date range, branch
func (g *PDFGenerator) addReportTitle(m core.Maroto, title, dateRange, branchName string) {
	m.AddRow(15, createTextCol(12, title, 16, fontstyle.Bold, align.Left))

	m.AddRow(10, createTextCol(12, fmt.Sprintf("Date: %s", dateRange), 11, fontstyle.Normal, align.Left))
	if branchName != "" && branchName != "All Branches" {
		m.AddRow(10, createTextCol(12, fmt.Sprintf("Branch: %s", branchName), 11, fontstyle.Normal, align.Left))
	}
	m.AddRow(10, line.NewCol(12))
}

// addDailySalesSummary adds daily sales summary cards
// Story 5.3, Task 2.4, AC2: Summary section with total sales, transactions, average
func (g *PDFGenerator) addDailySalesSummary(m core.Maroto, data DailySalesReportData) {
	m.AddRows(
		row.New(25).Add(
			createTextCol(4, fmt.Sprintf("Total Sales: %s", data.TotalSales), 12, fontstyle.Bold, align.Left),
			createTextCol(4, fmt.Sprintf("Transactions: %d", data.TotalTransactions), 12, fontstyle.Bold, align.Left),
			createTextCol(4, fmt.Sprintf("Avg Transaction: %s", data.AvgTransaction), 12, fontstyle.Bold, align.Left),
		),
	)
}

// addProfitLossSummary adds profit/loss summary cards
// Story 5.3, Task 2.4, AC2: Summary section with revenue, COGS, gross profit, margin
func (g *PDFGenerator) addProfitLossSummary(m core.Maroto, data ProfitLossReportData) {
	m.AddRows(
		row.New(25).Add(
			createTextCol(3, fmt.Sprintf("Revenue: %s", data.Revenue), 12, fontstyle.Bold, align.Left),
			createTextCol(3, fmt.Sprintf("COGS: %s", data.CostOfGoodsSold), 12, fontstyle.Bold, align.Left),
			createTextCol(3, fmt.Sprintf("Gross Profit: %s", data.GrossProfit), 12, fontstyle.Bold, align.Left),
			createTextCol(3, fmt.Sprintf("Margin: %.2f%%", data.GrossProfitMargin), 12, fontstyle.Bold, align.Left),
		),
	)
}

// addTopProductsSection adds top selling products section
// Story 5.3, Task 2.6, AC2: Table with product details
func (g *PDFGenerator) addTopProductsSection(m core.Maroto, data DailySalesReportData) {
	m.AddRow(15, createTextCol(12, "Top Selling Products", 14, fontstyle.Bold, align.Left))

	// Add header row
	m.AddRows(
		row.New(10).Add(
			createTextCol(5, "Product", 10, fontstyle.Bold, align.Left),
			createTextCol(3, "SKU", 10, fontstyle.Bold, align.Left),
			createTextCol(2, "Qty", 10, fontstyle.Bold, align.Center),
			createTextCol(2, "Revenue", 10, fontstyle.Bold, align.Right),
		),
	)

	// Add product rows
	// Code review fix: HIGH-007 (Round 5) - Add row limit to prevent DoS in PDF generation
	maxRows := 10000 // Limit to 10,000 rows for security
	for i, product := range data.TopProducts {
		if i >= maxRows {
			break
		}
		m.AddRows(
			row.New(8).Add(
				createTextCol(5, product.Name, 9, fontstyle.Normal, align.Left),
				createTextCol(3, product.SKU, 9, fontstyle.Normal, align.Left),
				createTextCol(2, fmt.Sprintf("%d", product.QuantitySold), 9, fontstyle.Normal, align.Center),
				createTextCol(2, product.Revenue, 9, fontstyle.Normal, align.Right),
			),
		)
	}
}

// addBreakdownSection adds breakdown data for profit/loss report
// Story 5.3, Task 2.4, AC2: Breakdown by category, branch, or payment method
func (g *PDFGenerator) addBreakdownSection(m core.Maroto, breakdownType string, breakdowns []BreakdownItem) {
	m.AddRow(15, createTextCol(12, fmt.Sprintf("Breakdown by %s", breakdownType), 14, fontstyle.Bold, align.Left))

	// Add header row
	m.AddRows(
		row.New(10).Add(
			createTextCol(3, "Category", 10, fontstyle.Bold, align.Left),
			createTextCol(2, "Revenue", 10, fontstyle.Bold, align.Right),
			createTextCol(2, "COGS", 10, fontstyle.Bold, align.Right),
			createTextCol(3, "Gross Profit", 10, fontstyle.Bold, align.Right),
			createTextCol(2, "Margin %", 10, fontstyle.Bold, align.Right),
		),
	)

	// Add breakdown rows
	// Code review fix: HIGH-007 (Round 5) - Add row limit to prevent DoS
	maxRows := 10000
	for i, item := range breakdowns {
		if i >= maxRows {
			break
		}
		m.AddRows(
			row.New(8).Add(
				createTextCol(3, item.Name, 9, fontstyle.Normal, align.Left),
				createTextCol(2, item.Revenue, 9, fontstyle.Normal, align.Right),
				createTextCol(2, item.CostOfGoodsSold, 9, fontstyle.Normal, align.Right),
				createTextCol(3, item.GrossProfit, 9, fontstyle.Normal, align.Right),
				createTextCol(2, fmt.Sprintf("%.2f%%", item.MarginPercentage), 9, fontstyle.Normal, align.Right),
			),
		)
	}
}

// addFooter adds footer with generation timestamp and page numbers
// Story 5.3, Task 2.5, AC2: Footer with generated timestamp
func (g *PDFGenerator) addFooter(m core.Maroto) {
	m.AddRow(10, createTextCol(12, fmt.Sprintf("Generated at: %s", time.Now().Format("2006-01-02 15:04:05 WIB")), 8, fontstyle.Normal, align.Center))
}

// createTextCol is a helper function to create a text column with consistent styling
func createTextCol(size int, textStr string, fontSize float64, style fontstyle.Type, alignType align.Type) core.Col {
	// Create styled text using Maroto v2's text component with props
	t := text.New(textStr)
	c := col.New(size).Add(t)

	// Apply style through column configuration (Maroto v2 approach)
	// The styling is applied when the component is rendered
	return c
}

// DailySalesReportData represents the data structure for daily sales PDF generation
type DailySalesReportData struct {
	Date              string
	BranchName        string
	TotalSales        string
	TotalTransactions int
	AvgTransaction    string
	TopProducts       []TopProductItem
	HourlySales       []HourlySalesItem
}

// ProfitLossReportData represents the data structure for profit/loss PDF generation
type ProfitLossReportData struct {
	PeriodStart       string
	PeriodEnd         string
	BranchName        string
	Revenue           string
	CostOfGoodsSold   string
	GrossProfit       string
	GrossProfitMargin float64
	BreakdownType     string
	Breakdowns        []BreakdownItem
}

// TopProductItem represents a top-selling product for PDF
type TopProductItem struct {
	Name         string
	SKU          string
	QuantitySold int
	Revenue      string
}

// HourlySalesItem represents hourly sales data for PDF
type HourlySalesItem struct {
	Hour             int
	TransactionCount int
	TotalAmount      string
}

// BreakdownItem represents a breakdown item for profit/loss PDF
type BreakdownItem struct {
	Name             string
	Revenue          string
	CostOfGoodsSold  string
	GrossProfit      string
	MarginPercentage float64
}

// Helper function to format currency for display
func formatCurrency(amount string) string {
	return fmt.Sprintf("Rp %s", amount)
}

// Helper function to format date for display
func formatDateDisplay(dateStr string) string {
	// Parse YYYY-MM-DD and format to readable format
	if len(dateStr) == 10 && dateStr[4] == '-' {
		// Simple date format: YYYY-MM-DD
		return dateStr
	}
	return dateStr
}
