package whitelist

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/user"
)

var (
	// ErrInvalidRole is returned when an invalid role is provided
	ErrInvalidRole = errors.New("invalid role for whitelist entry")
	// ErrDomainRequired is returned when domain is empty
	ErrDomainRequired = errors.New("domain is required")
)

// Service defines the interface for whitelist business logic
// Story 1.9, Task 2: Whitelist management service
type Service interface {
	// AddDomain adds a new email domain to the whitelist
	AddDomain(ctx context.Context, req AddWhitelistEntryRequest) (*WhitelistEntry, error)

	// GetDomain retrieves a whitelist entry by ID
	GetDomain(ctx context.Context, id uint) (*WhitelistEntry, error)

	// ListDomains retrieves all whitelist entries
	ListDomains(ctx context.Context) ([]WhitelistEntry, error)

	// UpdateDomain updates a whitelist entry
	UpdateDomain(ctx context.Context, id uint, req UpdateWhitelistEntryRequest) (*WhitelistEntry, error)

	// DeleteDomain removes a whitelist entry
	DeleteDomain(ctx context.Context, id uint) error

	// ValidateDomainWhitelisted checks if a domain is whitelisted
	ValidateDomainWhitelisted(ctx context.Context, domain string) (*WhitelistEntry, error)
}

type service struct {
	repo Repository
}

// NewService creates a new whitelist service
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

// AddDomain adds a new email domain to the whitelist
func (s *service) AddDomain(ctx context.Context, req AddWhitelistEntryRequest) (*WhitelistEntry, error) {
	// Validate domain is not empty
	if req.Domain == "" {
		return nil, ErrDomainRequired
	}

	// Validate domain format
	if !isValidDomainFormat(req.Domain) {
		return nil, fmt.Errorf("invalid domain format: '%s'", req.Domain)
	}

	// Normalize domain to lowercase for case-insensitive matching
	normalizedDomain := strings.ToLower(req.Domain)

	// Validate role
	if !user.IsValidRoleForCreate(req.DefaultRole) {
		return nil, ErrInvalidRole
	}

	// Check if domain already exists (case-insensitive)
	existing, err := s.repo.FindByDomain(ctx, normalizedDomain)
	if err == nil && existing != nil {
		return nil, ErrDomainAlreadyExists
	}

	// Create whitelist entry with normalized domain
	entry := &WhitelistEntry{
		Domain:      normalizedDomain,
		DefaultRole: req.DefaultRole,
		Description: req.Description,
	}

	if err := s.repo.Create(ctx, entry); err != nil {
		return nil, fmt.Errorf("failed to create whitelist entry: %w", err)
	}

	return entry, nil
}

// GetDomain retrieves a whitelist entry by ID
func (s *service) GetDomain(ctx context.Context, id uint) (*WhitelistEntry, error) {
	return s.repo.FindByID(ctx, id)
}

// ListDomains retrieves all whitelist entries
func (s *service) ListDomains(ctx context.Context) ([]WhitelistEntry, error) {
	return s.repo.List(ctx)
}

// UpdateDomain updates a whitelist entry
func (s *service) UpdateDomain(ctx context.Context, id uint, req UpdateWhitelistEntryRequest) (*WhitelistEntry, error) {
	// Find existing entry
	entry, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if req.DefaultRole != "" {
		if !user.IsValidRoleForCreate(req.DefaultRole) {
			return nil, ErrInvalidRole
		}
		entry.DefaultRole = req.DefaultRole
	}

	if req.Description != "" {
		entry.Description = req.Description
	}

	// Save updates
	if err := s.repo.Update(ctx, entry); err != nil {
		return nil, fmt.Errorf("failed to update whitelist entry: %w", err)
	}

	return entry, nil
}

// DeleteDomain removes a whitelist entry
func (s *service) DeleteDomain(ctx context.Context, id uint) error {
	return s.repo.Delete(ctx, id)
}

// ValidateDomainWhitelisted checks if a domain is whitelisted
func (s *service) ValidateDomainWhitelisted(ctx context.Context, domain string) (*WhitelistEntry, error) {
	if domain == "" {
		return nil, ErrDomainRequired
	}

	entry, err := s.repo.FindByDomain(ctx, domain)
	if err != nil {
		if errors.Is(err, ErrWhitelistEntryNotFound) {
			return nil, fmt.Errorf("domain '%s' is not whitelisted", domain)
		}
		return nil, err
	}

	return entry, nil
}

// isValidDomainFormat checks if a domain has a valid format
// Story 1.9: Domain format validation for whitelist entries
func isValidDomainFormat(domain string) bool {
	if domain == "" {
		return false
	}

	// Basic domain format validation
	// Must contain at least one dot, no spaces, valid characters
	parts := strings.Split(domain, ".")
	if len(parts) < 2 {
		return false
	}

	for _, part := range parts {
		if part == "" {
			return false
		}
		// Check for valid hostname characters (alphanumeric and hyphens)
		for _, r := range part {
			if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
				(r >= '0' && r <= '9') || r == '-') {
				return false
			}
		}
		// Part cannot start or end with hyphen
		if strings.HasPrefix(part, "-") || strings.HasSuffix(part, "-") {
			return false
		}
	}

	return true
}
