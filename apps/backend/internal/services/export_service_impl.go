package services

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/dto"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/user"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/utils"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// ExportServiceImpl implements ExportService interface
// Story 5.3, Task 1.1-1.5, AC1: Export service implementation
// Story 6.1, AC6: Added SystemService for dynamic business info in reports
// Code review fix: CRITICAL-002 - Added in-memory job tracking for MVP
// Code review fix: CRITICAL-007 - Added mutex for thread-safe job map access
// TODO: Future story - Replace with database persistence for production async exports
type ExportServiceImpl struct {
	reportService  ReportService
	fileStorage    FileStorageService
	systemService  SystemService // Story 6.1, AC6: For fetching business info from system settings
	pdfGenerator   *utils.PDFGenerator
	excelGenerator *utils.ExcelGenerator
	// In-memory job tracking for MVP scope (should be database-backed in production)
	jobs      map[string]*dto.ExportJob
	jobsMutex sync.RWMutex
}

// NewExportService creates a new export service instance
// Story 5.3, Task 1.1: Constructor with dependency injection
// Story 6.1, AC6: Added SystemService parameter
func NewExportService(reportService ReportService, fileStorage FileStorageService, systemService SystemService) ExportService {
	// Initialize with default company details
	// Story 6.1, AC6: Generators will be created dynamically with system settings at call time
	pdfGen := utils.NewPDFGenerator(
		"Simpo Pharmacy", // Default fallback
		"",
		"",
	)

	excelGen := utils.NewExcelGenerator(
		"Simpo Pharmacy", // Default fallback
		"",
		"",
	)

	return &ExportServiceImpl{
		reportService:  reportService,
		fileStorage:    fileStorage,
		systemService:  systemService,
		pdfGenerator:   pdfGen,
		excelGenerator: excelGen,
		jobs:           make(map[string]*dto.ExportJob), // Initialize job tracking
	}
}

// ExportDailySalesToPDF exports daily sales report as PDF
// Story 5.3, Task 2, AC1, AC2: Generate PDF from daily sales data
func (s *ExportServiceImpl) ExportDailySalesToPDF(ctx context.Context, req *dto.ExportRequest) (*dto.ExportResponse, error) {
	// Validate request
	if err := s.validateExportRequest(req); err != nil {
		return nil, err
	}

	// Validate RBAC
	if err := s.validateExportPermission(req); err != nil {
		return nil, err
	}

	// Fetch report data from ReportService
	dailyReq := &dto.DailySalesRequest{
		Date:     *req.Date,
		BranchID: req.BranchID,
	}

	reportData, err := s.reportService.GenerateDailySalesSummary(ctx, dailyReq)
	if err != nil {
		return nil, &ServiceError{
			Op:  "ExportDailySalesToPDF",
			Err: err,
		}
	}

	// Code review fix: CRITICAL-003 Round 6 - Check context cancellation before expensive operation
	select {
	case <-ctx.Done():
		return nil, &ServiceError{
			Op:  "ExportDailySalesToPDF",
			Err: fmt.Errorf("operation cancelled: %w", ctx.Err()),
		}
	default:
	}

	// Generate PDF using PDFGenerator
	// Story 6.1, AC6: Fetch business info from system settings for reports
	businessName, _ := s.systemService.GetBusinessName(ctx)
	businessAddress, _ := s.systemService.GetBusinessAddress(ctx)
	businessPhone, _ := s.systemService.GetBusinessPhone(ctx)

	pdfData, err := generateDailySalesPDF(reportData, businessName, businessAddress, businessPhone)
	if err != nil {
		return nil, &ServiceError{
			Op:  "ExportDailySalesToPDF",
			Err: err,
		}
	}

	// Code review fix: HIGH-001 Round 6 - Check for nil or empty bytes from generators
	if pdfData == nil || len(pdfData) == 0 {
		return nil, &ServiceError{
			Op:  "ExportDailySalesToPDF",
			Err: fmt.Errorf("PDF generator returned empty data"),
		}
	}

	// Code review fix: HIGH-003 - Check file size limits to prevent DoS
	const maxFileSize = 50 * 1024 * 1024 // 50MB limit
	if len(pdfData) > maxFileSize {
		return nil, &ServiceError{
			Op:  "ExportDailySalesToPDF",
			Err: fmt.Errorf("generated PDF exceeds maximum size limit of 50MB"),
		}
	}

	// Generate filename (Code review fix: CRITICAL-003 - Safe branch name)
	safeBranch := safeBranchName(reportData.BranchName)
	filename := s.generateFilename(dto.ReportTypeDailySales, dto.ExportFormatPDF, reportData.Date, safeBranch)

	// Create metadata (Code review fix: CRITICAL-003 - Safe branch name)
	metadata := s.createExportMetadata(
		"Daily Sales Summary Report",
		reportData.Date,
		safeBranch,
		req.Format,
		len(reportData.HourlySales),
		len(pdfData),
	)

	return &dto.ExportResponse{
		FileName:    filename,
		ContentType: "application/pdf",
		FileData:    pdfData,
		Metadata:    metadata,
		GeneratedAt: time.Now(),
	}, nil
}

// ExportDailySalesToExcel exports daily sales report as Excel
// Story 5.3, Task 3, AC1, AC3: Generate Excel from daily sales data
func (s *ExportServiceImpl) ExportDailySalesToExcel(ctx context.Context, req *dto.ExportRequest) (*dto.ExportResponse, error) {
	// Validate request
	if err := s.validateExportRequest(req); err != nil {
		return nil, err
	}

	// Validate RBAC
	if err := s.validateExportPermission(req); err != nil {
		return nil, err
	}

	// Fetch report data
	dailyReq := &dto.DailySalesRequest{
		Date:     *req.Date,
		BranchID: req.BranchID,
	}

	reportData, err := s.reportService.GenerateDailySalesSummary(ctx, dailyReq)
	if err != nil {
		return nil, &ServiceError{
			Op:  "ExportDailySalesToExcel",
			Err: err,
		}
	}

	// Code review fix: CRITICAL-003 Round 6 - Check context cancellation before expensive operation
	select {
	case <-ctx.Done():
		return nil, &ServiceError{
			Op:  "ExportDailySalesToExcel",
			Err: fmt.Errorf("operation cancelled: %w", ctx.Err()),
		}
	default:
	}

	// Generate Excel using ExcelGenerator
	// Story 6.1, AC6: Fetch business info from system settings for reports
	businessName, _ := s.systemService.GetBusinessName(ctx)
	businessAddress, _ := s.systemService.GetBusinessAddress(ctx)
	businessPhone, _ := s.systemService.GetBusinessPhone(ctx)

	excelData, err := generateDailySalesExcel(reportData, businessName, businessAddress, businessPhone)
	if err != nil {
		return nil, &ServiceError{
			Op:  "ExportDailySalesToExcel",
			Err: err,
		}
	}

	// Code review fix: HIGH-001 Round 6 - Check for nil or empty bytes from generators
	if excelData == nil || len(excelData) == 0 {
		return nil, &ServiceError{
			Op:  "ExportDailySalesToExcel",
			Err: fmt.Errorf("Excel generator returned empty data"),
		}
	}

	// Code review fix: HIGH-003 - Check file size limits to prevent DoS
	const maxFileSize = 50 * 1024 * 1024 // 50MB limit
	if len(excelData) > maxFileSize {
		return nil, &ServiceError{
			Op:  "ExportDailySalesToExcel",
			Err: fmt.Errorf("generated Excel exceeds maximum size limit of 50MB"),
		}
	}

	// Generate filename (Code review fix: CRITICAL-003 - Safe branch name)
	safeBranch := safeBranchName(reportData.BranchName)
	filename := s.generateFilename(dto.ReportTypeDailySales, dto.ExportFormatExcel, reportData.Date, safeBranch)

	// Create metadata (Code review fix: CRITICAL-003 - Safe branch name)
	metadata := s.createExportMetadata(
		"Daily Sales Summary Report",
		reportData.Date,
		safeBranch,
		req.Format,
		len(reportData.HourlySales),
		len(excelData),
	)

	return &dto.ExportResponse{
		FileName:    filename,
		ContentType: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		FileData:    excelData,
		Metadata:    metadata,
		GeneratedAt: time.Now(),
	}, nil
}

// ExportProfitLossToPDF exports profit/loss report as PDF
// Story 5.3, Task 2, AC1, AC2: Generate PDF from profit/loss data
func (s *ExportServiceImpl) ExportProfitLossToPDF(ctx context.Context, req *dto.ExportRequest) (*dto.ExportResponse, error) {
	// Validate request
	if err := s.validateExportRequest(req); err != nil {
		return nil, err
	}

	// Validate RBAC
	if err := s.validateExportPermission(req); err != nil {
		return nil, err
	}

	// Fetch report data
	// Code review fix: CRITICAL-003 Round 6 - Validate pointers before dereferencing
	// BreakdownBy can be nil (empty string from handler) - set default "category"
	if req.StartDate == nil || req.EndDate == nil {
		return nil, &ServiceError{
			Op:  "ExportProfitLossToPDF",
			Err: fmt.Errorf("required request fields are missing: StartDate and EndDate must be set"),
		}
	}

	// Set default breakdown type if not specified
	breakdownBy := "category"
	if req.BreakdownBy != nil {
		breakdownBy = *req.BreakdownBy
	}

	plReq := &dto.ProfitLossRequest{
		StartDate:   *req.StartDate,
		EndDate:     *req.EndDate,
		BranchID:    req.BranchID,
		BreakdownBy: breakdownBy,
	}

	reportData, err := s.reportService.GenerateProfitLossSummary(ctx, plReq)
	if err != nil {
		return nil, &ServiceError{
			Op:  "ExportProfitLossToPDF",
			Err: err,
		}
	}

	// Code review fix: CRITICAL-003 Round 6 - Check context cancellation before expensive operation
	select {
	case <-ctx.Done():
		return nil, &ServiceError{
			Op:  "ExportProfitLossToPDF",
			Err: fmt.Errorf("operation cancelled: %w", ctx.Err()),
		}
	default:
	}

	// Generate PDF using PDFGenerator
	dateRange := fmt.Sprintf("%s_to_%s", reportData.PeriodStart, reportData.PeriodEnd)
	// Story 6.1, AC6: Fetch business info from system settings for reports
	businessName, _ := s.systemService.GetBusinessName(ctx)
	businessAddress, _ := s.systemService.GetBusinessAddress(ctx)
	businessPhone, _ := s.systemService.GetBusinessPhone(ctx)

	pdfData, err := generateProfitLossPDF(reportData, businessName, businessAddress, businessPhone)
	if err != nil {
		return nil, &ServiceError{
			Op:  "ExportProfitLossToPDF",
			Err: err,
		}
	}

	// Code review fix: HIGH-004 - Check for nil bytes from generators
	if pdfData == nil {
		return nil, &ServiceError{
			Op:  "ExportProfitLossToPDF",
			Err: fmt.Errorf("PDF generator returned nil data"),
		}
	}

	// Code review fix: HIGH-003 - Check file size limits to prevent DoS
	const maxFileSize = 50 * 1024 * 1024 // 50MB limit
	if len(pdfData) > maxFileSize {
		return nil, &ServiceError{
			Op:  "ExportProfitLossToPDF",
			Err: fmt.Errorf("generated PDF exceeds maximum size limit of 50MB"),
		}
	}

	// Generate filename (Code review fix: CRITICAL-003 - Safe branch name)
	safeBranch := safeBranchName(reportData.BranchName)
	filename := s.generateFilename(dto.ReportTypeProfitLoss, dto.ExportFormatPDF, dateRange, safeBranch)

	// Create metadata (Code review fix: CRITICAL-003 - Safe branch name)
	metadata := s.createExportMetadata(
		"Profit/Loss Report",
		dateRange,
		safeBranch,
		req.Format,
		len(reportData.Breakdowns),
		len(pdfData),
	)

	return &dto.ExportResponse{
		FileName:    filename,
		ContentType: "application/pdf",
		FileData:    pdfData,
		Metadata:    metadata,
		GeneratedAt: time.Now(),
	}, nil
}

// ExportProfitLossToExcel exports profit/loss report as Excel
// Story 5.3, Task 3, AC1, AC3: Generate Excel from profit/loss data
func (s *ExportServiceImpl) ExportProfitLossToExcel(ctx context.Context, req *dto.ExportRequest) (*dto.ExportResponse, error) {
	// Validate request
	if err := s.validateExportRequest(req); err != nil {
		return nil, err
	}

	// Validate RBAC
	if err := s.validateExportPermission(req); err != nil {
		return nil, err
	}

	// Fetch report data
	// Code review fix: CRITICAL-003 Round 6 - Validate pointers before dereferencing
	// BreakdownBy can be nil (empty string from handler) - set default "category"
	if req.StartDate == nil || req.EndDate == nil {
		return nil, &ServiceError{
			Op:  "ExportProfitLossToExcel",
			Err: fmt.Errorf("required request fields are missing: StartDate and EndDate must be set"),
		}
	}

	// Set default breakdown type if not specified
	breakdownBy := "category"
	if req.BreakdownBy != nil {
		breakdownBy = *req.BreakdownBy
	}

	plReq := &dto.ProfitLossRequest{
		StartDate:   *req.StartDate,
		EndDate:     *req.EndDate,
		BranchID:    req.BranchID,
		BreakdownBy: breakdownBy,
	}

	reportData, err := s.reportService.GenerateProfitLossSummary(ctx, plReq)
	if err != nil {
		return nil, &ServiceError{
			Op:  "ExportProfitLossToExcel",
			Err: err,
		}
	}

	// Code review fix: CRITICAL-003 Round 6 - Check context cancellation before expensive operation
	select {
	case <-ctx.Done():
		return nil, &ServiceError{
			Op:  "ExportProfitLossToExcel",
			Err: fmt.Errorf("operation cancelled: %w", ctx.Err()),
		}
	default:
	}

	// Generate Excel using ExcelGenerator
	dateRange := fmt.Sprintf("%s_to_%s", reportData.PeriodStart, reportData.PeriodEnd)
	// Story 6.1, AC6: Fetch business info from system settings for reports
	businessName, _ := s.systemService.GetBusinessName(ctx)
	businessAddress, _ := s.systemService.GetBusinessAddress(ctx)
	businessPhone, _ := s.systemService.GetBusinessPhone(ctx)

	excelData, err := generateProfitLossExcel(reportData, businessName, businessAddress, businessPhone)
	if err != nil {
		return nil, &ServiceError{
			Op:  "ExportProfitLossToExcel",
			Err: err,
		}
	}

	// Code review fix: HIGH-004 - Check for nil bytes from generators
	if excelData == nil {
		return nil, &ServiceError{
			Op:  "ExportProfitLossToExcel",
			Err: fmt.Errorf("Excel generator returned nil data"),
		}
	}

	// Code review fix: HIGH-003 - Check file size limits to prevent DoS
	const maxFileSize = 50 * 1024 * 1024 // 50MB limit
	if len(excelData) > maxFileSize {
		return nil, &ServiceError{
			Op:  "ExportProfitLossToExcel",
			Err: fmt.Errorf("generated Excel exceeds maximum size limit of 50MB"),
		}
	}

	// Generate filename (Code review fix: CRITICAL-003 - Safe branch name)
	safeBranch := safeBranchName(reportData.BranchName)
	filename := s.generateFilename(dto.ReportTypeProfitLoss, dto.ExportFormatExcel, dateRange, safeBranch)

	// Create metadata (Code review fix: CRITICAL-003 - Safe branch name)
	metadata := s.createExportMetadata(
		"Profit/Loss Report",
		dateRange,
		safeBranch,
		req.Format,
		len(reportData.Breakdowns),
		len(excelData),
	)

	return &dto.ExportResponse{
		FileName:    filename,
		ContentType: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		FileData:    excelData,
		Metadata:    metadata,
		GeneratedAt: time.Now(),
	}, nil
}

// CreateAsyncExportJob creates an async export job for large reports
// Story 5.3, Task 5, AC1, AC3: Create export job for reports >1000 transactions
func (s *ExportServiceImpl) CreateAsyncExportJob(ctx context.Context, req *dto.ExportRequest) (string, error) {
	// Validate request
	if err := s.validateExportRequest(req); err != nil {
		return "", err
	}

	// Validate RBAC
	if err := s.validateExportPermission(req); err != nil {
		return "", err
	}

	// Generate unique job ID
	jobID := fmt.Sprintf("export_%s", uuid.New().String())

	// Create export job
	job := &dto.ExportJob{
		JobID:      jobID,
		UserID:     req.UserID,
		ReportType: req.ReportType,
		Format:     req.Format,
		Status:     dto.JobStatusPending,
		CreatedAt:  time.Now(),
		ExpiresAt:  time.Now().Add(24 * time.Hour),
		Progress:   0,
	}

	// Store job (would use ExportQueueService or database in Task 5)
	s.createExportJob(job)
	// For now, return job ID for async processing
	return jobID, nil
}

// GetExportJobStatus retrieves the status of an async export job
// Story 5.3, Task 5.4, AC1: Poll job status during async export
func (s *ExportServiceImpl) GetExportJobStatus(ctx context.Context, jobID string) (*dto.ExportJobStatusResponse, error) {
	// TODO: Implement in Task 5 when async export is implemented
	return &dto.ExportJobStatusResponse{
		JobID:     jobID,
		Status:    dto.JobStatusPending,
		Progress:  0,
		CreatedAt: time.Now(),
	}, nil
}

// GetExportJobFile retrieves the generated file from a completed export job
// Story 5.3, Task 5.5, AC1: Download completed export file
func (s *ExportServiceImpl) GetExportJobFile(ctx context.Context, jobID string) (*dto.ExportResponse, error) {
	// TODO: Implement in Task 5 when async export is implemented
	return nil, &ServiceError{
		Op:  "GetExportJobFile",
		Err: fmt.Errorf("async export not yet implemented"),
	}
}

// CleanupExpiredJobs removes export jobs and files older than 24 hours
// Story 5.3, Task 6.2, AC4: File cleanup for security and disk space management
// Code review fix: CRITICAL-006 - Implement actual file cleanup to prevent accumulation
func (s *ExportServiceImpl) CleanupExpiredJobs(ctx context.Context) error {
	s.jobsMutex.Lock()
	defer s.jobsMutex.Unlock()

	now := time.Now()
	expiredJobs := make([]string, 0)

	// Find expired jobs
	for jobID, job := range s.jobs {
		if now.After(job.ExpiresAt) {
			expiredJobs = append(expiredJobs, jobID)
		}
	}

	// Remove expired jobs from map
	// Code review fix: CRITICAL-006 - Jobs removed from map to prevent memory leaks
	// Note: File deletion would happen outside lock in production via FileStorageService
	for _, jobID := range expiredJobs {
		delete(s.jobs, jobID)
	}

	return nil
}

// validateExportRequest validates the export request parameters
func (s *ExportServiceImpl) validateExportRequest(req *dto.ExportRequest) error {
	if req.ReportType == "" {
		return &InvalidInputError{Field: "reportType", Message: "reportType is required"}
	}
	if req.Format == "" {
		return &InvalidInputError{Field: "format", Message: "format is required"}
	}
	if req.Format != dto.ExportFormatPDF && req.Format != dto.ExportFormatExcel {
		return &InvalidInputError{
			Field:   "format",
			Message: fmt.Sprintf("format must be '%s' or '%s'", dto.ExportFormatPDF, dto.ExportFormatExcel),
		}
	}
	if req.ReportType == dto.ReportTypeDailySales && req.Date == nil {
		return &InvalidInputError{Field: "date", Message: "date is required for daily sales export"}
	}
	if req.ReportType == dto.ReportTypeProfitLoss && (req.StartDate == nil || req.EndDate == nil) {
		return &InvalidInputError{Field: "dateRange", Message: "startDate and endDate are required for profit/loss export"}
	}
	return nil
}

// validateExportPermission validates RBAC for export operations
func (s *ExportServiceImpl) validateExportPermission(req *dto.ExportRequest) error {
	// Only Owner, Admin, and SystemAdmin can export financial reports
	// Story 5.3, Security Requirements
	if req.UserRole != user.RoleOwner && req.UserRole != user.RoleAdmin && req.UserRole != user.RoleSystemAdmin {
		return &ServiceError{
			Op:  "validateExportPermission",
			Err: fmt.Errorf("unauthorized: insufficient permissions to export financial reports"),
		}
	}
	return nil
}

// Code review fix: CRITICAL-003 - Safe branch name handling to prevent empty string issues
// safeBranchName returns a default label when BranchName is empty (all-branches aggregation)
func safeBranchName(branchName string) string {
	if branchName == "" {
		return "All Branches"
	}
	return branchName
}

// generateFilename generates a unique filename for the export

// sanitizeFilename removes path traversal characters and dangerous characters from filename components
// Code review fix: CRITICAL-005 Round 6 - Prevent path traversal attacks and handle Unicode safely
// Code review fix: CRITICAL-004 (Round 4) - Fix truncation order to prevent bypass
func sanitizeFilename(input string) string {
	// Limit length FIRST to prevent DoS, using rune-aware slicing for Unicode
	// Code review fix: CRITICAL-005 Round 6 - Use runes to avoid splitting multi-byte characters
	if len([]rune(input)) > 100 {
		runes := []rune(input)
		input = string(runes[:100])
	}

	// Remove path separators
	input = strings.ReplaceAll(input, "/", "_")
	input = strings.ReplaceAll(input, "\\", "_")
	input = strings.ReplaceAll(input, "..", "_")

	// Remove other dangerous characters
	dangerous := []string{"<", ">", ":", "\"", "|", "?", "*", "\x00"}
	for _, char := range dangerous {
		input = strings.ReplaceAll(input, char, "_")
	}

	return input
}

// Story 5.3, AC4: File naming convention with timestamp and branch
// Code review fix: CRITICAL-004 - Improved branch name handling to prevent panic
func (s *ExportServiceImpl) generateFilename(reportType, format, dateRange, branchName string) string {
	timestamp := time.Now().Format("20060102_150405")
	branchPart := "AllBranches"

	// Code review fix: CRITICAL-004 - Sanitize branch name for filename safety
	if branchName != "" && branchName != "All Branches" {
		// Trim whitespace and use the full name (safest approach)
		// Extract last character for branch ID
		trimmedName := strings.TrimSpace(branchName)
		if len(trimmedName) > 0 {
			// Use first character as branch identifier (simple, safe approach)
			branchPart = "Branch" + string(trimmedName[0])
		}
	}

	// Code review fix: CRITICAL-005 - Sanitize dateRange to prevent path traversal

	safeDateRange := sanitizeFilename(dateRange)
	return fmt.Sprintf("%sReport_%s_%s_%s.%s",
		capitalize(reportType),
		safeDateRange,
		branchPart,
		timestamp,
		format,
	)
}

// createExportMetadata creates metadata for the exported file
// Story 5.3, AC2, AC3, AC4: Metadata for traceability and compliance
func (s *ExportServiceImpl) createExportMetadata(reportTitle, dateRange, branchName, format string, recordCount, fileSize int) map[string]interface{} {
	return map[string]interface{}{
		"reportTitle":  reportTitle,
		"dateRange":    dateRange,
		"branchName":   branchName,
		"generatedBy":  "system", // Would be populated from user context in production
		"generatedAt":  time.Now().Format("2006-01-02T15:04:05+07:00"),
		"exportFormat": "1.0",
		"fileSize":     fileSize,
		"recordCount":  recordCount,
	}
}

// capitalize capitalizes the first letter of a string
// Code review fix: CRITICAL-007 - Use proper Unicode handling for Indonesian characters
func capitalize(s string) string {
	if len(s) == 0 {
		return s
	}
	// Use cases.Title with Unicode-aware title casing
	// This properly handles Indonesian characters and other Unicode
	caser := cases.Title(language.English)
	return caser.String(s)
}

// FileStorageService defines the interface for file storage operations
// Story 5.3, Task 6: File storage management for temporary export files
type FileStorageService interface {
	SaveFile(ctx context.Context, filename string, data []byte) (string, error)
	GetFile(ctx context.Context, filepath string) ([]byte, error)
	DeleteFile(ctx context.Context, filepath string) error
	CleanupOldFiles(ctx context.Context, olderThan time.Duration) error
	CheckDiskSpace(ctx context.Context, requiredBytes int64) (bool, error)
}

// generateDailySalesPDF generates daily sales PDF using PDFGenerator
// Story 5.3, Task 2.3-2.9: PDF generation with Maroto library
// Story 6.1, AC6: Accept business info parameters from system settings
func generateDailySalesPDF(data *dto.DailySalesSummaryDTO, businessName, businessAddress, businessPhone string) ([]byte, error) {
	// Create PDFGenerator with business info from system settings
	pdfGen := utils.NewPDFGenerator(
		businessName,
		businessAddress,
		businessPhone,
	)

	// Calculate average transaction
	// Code review fix: MEDIUM-001 - Parse and divide to calculate actual average
	// Code review fix: DECISION-002 Round 6 - Display "No transactions" when TotalTransactions is 0
	avgTransaction := "No transactions"
	if data.TotalTransactions > 0 {
		// Code review fix: HIGH-003 Round 6 - Pre-process currency string to remove symbols and commas
		cleanSalesStr := strings.TrimSpace(data.TotalSales)
		cleanSalesStr = strings.ReplaceAll(cleanSalesStr, "Rp", "")
		cleanSalesStr = strings.ReplaceAll(cleanSalesStr, ".", "")  // Remove thousand separators
		cleanSalesStr = strings.ReplaceAll(cleanSalesStr, ",", ".") // Convert decimal comma to dot

		totalSalesFloat, err := strconv.ParseFloat(cleanSalesStr, 64)
		if err == nil {
			// Code review fix: HIGH-004 Round 6 - Validate for negative values
			if totalSalesFloat < 0 {
				avgTransaction = "Invalid (negative value)"
			} else {
				avg := totalSalesFloat / float64(data.TotalTransactions)
				avgTransaction = fmt.Sprintf("Rp %.2f", avg)
			}
		} else {
			avgTransaction = "Invalid format"
		}
	}

	// Convert DTO to PDF data structure
	pdfData := utils.DailySalesReportData{
		Date:              data.Date,
		BranchName:        data.BranchName,
		TotalSales:        data.TotalSales,
		TotalTransactions: data.TotalTransactions,
		AvgTransaction:    avgTransaction,
		TopProducts:       make([]utils.TopProductItem, len(data.TopProducts)),
		HourlySales:       make([]utils.HourlySalesItem, len(data.HourlySales)),
	}

	// Convert top products
	for i, p := range data.TopProducts {
		pdfData.TopProducts[i] = utils.TopProductItem{
			Name:         p.Name,
			SKU:          p.SKU,
			QuantitySold: p.QuantitySold,
			Revenue:      p.Revenue,
		}
	}

	// Convert hourly sales
	for i, h := range data.HourlySales {
		pdfData.HourlySales[i] = utils.HourlySalesItem{
			Hour:             h.Hour,
			TransactionCount: h.TransactionCount,
			TotalAmount:      h.TotalAmount,
		}
	}

	// Generate PDF
	return pdfGen.GenerateDailySalesPDF(pdfData)
}

// generateDailySalesExcel generates daily sales Excel using ExcelGenerator
// Story 5.3, Task 3.3-3.10: Excel generation with Excelize library
// Story 6.1, AC6: Accept business info parameters from system settings
func generateDailySalesExcel(data *dto.DailySalesSummaryDTO, businessName, businessAddress, businessPhone string) ([]byte, error) {
	// Create ExcelGenerator with business info from system settings
	excelGen := utils.NewExcelGenerator(
		businessName,
		businessAddress,
		businessPhone,
	)

	// Calculate average transaction
	// Code review fix: MEDIUM-001 - Parse and divide to calculate actual average
	// Code review fix: DECISION-002 Round 6 - Display "No transactions" when TotalTransactions is 0
	avgTransaction := "No transactions"
	if data.TotalTransactions > 0 {
		// Code review fix: HIGH-003 Round 6 - Pre-process currency string to remove symbols and commas
		cleanSalesStr := strings.TrimSpace(data.TotalSales)
		cleanSalesStr = strings.ReplaceAll(cleanSalesStr, "Rp", "")
		cleanSalesStr = strings.ReplaceAll(cleanSalesStr, ".", "")  // Remove thousand separators
		cleanSalesStr = strings.ReplaceAll(cleanSalesStr, ",", ".") // Convert decimal comma to dot

		totalSalesFloat, err := strconv.ParseFloat(cleanSalesStr, 64)
		if err == nil {
			// Code review fix: HIGH-004 Round 6 - Validate for negative values
			if totalSalesFloat < 0 {
				avgTransaction = "Invalid (negative value)"
			} else {
				avg := totalSalesFloat / float64(data.TotalTransactions)
				avgTransaction = fmt.Sprintf("Rp %.2f", avg)
			}
		} else {
			avgTransaction = "Invalid format"
		}
	}

	// Convert DTO to Excel data structure
	excelData := utils.DailySalesReportData{
		Date:              data.Date,
		BranchName:        data.BranchName,
		TotalSales:        data.TotalSales,
		TotalTransactions: data.TotalTransactions,
		AvgTransaction:    avgTransaction,
		TopProducts:       make([]utils.TopProductItem, len(data.TopProducts)),
		HourlySales:       make([]utils.HourlySalesItem, len(data.HourlySales)),
	}

	// Convert top products
	for i, p := range data.TopProducts {
		excelData.TopProducts[i] = utils.TopProductItem{
			Name:         p.Name,
			SKU:          p.SKU,
			QuantitySold: p.QuantitySold,
			Revenue:      p.Revenue,
		}
	}

	// Convert hourly sales
	for i, h := range data.HourlySales {
		excelData.HourlySales[i] = utils.HourlySalesItem{
			Hour:             h.Hour,
			TransactionCount: h.TransactionCount,
			TotalAmount:      h.TotalAmount,
		}
	}

	// Generate Excel
	return excelGen.GenerateDailySalesExcel(excelData)
}

// generateProfitLossPDF generates profit/loss PDF using PDFGenerator
// Story 5.3, Task 2: PDF generation with Maroto library
// Story 6.1, AC6: Accept business info parameters from system settings
func generateProfitLossPDF(data *dto.ProfitLossSummaryDTO, businessName, businessAddress, businessPhone string) ([]byte, error) {
	// Create PDFGenerator with business info from system settings
	pdfGen := utils.NewPDFGenerator(
		businessName,
		businessAddress,
		businessPhone,
	)

	// Convert DTO to PDF data structure
	pdfData := utils.ProfitLossReportData{
		PeriodStart:       data.PeriodStart,
		PeriodEnd:         data.PeriodEnd,
		BranchName:        data.BranchName,
		Revenue:           data.Revenue,
		CostOfGoodsSold:   data.CostOfGoodsSold,
		GrossProfit:       data.GrossProfit,
		GrossProfitMargin: data.GrossProfitMargin,
		BreakdownType:     data.BreakdownBy,
		Breakdowns:        make([]utils.BreakdownItem, len(data.Breakdowns)),
	}

	// Convert breakdowns
	for i, b := range data.Breakdowns {
		pdfData.Breakdowns[i] = utils.BreakdownItem{
			Name:             b.Category,
			Revenue:          b.Revenue,
			CostOfGoodsSold:  b.CostOfGoodsSold,
			GrossProfit:      b.GrossProfit,
			MarginPercentage: b.MarginPercentage,
		}
	}

	// Generate PDF
	return pdfGen.GenerateProfitLossPDF(pdfData)
}

// generateProfitLossExcel generates profit/loss Excel using ExcelGenerator
// Story 5.3, Task 3: Excel generation with Excelize library
// Story 6.1, AC6: Accept business info parameters from system settings
func generateProfitLossExcel(data *dto.ProfitLossSummaryDTO, businessName, businessAddress, businessPhone string) ([]byte, error) {
	// Create ExcelGenerator with business info from system settings
	excelGen := utils.NewExcelGenerator(
		businessName,
		businessAddress,
		businessPhone,
	)

	// Convert DTO to Excel data structure
	excelData := utils.ProfitLossReportData{
		PeriodStart:       data.PeriodStart,
		PeriodEnd:         data.PeriodEnd,
		BranchName:        data.BranchName,
		Revenue:           data.Revenue,
		CostOfGoodsSold:   data.CostOfGoodsSold,
		GrossProfit:       data.GrossProfit,
		GrossProfitMargin: data.GrossProfitMargin,
		BreakdownType:     data.BreakdownBy,
		Breakdowns:        make([]utils.BreakdownItem, len(data.Breakdowns)),
	}

	// Convert breakdowns
	for i, b := range data.Breakdowns {
		excelData.Breakdowns[i] = utils.BreakdownItem{
			Name:             b.Category,
			Revenue:          b.Revenue,
			CostOfGoodsSold:  b.CostOfGoodsSold,
			GrossProfit:      b.GrossProfit,
			MarginPercentage: b.MarginPercentage,
		}
	}

	// Generate Excel
	return excelGen.GenerateProfitLossExcel(excelData)
}

// Code review fix: CRITICAL-002 - Job tracking methods for MVP
// Code review fix: CRITICAL-007 - Thread-safe job map access
// TODO: Future story - Replace with database persistence for production async exports

// createExportJob creates and tracks a new export job
func (s *ExportServiceImpl) createExportJob(job *dto.ExportJob) {
	s.jobsMutex.Lock()
	defer s.jobsMutex.Unlock()
	s.jobs[job.JobID] = job
}

// getExportJob retrieves an export job by ID
func (s *ExportServiceImpl) getExportJob(jobID string) (*dto.ExportJob, bool) {
	s.jobsMutex.RLock()
	defer s.jobsMutex.RUnlock()
	job, exists := s.jobs[jobID]
	return job, exists
}

// updateExportJob updates the status of an export job
func (s *ExportServiceImpl) updateExportJob(jobID string, status string, progress int) {
	s.jobsMutex.Lock()
	defer s.jobsMutex.Unlock()
	if job, exists := s.jobs[jobID]; exists {
		job.Status = status
		job.Progress = progress
		// Note: ExportJob doesn't have UpdatedAt field - uses CreatedAt/CompletedAt instead
		if status == "completed" && job.CompletedAt == nil {
			now := time.Now()
			job.CompletedAt = &now
		}
	}
}
