package whitelist

import (
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"
)

var (
	// ErrDomainAlreadyExists is returned when attempting to add a duplicate domain
	ErrDomainAlreadyExists = errors.New("domain already exists in whitelist")
	// ErrWhitelistEntryNotFound is returned when a whitelist entry is not found
	ErrWhitelistEntryNotFound = errors.New("whitelist entry not found")
)

// Repository defines the interface for whitelist data operations
// Story 1.9, Task 1: Repository layer for email whitelist
type Repository interface {
	// Create creates a new whitelist entry
	Create(ctx context.Context, entry *WhitelistEntry) error

	// FindByID finds a whitelist entry by ID
	FindByID(ctx context.Context, id uint) (*WhitelistEntry, error)

	// FindByDomain finds a whitelist entry by domain
	FindByDomain(ctx context.Context, domain string) (*WhitelistEntry, error)

	// List retrieves all whitelist entries
	List(ctx context.Context) ([]WhitelistEntry, error)

	// Update updates a whitelist entry
	Update(ctx context.Context, entry *WhitelistEntry) error

	// Delete deletes a whitelist entry by ID
	Delete(ctx context.Context, id uint) error

	// Exists checks if a domain exists in the whitelist
	Exists(ctx context.Context, domain string) (bool, error)
}

type repository struct {
	db *gorm.DB
}

// NewRepository creates a new whitelist repository
func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

// Create creates a new whitelist entry
func (r *repository) Create(ctx context.Context, entry *WhitelistEntry) error {
	result := r.db.WithContext(ctx).Create(entry)
	if result.Error != nil {
		// Check for unique constraint violation
		// SQLite: "UNIQUE constraint failed: email_whitelist.domain"
		// PostgreSQL: duplicate key value violates unique constraint
		errMsg := result.Error.Error()
		if strings.Contains(errMsg, "UNIQUE constraint failed") ||
			strings.Contains(errMsg, "duplicate key") ||
			errors.Is(result.Error, gorm.ErrDuplicatedKey) {
			return ErrDomainAlreadyExists
		}
		return result.Error
	}
	return nil
}

// FindByID finds a whitelist entry by ID
func (r *repository) FindByID(ctx context.Context, id uint) (*WhitelistEntry, error) {
	var entry WhitelistEntry
	result := r.db.WithContext(ctx).First(&entry, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrWhitelistEntryNotFound
		}
		return nil, result.Error
	}
	return &entry, nil
}

// FindByDomain finds a whitelist entry by domain
func (r *repository) FindByDomain(ctx context.Context, domain string) (*WhitelistEntry, error) {
	var entry WhitelistEntry
	result := r.db.WithContext(ctx).Where("domain = ?", domain).First(&entry)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrWhitelistEntryNotFound
		}
		return nil, result.Error
	}
	return &entry, nil
}

// List retrieves all whitelist entries
func (r *repository) List(ctx context.Context) ([]WhitelistEntry, error) {
	var entries []WhitelistEntry
	result := r.db.WithContext(ctx).Find(&entries)
	if result.Error != nil {
		return nil, result.Error
	}
	return entries, nil
}

// Update updates a whitelist entry
func (r *repository) Update(ctx context.Context, entry *WhitelistEntry) error {
	result := r.db.WithContext(ctx).Save(entry)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrWhitelistEntryNotFound
	}
	return nil
}

// Delete deletes a whitelist entry by ID
func (r *repository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&WhitelistEntry{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrWhitelistEntryNotFound
	}
	return nil
}

// Exists checks if a domain exists in the whitelist
func (r *repository) Exists(ctx context.Context, domain string) (bool, error) {
	var count int64
	result := r.db.WithContext(ctx).Model(&WhitelistEntry{}).Where("domain = ?", domain).Count(&count)
	if result.Error != nil {
		return false, result.Error
	}
	return count > 0, nil
}
