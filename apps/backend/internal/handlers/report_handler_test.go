package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/dto"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/services"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/user"
)

// MockReportServiceForHandler is a mock implementation for testing
// Story 5.1, Task 5.2: Testing ReportHandler with mock service
type MockReportServiceForHandler struct {
	GenerateDailySalesSummaryFunc func(ctx context.Context, req *dto.DailySalesRequest) (*dto.DailySalesSummaryDTO, error)
}

func (m *MockReportServiceForHandler) GenerateDailySalesSummary(ctx context.Context, req *dto.DailySalesRequest) (*dto.DailySalesSummaryDTO, error) {
	if m.GenerateDailySalesSummaryFunc != nil {
		return m.GenerateDailySalesSummaryFunc(ctx, req)
	}
	return &dto.DailySalesSummaryDTO{}, nil
}

func (m *MockReportServiceForHandler) GenerateDailySales(ctx context.Context, branchID uint, startDate, endDate time.Time) (*services.SalesReport, error) {
	return nil, nil
}

func (m *MockReportServiceForHandler) GenerateProfitLoss(ctx context.Context, branchID uint, startDate, endDate time.Time) (*services.ProfitLossReport, error) {
	return nil, nil
}

func (m *MockReportServiceForHandler) ExportReport(ctx context.Context, reportType string, format string) ([]byte, error) {
	return nil, nil
}

// setupReportTestRouter creates a test router with report handler and auth context
// Story 5.1, Task 5.1: Test helper for report handler testing
func setupReportTestRouter(reportService services.ReportService, userRole string, branchID uint) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Add auth context middleware to simulate authenticated user
	router.Use(func(c *gin.Context) {
		c.Set("user_role", userRole)
		c.Set("branch_id", branchID)
		c.Next()
	})

	handler := NewReportHandler(reportService)
	router.GET("/api/v1/reports/daily", handler.GetDailySalesReport)

	return router
}

// TestReportHandler_GetDailySalesReport_Success tests successful report generation
// Story 5.1, Task 4.1-4.6: Full workflow test with valid date and branch
func TestReportHandler_GetDailySalesReport_Success(t *testing.T) {
	// Arrange: Create mock service with test data
	mockService := &MockReportServiceForHandler{
		GenerateDailySalesSummaryFunc: func(ctx context.Context, req *dto.DailySalesRequest) (*dto.DailySalesSummaryDTO, error) {
			return &dto.DailySalesSummaryDTO{
				Date:              req.Date,
				BranchID:          1,
				BranchName:        "Test Branch",
				TotalSales:        "1000000.00",
				TotalTransactions: 10,
				PaymentBreakdown: []dto.PaymentBreakdown{
					{PaymentMethod: "CASH", Amount: "500000.00", TransactionCount: 5, Percentage: 50.0},
				},
				GeneratedAt: time.Now(),
			}, nil
		},
	}

	router := setupReportTestRouter(mockService, user.RoleOwner, 1)

	// Act: Make request with valid parameters
	req, _ := http.NewRequest("GET", "/api/v1/reports/daily?date=2026-05-23&branch_id=1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert: Verify success response
	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.DailySalesSummaryDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "2026-05-23", response.Date)
	assert.Equal(t, uint(1), response.BranchID)
	assert.Equal(t, "Test Branch", response.BranchName)
	assert.Equal(t, "1000000.00", response.TotalSales)
	assert.Equal(t, 10, response.TotalTransactions)
}

// TestReportHandler_GetDailySalesReport_AllBranches tests report for all branches (no branch_id)
// Story 5.1, AC2: Branch filtering with nil branch_id parameter
func TestReportHandler_GetDailySalesReport_AllBranches(t *testing.T) {
	// Arrange
	mockService := &MockReportServiceForHandler{
		GenerateDailySalesSummaryFunc: func(ctx context.Context, req *dto.DailySalesRequest) (*dto.DailySalesSummaryDTO, error) {
			assert.Nil(t, req.BranchID, "BranchID should be nil for all branches")
			return &dto.DailySalesSummaryDTO{
				Date:              req.Date,
				BranchID:          0,
				BranchName:        "All Branches",
				TotalSales:        "2000000.00",
				TotalTransactions: 20,
				GeneratedAt:       time.Now(),
			}, nil
		},
	}

	router := setupReportTestRouter(mockService, user.RoleOwner, 1)

	// Act: Request without branch_id parameter
	req, _ := http.NewRequest("GET", "/api/v1/reports/daily?date=2026-05-23", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	var response dto.DailySalesSummaryDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "All Branches", response.BranchName)
}

// TestReportHandler_GetDailySalesReport_MissingDate tests missing date parameter
// Story 5.1, Task 4.3: Date parameter is required
func TestReportHandler_GetDailySalesReport_MissingDate(t *testing.T) {
	// Arrange
	mockService := &MockReportServiceForHandler{}
	router := setupReportTestRouter(mockService, user.RoleOwner, 1)

	// Act: Request without date parameter
	req, _ := http.NewRequest("GET", "/api/v1/reports/daily", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert: Should return 400 Bad Request
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var errorResp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &errorResp)
	require.NoError(t, err)
	assert.Equal(t, "Validation Failed", errorResp["title"])
	assert.Contains(t, errorResp["detail"], "Date")
}

// TestReportHandler_GetDailySalesReport_InvalidDateFormat tests invalid date format
// Story 5.1, Task 4.3: Date format validation (YYYY-MM-DD)
func TestReportHandler_GetDailySalesReport_InvalidDateFormat(t *testing.T) {
	// Arrange
	mockService := &MockReportServiceForHandler{}
	router := setupReportTestRouter(mockService, user.RoleOwner, 1)

	// Act: Request with invalid date format
	req, _ := http.NewRequest("GET", "/api/v1/reports/daily?date=invalid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert: Should return 400 Bad Request
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var errorResp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &errorResp)
	require.NoError(t, err)
	assert.Equal(t, "Validation Failed", errorResp["title"])
}

// TestReportHandler_GetDailySalesReport_CashierForbidden tests RBAC for cashier role
// Story 5.1, Task 4.5: Cashiers should not have access to financial reports
func TestReportHandler_GetDailySalesReport_CashierForbidden(t *testing.T) {
	// Arrange
	mockService := &MockReportServiceForHandler{}
	router := setupReportTestRouter(mockService, user.RoleCashier, 1)

	// Act: Request as cashier
	req, _ := http.NewRequest("GET", "/api/v1/reports/daily?date=2026-05-23", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert: Should return 403 Forbidden
	assert.Equal(t, http.StatusForbidden, w.Code)

	var errorResp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &errorResp)
	require.NoError(t, err)
	assert.Equal(t, "Access Denied", errorResp["title"])
	assert.Contains(t, errorResp["detail"], "permission")
}
