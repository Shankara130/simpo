package user

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
)

var (
	// ErrVerificationTokenNotFound is returned when a verification token is not found
	ErrVerificationTokenNotFound = errors.New("verification token not found")
	// ErrVerificationTokenExpired is returned when a token has expired
	ErrVerificationTokenExpired = errors.New("verification token has expired")
)

// verificationRepository implements the VerificationRepository interface
type verificationRepository struct {
	db *gorm.DB
}

// NewVerificationRepository creates a new verification repository
func NewVerificationRepository(db *gorm.DB) VerificationRepository {
	return &verificationRepository{db: db}
}

// CreateToken creates a new verification token
func (r *verificationRepository) CreateToken(ctx context.Context, token, email string, expiresAt time.Time) error {
	verificationToken := &EmailVerificationToken{
		Token:     token,
		Email:     email,
		ExpiresAt: expiresAt,
	}

	result := r.db.WithContext(ctx).Create(verificationToken)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// FindByToken finds a verification token by token string
func (r *verificationRepository) FindByToken(ctx context.Context, token string) (*EmailVerificationToken, error) {
	var verificationToken EmailVerificationToken
	result := r.db.WithContext(ctx).Where("token = ?", token).First(&verificationToken)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrVerificationTokenNotFound
		}
		return nil, result.Error
	}
	return &verificationToken, nil
}

// DeleteToken deletes a verification token
func (r *verificationRepository) DeleteToken(ctx context.Context, token string) error {
	result := r.db.WithContext(ctx).Where("token = ?", token).Delete(&EmailVerificationToken{})
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// DeleteExpiredTokens removes all expired tokens
func (r *verificationRepository) DeleteExpiredTokens(ctx context.Context) error {
	result := r.db.WithContext(ctx).Where("expires_at < ?", time.Now()).Delete(&EmailVerificationToken{})
	if result.Error != nil {
		return result.Error
	}
	return nil
}
