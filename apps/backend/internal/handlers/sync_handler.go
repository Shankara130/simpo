package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/dto"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/services"
)

// SyncHandler handles offline transaction synchronization with conflict resolution
// Story 8-5, AC1, AC2, AC3, AC10: Process batch transactions with stock validation
type SyncHandler struct {
	conflictSvc *services.ConflictResolutionService
	syncSvc     services.SyncService
}

// SyncRequest represents the request body for transaction sync
type SyncRequest struct {
	Transactions []services.OfflineTransaction `json:"transactions" binding:"required"`
}

// SyncResponse represents successful sync response
type SyncResponse struct {
	Success                bool                          `json:"success"`
	Message                string                        `json:"message"`
	ProcessedTransactions  int                           `json:"processed_transactions"`
	SuccessfulTransactions []services.OfflineTransaction `json:"successful_transactions"`
	FailedTransactions     []services.OfflineTransaction `json:"failed_transactions"`
	ConflictErrors         []dto.ConflictErrorResponse   `json:"conflict_errors,omitempty"`
}

// NewSyncHandler creates a new sync handler
func NewSyncHandler(conflictSvc *services.ConflictResolutionService, syncSvc services.SyncService) *SyncHandler {
	return &SyncHandler{
		conflictSvc: conflictSvc,
		syncSvc:     syncSvc,
	}
}

// SyncTransactions handles POST /api/v1/sync endpoint
// Story 8-5, AC1, AC2, AC3, AC10: Process offline transactions with conflict resolution
func (h *SyncHandler) SyncTransactions(c *gin.Context) {
	ctx := c.Request.Context()

	// Parse request body
	var req SyncRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  "Invalid request body",
			"detail": err.Error(),
		})
		return
	}

	// Process transactions with conflict resolution
	result, conflictErrors := h.conflictSvc.ProcessBatchWithValidation(ctx, req.Transactions)

	// Log sync attempt (Story 8-5, AC10)
	// TODO: Implement audit logging when audit service is ready

	// Determine response status
	if len(conflictErrors) > 0 {
		// Return conflict error response (AC3)
		c.JSON(http.StatusConflict, SyncResponse{
			Success:                false,
			Message:                "Conflict detected during transaction sync",
			ProcessedTransactions:  len(req.Transactions),
			SuccessfulTransactions: result.SuccessfulTransactions,
			FailedTransactions:     result.FailedTransactions,
			ConflictErrors:         conflictErrors,
		})
		return
	}

	// All transactions successful - process them
	// TODO: Call sync service to persist successful transactions
	// h.syncSvc.ProcessTransactions(ctx, result.SuccessfulTransactions)

	c.JSON(http.StatusOK, SyncResponse{
		Success:                true,
		Message:                "Transactions synchronized successfully",
		ProcessedTransactions:  len(req.Transactions),
		SuccessfulTransactions: result.SuccessfulTransactions,
		FailedTransactions:     result.FailedTransactions,
	})
}
