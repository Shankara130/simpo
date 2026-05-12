package whitelist

// AddWhitelistEntryRequest represents a request to add a new whitelist entry
// Story 1.9, AC1: Admin configures email domain whitelist
type AddWhitelistEntryRequest struct {
	// Domain is the email domain to whitelist (e.g., "simpo.pharmacy")
	// Required: true
	// Example: "simpo.pharmacy"
	Domain string `json:"domain" binding:"required" example:"simpo.pharmacy"`

	// DefaultRole is the default role assigned to users registering with this domain
	// Required: true
	// Enum: SYSTEM_ADMIN, OWNER, CASHIER
	// Example: "CASHIER"
	DefaultRole string `json:"default_role" binding:"required,oneof=SYSTEM_ADMIN OWNER CASHIER" example:"CASHIER"`

	// Description is an optional description of the whitelist entry
	// Required: false
	// Example: "Simpo Pharmacy staff domain"
	Description string `json:"description" example:"Simpo Pharmacy staff domain"`
}

// UpdateWhitelistEntryRequest represents a request to update a whitelist entry
// Story 1.9, AC3: Update whitelist entry
type UpdateWhitelistEntryRequest struct {
	// DefaultRole is the default role assigned to users registering with this domain
	// Required: false (updates only provided field)
	// Enum: SYSTEM_ADMIN, OWNER, CASHIER
	// Example: "OWNER"
	DefaultRole string `json:"default_role" binding:"omitempty,oneof=SYSTEM_ADMIN OWNER CASHIER" example:"OWNER"`

	// Description is an optional description of the whitelist entry
	// Required: false
	// Example: "Updated description"
	Description string `json:"description" example:"Updated description"`
}

// WhitelistEntryResponse represents a whitelist entry response
// Story 1.9, AC2: View email domain whitelist
type WhitelistEntryResponse struct {
	// ID is the unique identifier for the whitelist entry
	// Example: 1
	ID uint `json:"id" example:"1"`

	// Domain is the whitelisted email domain
	// Example: "simpo.pharmacy"
	Domain string `json:"domain" example:"simpo.pharmacy"`

	// DefaultRole is the default role for this domain
	// Example: "CASHIER"
	DefaultRole string `json:"default_role" example:"CASHIER"`

	// Description is the description of the whitelist entry
	// Example: "Simpo Pharmacy staff domain"
	Description string `json:"description" example:"Simpo Pharmacy staff domain"`

	// CreatedAt is the timestamp when the entry was created
	// Example: "2026-05-12T00:00:00Z"
	CreatedAt string `json:"created_at" example:"2026-05-12T00:00:00Z"`

	// UpdatedAt is the timestamp when the entry was last updated
	// Example: "2026-05-12T00:00:00Z"
	UpdatedAt string `json:"updated_at" example:"2026-05-12T00:00:00Z"`
}

// ToWhitelistEntryResponse converts WhitelistEntry model to DTO
// Story 1.9, AC2: Response format
func ToWhitelistEntryResponse(entry *WhitelistEntry) WhitelistEntryResponse {
	return WhitelistEntryResponse{
		ID:          entry.ID,
		Domain:      entry.Domain,
		DefaultRole: entry.DefaultRole,
		Description: entry.Description,
		CreatedAt:   entry.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   entry.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}
