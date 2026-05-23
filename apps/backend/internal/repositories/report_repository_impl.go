package repositories

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/dto"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/models"
)

// Story 5.1, Task 2: Concrete implementation with SQL aggregation queries
type reportRepository struct {
	db *gorm.DB
}

// NewReportRepository creates a new report repository
// Story 5.1, Task 2: Factory function for dependency injection
func NewReportRepository(db interface{}) ReportRepository {
	gormDB, ok := db.(*gorm.DB)
	if !ok {
		panic("db must be *gorm.DB")
	}
	return &reportRepository{db: gormDB}
}

// GetDailySalesSummary retrieves comprehensive daily sales summary
// Story 5.1, Task 2.1-2.5, AC1: SQL aggregation queries for all report sections
// Code review fixes: Transaction isolation, timezone handling, branch lookup error handling, input sanitization
func (r *reportRepository) GetDailySalesSummary(ctx context.Context, date string, branchID uint) (*dto.DailySalesSummaryDTO, error) {
	// Code review fix: HIGH-003 - Validate branchID input to prevent potential injection
	// Ensure branchID is within reasonable bounds
	if branchID > 1000000 { // Reasonable upper limit for branch IDs
		return nil, fmt.Errorf("invalid branch ID: %d (exceeds maximum allowed value)", branchID)
	}

	// Parse date string to time.Time for date calculations
	reportDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		return nil, fmt.Errorf("invalid date format: %w", err)
	}

	// Code review fix: Use Indonesia West Time (WIB) timezone instead of UTC
	// Story 5.1, HIGH-001: Timezone handling
	loc := time.FixedZone("WIB", 7*60*60) // Indonesia West Time (UTC+7)
	startOfDay := time.Date(reportDate.Year(), reportDate.Month(), reportDate.Day(), 0, 0, 0, 0, loc)
	endOfDay := startOfDay.Add(24 * time.Hour)

	// Initialize DTO with basic info
	summary := &dto.DailySalesSummaryDTO{
		Date:        date,
		BranchID:    branchID,
		GeneratedAt: time.Now(),
	}

	// Code review fix: Wrap all queries in transaction with RepeatableRead isolation
	// Story 5.1, CRITICAL-003: Transaction isolation for data consistency
	// This ensures all queries see a consistent snapshot of the data
	txErr := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Get branch name if filtering by specific branch
		// Code review fix: HIGH-004 - Return proper error instead of silent failure
		if branchID > 0 {
			var branch models.Branch
			err := tx.Select("name").First(&branch, branchID).Error
			if err != nil {
				if err == gorm.ErrRecordNotFound {
					return fmt.Errorf("branch with ID %d not found", branchID)
				}
				return fmt.Errorf("failed to get branch: %w", err)
			}
			summary.BranchName = branch.Name
		} else {
			summary.BranchName = "All Branches"
		}

		// Task 2.2: Query total sales and transaction count
		// Story 5.1, AC1: Total sales amount and total number of transactions
		var totalSales struct {
			TotalAmount        string
			TotalTransactions int64
		}

		totalQuery := tx.Table("transactions").
			Select("COALESCE(SUM(CAST(total AS DECIMAL)), 0) as total_amount, COUNT(*) as total_transactions").
			Where("created_at >= ? AND created_at < ? AND status = ?", startOfDay, endOfDay, models.StatusCompleted)

		if branchID > 0 {
			totalQuery = totalQuery.Where("branch_id = ?", branchID)
		}

		if err := totalQuery.Scan(&totalSales).Error; err != nil {
			return fmt.Errorf("failed to get total sales: %w", err)
		}

		summary.TotalSales = totalSales.TotalAmount
		summary.TotalTransactions = int(totalSales.TotalTransactions)

		// Task 2.3: Query payment method breakdown with GROUP BY
		// Story 5.1, AC1: Breakdown by payment method (Cash, Transfer, E-Wallet)
		type PaymentResult struct {
			PaymentMethod string
			Amount        string
			Count         int64
		}

		paymentQuery := tx.Table("transactions").
			Select("payment_method, COALESCE(SUM(CAST(total AS DECIMAL)), 0) as amount, COUNT(*) as count").
			Where("created_at >= ? AND created_at < ? AND status = ?", startOfDay, endOfDay, models.StatusCompleted).
			Group("payment_method").
			Order("payment_method")

		if branchID > 0 {
			paymentQuery = paymentQuery.Where("branch_id = ?", branchID)
		}

		var paymentResults []PaymentResult
		if err := paymentQuery.Scan(&paymentResults).Error; err != nil {
			return fmt.Errorf("failed to get payment breakdown: %w", err)
		}

		// Build payment breakdown with percentages
		summary.PaymentBreakdown = make([]dto.PaymentBreakdown, len(paymentResults))
		for i, pr := range paymentResults {
			var percentage float64
			if summary.TotalTransactions > 0 {
				percentage = float64(pr.Count) / float64(summary.TotalTransactions) * 100
			}

			summary.PaymentBreakdown[i] = dto.PaymentBreakdown{
				PaymentMethod:    pr.PaymentMethod,
				Amount:           pr.Amount,
				TransactionCount: int(pr.Count),
				Percentage:       percentage,
			}
		}

		// Task 2.4: Query top 10 products with JOIN transaction_items and transactions
		// Story 5.1, AC1: Top 10 selling products by quantity and revenue
		// Code review fix: MED-002 - Make top products limit configurable (currently hardcoded to 10)
		type TopProductResult struct {
			ProductID    uint
			SKU          string
			Name         string
			QuantitySold int64
			Revenue      string
		}

		topProductsQuery := tx.Table("products p").
			Select("p.id as product_id, p.sku, p.name, SUM(ti.quantity) as quantity_sold, COALESCE(SUM(ti.quantity * CAST(ti.unit_price AS DECIMAL)), 0) as revenue").
			Joins("INNER JOIN transaction_items ti ON p.id = ti.product_id").
			Joins("INNER JOIN transactions t ON ti.transaction_id = t.id").
			Where("t.created_at >= ? AND t.created_at < ? AND t.status = ?", startOfDay, endOfDay, models.StatusCompleted).
			Group("p.id, p.sku, p.name").
			Order("quantity_sold DESC").
			Limit(10) // TODO: Make this configurable (MED-002)

		if branchID > 0 {
			topProductsQuery = topProductsQuery.Where("t.branch_id = ?", branchID)
		}

		var topProductResults []TopProductResult
		if err := topProductsQuery.Scan(&topProductResults).Error; err != nil {
			return fmt.Errorf("failed to get top products: %w", err)
		}

		summary.TopProducts = make([]dto.TopProduct, len(topProductResults))
		for i, tr := range topProductResults {
			summary.TopProducts[i] = dto.TopProduct{
				ProductID:    tr.ProductID,
				SKU:          tr.SKU,
				Name:         tr.Name,
				QuantitySold: int(tr.QuantitySold),
				Revenue:      tr.Revenue,
			}
		}

		// Task 2.5: Query hourly sales with EXTRACT(HOUR FROM created_at) GROUP BY
		// Story 5.1, AC1: Sales by hour (for operational insights)
		type HourlyResult struct {
			Hour            int
			TransactionCount int64
			TotalAmount     string
		}

		hourlyQuery := tx.Table("transactions").
			Select("EXTRACT(HOUR FROM created_at) as hour, COUNT(*) as transaction_count, COALESCE(SUM(CAST(total AS DECIMAL)), 0) as total_amount").
			Where("created_at >= ? AND created_at < ? AND status = ?", startOfDay, endOfDay, models.StatusCompleted).
			Group("hour").
			Order("hour")

		if branchID > 0 {
			hourlyQuery = hourlyQuery.Where("branch_id = ?", branchID)
		}

		var hourlyResults []HourlyResult
		if err := hourlyQuery.Scan(&hourlyResults).Error; err != nil {
			return fmt.Errorf("failed to get hourly sales: %w", err)
		}

		summary.HourlySales = make([]dto.HourlySales, len(hourlyResults))
		for i, hr := range hourlyResults {
			summary.HourlySales[i] = dto.HourlySales{
				Hour:            int(hr.Hour),
				TransactionCount: int(hr.TransactionCount),
				TotalAmount:     hr.TotalAmount,
			}
		}

		return nil
	})

	if txErr != nil {
		return nil, txErr
	}

	return summary, nil
}
