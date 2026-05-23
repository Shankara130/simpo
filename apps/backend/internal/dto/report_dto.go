package dto

import "time"

// DailySalesSummaryDTO represents the complete daily sales summary report
// Story 5.1, Task 1.1, AC1: Daily sales summary with all report sections
type DailySalesSummaryDTO struct {
	Date              string            `json:"date"`                        // Report date in YYYY-MM-DD format
	BranchID          uint              `json:"branchId"`                    // Branch ID (0 means all branches aggregated)
	BranchName        string            `json:"branchName"`                  // Branch name (empty if aggregating all branches)
	TotalSales        string            `json:"totalSales"`                  // Total sales amount (decimal string for precision)
	TotalTransactions int               `json:"totalTransactions"`           // Total number of transactions
	PaymentBreakdown  []PaymentBreakdown `json:"paymentBreakdown"`           // Breakdown by payment method (Task 1.2)
	TopProducts       []TopProduct      `json:"topProducts"`                 // Top 10 selling products (Task 1.3)
	HourlySales       []HourlySales     `json:"hourlySales"`                 // Sales by hour for operational insights (Task 1.4)
	GeneratedAt       time.Time         `json:"generatedAt"`                 // Report generation timestamp (ISO 8601)
}

// PaymentBreakdown represents payment method statistics
// Story 5.1, Task 1.2, AC1: Payment method breakdown with amount, count, and percentage
type PaymentBreakdown struct {
	PaymentMethod    string  `json:"paymentMethod"`    // Payment method: CASH, TRANSFER, E-WALLET
	Amount           string  `json:"amount"`           // Total amount for this payment method (decimal string)
	TransactionCount int     `json:"transactionCount"` // Number of transactions with this payment method
	Percentage       float64 `json:"percentage"`       // Percentage of total transactions (0-100)
}

// TopProduct represents a top-selling product in the daily report
// Story 5.1, Task 1.3, AC1: Top 10 products by quantity and revenue
type TopProduct struct {
	ProductID    uint   `json:"productId"`    // Product ID
	SKU          string `json:"sku"`          // Product SKU
	Name         string `json:"name"`         // Product name
	QuantitySold int    `json:"quantitySold"` // Total quantity sold
	Revenue      string `json:"revenue"`      // Total revenue from this product (decimal string)
}

// HourlySales represents sales data for a specific hour
// Story 5.1, Task 1.4, AC1: Sales by hour for operational insights
type HourlySales struct {
	Hour            int    `json:"hour"`            // Hour of day (0-23)
	TransactionCount int    `json:"transactionCount"` // Number of transactions in this hour
	TotalAmount     string `json:"totalAmount"`     // Total sales amount in this hour (decimal string)
}

// DailySalesRequest represents the request parameters for generating a daily sales summary
// Story 5.1, Task 3.2, AC1, AC2: Date and optional branch_id parameters
type DailySalesRequest struct {
	Date     string `json:"date" binding:"required"`      // Report date in YYYY-MM-DD format (required)
	BranchID *uint  `json:"branchId"`                    // Branch ID filter (optional, null means all branches)
}
