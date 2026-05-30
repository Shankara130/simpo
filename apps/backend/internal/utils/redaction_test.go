package utils

import (
	"testing"
)

// TestRedactor tests basic redaction functionality
func TestRedactor(t *testing.T) {
	redactor := NewRedactor()

	// Test password redaction
	input := map[string]interface{}{
		"username": "testuser",
		"password": "secret123",
		"token":    "abc123xyz",
	}

	result := redactor.Redact(input)

	// Verify username is not redacted
	if result["username"] != "testuser" {
		t.Errorf("Expected username to be unchanged, got: %v", result["username"])
	}

	// Verify password is redacted
	password, ok := result["password"].(string)
	if !ok || password == "" || password == "secret123" {
		t.Errorf("Expected password to be redacted (not equal to original), got: %v", password)
	}

	// Verify token is redacted
	token, ok := result["token"].(string)
	if !ok || token == "" || token == "abc123xyz" {
		t.Errorf("Expected token to be redacted (not equal to original), got: %v", token)
	}
}

// TestRedactorString tests string redaction
func TestRedactorString(t *testing.T) {
	redactor := NewRedactor()

	tests := []struct {
		name     string
		input    string
		wantRedacted bool
	}{
		{"plain text", "hello world", false},
		{"password field", "password=secret123", true},
		{"token in JSON", `{"token":"abc123"}`, true},
		{"Bearer token", "authorization: bearer xyz", true}, // Lowercase for matching
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := redactor.RedactString(tt.input)
			isRedacted := result != tt.input
			if isRedacted != tt.wantRedacted {
				t.Errorf("RedactString(%q) redacted=%v, want %v", tt.input, isRedacted, tt.wantRedacted)
			}
		})
	}
}

// TestRedactorHeader tests header redaction
func TestRedactorHeader(t *testing.T) {
	redactor := NewRedactor()

	tests := []struct {
		name       string
		header     string
		value      string
		wantRedacted bool
	}{
		{"Authorization header", "Authorization", "Bearer secret123", true},
		{"Cookie header", "Cookie", "session=abc123", true}, // Cookie should be redacted
		{"regular header", "Content-Type", "application/json", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := redactor.RedactHeader(tt.header, tt.value)
			isRedacted := result != tt.value && result == "****" || result == "Bearer ****"

			if isRedacted != tt.wantRedacted {
				t.Errorf("RedactHeader(%q, %q) redacted=%v, want %v (got: %q)", tt.header, tt.value, isRedacted, tt.wantRedacted, result)
			}
		})
	}
}

// TestRedactorWithCustomPatterns tests custom redaction patterns
func TestRedactorWithCustomPatterns(t *testing.T) {
	redactor := NewRedactor()
	redactor.AddPattern("api_key")

	input := map[string]interface{}{
		"api_key": "secret_key_123",
		"other":   "value",
	}

	result := redactor.Redact(input)

	apiKey, ok := result["api_key"].(string)
	if !ok || apiKey == "secret_key_123" {
		t.Errorf("Expected api_key to be redacted, got: %v", apiKey)
	}

	if result["other"] != "value" {
		t.Errorf("Expected 'other' field to be unchanged")
	}
}
