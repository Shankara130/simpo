package utils

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
)

// LogTransactionEvent logs a transaction-related business event
func LogTransactionEvent(ctx context.Context, logger *slog.Logger, event string, transactionID string, details map[string]interface{}) {
	args := []any{
		slog.String("event_type", event),
		slog.String("transaction_id", transactionID),
	}

	// Add request ID from context if available
	if requestID, ok := getRequestID(ctx); ok {
		args = append(args, slog.String("request_id", requestID))
	}

	// Add additional details
	for key, value := range details {
		args = append(args, slog.Any(key, value))
	}

	logger.Info("transaction.event", args...)
}

// LogStockChangeEvent logs a stock change business event
func LogStockChangeEvent(ctx context.Context, logger *slog.Logger, productID uint, oldQty, newQty int, reason string) {
	args := []any{
		slog.String("event_type", "stock.changed"),
		slog.Int64("product_id", int64(productID)),
		slog.Int("old_qty", oldQty),
		slog.Int("new_qty", newQty),
		slog.Int("qty_change", oldQty-newQty),
		slog.String("reason", reason),
	}

	// Add request ID from context if available
	if requestID, ok := getRequestID(ctx); ok {
		args = append(args, slog.String("request_id", requestID))
	}

	logger.Info("stock.event", args...)
}

// LogUserActionEvent logs a user action for audit
func LogUserActionEvent(ctx context.Context, logger *slog.Logger, userID uint, action string, details map[string]interface{}) {
	args := []any{
		slog.String("event_type", "user.action"),
		slog.Int64("user_id", int64(userID)),
		slog.String("action", action),
	}

	// Add request ID from context if available
	if requestID, ok := getRequestID(ctx); ok {
		args = append(args, slog.String("request_id", requestID))
	}

	// Add additional details
	for key, value := range details {
		args = append(args, slog.Any(key, value))
	}

	logger.Info("user.event", args...)
}

// LogSystemEvent logs a system operation event
func LogSystemEvent(ctx context.Context, logger *slog.Logger, event string, details map[string]interface{}) {
	args := []any{
		slog.String("event_type", event),
	}

	// Add request ID from context if available
	if requestID, ok := getRequestID(ctx); ok {
		args = append(args, slog.String("request_id", requestID))
	}

	// Add additional details
	for key, value := range details {
		args = append(args, slog.Any(key, value))
	}

	logger.Info("system.event", args...)
}

// getRequestID extracts request ID from context if available
func getRequestID(ctx context.Context) (string, bool) {
	// Try common context key patterns
	if requestID, ok := ctx.Value("request_id").(string); ok && requestID != "" {
		return requestID, true
	}

	// Try UUID from context
	if requestID, ok := ctx.Value("X-Request-ID").(string); ok && requestID != "" {
		return requestID, true
	}

	// Generate new request ID if none exists
	requestID := uuid.New().String()
	return requestID, true
}

// LogError logs an error with context
func LogError(ctx context.Context, logger *slog.Logger, operation string, err error, details map[string]interface{}) {
	args := []any{
		slog.String("operation", operation),
		slog.String("error", err.Error()),
	}

	// Add request ID from context if available
	if requestID, ok := getRequestID(ctx); ok {
		args = append(args, slog.String("request_id", requestID))
	}

	// Add additional details
	for key, value := range details {
		args = append(args, slog.Any(key, value))
	}

	logger.Error("operation.error", args...)
}
