package services

import (
	"context"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/dto"
)

// ExportService defines the interface for report export operations
// Story 5.3, Task 1.1-1.2, AC1: Export service interface for PDF and Excel generation
type ExportService interface {
	// ExportDailySalesToPDF exports daily sales report as PDF
	// Story 5.3, Task 1.2, Task 2, AC1, AC2: Generate PDF from daily sales data
	// Returns ExportResponse with PDF file data and metadata
	ExportDailySalesToPDF(ctx context.Context, req *dto.ExportRequest) (*dto.ExportResponse, error)

	// ExportDailySalesToExcel exports daily sales report as Excel
	// Story 5.3, Task 1.2, Task 3, AC1, AC3: Generate Excel from daily sales data
	// Returns ExportResponse with Excel file data and metadata
	ExportDailySalesToExcel(ctx context.Context, req *dto.ExportRequest) (*dto.ExportResponse, error)

	// ExportProfitLossToPDF exports profit/loss report as PDF
	// Story 5.3, Task 1.2, Task 2, AC1, AC2: Generate PDF from profit/loss data
	// Returns ExportResponse with PDF file data and metadata
	ExportProfitLossToPDF(ctx context.Context, req *dto.ExportRequest) (*dto.ExportResponse, error)

	// ExportProfitLossToExcel exports profit/loss report as Excel
	// Story 5.3, Task 1.2, Task 3, AC1, AC3: Generate Excel from profit/loss data
	// Returns ExportResponse with Excel file data and metadata
	ExportProfitLossToExcel(ctx context.Context, req *dto.ExportRequest) (*dto.ExportResponse, error)

	// CreateAsyncExportJob creates an async export job for large reports
	// Story 5.3, Task 5, AC1, AC3: Create export job for reports >1000 transactions
	// Returns job ID for status tracking
	CreateAsyncExportJob(ctx context.Context, req *dto.ExportRequest) (string, error)

	// GetExportJobStatus retrieves the status of an async export job
	// Story 5.3, Task 5.4, AC1: Poll job status during async export
	GetExportJobStatus(ctx context.Context, jobID string) (*dto.ExportJobStatusResponse, error)

	// GetExportJobFile retrieves the generated file from a completed export job
	// Story 5.3, Task 5.5, AC1: Download completed export file
	GetExportJobFile(ctx context.Context, jobID string) (*dto.ExportResponse, error)

	// CleanupExpiredJobs removes export jobs and files older than 24 hours
	// Story 5.3, Task 6.2, AC4: File cleanup for security and disk space management
	CleanupExpiredJobs(ctx context.Context) error
}

// ExportQueueService defines the interface for async export queue management
// Story 5.3, Task 5, AC1: Export queue for processing large reports
type ExportQueueService interface {
	// EnqueueExportJob adds an export job to the processing queue
	EnqueueExportJob(ctx context.Context, job *dto.ExportJob) error

	// ProcessExportQueue processes pending export jobs from the queue
	// Should be run as a background worker/goroutine
	ProcessExportQueue(ctx context.Context) error

	// GetQueueSize returns the current number of jobs in the queue
	GetQueueSize(ctx context.Context) (int, error)
}
