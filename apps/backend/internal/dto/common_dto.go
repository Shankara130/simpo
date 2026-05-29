package dto

// ErrorResponse represents RFC 7807 Problem Details for HTTP APIs
// Used for standardized error responses across all endpoints (Story 9.2, AC7)
// See: https://tools.ietf.org/html/rfc7807
type ErrorResponse struct {
	Type     string `json:"type" example:"https://api.simpo.com/errors/validation-error"`
	Title    string `json:"title" example:"Validation Error"`
	Status   int    `json:"status" example:"400"`
	Detail   string `json:"detail" example:"Invalid request parameters"`
	Instance string `json:"instance" example:"/api/v1/transactions"`
}

// PaginationRequest represents pagination parameters for list endpoints
// Used across all list endpoints for consistent pagination (Story 9.2, AC4)
type PaginationRequest struct {
	Page  int `form:"page" example:"1" binding:"min=1"`
	Limit int `form:"limit" example:"20" binding:"min=1,max=100"`
}

// PaginationResponse represents pagination metadata in list responses
// Returned with all paginated list endpoints (Story 9.2, AC4)
type PaginationResponse struct {
	Page       int   `json:"page" example:"1"`
	Limit      int   `json:"limit" example:"20"`
	Total      int64 `json:"total" example:"150"`
	TotalPages int   `json:"totalPages" example:"8"`
}
