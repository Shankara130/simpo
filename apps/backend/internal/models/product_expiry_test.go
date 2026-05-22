package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestProduct_IsExpired_WithExpiredProduct(t *testing.T) {
	// Product with expiry date in the past
	pastDate := time.Now().Add(-24 * time.Hour) // 1 day ago
	product := Product{
		ID:         1,
		SKU:        "TEST123",
		Name:       "Expired Medicine",
		ExpiryDate: &pastDate,
	}

	// Should be expired
	assert.True(t, product.IsExpired(), "Product with past expiry date should be expired")
}

func TestProduct_IsExpired_WithExpiringToday(t *testing.T) {
	// Product with expiry date equal to today (start of day)
	today := time.Now().Truncate(24 * time.Hour)
	product := Product{
		ID:         2,
		SKU:        "TEST456",
		Name:       "Expiring Today",
		ExpiryDate: &today,
	}

	// Should be expired (equal to today counts as expired)
	assert.True(t, product.IsExpired(), "Product with today's expiry date should be expired")
}

func TestProduct_IsExpired_WithFutureExpiry(t *testing.T) {
	// Product with expiry date in the future
	futureDate := time.Now().Add(30 * 24 * time.Hour) // 30 days from now
	product := Product{
		ID:         3,
		SKU:        "TEST789",
		Name:       "Valid Medicine",
		ExpiryDate: &futureDate,
	}

	// Should NOT be expired
	assert.False(t, product.IsExpired(), "Product with future expiry date should not be expired")
}

func TestProduct_IsExpired_WithNoExpiryDate(t *testing.T) {
	// Product without expiry date set
	product := Product{
		ID:         4,
		SKU:        "TEST000",
		Name:       "No Expiry Date",
		ExpiryDate: nil,
	}

	// Should NOT be expired (nil expiry date means no expiry)
	assert.False(t, product.IsExpired(), "Product without expiry date should not be expired")
}

func TestProduct_IsExpired_JustBeforeExpiry(t *testing.T) {
	// Product expiring in 1 minute - still valid
	almostExpired := time.Now().Add(1 * time.Minute)
	product := Product{
		ID:         5,
		SKU:        "TEST111",
		Name:       "Almost Expired",
		ExpiryDate: &almostExpired,
	}

	// Should NOT be expired (still has 1 minute)
	assert.False(t, product.IsExpired(), "Product expiring in future should not be expired")
}

func TestProduct_IsExpired_JustAfterExpiry(t *testing.T) {
	// Product expired 1 minute ago
	justExpired := time.Now().Add(-1 * time.Minute)
	product := Product{
		ID:         6,
		SKU:        "TEST222",
		Name:       "Just Expired",
		ExpiryDate: &justExpired,
	}

	// Should be expired
	assert.True(t, product.IsExpired(), "Product that just expired should be expired")
}
