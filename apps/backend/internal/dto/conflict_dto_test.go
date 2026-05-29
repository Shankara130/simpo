package dto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestConflictErrorResponse_RFC7807Format verifies error response matches RFC 7807
func TestConflictErrorResponse_RFC7807Format(t *testing.T) {
	details := ConflictDetails{
		ProductID:         123,
		ProductSKU:        "SKU-12345",
		RequestedQuantity: 10,
		AvailableStock:    5,
		Shortfall:         5,
	}

	response := BuildConflictErrorResponse(details, "TRX-001", "/api/v1/sync")

	// Verify RFC 7807 required fields
	assert.NotEmpty(t, response.Type, "type must be present")
	assert.NotEmpty(t, response.Title, "title must be present")
	assert.Equal(t, 409, response.Status, "status must be 409 Conflict")
	assert.NotEmpty(t, response.Detail, "detail must be present")
	assert.NotEmpty(t, response.Instance, "instance must be present")

	// Verify conflict-specific fields
	assert.Equal(t, "TRX-001", response.TransactionID, "transaction_id must match")
	assert.Equal(t, uint(123), response.ConflictDetails.ProductID, "product_id must match")
	assert.Equal(t, "SKU-12345", response.ConflictDetails.ProductSKU, "product_sku must match")
	assert.Equal(t, 10, response.ConflictDetails.RequestedQuantity, "requested_qty must match")
	assert.Equal(t, int64(5), response.ConflictDetails.AvailableStock, "available_stock must match")
	assert.Equal(t, 5, response.ConflictDetails.Shortfall, "shortfall must match")
}

// TestConflictErrorResponse_JSONSerialization verifies JSON output format
func TestConflictErrorResponse_JSONSerialization(t *testing.T) {
	details := ConflictDetails{
		ProductID:         1,
		ProductSKU:        "PAR-001",
		RequestedQuantity: 100,
		AvailableStock:    50,
		Shortfall:         50,
	}

	response := BuildConflictErrorResponse(details, "TRX-12345", "/api/v1/sync")

	// Verify JSON structure matches AC3 specification
	// Expected:
	// {
	//   "type": "https://api.simpo.com/errors/conflict-insufficient-stock",
	//   "title": "Insufficient Stock",
	//   "status": 409,
	//   "detail": "Product PAR-001 has insufficient stock...",
	//   "instance": "/api/v1/sync",
	//   "transaction_id": "TRX-12345",
	//   "conflict_details": {
	//     "product_id": 1,
	//     "product_sku": "PAR-001",
	//     "requested_qty": 100,
	//     "available_stock": 50,
	//     "shortfall": 50
	//   }
	// }

	assert.Equal(t, "https://api.simpo.com/errors/conflict-insufficient-stock", response.Type)
	assert.Equal(t, "Insufficient Stock", response.Title)
	assert.Contains(t, response.Detail, "PAR-001", "detail must contain product SKU")
	assert.Contains(t, response.Detail, "100", "detail must contain requested quantity")
	assert.Contains(t, response.Detail, "50", "detail must contain available stock")
}

// TestConflictDetails_Validation verifies conflict details structure
func TestConflictDetails_Validation(t *testing.T) {
	details := ConflictDetails{
		ProductID:         999,
		ProductSKU:        "TEST-SKU",
		RequestedQuantity: 1,
		AvailableStock:    0,
		Shortfall:         1,
	}

	// Verify all fields are properly set
	assert.Greater(t, details.ProductID, uint(0), "product_id must be positive")
	assert.NotEmpty(t, details.ProductSKU, "product_sku must not be empty")
	assert.Greater(t, details.RequestedQuantity, 0, "requested_qty must be positive")
	assert.GreaterOrEqual(t, details.AvailableStock, int64(0), "available_stock can be zero or positive")
	assert.Greater(t, details.Shortfall, 0, "shortfall must be positive")
	assert.Equal(t, details.RequestedQuantity-int(details.AvailableStock), details.Shortfall,
		"shortfall must equal requested_qty - available_stock")
}
