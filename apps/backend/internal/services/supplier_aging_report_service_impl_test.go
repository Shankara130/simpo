package services

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/dto"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/models"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/repositories"
)

// Mock repositories for testing
type mockPurchaseInvoiceRepository struct {
	invoices []models.PurchaseInvoice
	err     error
}

func (m *mockPurchaseInvoiceRepository) Create(ctx context.Context, invoice *models.PurchaseInvoice, createdBy uint, items []models.PurchaseInvoiceItem) error {
	return nil
}

func (m *mockPurchaseInvoiceRepository) GetByID(ctx context.Context, id uint) (*models.PurchaseInvoice, error) {
	for i := range m.invoices {
		if m.invoices[i].ID == id {
			return &m.invoices[i], nil
		}
	}
	return nil, nil
}

func (m *mockPurchaseInvoiceRepository) GetByInvoiceNumber(ctx context.Context, invoiceNumber string) (*models.PurchaseInvoice, error) {
	return nil, nil
}

func (m *mockPurchaseInvoiceRepository) List(ctx context.Context, filter *repositories.PurchaseInvoiceFilter) ([]*models.PurchaseInvoice, int64, error) {
	if m.err != nil {
		return nil, 0, m.err
	}

	// Filter invoices based on payment status
	var result []*models.PurchaseInvoice
	for _, inv := range m.invoices {
		if filter.PaymentStatus != nil && inv.PaymentStatus == *filter.PaymentStatus {
			result = append(result, &inv)
		}
	}

	return result, int64(len(result)), nil
}

func (m *mockPurchaseInvoiceRepository) Update(ctx context.Context, invoice *models.PurchaseInvoice, updatedBy uint) error {
	return nil
}

func (m *mockPurchaseInvoiceRepository) Delete(ctx context.Context, id uint, deletedBy uint) error {
	return nil
}

func (m *mockPurchaseInvoiceRepository) UpdatePaymentStatus(ctx context.Context, invoiceID uint) error {
	return nil
}

type mockSupplierPaymentRepository struct {
	paymentTotals map[uint]float64
	err          error
}

func (m *mockSupplierPaymentRepository) GetTotalPaidByInvoice(ctx context.Context, invoiceID uint) (float64, error) {
	if m.err != nil {
		return 0, m.err
	}
	if amount, ok := m.paymentTotals[invoiceID]; ok {
		return amount, nil
	}
	return 0, nil
}

func (m *mockSupplierPaymentRepository) Create(ctx context.Context, payment *models.SupplierPayment) error {
	return nil
}

func (m *mockSupplierPaymentRepository) GetByID(ctx context.Context, id uint) (*models.SupplierPayment, error) {
	return nil, nil
}

func (m *mockSupplierPaymentRepository) GetByInvoiceID(ctx context.Context, invoiceID uint) ([]*models.SupplierPayment, error) {
	return nil, nil
}

func (m *mockSupplierPaymentRepository) List(ctx context.Context, filter *repositories.SupplierPaymentFilter) ([]*models.SupplierPayment, int64, error) {
	return nil, 0, nil
}

func (m *mockSupplierPaymentRepository) Update(ctx context.Context, payment *models.SupplierPayment) error {
	return nil
}

func (m *mockSupplierPaymentRepository) Delete(ctx context.Context, id uint) error {
	return nil
}

type mockSupplierRepository struct {
	suppliers map[uint]models.Supplier
	err       error
}

func (m *mockSupplierRepository) Create(ctx context.Context, supplier *models.Supplier, createdBy uint) error {
	return nil
}

func (m *mockSupplierRepository) GetByID(ctx context.Context, id uint) (*models.Supplier, error) {
	if m.err != nil {
		return nil, m.err
	}
	if supplier, ok := m.suppliers[id]; ok {
		return &supplier, nil
	}
	return nil, nil
}

func (m *mockSupplierRepository) GetByName(ctx context.Context, name string) (*models.Supplier, error) {
	return nil, nil
}

func (m *mockSupplierRepository) List(ctx context.Context, filter *repositories.SupplierFilter) ([]*models.Supplier, int64, error) {
	return nil, 0, nil
}

func (m *mockSupplierRepository) Update(ctx context.Context, supplier *models.Supplier, updatedBy uint) error {
	return nil
}

func (m *mockSupplierRepository) Delete(ctx context.Context, id uint) error {
	return nil
}

func (m *mockSupplierRepository) Deactivate(ctx context.Context, id uint, deactivatedBy uint) error {
	return nil
}

type mockAgingReportAuditService struct{}

func (m *mockAgingReportAuditService) LogAction(ctx context.Context, resourceType, resourceID string, action string, details map[string]interface{}) error {
	return nil
}

func (m *mockAgingReportAuditService) LogLoginAttempt(ctx context.Context, entry AuditLogEntry) error {
	return nil
}

func (m *mockAgingReportAuditService) LogAuthorizationFailure(ctx context.Context, entry AuditLogEntry) error {
	return nil
}

func (m *mockAgingReportAuditService) LogUserCreation(ctx context.Context, adminID uint, createdUserID uint, adminUsername string, createdUsername string, ipAddress string) error {
	return nil
}

func (m *mockAgingReportAuditService) LogWhitelistChange(ctx context.Context, adminID uint, adminUsername string, domain string, action models.AuditAction, ipAddress string) error {
	return nil
}

func (m *mockAgingReportAuditService) LogSelfRegistration(ctx context.Context, userID uint, email string, domain string, ipAddress string) error {
	return nil
}

func (m *mockAgingReportAuditService) LogEmailVerification(ctx context.Context, userID uint, email string, ipAddress string) error {
	return nil
}

func (m *mockAgingReportAuditService) LogUserDeactivation(ctx context.Context, adminID uint, deactivatedUserID uint, adminUsername string, deactivatedUsername string, reason string, ipAddress string) error {
	return nil
}

func (m *mockAgingReportAuditService) LogStockAdjustment(ctx context.Context, adminID uint, adminUsername string, productID uint, productSKU string, oldQty int64, newQty int64, reason string, ipAddress string) error {
	return nil
}

func (m *mockAgingReportAuditService) LogBlockedSaleAttempt(ctx context.Context, userID uint, username string, productID uint, productSKU string, productName string, expiryDate string, reason string, ipAddress string) error {
	return nil
}

func (m *mockAgingReportAuditService) LogReportExport(ctx context.Context, userID uint, username string, reportType string, format string, dateRange string, outcome string, ipAddress string) error {
	return nil
}

func (m *mockAgingReportAuditService) LogSettingsUpdate(ctx context.Context, adminID uint, adminUsername string, changesJSON string, ipAddress string) error {
	return nil
}

func (m *mockAgingReportAuditService) LogBackupCreated(ctx context.Context, adminID uint, adminUsername string, backupFile string, size int64, ipAddress string) error {
	return nil
}

func (m *mockAgingReportAuditService) LogBackupRestored(ctx context.Context, adminID uint, adminUsername string, backupFile string, ipAddress string) error {
	return nil
}

func (m *mockAgingReportAuditService) LogBackupDeleted(ctx context.Context, adminID uint, adminUsername string, backupFile string, ipAddress string) error {
	return nil
}

func (m *mockAgingReportAuditService) LogRoleUpdated(ctx context.Context, adminID uint, adminUsername string, targetUserID uint, targetUsername string, oldRole string, newRole string, ipAddress string) error {
	return nil
}

func (m *mockAgingReportAuditService) LogPermissionGranted(ctx context.Context, adminID uint, adminUsername string, targetUserID uint, targetUsername string, permission string, ipAddress string) error {
	return nil
}

func (m *mockAgingReportAuditService) LogPermissionRevoked(ctx context.Context, adminID uint, adminUsername string, targetUserID uint, targetUsername string, permission string, ipAddress string) error {
	return nil
}

func (m *mockAgingReportAuditService) LogBranchCreated(ctx context.Context, adminID uint, adminUsername string, branchName string, branchLocation string, ipAddress string) error {
	return nil
}

func (m *mockAgingReportAuditService) LogBranchUpdated(ctx context.Context, adminID uint, adminUsername string, branchID uint, branchName string, changes string, ipAddress string) error {
	return nil
}

func (m *mockAgingReportAuditService) LogBranchDeactivated(ctx context.Context, adminID uint, adminUsername string, branchID uint, branchName string, reason string, ipAddress string) error {
	return nil
}

func (m *mockAgingReportAuditService) LogSystemStartup(ctx context.Context, systemID string, serverInfo string, ipAddress string) error {
	return nil
}

func (m *mockAgingReportAuditService) LogSystemShutdown(ctx context.Context, systemID string, reason string, ipAddress string) error {
	return nil
}

func (m *mockAgingReportAuditService) LogMaintenanceModeEnabled(ctx context.Context, adminID uint, adminUsername string, reason string, ipAddress string) error {
	return nil
}

func (m *mockAgingReportAuditService) LogMaintenanceModeDisabled(ctx context.Context, adminID uint, adminUsername string, reason string, ipAddress string) error {
	return nil
}

func (m *mockAgingReportAuditService) LogConflictResolution(ctx context.Context, eventType string, transactionID string, originalError string, resolutionType string, resolvedBy string, resolvedAt time.Time, conflictDetails string, ipAddress string) error {
	return nil
}

func (m *mockAgingReportAuditService) ResetMetrics() {}

func (m *mockAgingReportAuditService) Shutdown(ctx context.Context) error {
	return nil
}

// TestGenerateAgingReport_Success tests successful aging report generation
// Story 10.6, Task 2: Service layer generates aging buckets correctly
func TestGenerateAgingReport_Success(t *testing.T) {
	// Setup test data
	ctx := context.Background()
	asOfDate := "2026-05-31"

	// Create test invoices with different aging scenarios
	invoices := []models.PurchaseInvoice{
		{
			ID:            1,
			SupplierID:    1,
			InvoiceNumber: "INV-001",
			InvoiceDate:   time.Date(2026, 5, 15, 0, 0, 0, 0, time.UTC),
			TotalAmount:   5000000,
			PaymentStatus: "unpaid",
		},
		{
			ID:            2,
			SupplierID:    1,
			InvoiceNumber: "INV-002",
			InvoiceDate:   time.Date(2026, 4, 10, 0, 0, 0, 0, time.UTC),
			TotalAmount:   3000000,
			PaymentStatus: "partial",
		},
		{
			ID:            3,
			SupplierID:    2,
			InvoiceNumber: "INV-003",
			InvoiceDate:   time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
			TotalAmount:   4000000,
			PaymentStatus: "unpaid",
		},
	}

	suppliers := map[uint]models.Supplier{
		1: {
			ID:            1,
			Name:          "PT. Pharmasi Jaya",
			ContactPerson: "John Doe",
			Phone:         "+62-21-555-1234",
			Email:         "orders@pharmasi-jaya.co.id",
			Address:       "Jl. Industri No. 123, Jakarta",
		},
		2: {
			ID:            2,
			Name:          "CV. Medika Sehat",
			ContactPerson: "Jane Smith",
			Phone:         "+62-21-555-5678",
			Email:         "sales@medika-sehat.co.id",
			Address:       "Jl. Kesehatan No. 45, Jakarta",
		},
	}

	// Payment totals: invoice 2 has partial payment of 1500000
	paymentTotals := map[uint]float64{
		2: 1500000,
	}

	// Create mocks
	invoiceRepo := &mockPurchaseInvoiceRepository{invoices: invoices}
	paymentRepo := &mockSupplierPaymentRepository{paymentTotals: paymentTotals}
	supplierRepo := &mockSupplierRepository{suppliers: suppliers}
	auditService := &mockAgingReportAuditService{}

	// Create service
	service := NewSupplierAgingReportService(invoiceRepo, paymentRepo, supplierRepo, auditService)

	// Execute
	request := &dto.SupplierAgingReportRequest{
		AsOfDate: asOfDate,
	}

	response, err := service.GenerateAgingReport(ctx, request)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, "2026-05-31", response.AsOfDate)
	assert.Equal(t, "IDR", response.Currency)
	assert.Len(t, response.Suppliers, 2)

	// Verify supplier 1 (Pharmasi Jaya)
	supplier1 := findSupplier(response.Suppliers, 1)
	assert.NotNil(t, supplier1)
	assert.Equal(t, "PT. Pharmasi Jaya", supplier1.SupplierName)
	// Invoice 1 (16 days overdue) + Invoice 2 (partial, 51 days overdue)
	// Total outstanding should be: 5000000 + (3000000 - 1500000) = 6500000
	assert.Equal(t, float64(6500000), supplier1.TotalOutstanding)
	assert.Equal(t, 2, supplier1.InvoiceCount)

	// Verify supplier 2 (Medika Sehat)
	supplier2 := findSupplier(response.Suppliers, 2)
	assert.NotNil(t, supplier2)
	assert.Equal(t, "CV. Medika Sehat", supplier2.SupplierName)
	// Invoice 3 is 91 days overdue (90+ bucket)
	assert.Equal(t, float64(4000000), supplier2.TotalOutstanding)
	assert.Equal(t, 1, supplier2.InvoiceCount)

	// Verify grand totals
	assert.Equal(t, 2, response.GrandTotals.TotalSuppliers)
	assert.Equal(t, 3, response.GrandTotals.TotalInvoices)
	// Total outstanding: 6500000 + 4000000 = 10500000
	assert.Equal(t, float64(10500000), response.GrandTotals.TotalOutstanding)
}

// TestGenerateAgingReport_WithSupplierFilter tests filtering by supplier ID
// Story 10.6, Task 2: Service respects supplier filter
func TestGenerateAgingReport_WithSupplierFilter(t *testing.T) {
	ctx := context.Background()
	supplierID := uint(1)

	invoices := []models.PurchaseInvoice{
		{
			ID:            1,
			SupplierID:    1,
			InvoiceNumber: "INV-001",
			InvoiceDate:   time.Date(2026, 5, 15, 0, 0, 0, 0, time.UTC),
			TotalAmount:   5000000,
			PaymentStatus: "unpaid",
		},
		{
			ID:            2,
			SupplierID:    2,
			InvoiceNumber: "INV-002",
			InvoiceDate:   time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
			TotalAmount:   3000000,
			PaymentStatus: "unpaid",
		},
	}

	suppliers := map[uint]models.Supplier{
		1: {ID: 1, Name: "PT. Pharmasi Jaya"},
		2: {ID: 2, Name: "CV. Medika Sehat"},
	}

	invoiceRepo := &mockPurchaseInvoiceRepository{invoices: invoices}
	paymentRepo := &mockSupplierPaymentRepository{paymentTotals: make(map[uint]float64)}
	supplierRepo := &mockSupplierRepository{suppliers: suppliers}
	auditService := &mockAuditService{}

	service := NewSupplierAgingReportService(invoiceRepo, paymentRepo, supplierRepo, auditService)

	request := &dto.SupplierAgingReportRequest{
		AsOfDate:   "2026-05-31",
		SupplierID: &supplierID,
	}

	response, err := service.GenerateAgingReport(ctx, request)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	// Should only return supplier 1
	assert.Len(t, response.Suppliers, 1)
	assert.Equal(t, "PT. Pharmasi Jaya", response.Suppliers[0].SupplierName)
}

// TestGenerateAgingReport_InvalidDate tests validation of asOfDate parameter
// Story 10.6, Task 2: Service validates date format
func TestGenerateAgingReport_InvalidDate(t *testing.T) {
	ctx := context.Background()

	invoiceRepo := &mockPurchaseInvoiceRepository{invoices: []models.PurchaseInvoice{}}
	paymentRepo := &mockSupplierPaymentRepository{paymentTotals: make(map[uint]float64)}
	supplierRepo := &mockSupplierRepository{suppliers: make(map[uint]models.Supplier)}
	auditService := &mockAuditService{}

	service := NewSupplierAgingReportService(invoiceRepo, paymentRepo, supplierRepo, auditService)

	tests := []struct {
		name     string
		asOfDate string
		wantErr  bool
	}{
		{"Invalid format", "2026/05/31", true},
		{"Empty date", "", true},
		{"Valid date", "2026-05-31", false},
		{"Invalid date", "not-a-date", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := &dto.SupplierAgingReportRequest{AsOfDate: tt.asOfDate}
			_, err := service.GenerateAgingReport(ctx, request)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestGenerateAgingReport_EmptyRequest tests nil request handling
// Story 10.6, Task 2: Service handles nil request gracefully
func TestGenerateAgingReport_EmptyRequest(t *testing.T) {
	ctx := context.Background()

	invoiceRepo := &mockPurchaseInvoiceRepository{invoices: []models.PurchaseInvoice{}}
	paymentRepo := &mockSupplierPaymentRepository{paymentTotals: make(map[uint]float64)}
	supplierRepo := &mockSupplierRepository{suppliers: make(map[uint]models.Supplier)}
	auditService := &mockAuditService{}

	service := NewSupplierAgingReportService(invoiceRepo, paymentRepo, supplierRepo, auditService)

	_, err := service.GenerateAgingReport(ctx, nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "request cannot be nil")
}

// TestCalculateAgingBuckets_Categorization tests aging bucket categorization logic
// Story 10.6, Task 2: Service correctly categorizes invoices into aging buckets
func TestCalculateAgingBuckets_Categorization(t *testing.T) {
	ctx := context.Background()

	// Create invoices with different overdue scenarios
	invoices := []models.PurchaseInvoice{
		{
			ID:            1,
			InvoiceNumber: "INV-001",
			InvoiceDate:   time.Date(2026, 5, 20, 0, 0, 0, 0, time.UTC), // 11 days ago -> Current bucket
			TotalAmount:   1000000,
			PaymentStatus: "unpaid",
		},
		{
			ID:            2,
			InvoiceNumber: "INV-002",
			InvoiceDate:   time.Date(2026, 4, 15, 0, 0, 0, 0, time.UTC), // 46 days ago -> 31-60 bucket
			TotalAmount:   2000000,
			PaymentStatus: "unpaid",
		},
		{
			ID:            3,
			InvoiceNumber: "INV-003",
			InvoiceDate:   time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC), // 72 days ago -> 61-90 bucket
			TotalAmount:   3000000,
			PaymentStatus: "unpaid",
		},
		{
			ID:            4,
			InvoiceNumber: "INV-004",
			InvoiceDate:   time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC), // 119 days ago -> 90+ bucket
			TotalAmount:   4000000,
			PaymentStatus: "unpaid",
		},
	}

	// Mock payment totals (all unpaid, so no payments)
	paymentTotals := map[uint]float64{}

	paymentRepo := &mockSupplierPaymentRepository{paymentTotals: paymentTotals}

	// Create service instance to access the method
	invoiceRepo := &mockPurchaseInvoiceRepository{invoices: invoices}
	supplierRepo := &mockSupplierRepository{suppliers: make(map[uint]models.Supplier)}
	auditService := &mockAuditService{}
	service := NewSupplierAgingReportService(invoiceRepo, paymentRepo, supplierRepo, auditService)

	// Use type assertion to access private method (we'll test the public interface instead)
	// Instead, we'll verify the categorization through the public GenerateAgingReport
	supplier := map[uint]models.Supplier{
		1: {ID: 1, Name: "Test Supplier"},
	}
	supplierRepo.suppliers = supplier
	for _, inv := range invoices {
		inv.SupplierID = 1
	}

	request := &dto.SupplierAgingReportRequest{AsOfDate: "2026-05-31"}
	response, err := service.GenerateAgingReport(ctx, request)

	assert.NoError(t, err)
	assert.Len(t, response.Suppliers, 1)

	aging := response.Suppliers[0].AgingBuckets
	assert.Equal(t, float64(1000000), aging.Current)
	assert.Equal(t, 1, aging.CurrentCount)
	assert.Equal(t, float64(2000000), aging.Days31to60)
	assert.Equal(t, 1, aging.Days31to60Count)
	assert.Equal(t, float64(3000000), aging.Days61to90)
	assert.Equal(t, 1, aging.Days61to90Count)
	assert.Equal(t, float64(4000000), aging.DaysOver90)
	assert.Equal(t, 1, aging.DaysOver90Count)
}

// TestCalculateTotalOutstanding tests total outstanding calculation
// Story 10.6, Task 2: Service correctly sums all aging buckets
func TestCalculateTotalOutstanding(t *testing.T) {
	bucket := dto.AgingBucket{
		Current:    5000000,
		Days31to60: 3000000,
		Days61to90: 2000000,
		DaysOver90: 1000000,
	}

	// We'll test this through the service
	paymentRepo := &mockSupplierPaymentRepository{paymentTotals: make(map[uint]float64)}
	invoiceRepo := &mockPurchaseInvoiceRepository{invoices: []models.PurchaseInvoice{}}
	supplierRepo := &mockSupplierRepository{suppliers: make(map[uint]models.Supplier)}
	auditService := &mockAuditService{}
	service := NewSupplierAgingReportService(invoiceRepo, paymentRepo, supplierRepo, auditService)

	// Create a response with known aging buckets
	response := &dto.SupplierAgingReportResponse{
		Suppliers: []dto.SupplierAgingSummary{
			{
				AgingBuckets: bucket,
			},
		},
	}

	// Verify grand totals calculation
	totals := service.(*supplierAgingReportServiceImpl).calculateGrandTotals(response.Suppliers)
	assert.Equal(t, float64(11000000), totals.TotalOutstanding)
}

// TestExportAgingReportPDF_Placeholder tests PDF export placeholder
// Story 10.6, Task 6: PDF export returns placeholder until full implementation
func TestExportAgingReportPDF_Placeholder(t *testing.T) {
	ctx := context.Background()

	invoiceRepo := &mockPurchaseInvoiceRepository{invoices: []models.PurchaseInvoice{}}
	paymentRepo := &mockSupplierPaymentRepository{paymentTotals: make(map[uint]float64)}
	supplierRepo := &mockSupplierRepository{suppliers: make(map[uint]models.Supplier)}
	auditService := &mockAuditService{}

	service := NewSupplierAgingReportService(invoiceRepo, paymentRepo, supplierRepo, auditService)

	request := &dto.SupplierAgingReportRequest{AsOfDate: "2026-05-31"}
	pdfBytes, filename, err := service.ExportAgingReportPDF(ctx, request)

	assert.NoError(t, err)
	assert.NotEmpty(t, pdfBytes)
	assert.Contains(t, filename, "supplier-aging-report-2026-05-31.pdf")
	assert.Contains(t, filename, ".pdf")
}

// TestExportAgingReportExcel_Placeholder tests Excel export placeholder
// Story 10.6, Task 7: Excel export returns placeholder until full implementation
func TestExportAgingReportExcel_Placeholder(t *testing.T) {
	ctx := context.Background()

	invoiceRepo := &mockPurchaseInvoiceRepository{invoices: []models.PurchaseInvoice{}}
	paymentRepo := &mockSupplierPaymentRepository{paymentTotals: make(map[uint]float64)}
	supplierRepo := &mockSupplierRepository{suppliers: make(map[uint]models.Supplier)}
	auditService := &mockAuditService{}

	service := NewSupplierAgingReportService(invoiceRepo, paymentRepo, supplierRepo, auditService)

	request := &dto.SupplierAgingReportRequest{AsOfDate: "2026-05-31"}
	excelBytes, filename, err := service.ExportAgingReportExcel(ctx, request)

	assert.NoError(t, err)
	assert.NotEmpty(t, excelBytes)
	assert.Contains(t, filename, "supplier-aging-report-2026-05-31.xlsx")
	assert.Contains(t, filename, ".xlsx")
}

// Helper function to find supplier by ID
func findSupplier(suppliers []dto.SupplierAgingSummary, id uint) *dto.SupplierAgingSummary {
	for i := range suppliers {
		if suppliers[i].SupplierID == id {
			return &suppliers[i]
		}
	}
	return nil
}
