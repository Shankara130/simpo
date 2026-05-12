package user

import (
	"time"
)

// EmailVerificationToken represents an email verification token
// Story 1.9, AC6: Email verification token system
type EmailVerificationToken struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Token     string    `gorm:"uniqueIndex;not null" json:"token"`
	Email     string    `gorm:"not null;index" json:"email"`
	ExpiresAt time.Time `gorm:"not null;index" json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

// TableName specifies the table name for EmailVerificationToken model
func (EmailVerificationToken) TableName() string {
	return "email_verification_tokens"
}

// IsExpired checks if the token has expired
func (t *EmailVerificationToken) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}
