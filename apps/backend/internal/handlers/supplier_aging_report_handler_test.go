package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/dto"
)

// mockSupplierAgingReportService is a mock implementation for testing
type mockSupplierAgingReportService struct {
	report *dto.SupplierAgingReportResponse
	err    error
}

func (m *mockSupplierAgingReportService) GenerateAgingReport(ctx context.Context, request *dto.SupplierAgingReportRequest) (*dto.SupplierAgingReportResponse, error) {
	if m.err != nil {
		return nil, m.err
	}

	// Check for invalid date format in request
	if request.AsOfDate == "2026/05/31" {
		return nil, fmt.Errorf("invalid asOfDate format: cannot parse \"2026/05/31\"")
	}

	return m.report, nil
}

func (m *mockSupplierAgingReportService) ExportAgingReportPDF(ctx context.Context, request *dto.SupplierAgingReportRequest) ([]byte, string, error) {
	if m.err != nil {
		return nil, "", m.err
	}
	filename := "supplier-aging-report-" + request.AsOfDate + ".pdf"
	return []byte("PDF content"), filename, nil
}

func (m *mockSupplierAgingReportService) ExportAgingReportExcel(ctx context.Context, request *dto.SupplierAgingReportRequest) ([]byte, string, error) {
	if m.err != nil {
		return nil, "", m.err
	}
	filename := "supplier-aging-report-" + request.AsOfDate + ".xlsx"
	return []byte("Excel content"), filename, nil
}

// setupAgingReportTestRouter creates a test router with the aging report handler
func setupAgingReportTestRouter(handler *SupplierAgingReportHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Add authentication middleware mock
	router.Use(func(c *gin.Context) {
		c.Set("userID", uint(1))
		c.Set("username", "testowner")
		c.Set("role", "owner")
		c.Set("branchID", uint(1))
		c.Next()
	})

	router.POST("/api/v1/reports/supplier-aging", handler.GenerateAgingReport)
	router.POST("/api/v1/reports/supplier-aging/export/pdf", handler.ExportAgingReportPDF)
	router.POST("/api/v1/reports/supplier-aging/export/excel", handler.ExportAgingReportExcel)

	return router
}

// TestGenerateAgingReport_Success tests successful aging report generation
func TestGenerateAgingReport_Success(t *testing.T) {
	// Setup
	mockResponse := &dto.SupplierAgingReportResponse{
		AsOfDate:         "2026-05-31",
		ReportGeneratedAt: "2026-05-31T10:00:00Z",
		Currency:         "IDR",
		Suppliers: []dto.SupplierAgingSummary{
			{
				SupplierID:       1,
				SupplierName:     "PT. Test Supplier",
				TotalOutstanding: 10000000,
				InvoiceCount:     2,
			},
		},
		GrandTotals: dto.AgingGrandTotals{
			TotalOutstanding: 10000000,
			TotalInvoices:    2,
			TotalSuppliers:   1,
		},
	}

	mockService := &mockSupplierAgingReportService{report: mockResponse}
	handler := NewSupplierAgingReportHandler(mockService)
	router := setupAgingReportTestRouter(handler)

	// Create request
	requestBody := dto.SupplierAgingReportRequest{
		AsOfDate: "2026-05-31",
	}
	body, _ := json.Marshal(requestBody)

	// Create request
	req, _ := http.NewRequest("POST", "/api/v1/reports/supplier-aging", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	// Record response
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.SupplierAgingReportResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "2026-05-31", response.AsOfDate)
	assert.Equal(t, "IDR", response.Currency)
	assert.Len(t, response.Suppliers, 1)
}

// TestGenerateAgingReport_InvalidJSON tests handling of invalid JSON
func TestGenerateAgingReport_InvalidJSON(t *testing.T) {
	// Setup
	mockService := &mockSupplierAgingReportService{}
	handler := NewSupplierAgingReportHandler(mockService)
	router := setupAgingReportTestRouter(handler)

	// Create request with invalid JSON
	req, _ := http.NewRequest("POST", "/api/v1/reports/supplier-aging", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	// Record response
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var errorResp dto.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errorResp)
	assert.NoError(t, err)
	assert.Equal(t, "/errors/validation-error", errorResp.Type)
}

// TestGenerateAgingReport_MissingAsOfDate tests validation of asOfDate parameter
func TestGenerateAgingReport_MissingAsOfDate(t *testing.T) {
	// Setup
	mockService := &mockSupplierAgingReportService{}
	handler := NewSupplierAgingReportHandler(mockService)
	router := setupAgingReportTestRouter(handler)

	// Create request without asOfDate
	requestBody := map[string]interface{}{}
	body, _ := json.Marshal(requestBody)

	req, _ := http.NewRequest("POST", "/api/v1/reports/supplier-aging", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	// Record response
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestGenerateAgingReport_InvalidDateFormat tests validation of date format
func TestGenerateAgingReport_InvalidDateFormat(t *testing.T) {
	// Setup
	mockService := &mockSupplierAgingReportService{}
	handler := NewSupplierAgingReportHandler(mockService)
	router := setupAgingReportTestRouter(handler)

	// Create request with invalid date format
	requestBody := dto.SupplierAgingReportRequest{
		AsOfDate: "2026/05/31", // Wrong format
	}
	body, _ := json.Marshal(requestBody)

	req, _ := http.NewRequest("POST", "/api/v1/reports/supplier-aging", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	// Record response
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var errorResp dto.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errorResp)
	assert.NoError(t, err)
	assert.Contains(t, errorResp.Detail, "Tanggal laporan tidak valid")
}

// TestGenerateAgingReport_WithSupplierFilter tests filtering by supplier
func TestGenerateAgingReport_WithSupplierFilter(t *testing.T) {
	// Setup
	supplierID := uint(1)
	mockResponse := &dto.SupplierAgingReportResponse{
		AsOfDate: "2026-05-31",
		Suppliers: []dto.SupplierAgingSummary{
			{
				SupplierID:   1,
				SupplierName: "PT. Test Supplier",
			},
		},
		GrandTotals: dto.AgingGrandTotals{
			TotalSuppliers: 1,
		},
	}

	mockService := &mockSupplierAgingReportService{report: mockResponse}
	handler := NewSupplierAgingReportHandler(mockService)
	router := setupAgingReportTestRouter(handler)

	// Create request with supplier filter
	requestBody := dto.SupplierAgingReportRequest{
		AsOfDate:   "2026-05-31",
		SupplierID: &supplierID,
	}
	body, _ := json.Marshal(requestBody)

	req, _ := http.NewRequest("POST", "/api/v1/reports/supplier-aging", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	// Record response
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.SupplierAgingReportResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response.Suppliers, 1)
}

// TestExportAgingReportPDF_Success tests successful PDF export
func TestExportAgingReportPDF_Success(t *testing.T) {
	// Setup
	mockService := &mockSupplierAgingReportService{}
	handler := NewSupplierAgingReportHandler(mockService)
	router := setupAgingReportTestRouter(handler)

	// Create request
	requestBody := dto.SupplierAgingReportRequest{
		AsOfDate: "2026-05-31",
	}
	body, _ := json.Marshal(requestBody)

	req, _ := http.NewRequest("POST", "/api/v1/reports/supplier-aging/export/pdf", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	// Record response
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/pdf", w.Header().Get("Content-Type"))
	assert.Contains(t, w.Header().Get("Content-Disposition"), "supplier-aging-report-2026-05-31.pdf")
	assert.Equal(t, "PDF content", w.Body.String())
}

// TestExportAgingReportExcel_Success tests successful Excel export
func TestExportAgingReportExcel_Success(t *testing.T) {
	// Setup
	mockService := &mockSupplierAgingReportService{}
	handler := NewSupplierAgingReportHandler(mockService)
	router := setupAgingReportTestRouter(handler)

	// Create request
	requestBody := dto.SupplierAgingReportRequest{
		AsOfDate: "2026-05-31",
	}
	body, _ := json.Marshal(requestBody)

	req, _ := http.NewRequest("POST", "/api/v1/reports/supplier-aging/export/excel", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	// Record response
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", w.Header().Get("Content-Type"))
	assert.Contains(t, w.Header().Get("Content-Disposition"), "supplier-aging-report-2026-05-31.xlsx")
	assert.Equal(t, "Excel content", w.Body.String())
}

// TestExtractUserContext_Unauthenticated tests handling of unauthenticated requests
func TestExtractUserContext_Unauthenticated(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockService := &mockSupplierAgingReportService{}
	handler := NewSupplierAgingReportHandler(mockService)

	// Router without authentication middleware
	router.POST("/api/v1/reports/supplier-aging", handler.GenerateAgingReport)

	// Create request without auth context
	requestBody := dto.SupplierAgingReportRequest{
		AsOfDate: "2026-05-31",
	}
	body, _ := json.Marshal(requestBody)

	req, _ := http.NewRequest("POST", "/api/v1/reports/supplier-aging", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	// Record response
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert - should return 401 because userID is not in context
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var errorResp dto.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errorResp)
	assert.NoError(t, err)
	assert.Equal(t, "/errors/unauthorized", errorResp.Type)
}
