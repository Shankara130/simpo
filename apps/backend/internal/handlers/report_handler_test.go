package handlers

import (
	"context"
	"encoding/json"
	"fmt"
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
	GenerateDailySalesSummaryFunc  func(ctx context.Context, req *dto.DailySalesRequest) (*dto.DailySalesSummaryDTO, error)
	GenerateProfitLossSummaryFunc   func(ctx context.Context, req *dto.ProfitLossRequest) (*dto.ProfitLossSummaryDTO, error)
}

func (m *MockReportServiceForHandler) GenerateDailySalesSummary(ctx context.Context, req *dto.DailySalesRequest) (*dto.DailySalesSummaryDTO, error) {
	if m.GenerateDailySalesSummaryFunc != nil {
		return m.GenerateDailySalesSummaryFunc(ctx, req)
	}
	return &dto.DailySalesSummaryDTO{}, nil
}

func (m *MockReportServiceForHandler) GenerateProfitLossSummary(ctx context.Context, req *dto.ProfitLossRequest) (*dto.ProfitLossSummaryDTO, error) {
	if m.GenerateProfitLossSummaryFunc != nil {
		return m.GenerateProfitLossSummaryFunc(ctx, req)
	}
	return &dto.ProfitLossSummaryDTO{}, nil
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

	// Create mock export service for testing
	mockExportService := &mockExportServiceForTest{}

	// Create mock audit service for testing (Code review fix: CRITICAL-004)
	mockAuditService := services.NewAuditService()

	handler := NewReportHandler(reportService, mockExportService, mockAuditService)
	router.GET("/api/v1/reports/daily", handler.GetDailySalesReport)

	return router
}

// mockExportServiceForTest is a mock implementation of ExportService for testing
// Story 5.3, Task 9.1: Mock export service for testing export handlers
type mockExportServiceForTest struct {
	pdfData    []byte
	excelData  []byte
	fileName   string
	shouldFail bool
}

func (m *mockExportServiceForTest) ExportDailySalesToPDF(ctx context.Context, req *dto.ExportRequest) (*dto.ExportResponse, error) {
	if m.shouldFail {
		return nil, &services.ServiceError{Op: "ExportDailySalesToPDF", Err: fmt.Errorf("export generation failed")}
	}
	data := m.pdfData
	if data == nil {
		data = []byte("%PDF-1.4 default test pdf")
	}
	filename := m.fileName
	if filename == "" {
		filename = "DailySalesReport_test.pdf"
	}
	return &dto.ExportResponse{
		FileName:    filename,
		ContentType: "application/pdf",
		FileData:    data,
	}, nil
}

func (m *mockExportServiceForTest) ExportDailySalesToExcel(ctx context.Context, req *dto.ExportRequest) (*dto.ExportResponse, error) {
	if m.shouldFail {
		return nil, &services.ServiceError{Op: "ExportDailySalesToExcel", Err: fmt.Errorf("export generation failed")}
	}
	data := m.excelData
	if data == nil {
		data = []byte{0x50, 0x4B, 0x03, 0x04} // Default Excel magic bytes
	}
	filename := m.fileName
	if filename == "" {
		filename = "DailySalesReport_test.xlsx"
	}
	return &dto.ExportResponse{
		FileName:    filename,
		ContentType: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		FileData:    data,
	}, nil
}

func (m *mockExportServiceForTest) ExportProfitLossToPDF(ctx context.Context, req *dto.ExportRequest) (*dto.ExportResponse, error) {
	if m.shouldFail {
		return nil, &services.ServiceError{Op: "ExportProfitLossToPDF", Err: fmt.Errorf("export generation failed")}
	}
	data := m.pdfData
	if data == nil {
		data = []byte("%PDF-1.4 default test pdf")
	}
	filename := m.fileName
	if filename == "" {
		filename = "ProfitLossReport_test.pdf"
	}
	return &dto.ExportResponse{
		FileName:    filename,
		ContentType: "application/pdf",
		FileData:    data,
	}, nil
}

func (m *mockExportServiceForTest) ExportProfitLossToExcel(ctx context.Context, req *dto.ExportRequest) (*dto.ExportResponse, error) {
	if m.shouldFail {
		return nil, &services.ServiceError{Op: "ExportProfitLossToExcel", Err: fmt.Errorf("export generation failed")}
	}
	data := m.excelData
	if data == nil {
		data = []byte{0x50, 0x4B, 0x03, 0x04} // Default Excel magic bytes
	}
	filename := m.fileName
	if filename == "" {
		filename = "ProfitLossReport_test.xlsx"
	}
	return &dto.ExportResponse{
		FileName:    filename,
		ContentType: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		FileData:    data,
	}, nil
}

func (m *mockExportServiceForTest) CreateAsyncExportJob(ctx context.Context, req *dto.ExportRequest) (string, error) {
	return "test-job-id", nil
}

func (m *mockExportServiceForTest) GetExportJobStatus(ctx context.Context, jobID string) (*dto.ExportJobStatusResponse, error) {
	return &dto.ExportJobStatusResponse{}, nil
}

func (m *mockExportServiceForTest) GetExportJobFile(ctx context.Context, jobID string) (*dto.ExportResponse, error) {
	return &dto.ExportResponse{}, nil
}

func (m *mockExportServiceForTest) CleanupExpiredJobs(ctx context.Context) error {
	return nil
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

// setupExportTestRouter creates a test router with export routes
// Story 5.3, Task 9.6: Test helper for export handler testing
func setupExportTestRouter(userRole string, branchID uint, exportService services.ExportService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Add auth context middleware to simulate authenticated user
	router.Use(func(c *gin.Context) {
		c.Set("user_role", userRole)
		c.Set("user_id", uint(1))
		c.Set("branch_id", branchID)
		c.Next()
	})

	// Create mock report service for testing
	mockReportService := &MockReportServiceForHandler{}

	// Create mock audit service for testing (Code review fix: CRITICAL-004)
	mockAuditService := services.NewAuditService()

	handler := NewReportHandler(mockReportService, exportService, mockAuditService)
	router.GET("/api/v1/reports/daily/export", handler.ExportDailySalesReport)
	router.GET("/api/v1/reports/profit-loss/export", handler.ExportProfitLossReport)

	return router
}

// TestReportHandler_ExportDailySalesReport_PDF_Success tests successful PDF export
// Story 5.3, Task 9.2: Test PDF export with valid data
func TestReportHandler_ExportDailySalesReport_PDF_Success(t *testing.T) {
	// Arrange: Create mock export service that returns PDF data
	mockExportService := &mockExportServiceForTest{
		pdfData:   []byte("%PDF-1.4 fake pdf data"),
		fileName: "DailySalesReport_2026-05-24.pdf",
	}

	router := setupExportTestRouter(user.RoleOwner, 1, mockExportService)

	// Act: Make export request with valid parameters
	req, _ := http.NewRequest("GET", "/api/v1/reports/daily/export?date=2026-05-24&format=pdf", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert: Verify success response with PDF headers
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/pdf", w.Header().Get("Content-Type"))
	assert.Contains(t, w.Header().Get("Content-Disposition"), "DailySalesReport_2026-05-24")
	assert.Equal(t, []byte("%PDF-1.4 fake pdf data"), w.Body.Bytes())
}

// TestReportHandler_ExportDailySalesReport_Excel_Success tests successful Excel export
// Story 5.3, Task 9.2: Test Excel export with valid data
func TestReportHandler_ExportDailySalesReport_Excel_Success(t *testing.T) {
	// Arrange: Create mock export service that returns Excel data
	mockExportService := &mockExportServiceForTest{
		excelData: []byte{0x50, 0x4B, 0x03, 0x04}, // ZIP/Excel magic bytes
		fileName:  "DailySalesReport_2026-05-24.xlsx",
	}

	router := setupExportTestRouter(user.RoleOwner, 1, mockExportService)

	// Act: Make export request with Excel format
	req, _ := http.NewRequest("GET", "/api/v1/reports/daily/export?date=2026-05-24&format=xlsx", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert: Verify success response with Excel headers
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", w.Header().Get("Content-Type"))
	assert.Contains(t, w.Header().Get("Content-Disposition"), "DailySalesReport_2026-05-24")
}

// TestReportHandler_ExportDailySalesReport_MissingDate tests missing date parameter
// Story 5.3, Task 9.6: Test validation - date parameter is required
func TestReportHandler_ExportDailySalesReport_MissingDate(t *testing.T) {
	// Arrange
	mockExportService := &mockExportServiceForTest{}
	router := setupExportTestRouter(user.RoleOwner, 1, mockExportService)

	// Act: Request without date parameter
	req, _ := http.NewRequest("GET", "/api/v1/reports/daily/export?format=pdf", nil)
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

// TestReportHandler_ExportDailySalesReport_InvalidFormat tests invalid format parameter
// Story 5.3, Task 9.6: Test validation - format must be pdf or xlsx
func TestReportHandler_ExportDailySalesReport_InvalidFormat(t *testing.T) {
	// Arrange
	mockExportService := &mockExportServiceForTest{}
	router := setupExportTestRouter(user.RoleOwner, 1, mockExportService)

	// Act: Request with invalid format
	req, _ := http.NewRequest("GET", "/api/v1/reports/daily/export?date=2026-05-24&format=docx", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert: Should return 400 Bad Request
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var errorResp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &errorResp)
	require.NoError(t, err)
	assert.Equal(t, "Invalid Format Parameter", errorResp["title"])
	assert.Contains(t, errorResp["detail"], "pdf")
	assert.Contains(t, errorResp["detail"], "xlsx")
}

// TestReportHandler_ExportDailySalesReport_CashierForbidden tests RBAC for export
// Story 5.3, Task 9.6: Cashiers should not have access to export
func TestReportHandler_ExportDailySalesReport_CashierForbidden(t *testing.T) {
	// Arrange
	mockExportService := &mockExportServiceForTest{}
	router := setupExportTestRouter(user.RoleCashier, 1, mockExportService)

	// Act: Request as cashier
	req, _ := http.NewRequest("GET", "/api/v1/reports/daily/export?date=2026-05-24&format=pdf", nil)
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

// TestReportHandler_ExportProfitLossReport_Success tests profit/loss export
// Story 5.3, Task 9.3: Test PDF export with profit/loss data
func TestReportHandler_ExportProfitLossReport_Success(t *testing.T) {
	// Arrange: Create mock export service that returns PDF data
	mockExportService := &mockExportServiceForTest{
		pdfData:   []byte("%PDF-1.4 fake profit loss pdf"),
		fileName:  "ProfitLossReport_2026-05-01_to_2026-05-24.pdf",
	}

	router := setupExportTestRouter(user.RoleOwner, 1, mockExportService)

	// Act: Make export request with profit/loss parameters
	req, _ := http.NewRequest("GET", "/api/v1/reports/profit-loss/export?start_date=2026-05-01&end_date=2026-05-24&format=pdf", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert: Verify success response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/pdf", w.Header().Get("Content-Type"))
	assert.Contains(t, w.Header().Get("Content-Disposition"), "ProfitLossReport_2026-05-01_to_2026-05-24")
}
