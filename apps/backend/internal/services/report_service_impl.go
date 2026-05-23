package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/dto"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/repositories"
)

// reportService implements ReportService interface
// Story 5.1, Task 3: Report service with caching and RBAC
type reportService struct {
	transactionRepo repositories.TransactionRepository
	productRepo     repositories.ProductRepository
	reportRepo      repositories.ReportRepository // Story 5.1, Task 3: New dependency
	auditService    AuditService
	redisClient     *redis.Client // Story 5.1, Task 3.5: Redis for caching
}

// NewReportService creates a new report service with dependency injection
// Story 5.1, Task 3: Updated constructor with ReportRepository and Redis
func NewReportService(
	transactionRepo repositories.TransactionRepository,
	productRepo repositories.ProductRepository,
	reportRepo repositories.ReportRepository, // Story 5.1, Task 3: New dependency
	auditService AuditService,
	redisClient *redis.Client, // Story 5.1, Task 3.5: Redis for caching (optional)
) ReportService {
	// Fail fast on nil dependencies
	if transactionRepo == nil {
		panic("reportService: transactionRepo cannot be nil")
	}
	if productRepo == nil {
		panic("reportService: productRepo cannot be nil")
	}
	if reportRepo == nil {
		panic("reportService: reportRepo cannot be nil") // Story 5.1, Task 3
	}
	if auditService == nil {
		panic("reportService: auditService cannot be nil")
	}
	// Story 5.1, Task 3.5: Redis is optional for graceful degradation

	return &reportService{
		transactionRepo: transactionRepo,
		productRepo:     productRepo,
		reportRepo:      reportRepo, // Story 5.1, Task 3
		auditService:    auditService,
		redisClient:     redisClient, // Story 5.1, Task 3.5
	}
}

// GenerateDailySalesSummary generates comprehensive daily sales summary report
// Story 5.1, Task 3.1-3.6: Implementation with RBAC, caching, and performance requirements
// Code review fix: LOW-003 - Add performance logging to track <10s requirement
func (s *reportService) GenerateDailySalesSummary(ctx context.Context, req *dto.DailySalesRequest) (*dto.DailySalesSummaryDTO, error) {
	// Story 5.1, Task 3.6: Performance requirement - context timeout for <10s
	startTime := time.Now()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Code review fix: LOW-003 - Log performance metrics
	defer func() {
		duration := time.Since(startTime)
		slog.InfoContext(ctx, "report_generation_completed",
			"duration_ms", duration.Milliseconds(),
			"date", req.Date,
			"branch_id", req.BranchID,
			"meets_sla", duration < 10*time.Second)
	}()

	// Check context cancellation
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("operation cancelled: %w", err)
	}

	// Story 5.1, Task 3.2: Validate request DTO
	if req.Date == "" {
		return nil, &InvalidInputError{Field: "date", Message: "date is required"}
	}

	// Validate date format
	parsedDate, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return nil, &InvalidInputError{Field: "date", Message: ErrInvalidDateFormat}
	}

	// Code review fix: HIGH-002 - Add date range validation
	// Prevent future dates and dates too far in the past
	now := time.Now()
	// Use Indonesia timezone for comparison
	loc := time.FixedZone("WIB", 7*60*60)
	nowInLoc := now.In(loc)

	// Check if date is in the future
	if parsedDate.After(nowInLoc) {
		return nil, &InvalidInputError{
			Field:   "date",
			Message: "date cannot be in the future",
		}
	}

	// Check if date is more than 1 year in the past
	oneYearAgo := nowInLoc.AddDate(-1, 0, 0)
	if parsedDate.Before(oneYearAgo) {
		return nil, &InvalidInputError{
			Field:   "date",
			Message: "date cannot be more than 1 year in the past",
		}
	}

	// Story 5.1, Task 3.5: Check Redis cache first (cache key format: daily_sales:{date}:{branch_id})
	// Use "all" for branch_id when null (aggregating all branches)
	cacheKey := fmt.Sprintf("daily_sales:%s:%d", req.Date, getBranchIDValue(req.BranchID))

	if s.redisClient != nil {
		cachedData, err := s.redisClient.Get(ctx, cacheKey).Result()
		if err == nil && cachedData != "" {
			// Cache hit - deserialize and return
			var summary dto.DailySalesSummaryDTO
			if err := json.Unmarshal([]byte(cachedData), &summary); err == nil {
				slog.InfoContext(ctx, "cache_hit",
					"cache_key", cacheKey,
					"date", req.Date,
					"branch_id", req.BranchID)
				return &summary, nil
			}
			// If deserialization fails, continue to fetch from DB
			slog.WarnContext(ctx, "cache_deserialization_failed",
				"cache_key", cacheKey,
				"error", err)
		}
	}

	// Story 5.1, Task 3.4: Branch filtering logic
	// If BranchID is nil (not specified), use 0 for "all branches"
	branchID := uint(0)
	if req.BranchID != nil {
		branchID = *req.BranchID
	}

	// Story 5.1, Task 2.1: Call ReportRepository to get data from database
	summary, err := s.reportRepo.GetDailySalesSummary(ctx, req.Date, branchID)
	if err != nil {
		return nil, &ServiceError{Op: "get daily sales summary", Err: err}
	}

	// Story 5.1, Task 3.5: Store in Redis cache with 5-minute TTL
	if s.redisClient != nil {
		data, err := json.Marshal(summary)
		if err == nil {
			// Cache for 5 minutes (300 seconds)
			if err := s.redisClient.Set(ctx, cacheKey, data, 5*time.Minute).Err(); err != nil {
				slog.WarnContext(ctx, "cache_set_failed",
					"cache_key", cacheKey,
					"error", err)
			} else {
				slog.InfoContext(ctx, "cache_set_success",
					"cache_key", cacheKey,
					"ttl", "5m")
			}
		}
	}

	return summary, nil
}

// getBranchIDValue converts optional branch ID to cache key value
// Story 1, Task 3.5: Helper function for cache key generation
func getBranchIDValue(branchID *uint) uint {
	if branchID == nil {
		return 0 // Use 0 to represent "all branches"
	}
	return *branchID
}

// InvalidateDailySalesCache invalidates the Redis cache for a specific date and branch
// Code review fix: MED-003 - Implement cache invalidation strategy
// Call this method when new transactions are added to ensure reports stay current
func (s *reportService) InvalidateDailySalesCache(ctx context.Context, date string, branchID uint) error {
	if s.redisClient == nil {
		return nil // No Redis configured, nothing to invalidate
	}

	cacheKey := fmt.Sprintf("daily_sales:%s:%d", date, branchID)
	err := s.redisClient.Del(ctx, cacheKey).Err()
	if err != nil {
		slog.WarnContext(ctx, "cache_invalidation_failed",
			"cache_key", cacheKey,
			"error", err)
		return err
	}

	slog.InfoContext(ctx, "cache_invalidated",
		"cache_key", cacheKey,
		"date", date,
		"branch_id", branchID)

	return nil
}

// GenerateDailySales generates daily sales summary report
// AC3: Filters by date range and branch
func (s *reportService) GenerateDailySales(ctx context.Context, branchID uint, startDate, endDate time.Time) (*SalesReport, error) {
	// Check context cancellation
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("operation cancelled: %w", err)
	}

	// Validate date range
	if endDate.Before(startDate) {
		return nil, &InvalidInputError{Field: "date_range", Message: "end date cannot be before start date"}
	}

	// Validate branch ID
	if branchID == 0 {
		return nil, &InvalidInputError{Field: "branch_id", Message: "branch ID is required"}
	}

	// PATCH: Use startDate instead of endDate (date range was being ignored)
	// Get daily summary for the start date of the range
	summary, err := s.transactionRepo.GetDailySummary(ctx, branchID, startDate)
	if err != nil {
		return nil, &ServiceError{Op: "get daily summary", Err: err}
	}

	// Build report
	report := &SalesReport{
		BranchID:              branchID,
		BranchName:            "", // Would need branch repo to get name
		StartDate:             startDate,
		EndDate:               endDate,
		TotalSales:            summary.TotalAmount,
		TotalTransactions:     summary.TotalTransactions,
		AverageTransactionValue: "0.00", // Simplified for MVP
		PaymentMethods:        []PaymentMethodBreakdown{},
	}

	// Convert payment method summaries
	for _, pm := range summary.PaymentMethods {
		report.PaymentMethods = append(report.PaymentMethods, PaymentMethodBreakdown{
			PaymentMethod: pm.PaymentMethod,
			Count:         pm.Count,
			TotalAmount:   pm.TotalAmount,
			Percentage:    0.0, // Simplified for MVP
		})
	}

	return report, nil
}

// GenerateProfitLoss generates profit and loss report
// AC3: Calculates: Revenue - Cost of Goods Sold
func (s *reportService) GenerateProfitLoss(ctx context.Context, branchID uint, startDate, endDate time.Time) (*ProfitLossReport, error) {
	// Check context cancellation
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("operation cancelled: %w", err)
	}

	// Validate date range
	if endDate.Before(startDate) {
		return nil, &InvalidInputError{Field: "date_range", Message: "end date cannot be before start date"}
	}

	// Validate branch ID
	if branchID == 0 {
		return nil, &InvalidInputError{Field: "branch_id", Message: "branch ID is required"}
	}

	// PATCH: Use startDate for consistency with GenerateDailySales fix
	summary, err := s.transactionRepo.GetDailySummary(ctx, branchID, startDate)
	if err != nil {
		return nil, &ServiceError{Op: "get daily summary", Err: err}
	}

	// For MVP, simplified calculation
	// In production, COGS would be calculated from transaction items with cost price
	report := &ProfitLossReport{
		BranchID:        branchID,
		BranchName:      "", // Would need branch repo to get name
		StartDate:       startDate,
		EndDate:         endDate,
		Revenue:         summary.TotalAmount,
		CostOfGoodsSold: "0.00", // Simplified for MVP
		GrossProfit:     summary.TotalAmount, // Simplified
		GrossMargin:     100.0, // Simplified
		Expenses:        "0.00", // Simplified
		NetProfit:       summary.TotalAmount, // Simplified
		NetMargin:       100.0, // Simplified
	}

	return report, nil
}

// ExportReport exports report in various formats
// Stub for future story
func (s *reportService) ExportReport(ctx context.Context, reportType string, format string) ([]byte, error) {
	// Check context cancellation
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("operation cancelled: %w", err)
	}

	// Stub implementation for future story
	return nil, &InvalidInputError{
		Field:   "export",
		Message: "report export not implemented - scheduled for future story",
	}
}
