package dto

// RecordPaymentRequest defines the data required to record a supplier payment
// Story 10.4, AC1: Request DTO for recording payment with validation tags
type RecordPaymentRequest struct {
	PurchaseInvoiceID uint    `json:"purchaseInvoiceId" binding:"required" example:"1"`
	PaymentDate      string  `json:"paymentDate" binding:"required" example:"2026-05-31"` // Format: YYYY-MM-DD
	PaymentAmount    float64 `json:"paymentAmount" binding:"required,gt=0" example:"1500000.00"`
	PaymentMethod    string  `json:"paymentMethod" binding:"required,oneof=cash transfer e-wallet check other" example:"transfer"`
	Notes            string  `json:"notes" binding:"omitempty,max=1000" example:"Payment for May 2026 invoice"`
	ReferenceNumber  string  `json:"referenceNumber" binding:"omitempty,max=100" example:"TRX-20260531-12345"`
}

// SupplierPaymentResponse defines the response structure for supplier payment details
// Story 10.4: Response DTO for supplier payment with invoice information
type SupplierPaymentResponse struct {
	ID                uint    `json:"id" example:"1"`
	PurchaseInvoiceID uint    `json:"purchaseInvoiceId" example:"1"`
	InvoiceNumber     string  `json:"invoiceNumber" example:"INV-2026-001"`
	InvoiceDate       string  `json:"invoiceDate" example:"2026-05-30"`
	PaymentDate       string  `json:"paymentDate" example:"2026-05-31"`
	PaymentAmount     float64 `json:"paymentAmount" example:"1500000.00"`
	PaymentMethod     string  `json:"paymentMethod" example:"transfer"`
	Notes             string  `json:"notes,omitempty" example:"Payment for May 2026 invoice"`
	ReferenceNumber   string  `json:"referenceNumber,omitempty" example:"TRX-20260531-12345"`
	BranchID          uint    `json:"branchId" example:"1"`
	CreatedBy         uint    `json:"createdBy" example:"1"`
	CreatedAt         string  `json:"createdAt" example:"2026-05-31T10:00:00Z"`
	UpdatedAt         string  `json:"updatedAt,omitempty" example:"2026-05-31T10:00:00Z"`
	// Invoice details (from eager loading)
	PaymentStatus     string  `json:"paymentStatus" example:"partial"`     // Overall invoice payment status
	RemainingBalance  float64 `json:"remainingBalance" example:"500000.00"` // Remaining amount to pay
}

// SupplierPaymentListResponse defines the paginated response structure for listing supplier payments
// Story 10.4: Response DTO with pagination support
type SupplierPaymentListResponse struct {
	Data       []*SupplierPaymentResponse `json:"data"`
	Pagination PaginationResponse             `json:"pagination"`
}

// SupplierPaymentFilter defines filtering options for supplier payment queries
// Story 10.4: Filter struct for frontend payment filtering
type SupplierPaymentFilter struct {
	PurchaseInvoiceID *uint   `json:"purchaseInvoiceId,omitempty" example:"1"`
	StartDate          *string `json:"startDate,omitempty" example:"2026-05-01"`
	EndDate            *string `json:"endDate,omitempty" example:"2026-05-31"`
	PaymentMethod     *string `json:"paymentMethod,omitempty" example:"transfer"`
	BranchID           *uint   `json:"branchId,omitempty" example:"1"`
	Page               int    `json:"page,omitempty" example:"1"`
	Limit              int    `json:"limit,omitempty" example:"20"`
	SortBy             string `json:"sortBy,omitempty" example:"payment_date"`
	SortOrder          string `json:"sortOrder,omitempty" example:"desc"`
}

// PaymentHistoryResponse represents a payment in the history with invoice details
// Story 10.4, AC2: Response DTO for payment history with invoice context
type PaymentHistoryResponse struct {
	ID                uint    `json:"id" example:"1"`
	PaymentDate       string  `json:"paymentDate" example:"2026-05-31"`
	PaymentAmount     float64 `json:"paymentAmount" example:"1500000.00"`
	PaymentMethod     string  `json:"paymentMethod" example:"transfer"`
	Notes             string  `json:"notes,omitempty" example:"Payment for May 2026 invoice"`
	ReferenceNumber   string  `json:"referenceNumber,omitempty" example:"TRX-20260531-12345"`
	InvoiceNumber     string  `json:"invoiceNumber" example:"INV-2026-001"`
	InvoiceDate       string  `json:"invoiceDate" example:"2026-05-30"`
	InvoiceTotalAmount float64 `json:"invoiceTotalAmount" example:"1500000.00"`
	RemainingBalance  float64 `json:"remainingBalance" example:"0.00"`
}

// PaymentHistoryListResponse defines the paginated response structure for payment history
// Story 10.4, AC2: Response DTO with pagination for supplier payment history
type PaymentHistoryListResponse struct {
	Data       []*PaymentHistoryResponse `json:"data"`
	Pagination PaginationResponse         `json:"pagination"`
}
