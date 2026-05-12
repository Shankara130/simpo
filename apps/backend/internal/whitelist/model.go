package whitelist

import "time"

// WhitelistEntry represents an approved email domain for staff self-registration
// Story 1.9: Email domain whitelist for staff registration
type WhitelistEntry struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Domain      string    `gorm:"uniqueIndex;not null" json:"domain"` // e.g., "simpo.pharmacy"
	DefaultRole string    `gorm:"not null;default:CASHIER" json:"default_role"` // SYSTEM_ADMIN, OWNER, CASHIER
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TableName specifies the table name for WhitelistEntry model
func (WhitelistEntry) TableName() string {
	return "email_whitelist"
}
