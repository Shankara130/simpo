package whitelist

import (
	"testing"
	"time"
)

// TestWhitelistEntryTableName verifies the table name is correct
func TestWhitelistEntryTableName(t *testing.T) {
	entry := WhitelistEntry{}
	if entry.TableName() != "email_whitelist" {
		t.Errorf("Expected table name 'email_whitelist', got '%s'", entry.TableName())
	}
}

// TestWhitelistEntryFields verifies the struct has all required fields
func TestWhitelistEntryFields(t *testing.T) {
	now := time.Now()
	entry := WhitelistEntry{
		ID:          1,
		Domain:      "simpo.pharmacy",
		DefaultRole: "CASHIER",
		Description: "Test domain",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if entry.ID != 1 {
		t.Errorf("Expected ID 1, got %d", entry.ID)
	}
	if entry.Domain != "simpo.pharmacy" {
		t.Errorf("Expected domain 'simpo.pharmacy', got '%s'", entry.Domain)
	}
	if entry.DefaultRole != "CASHIER" {
		t.Errorf("Expected default role 'CASHIER', got '%s'", entry.DefaultRole)
	}
	if entry.Description != "Test domain" {
		t.Errorf("Expected description 'Test domain', got '%s'", entry.Description)
	}
}

// TestToWhitelistEntryResponse verifies DTO conversion
func TestToWhitelistEntryResponse(t *testing.T) {
	now := time.Date(2026, 5, 12, 0, 0, 0, 0, time.UTC)
	entry := &WhitelistEntry{
		ID:          1,
		Domain:      "simpo.pharmacy",
		DefaultRole: "CASHIER",
		Description: "Test domain",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	response := ToWhitelistEntryResponse(entry)

	if response.ID != 1 {
		t.Errorf("Expected ID 1, got %d", response.ID)
	}
	if response.Domain != "simpo.pharmacy" {
		t.Errorf("Expected domain 'simpo.pharmacy', got '%s'", response.Domain)
	}
	if response.DefaultRole != "CASHIER" {
		t.Errorf("Expected default role 'CASHIER', got '%s'", response.DefaultRole)
	}
	if response.CreatedAt != "2026-05-12T00:00:00Z" {
		t.Errorf("Expected created_at '2026-05-12T00:00:00Z', got '%s'", response.CreatedAt)
	}
	if response.UpdatedAt != "2026-05-12T00:00:00Z" {
		t.Errorf("Expected updated_at '2026-05-12T00:00:00Z', got '%s'", response.UpdatedAt)
	}
}
