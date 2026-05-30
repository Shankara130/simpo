package dto

import "github.com/vahiiiid/go-rest-api-boilerplate/internal/models"

// Supplier DTOs for API request/response
// Story 10.1: Data transfer objects for supplier management endpoints

// CreateSupplierRequest represents the request payload for creating a new supplier
// Story 10.1, AC1: Validation tags for required fields
type CreateSupplierRequest struct {
	// Name is the supplier name (required, max 200 characters)
	// Example: "PT. Pharmasi Jaya"
	Name string `json:"name" binding:"required,max=200" example:"PT. Pharmasi Jaya"`

	// ContactPerson is the primary contact person name
	// Example: "Budi Santoso"
	ContactPerson string `json:"contactPerson,omitempty" binding:"omitempty,max=100" example:"Budi Santoso"`

	// Phone is the contact phone number (required)
	// Example: "+62-21-1234-5678"
	Phone string `json:"phone" binding:"required" example:"+62-21-1234-5678"`

	// Email is the contact email address
	// Example: "contact@pharmasi.com"
	Email string `json:"email,omitempty" binding:"omitempty,email,max=100" example:"contact@pharmasi.com"`

	// Address is the physical address of the supplier
	// Example: "Jl. Industri No. 123, Jakarta"
	Address string `json:"address,omitempty" binding:"omitempty,max=500" example:"Jl. Industri No. 123, Jakarta"`
}

// UpdateSupplierRequest represents the request payload for updating an existing supplier
// Story 10.1, AC2: All fields optional, reason required for audit trail
type UpdateSupplierRequest struct {
	// Name is the new supplier name
	// Example: "PT. Pharmasi Jaya Updated"
	Name string `json:"name,omitempty" binding:"omitempty,max=200" example:"PT. Pharmasi Jaya Updated"`

	// ContactPerson is the new contact person name
	// Example: "Jane Doe"
	ContactPerson string `json:"contactPerson,omitempty" binding:"omitempty,max=100" example:"Jane Doe"`

	// Phone is the new contact phone number
	// Example: "+62-21-9876-5432"
	Phone string `json:"phone,omitempty" example:"+62-21-9876-5432"`

	// Email is the new contact email address
	// Example: "updated@pharmasi.com"
	Email string `json:"email,omitempty" binding:"omitempty,email,max=100" example:"updated@pharmasi.com"`

	// Address is the new physical address
	// Example: "Jl. Industri No. 456, Jakarta"
	Address string `json:"address,omitempty" binding:"omitempty,max=500" example:"Jl. Industri No. 456, Jakarta"`

	// Reason is the reason for the update (required for audit trail)
	// Example: "Updating contact information"
	Reason string `json:"reason" binding:"required,min=5,max=500" example:"Updating contact information"`
}

// DeactivateSupplierRequest represents the request payload for deactivating a supplier
// Story 10.1, AC3: Reason required for audit trail
type DeactivateSupplierRequest struct {
	// Reason is the reason for deactivation (required for audit trail)
	// Example: "Supplier went out of business"
	Reason string `json:"reason" binding:"required,min=5,max=500" example:"Supplier went out of business"`
}

// SupplierResponse represents the response payload for supplier operations
// Story 10.1: Response DTO with supplier details
type SupplierResponse struct {
	// ID is the unique identifier for the supplier
	// Example: 1
	ID uint `json:"id" example:"1"`

	// Name is the supplier name
	// Example: "PT. Pharmasi Jaya"
	Name string `json:"name" example:"PT. Pharmasi Jaya"`

	// ContactPerson is the primary contact person name
	// Example: "Budi Santoso"
	ContactPerson string `json:"contactPerson,omitempty" example:"Budi Santoso"`

	// Phone is the contact phone number
	// Example: "+62-21-1234-5678"
	Phone string `json:"phone" example:"+62-21-1234-5678"`

	// Email is the contact email address
	// Example: "contact@pharmasi.com"
	Email string `json:"email,omitempty" example:"contact@pharmasi.com"`

	// Address is the physical address
	// Example: "Jl. Industri No. 123, Jakarta"
	Address string `json:"address,omitempty" example:"Jl. Industri No. 123, Jakarta"`

	// IsActive indicates if the supplier is active
	// Example: true
	IsActive bool `json:"isActive" example:"true"`

	// CreatedAt is the timestamp when the supplier was created
	// Example: "2026-05-30T10:00:00Z"
	CreatedAt string `json:"createdAt" example:"2026-05-30T10:00:00Z"`

	// UpdatedAt is the timestamp when the supplier was last updated
	// Example: "2026-05-30T10:00:00Z"
	UpdatedAt string `json:"updatedAt" example:"2026-05-30T10:00:00Z"`
}

// SupplierListResponse represents the paginated response for supplier listing
// Story 10.1, AC2: Response DTO with pagination metadata
type SupplierListResponse struct {
	// Data is the list of suppliers
	Data []SupplierResponse `json:"data"`

	// Pagination contains pagination metadata
	Pagination PaginationResponse `json:"pagination"`
}

// ToSupplierResponse converts a Supplier model to SupplierResponse DTO
// Story 10.1: Helper function for response conversion
func ToSupplierResponse(supplier *models.Supplier) SupplierResponse {
	return SupplierResponse{
		ID:           supplier.ID,
		Name:         supplier.Name,
		ContactPerson: supplier.ContactPerson,
		Phone:        supplier.Phone,
		Email:        supplier.Email,
		Address:      supplier.Address,
		IsActive:     supplier.IsActive,
		CreatedAt:    supplier.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:    supplier.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// ToSupplierListResponse converts a list of Supplier models to SupplierListResponse DTO
// Story 10.1: Helper function for paginated list response conversion
func ToSupplierListResponse(suppliers []*models.Supplier, total int64, page, limit int) SupplierListResponse {
	data := make([]SupplierResponse, len(suppliers))
	for i, supplier := range suppliers {
		data[i] = ToSupplierResponse(supplier)
	}

	totalPages := int(total) / limit
	if int(total)%limit != 0 {
		totalPages++
	}

	return SupplierListResponse{
		Data: data,
		Pagination: PaginationResponse{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
		},
	}
}
