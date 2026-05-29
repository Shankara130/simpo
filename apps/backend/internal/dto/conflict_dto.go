package dto

import (
	"fmt"
)

// ConflictDetails contains specifics about a stock conflict
type ConflictDetails struct {
	ProductID         uint `json:"product_id"`
	ProductSKU        string `json:"product_sku"`
	RequestedQuantity int `json:"requested_qty"`
	AvailableStock    int64 `json:"available_stock"`
	Shortfall         int `json:"shortfall"`
}

// ConflictErrorResponse is RFC 7807 formatted error response for insufficient stock
// Story 8-5, AC3: Return detailed error response with conflict information
type ConflictErrorResponse struct {
	Type            string          `json:"type"`
	Title           string          `json:"title"`
	Status          int             `json:"status"`
	Detail          string          `json:"detail"`
	Instance        string          `json:"instance"`
	TransactionID   string          `json:"transaction_id"`
	ConflictDetails ConflictDetails `json:"conflict_details"`
}

// BuildConflictErrorResponse creates an RFC 7807 formatted error response for stock conflicts
// Story 8-5, AC3: Error response must include type, title, status, detail, instance, transaction_id, conflict_details
func BuildConflictErrorResponse(details ConflictDetails, transactionID, instance string) ConflictErrorResponse {
	return ConflictErrorResponse{
		Type:          "https://api.simpo.com/errors/conflict-insufficient-stock",
		Title:         "Insufficient Stock",
		Status:        409,
		Detail:        fmt.Sprintf("Product %s has insufficient stock. Requested: %d, Available: %d",
			details.ProductSKU, details.RequestedQuantity, details.AvailableStock),
		Instance:      instance,
		TransactionID: transactionID,
		ConflictDetails: details,
	}
}

// Constants for RFC 7807 error types
const (
	// ErrorTypeInsufficientStock is the URI for insufficient stock errors
	ErrorTypeInsufficientStock = "https://api.simpo.com/errors/conflict-insufficient-stock"
)
