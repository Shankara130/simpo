package errors

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// ErrorHandler returns a Gin middleware that handles errors added to the context via c.Error().
// It converts APIError types to appropriate JSON responses and wraps unknown errors as internal server errors.
// Story 1.5, AC4: Returns RFC 7807 compliant error responses.
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			err := c.Errors.Last()
			requestID, _ := c.Get("request_id")
			reqID, _ := requestID.(string)

			if rateLimitErr, ok := err.Err.(*RateLimitError); ok {
				response := Response{
					Success: false,
					Error: &ErrorInfo{
						// RFC 7807 fields
						Type:     getErrorType(rateLimitErr.Code),
						Title:    "Rate Limit Exceeded",
						Status:   rateLimitErr.Status,
						Detail:   rateLimitErr.Message,
						Instance: getRequestPath(c),
						// Additional fields
						Code:       rateLimitErr.Code,
						Details:    rateLimitErr.Details,
						Timestamp:  time.Now(),
						Path:       getRequestPath(c),
						RequestID:  reqID,
						RetryAfter: &rateLimitErr.RetryAfter,
					},
				}
				c.JSON(rateLimitErr.Status, response)
				return
			}

			if apiErr, ok := err.Err.(*APIError); ok {
				response := Response{
					Success: false,
					Error: &ErrorInfo{
						// RFC 7807 fields (Story 1.5, AC4)
						Type:     getErrorType(apiErr.Code),
						Title:    getErrorTitle(apiErr.Status),
						Status:   apiErr.Status,
						Detail:   apiErr.Message,
						Instance: getRequestPath(c),
						// Additional fields
						Code:      apiErr.Code,
						Details:   apiErr.Details,
						Timestamp: time.Now(),
						Path:      getRequestPath(c),
						RequestID: reqID,
					},
				}
				c.JSON(apiErr.Status, response)
				return
			}

			// Unknown error - wrap as internal server error
			response := Response{
				Success: false,
				Error: &ErrorInfo{
					// RFC 7807 fields (Story 1.5, AC4)
					Type:     getErrorType(CodeInternal),
					Title:    "Internal Server Error",
					Status:   http.StatusInternalServerError,
					Detail:   "An unexpected error occurred",
					Instance: getRequestPath(c),
					// Additional fields
					Code:      CodeInternal,
					Details:   err.Err.Error(),
					Timestamp: time.Now(),
					Path:      getRequestPath(c),
					RequestID: reqID,
				},
			}
			c.JSON(http.StatusInternalServerError, response)
		}
	}
}

// getErrorType returns a URI reference for the error type (RFC 7807)
func getErrorType(code string) string {
	baseURL := "https://api.simpo.com/errors"
	return baseURL + "/" + code
}

// getErrorTitle returns a short, human-readable title for the HTTP status
func getErrorTitle(status int) string {
	switch status {
	case http.StatusBadRequest:
		return "Bad Request"
	case http.StatusUnauthorized:
		return "Unauthorized"
	case http.StatusForbidden:
		return "Forbidden"
	case http.StatusNotFound:
		return "Not Found"
	case http.StatusConflict:
		return "Conflict"
	case http.StatusTooManyRequests:
		return "Too Many Requests"
	case http.StatusInternalServerError:
		return "Internal Server Error"
	default:
		return "Error"
	}
}

func getRequestPath(c *gin.Context) string {
	if c.Request == nil || c.Request.URL == nil {
		return ""
	}
	return c.Request.URL.Path
}
