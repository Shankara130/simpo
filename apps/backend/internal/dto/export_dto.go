package dto

import "time"

// ==============================================================================
// Story 5.3: Report Export Functionality DTOs
// ==============================================================================

// ExportRequest represents the request parameters for exporting a report
// Story 5.3, Task 1.3, AC1: Request with report type and format
type ExportRequest struct {
	ReportType string    `json:"reportType" binding:"required"` // Report type: daily_sales, profit_loss
	Format     string    `json:"format" binding:"required"`     // Export format: pdf, xlsx
	Date       *string   `json:"date"`                           // Single date for daily sales (YYYY-MM-DD)
	StartDate  *string   `json:"startDate"`                      // Start date for profit/loss (YYYY-MM-DD)
	EndDate    *string   `json:"endDate"`                        // End date for profit/loss (YYYY-MM-DD)
	BranchID   *uint     `json:"branchId"`                       // Branch ID filter (optional)
	BreakdownBy *string  `json:"breakdownBy"`                     // Breakdown type for profit/loss (category, branch, payment_method)
	UserID    uint      `json:"-"`                               // User ID from context (not in JSON)
	UserRole  string    `json:"-"`                               // User role from context (not in JSON)
}

// ExportResponse represents the response from an export operation
// Story 5.3, Task 1.4, AC4: Response with file data and metadata
type ExportResponse struct {
	FileName    string                 `json:"fileName"`    // Generated filename
	ContentType string                 `json:"contentType"` // MIME type: application/pdf or application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
	FileData    []byte                 `json:"-"`          // Binary file data (not in JSON)
	Metadata    map[string]interface{} `json:"metadata"`   // Export metadata for traceability
	GeneratedAt time.Time              `json:"generatedAt"` // Generation timestamp
}

// ExportMetadata contains metadata for exported files
// Story 5.3, AC2, AC3, AC4: Metadata for traceability and compliance
type ExportMetadata struct {
	ReportTitle    string `json:"reportTitle"`    // Report title (e.g., "Daily Sales Summary Report")
	DateRange      string `json:"dateRange"`      // Formatted date range
	BranchName     string `json:"branchName"`     // Branch location
	GeneratedBy    string `json:"generatedBy"`    // User who generated the export
	GeneratedAt    string `json:"generatedAt"`    // Generation timestamp (formatted)
	ExportFormat   string `json:"exportFormat"`   // Export format version
	FileSize       int64  `json:"fileSize"`       // File size in bytes
	RecordCount    int    `json:"recordCount"`    // Number of records in export
}

// ExportJob represents an async export job
// Story 5.3, Task 5, AC1, AC3: Async export for large reports
type ExportJob struct {
	JobID       string    `json:"jobId"`       // Unique job identifier
	UserID      uint      `json:"userId"`      // User who requested the export
	ReportType  string    `json:"reportType"`  // Report type: daily_sales, profit_loss
	Format      string    `json:"format"`      // Export format: pdf, xlsx
	Status      string    `json:"status"`      // Job status: pending, processing, completed, failed
	FilePath    string    `json:"filePath"`    // Path to generated file (when completed)
	CreatedAt   time.Time `json:"createdAt"`   // Job creation timestamp
	CompletedAt *time.Time `json:"completedAt,omitempty"` // Job completion timestamp (nullable)
	ExpiresAt   time.Time `json:"expiresAt"`   // Job expiration timestamp (24 hours after creation)
	Progress    int       `json:"progress"`    // Progress percentage (0-100)
	Error       string    `json:"error,omitempty"` // Error message (if failed)
}

// ExportJobStatusResponse represents the status response for an export job
// Story 5.3, Task 5.4, AC1: Status endpoint response
type ExportJobStatusResponse struct {
	JobID        string     `json:"jobId"`        // Job identifier
	Status       string     `json:"status"`       // Current job status
	Progress     int        `json:"progress"`     // Progress percentage (0-100)
	FileURL      string     `json:"fileUrl,omitempty"` // Download URL (when completed)
	EstimatedSec int        `json:"estimatedSec,omitempty"` // Estimated time to completion (seconds)
	Error        string     `json:"error,omitempty"`     // Error message (if failed)
	CreatedAt    time.Time  `json:"createdAt"`    // Job creation timestamp
}

// Constants for export-related errors
const (
	// ErrInvalidExportFormat is returned when format is not pdf or xlsx
	// Story 5.3, Error Handling: Domain error for invalid format
	ErrInvalidExportFormat = "invalid export format: must be 'pdf' or 'xlsx'"

	// ErrUnauthorizedExport is returned when user lacks permission to export
	// Story 5.3, Security Requirements: Only Owner and Admin can export financial reports
	ErrUnauthorizedExport = "unauthorized: insufficient permissions to export financial reports"

	// ErrExportGenerationFailed is returned when PDF/Excel generation fails
	// Story 5.3, Error Handling: Service error for generation failures
	ErrExportGenerationFailed = "export generation failed"

	// ErrExportFileNotFound is returned when export file is missing
	// Story 5.3, Error Handling: Download attempt for non-existent file
	ErrExportFileNotFound = "export file not found or has expired"

	// ErrExportJobExpired is returned when export job is older than 24 hours
	// Story 5.3, Task 6.2: File cleanup after 24 hours
	ErrExportJobExpired = "export job has expired"

	// ErrInsufficientStorage is returned when disk space is low
	// Story 5.3, Task 6.5: Disk space monitoring
	ErrInsufficientStorage = "insufficient disk space for export generation"

	// ErrReportTooLarge is returned when report exceeds sync export limits
	// Story 5.3, Performance Requirements: Use async for reports >1000 transactions
	ErrReportTooLarge = "report too large for synchronous export, use async export"
)

// Valid export formats
const (
	ExportFormatPDF  = "pdf"
	ExportFormatExcel = "xlsx"
)

// Valid report types for export
const (
	ReportTypeDailySales = "daily_sales"
	ReportTypeProfitLoss = "profit_loss"
)

// Export job statuses
const (
	JobStatusPending    = "pending"
	JobStatusProcessing = "processing"
	JobStatusCompleted  = "completed"
	JobStatusFailed     = "failed"
)
