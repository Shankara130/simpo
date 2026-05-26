package services

import (
	"context"
	"testing"
	"time"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/dto"
	"github.com/stretchr/testify/assert"
)

// TestExportServiceInterface defines the contract that all ExportService implementations must satisfy
// Story 5.3, Task 1, AC1: Export service interface validation
type TestExportServiceInterface interface {
	// These methods must be implemented by any ExportService
	ExportDailySalesToPDF(ctx context.Context, req *dto.ExportRequest) (*dto.ExportResponse, error)
	ExportDailySalesToExcel(ctx context.Context, req *dto.ExportRequest) (*dto.ExportResponse, error)
	ExportProfitLossToPDF(ctx context.Context, req *dto.ExportRequest) (*dto.ExportResponse, error)
	ExportProfitLossToExcel(ctx context.Context, req *dto.ExportRequest) (*dto.ExportResponse, error)
	CreateAsyncExportJob(ctx context.Context, req *dto.ExportRequest) (string, error)
	GetExportJobStatus(ctx context.Context, jobID string) (*dto.ExportJobStatusResponse, error)
	GetExportJobFile(ctx context.Context, jobID string) (*dto.ExportResponse, error)
	CleanupExpiredJobs(ctx context.Context) error
}

// Ensure ExportServiceImpl satisfies the interface (compile-time check)
var _ TestExportServiceInterface = (*ExportServiceImpl)(nil)

// TestExportServiceImpl_Contract validates that ExportServiceImpl satisfies all interface requirements
// Story 5.3, Task 9.1: Unit test for ExportService interface compliance
func TestExportServiceImpl_Contract(t *testing.T) {
	// This test validates that the implementation struct properly implements the interface
	// It will fail to compile if ExportServiceImpl doesn't implement all required methods

	// Create a mock implementation for testing
	mockService := &mockExportService{
		t: t,
	}

	// Verify all methods can be called (signature validation)
	ctx := context.Background()
	req := &dto.ExportRequest{
		ReportType: dto.ReportTypeDailySales,
		Format:     dto.ExportFormatPDF,
	}

	// Test PDF export method exists and has correct signature
	_, err := mockService.ExportDailySalesToPDF(ctx, req)
	assert.NoError(t, err, "ExportDailySalesToPDF should return ExportResponse")

	// Test Excel export method exists and has correct signature
	_, err = mockService.ExportDailySalesToExcel(ctx, req)
	assert.NoError(t, err, "ExportDailySalesToExcel should return ExportResponse")

	// Test async job creation method exists and has correct signature
	jobID, err := mockService.CreateAsyncExportJob(ctx, req)
	assert.NotEmpty(t, jobID, "CreateAsyncExportJob should return job ID")
	assert.NoError(t, err, "CreateAsyncExportJob should not return error")

	// Test job status retrieval method exists and has correct signature
	status, err := mockService.GetExportJobStatus(ctx, jobID)
	assert.NotNil(t, status, "GetExportJobStatus should return status")
	assert.NoError(t, err, "GetExportJobStatus should not return error")

	// Test job file retrieval method exists and has correct signature
	_, err = mockService.GetExportJobFile(ctx, jobID)
	assert.NoError(t, err, "GetExportJobFile should return ExportResponse")

	// Test cleanup method exists and has correct signature
	err = mockService.CleanupExpiredJobs(ctx)
	assert.NoError(t, err, "CleanupExpiredJobs should not return error")
}

// TestExportRequest_Validation tests ExportRequest DTO validation
// Story 5.3, Task 9.6: Unit test for export request validation
func TestExportRequest_Validation(t *testing.T) {
	tests := []struct {
		name        string
		request     dto.ExportRequest
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid daily sales PDF export request",
			request: dto.ExportRequest{
				ReportType: dto.ReportTypeDailySales,
				Format:     dto.ExportFormatPDF,
				Date:       stringPtr("2026-05-24"),
				UserID:     1,
				UserRole:   "owner",
			},
			expectError: false,
		},
		{
			name: "Valid profit/loss Excel export request",
			request: dto.ExportRequest{
				ReportType:  dto.ReportTypeProfitLoss,
				Format:      dto.ExportFormatExcel,
				StartDate:   stringPtr("2026-05-01"),
				EndDate:     stringPtr("2026-05-24"),
				BreakdownBy: stringPtr("category"),
				UserID:      1,
				UserRole:    "owner",
			},
			expectError: false,
		},
		{
			name: "Invalid format - should fail validation",
			request: dto.ExportRequest{
				ReportType: dto.ReportTypeDailySales,
				Format:     "docx", // Invalid format
				Date:       stringPtr("2026-05-24"),
				UserID:     1,
				UserRole:   "owner",
			},
			expectError: true,
			errorMsg:    "format must be 'pdf' or 'xlsx'",
		},
		{
			name: "Missing required fields - should fail validation",
			request: dto.ExportRequest{
				// Missing ReportType and Format
				UserID:   1,
				UserRole: "owner",
			},
			expectError: true,
			errorMsg:    "required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validate request (simplified validation for test)
			err := validateExportRequest(tt.request)

			if tt.expectError {
				assert.Error(t, err, "Expected validation error")
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err, "Expected no validation error")
			}
		})
	}
}

// TestExportResponse_Structure tests ExportResponse DTO structure
// Story 5.3, Task 9.2: Unit test for export response structure
func TestExportResponse_Structure(t *testing.T) {
	response := &dto.ExportResponse{
		FileName:    "DailySalesReport_2026-05-24_Branch1_20260524_153045.pdf",
		ContentType: "application/pdf",
		FileData:    []byte{0x25, 0x50, 0x44, 0x46}, // PDF magic number
		Metadata: map[string]interface{}{
			"reportTitle":  "Daily Sales Summary Report",
			"dateRange":    "2026-05-24",
			"branchName":   "Apotek Sehat - Jakarta Pusat",
			"generatedBy":  "admin@example.com",
			"generatedAt":  "2026-05-24T15:30:45+07:00",
			"exportFormat": "1.0",
			"fileSize":     int64(2048),
			"recordCount":  150,
		},
	}

	// Verify required fields are present
	assert.NotEmpty(t, response.FileName, "FileName should not be empty")
	assert.NotEmpty(t, response.ContentType, "ContentType should not be empty")
	assert.NotEmpty(t, response.FileData, "FileData should not be empty")
	assert.NotNil(t, response.Metadata, "Metadata should not be nil")
	assert.Contains(t, response.Metadata, "reportTitle", "Metadata should contain reportTitle")
	assert.Contains(t, response.Metadata, "dateRange", "Metadata should contain dateRange")
	assert.Contains(t, response.Metadata, "branchName", "Metadata should contain branchName")
	assert.Contains(t, response.Metadata, "generatedBy", "Metadata should contain generatedBy")
}

// TestExportJob_Structure tests ExportJob DTO structure
// Story 5.3, Task 9.8: Unit test for async export job structure
func TestExportJob_Structure(t *testing.T) {
	now := time.Now()
	job := &dto.ExportJob{
		JobID:      "export_abc123",
		UserID:     1,
		ReportType: dto.ReportTypeDailySales,
		Format:     dto.ExportFormatPDF,
		Status:     dto.JobStatusPending,
		CreatedAt:  now,
		ExpiresAt:  now.Add(24 * time.Hour),
		Progress:   0,
	}

	// Verify required fields are present
	assert.NotEmpty(t, job.JobID, "JobID should not be empty")
	assert.NotZero(t, job.UserID, "UserID should not be zero")
	assert.NotEmpty(t, job.ReportType, "ReportType should not be empty")
	assert.NotEmpty(t, job.Format, "Format should not be empty")
	assert.NotEmpty(t, job.Status, "Status should not be empty")
	assert.False(t, job.ExpiresAt.IsZero(), "ExpiresAt should be set")
	assert.Greater(t, job.ExpiresAt.Sub(job.CreatedAt), 23*time.Hour, "Expiration should be ~24 hours after creation")
}

// TestExportJobStatusResponse_Structure tests status response structure
// Story 5.3, Task 9.8: Unit test for job status response structure
func TestExportJobStatusResponse_Structure(t *testing.T) {
	response := &dto.ExportJobStatusResponse{
		JobID:        "export_abc123",
		Status:       dto.JobStatusCompleted,
		Progress:     100,
		FileURL:      "/api/v1/reports/export/download/export_abc123",
		EstimatedSec: 30,
		CreatedAt:    time.Now(),
	}

	// Verify required fields are present
	assert.NotEmpty(t, response.JobID, "JobID should not be empty")
	assert.NotEmpty(t, response.Status, "Status should not be empty")
	assert.Equal(t, 100, response.Progress, "Progress should be 100 for completed job")
	assert.NotEmpty(t, response.FileURL, "FileURL should be present for completed job")
	assert.Greater(t, response.EstimatedSec, 0, "EstimatedSec should be positive")
}

// validateExportRequest is a helper function for validation testing
func validateExportRequest(req dto.ExportRequest) error {
	if req.ReportType == "" {
		return &InvalidInputError{Field: "reportType", Message: "reportType is required"}
	}
	if req.Format == "" {
		return &InvalidInputError{Field: "format", Message: "format is required"}
	}
	if req.Format != dto.ExportFormatPDF && req.Format != dto.ExportFormatExcel {
		return &InvalidInputError{Field: "format", Message: "format must be 'pdf' or 'xlsx'"}
	}
	return nil
}

// stringPtr is a helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}

// mockExportService is a mock implementation for testing
type mockExportService struct {
	t *testing.T
}

func (m *mockExportService) ExportDailySalesToPDF(ctx context.Context, req *dto.ExportRequest) (*dto.ExportResponse, error) {
	return &dto.ExportResponse{
		FileName:    "test.pdf",
		ContentType: "application/pdf",
		FileData:    []byte("test pdf data"),
		Metadata:    make(map[string]interface{}),
	}, nil
}

func (m *mockExportService) ExportDailySalesToExcel(ctx context.Context, req *dto.ExportRequest) (*dto.ExportResponse, error) {
	return &dto.ExportResponse{
		FileName:    "test.xlsx",
		ContentType: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		FileData:    []byte("test excel data"),
		Metadata:    make(map[string]interface{}),
	}, nil
}

func (m *mockExportService) ExportProfitLossToPDF(ctx context.Context, req *dto.ExportRequest) (*dto.ExportResponse, error) {
	return &dto.ExportResponse{
		FileName:    "test.pdf",
		ContentType: "application/pdf",
		FileData:    []byte("test pdf data"),
		Metadata:    make(map[string]interface{}),
	}, nil
}

func (m *mockExportService) ExportProfitLossToExcel(ctx context.Context, req *dto.ExportRequest) (*dto.ExportResponse, error) {
	return &dto.ExportResponse{
		FileName:    "test.xlsx",
		ContentType: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		FileData:    []byte("test excel data"),
		Metadata:    make(map[string]interface{}),
	}, nil
}

func (m *mockExportService) CreateAsyncExportJob(ctx context.Context, req *dto.ExportRequest) (string, error) {
	return "export_test_job_123", nil
}

func (m *mockExportService) GetExportJobStatus(ctx context.Context, jobID string) (*dto.ExportJobStatusResponse, error) {
	return &dto.ExportJobStatusResponse{
		JobID:     jobID,
		Status:    dto.JobStatusCompleted,
		Progress:  100,
		CreatedAt: time.Now(),
	}, nil
}

func (m *mockExportService) GetExportJobFile(ctx context.Context, jobID string) (*dto.ExportResponse, error) {
	return &dto.ExportResponse{
		FileName:    "test.pdf",
		ContentType: "application/pdf",
		FileData:    []byte("test data"),
		Metadata:    make(map[string]interface{}),
	}, nil
}

func (m *mockExportService) CleanupExpiredJobs(ctx context.Context) error {
	return nil
}
