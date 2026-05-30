package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"
)

// TestLogTransactionEvent tests transaction event logging
func TestLogTransactionEvent(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := slog.New(slog.NewJSONHandler(buf, &slog.HandlerOptions{Level: slog.LevelInfo}))
	transactionID := "TRX-20240530-0001"

	LogTransactionEvent(context.Background(), logger, "transaction.completed", transactionID, map[string]interface{}{
		"cashier_id": 123,
		"total":      150000,
	})

	logOutput := buf.String()
	if logOutput == "" {
		t.Fatal("Expected log output, got empty string")
	}

	// Parse first log line
	lines := strings.Split(strings.TrimSpace(logOutput), "\n")
	if len(lines) == 0 {
		t.Fatal("Expected at least one log line")
	}

	var logData map[string]interface{}
	if err := json.Unmarshal([]byte(lines[0]), &logData); err != nil {
		t.Fatalf("Failed to parse log JSON: %v", err)
	}

	if logData["msg"] != "transaction.event" {
		t.Errorf("Expected message 'transaction.event', got %v", logData["msg"])
	}

	if logData["transaction_id"] != transactionID {
		t.Errorf("Expected transaction_id %s, got %v", transactionID, logData["transaction_id"])
	}
}

// TestLogStockChangeEvent tests stock change event logging
func TestLogStockChangeEvent(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := slog.New(slog.NewJSONHandler(buf, &slog.HandlerOptions{Level: slog.LevelInfo}))

	LogStockChangeEvent(context.Background(), logger, 456, 10, 5, "sale")

	logOutput := buf.String()
	lines := strings.Split(strings.TrimSpace(logOutput), "\n")
	if len(lines) == 0 {
		t.Fatal("Expected at least one log line")
	}

	var logData map[string]interface{}
	if err := json.Unmarshal([]byte(lines[0]), &logData); err != nil {
		t.Fatalf("Failed to parse log JSON: %v", err)
	}

	if logData["msg"] != "stock.event" {
		t.Errorf("Expected message 'stock.event', got %v", logData["msg"])
	}

	if logData["product_id"] != float64(456) {
		t.Errorf("Expected product_id 456, got %v", logData["product_id"])
	}

	if logData["old_qty"] != float64(10) {
		t.Errorf("Expected old_qty 10, got %v", logData["old_qty"])
	}

	if logData["new_qty"] != float64(5) {
		t.Errorf("Expected new_qty 5, got %v", logData["new_qty"])
	}
}

// TestLogUserActionEvent tests user action event logging
func TestLogUserActionEvent(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := slog.New(slog.NewJSONHandler(buf, &slog.HandlerOptions{Level: slog.LevelInfo}))

	LogUserActionEvent(context.Background(), logger, 789, "user.created", map[string]interface{}{
		"target_user_id": 100,
		"role":          "cashier",
	})

	logOutput := buf.String()
	lines := strings.Split(strings.TrimSpace(logOutput), "\n")
	if len(lines) == 0 {
		t.Fatal("Expected at least one log line")
	}

	var logData map[string]interface{}
	if err := json.Unmarshal([]byte(lines[0]), &logData); err != nil {
		t.Fatalf("Failed to parse log JSON: %v", err)
	}

	if logData["msg"] != "user.event" {
		t.Errorf("Expected message 'user.event', got %v", logData["msg"])
	}

	if logData["user_id"] != float64(789) {
		t.Errorf("Expected user_id 789, got %v", logData["user_id"])
	}

	if logData["action"] != "user.created" {
		t.Errorf("Expected action 'user.created', got %v", logData["action"])
	}
}

// TestLogSystemEvent tests system event logging
func TestLogSystemEvent(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := slog.New(slog.NewJSONHandler(buf, &slog.HandlerOptions{Level: slog.LevelInfo}))

	LogSystemEvent(context.Background(), logger, "backup.completed", map[string]interface{}{
		"backup_file": "/backups/simpo_db_20240530.sql",
		"size_bytes":  1024000,
	})

	logOutput := buf.String()
	lines := strings.Split(strings.TrimSpace(logOutput), "\n")
	if len(lines) == 0 {
		t.Fatal("Expected at least one log line")
	}

	var logData map[string]interface{}
	if err := json.Unmarshal([]byte(lines[0]), &logData); err != nil {
		t.Fatalf("Failed to parse log JSON: %v", err)
	}

	if logData["msg"] != "system.event" {
		t.Errorf("Expected message 'system.event', got %v", logData["msg"])
	}

	if logData["event_type"] != "backup.completed" {
		t.Errorf("Expected event_type 'backup.completed', got %v", logData["event_type"])
	}
}

// TestBusinessEventWithRequestID tests that request IDs are included when available
func TestBusinessEventWithRequestID(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := slog.New(slog.NewJSONHandler(buf, &slog.HandlerOptions{Level: slog.LevelInfo}))
	ctx := context.Background()
	requestID := "test-req-123"

	// Add request ID to context
	ctx = context.WithValue(ctx, "request_id", requestID)

	LogTransactionEvent(ctx, logger, "transaction.completed", "TRX-001", map[string]interface{}{
		"total": 50000,
	})

	logOutput := buf.String()
	if logOutput == "" {
		t.Fatal("Expected log output, got empty string")
	}

	lines := strings.Split(strings.TrimSpace(logOutput), "\n")
	if len(lines) == 0 {
		t.Fatal("Expected at least one log line")
	}

	var logData map[string]interface{}
	if err := json.Unmarshal([]byte(lines[0]), &logData); err != nil {
		t.Fatalf("Failed to parse log JSON: %v", err)
	}

	if logData["request_id"] != requestID {
		t.Errorf("Expected request_id %s, got %v", requestID, logData["request_id"])
	}
}
