package user

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestCheckUsernameExists Tests for CheckUsernameExists repository method (Story 1.7)
func TestCheckUsernameExists(t *testing.T) {
	// Setup in-memory database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Run migrations
	err = db.AutoMigrate(&User{})
	require.NoError(t, err)

	repo := NewRepository(db)

	t.Run("Username exists returns true", func(t *testing.T) {
		// Create a test user
		user := &User{
			Username:     "testuser",
			Name:         "Test User",
			Email:        "test@example.com",
			PasswordHash: "hash",
			Role:         RoleCashier,
			Status:       UserStatusActive,
		}
		err := repo.Create(context.Background(), user)
		require.NoError(t, err)

		// Check if username exists
		exists, err := repo.CheckUsernameExists(context.Background(), "testuser")
		assert.NoError(t, err)
		assert.True(t, exists, "Username should exist")
	})

	t.Run("Username does not exist returns false", func(t *testing.T) {
		exists, err := repo.CheckUsernameExists(context.Background(), "nonexistent")
		assert.NoError(t, err)
		assert.False(t, exists, "Username should not exist")
	})

	t.Run("Empty username returns false", func(t *testing.T) {
		exists, err := repo.CheckUsernameExists(context.Background(), "")
		assert.NoError(t, err)
		assert.False(t, exists, "Empty username should return false")
	})
}

// TestCheckEmailExists Tests for CheckEmailExists repository method (Story 1.7)
func TestCheckEmailExists(t *testing.T) {
	// Setup in-memory database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Run migrations
	err = db.AutoMigrate(&User{})
	require.NoError(t, err)

	repo := NewRepository(db)

	t.Run("Email exists returns true", func(t *testing.T) {
		// Create a test user
		user := &User{
			Username:     "testuser2",
			Name:         "Test User 2",
			Email:        "test2@example.com",
			PasswordHash: "hash",
			Role:         RoleCashier,
			Status:       UserStatusActive,
		}
		err := repo.Create(context.Background(), user)
		require.NoError(t, err)

		// Check if email exists
		exists, err := repo.CheckEmailExists(context.Background(), "test2@example.com")
		assert.NoError(t, err)
		assert.True(t, exists, "Email should exist")
	})

	t.Run("Email does not exist returns false", func(t *testing.T) {
		exists, err := repo.CheckEmailExists(context.Background(), "nonexistent@example.com")
		assert.NoError(t, err)
		assert.False(t, exists, "Email should not exist")
	})

	t.Run("Empty email returns false", func(t *testing.T) {
		exists, err := repo.CheckEmailExists(context.Background(), "")
		assert.NoError(t, err)
		assert.False(t, exists, "Empty email should return false")
	})
}
