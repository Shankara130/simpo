package middleware

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestRedis creates a miniredis instance for testing
func setupTestRedis(t *testing.T) (*miniredis.Miniredis, *redis.Client) {
	t.Helper()

	mr, err := miniredis.Run()
	require.NoError(t, err)

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	return mr, client
}

// TestSessionManager_SaveSession tests saving session data to Redis
func TestSessionManager_SaveSession(t *testing.T) {
	mr, client := setupTestRedis(t)
	defer mr.Close()

	sm := NewSessionManager(client)
	ctx := context.Background()

	tokenID := "test-token-id-123"
	session := SessionInfo{
		UserID:       1,
		Username:     "testuser",
		Email:        "test@example.com",
		Role:         "OWNER",
		BranchID:     nil,
		TokenID:      tokenID,
		IssuedAt:     time.Now(),
		LastActivity: time.Now(),
	}

	err := sm.SaveSession(ctx, tokenID, session)
	assert.NoError(t, err)

	// Verify session was saved
	retrieved, err := sm.GetSession(ctx, session.UserID, tokenID)
	assert.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.Equal(t, session.UserID, retrieved.UserID)
	assert.Equal(t, session.Username, retrieved.Username)
	assert.Equal(t, session.Email, retrieved.Email)
	assert.Equal(t, session.Role, retrieved.Role)
	assert.Equal(t, session.TokenID, retrieved.TokenID)
}

// TestSessionManager_GetSession_NotFound tests retrieving non-existent session
func TestSessionManager_GetSession_NotFound(t *testing.T) {
	mr, client := setupTestRedis(t)
	defer mr.Close()

	sm := NewSessionManager(client)
	ctx := context.Background()

	session, err := sm.GetSession(ctx, 999, "non-existent-token")
	assert.NoError(t, err)
	assert.Nil(t, session)
}

// TestSessionManager_UpdateLastActivity tests updating last activity timestamp
func TestSessionManager_UpdateLastActivity(t *testing.T) {
	mr, client := setupTestRedis(t)
	defer mr.Close()

	sm := NewSessionManager(client)
	ctx := context.Background()

	tokenID := "test-token-id-123"
	originalActivity := time.Now().Add(-1 * time.Hour)
	session := SessionInfo{
		UserID:       1,
		Username:     "testuser",
		Email:        "test@example.com",
		Role:         "OWNER",
		TokenID:      tokenID,
		IssuedAt:     time.Now(),
		LastActivity: originalActivity,
	}

	// Save session
	err := sm.SaveSession(ctx, tokenID, session)
	require.NoError(t, err)

	// Update last activity
	err = sm.UpdateLastActivity(ctx, session.UserID, tokenID)
	assert.NoError(t, err)

	// Verify activity was updated
	retrieved, err := sm.GetSession(ctx, session.UserID, tokenID)
	assert.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.True(t, retrieved.LastActivity.After(originalActivity))
}

// TestSessionManager_DeleteSession tests deleting a session
func TestSessionManager_DeleteSession(t *testing.T) {
	mr, client := setupTestRedis(t)
	defer mr.Close()

	sm := NewSessionManager(client)
	ctx := context.Background()

	tokenID := "test-token-id-123"
	session := SessionInfo{
		UserID:       1,
		Username:     "testuser",
		Email:        "test@example.com",
		Role:         "OWNER",
		TokenID:      tokenID,
		IssuedAt:     time.Now(),
		LastActivity: time.Now(),
	}

	// Save session
	err := sm.SaveSession(ctx, tokenID, session)
	require.NoError(t, err)

	// Delete session
	err = sm.DeleteSession(ctx, session.UserID, tokenID)
	assert.NoError(t, err)

	// Verify session was deleted
	retrieved, err := sm.GetSession(ctx, session.UserID, tokenID)
	assert.NoError(t, err)
	assert.Nil(t, retrieved)
}

// TestSessionManager_RevokeToken tests revoking a token
func TestSessionManager_RevokeToken(t *testing.T) {
	mr, client := setupTestRedis(t)
	defer mr.Close()

	sm := NewSessionManager(client)
	ctx := context.Background()

	tokenID := "test-token-to-revoke"
	ttl := 1 * time.Hour

	err := sm.RevokeToken(ctx, tokenID, ttl)
	assert.NoError(t, err)

	// Verify token is revoked
	revoked, err := sm.IsTokenRevoked(ctx, tokenID)
	assert.NoError(t, err)
	assert.True(t, revoked)
}

// TestSessionManager_IsTokenRevoked_NotRevoked tests checking non-revoked token
func TestSessionManager_IsTokenRevoked_NotRevoked(t *testing.T) {
	mr, client := setupTestRedis(t)
	defer mr.Close()

	sm := NewSessionManager(client)
	ctx := context.Background()

	revoked, err := sm.IsTokenRevoked(ctx, "non-revoked-token")
	assert.NoError(t, err)
	assert.False(t, revoked)
}

// TestSessionManager_RevokeToken_TTL tests that revoked tokens expire after TTL
func TestSessionManager_RevokeToken_TTL(t *testing.T) {
	mr, client := setupTestRedis(t)
	defer mr.Close()

	sm := NewSessionManager(client)
	ctx := context.Background()

	tokenID := "test-token-with-short-ttl"
	ttl := 100 * time.Millisecond

	err := sm.RevokeToken(ctx, tokenID, ttl)
	assert.NoError(t, err)

	// Token should be revoked immediately
	revoked, err := sm.IsTokenRevoked(ctx, tokenID)
	assert.NoError(t, err)
	assert.True(t, revoked)

	// Fast forward time
	mr.FastForward(101 * time.Millisecond)

	// Token should no longer be revoked after TTL
	revoked, err = sm.IsTokenRevoked(ctx, tokenID)
	assert.NoError(t, err)
	assert.False(t, revoked)
}

// TestSessionManager_GetAllUserSessions tests retrieving all sessions for a user
func TestSessionManager_GetAllUserSessions(t *testing.T) {
	mr, client := setupTestRedis(t)
	defer mr.Close()

	sm := NewSessionManager(client)
	ctx := context.Background()

	userID := uint(1)

	// Create multiple sessions for the same user
	session1 := SessionInfo{
		UserID:       userID,
		Username:     "testuser",
		Email:        "test@example.com",
		Role:         "OWNER",
		TokenID:      "token-1",
		IssuedAt:     time.Now(),
		LastActivity: time.Now(),
	}

	session2 := SessionInfo{
		UserID:       userID,
		Username:     "testuser",
		Email:        "test@example.com",
		Role:         "OWNER",
		TokenID:      "token-2",
		IssuedAt:     time.Now(),
		LastActivity: time.Now(),
	}

	err := sm.SaveSession(ctx, session1.TokenID, session1)
	require.NoError(t, err)

	err = sm.SaveSession(ctx, session2.TokenID, session2)
	require.NoError(t, err)

	// Get all user sessions
	sessions, err := sm.GetAllUserSessions(ctx, userID)
	assert.NoError(t, err)
	assert.Len(t, sessions, 2)
}

// TestSessionManager_RevokeAllUserSessions tests revoking all user sessions
func TestSessionManager_RevokeAllUserSessions(t *testing.T) {
	mr, client := setupTestRedis(t)
	defer mr.Close()

	sm := NewSessionManager(client)
	ctx := context.Background()

	userID := uint(1)

	// Create multiple sessions for the same user
	session1 := SessionInfo{
		UserID:       userID,
		Username:     "testuser",
		Email:        "test@example.com",
		Role:         "OWNER",
		TokenID:      "token-1",
		IssuedAt:     time.Now(),
		LastActivity: time.Now(),
	}

	session2 := SessionInfo{
		UserID:       userID,
		Username:     "testuser",
		Email:        "test@example.com",
		Role:         "OWNER",
		TokenID:      "token-2",
		IssuedAt:     time.Now(),
		LastActivity: time.Now(),
	}

	err := sm.SaveSession(ctx, session1.TokenID, session1)
	require.NoError(t, err)

	err = sm.SaveSession(ctx, session2.TokenID, session2)
	require.NoError(t, err)

	// Revoke all sessions
	err = sm.RevokeAllUserSessions(ctx, userID)
	assert.NoError(t, err)

	// Verify all tokens are revoked
	revoked1, err := sm.IsTokenRevoked(ctx, session1.TokenID)
	assert.NoError(t, err)
	assert.True(t, revoked1)

	revoked2, err := sm.IsTokenRevoked(ctx, session2.TokenID)
	assert.NoError(t, err)
	assert.True(t, revoked2)

	// Verify all sessions are deleted
	sessions, err := sm.GetAllUserSessions(ctx, userID)
	assert.NoError(t, err)
	assert.Len(t, sessions, 0)
}


// TestSessionManager_NilRedis tests error handling when Redis is not available
// SECURITY FIX: Operations now fail loudly when Redis is unavailable
func TestSessionManager_NilRedis(t *testing.T) {
	sm := NewSessionManager(nil)
	ctx := context.Background()

	tokenID := "test-token-id"
	session := SessionInfo{
		UserID:       1,
		Username:     "testuser",
		Email:        "test@example.com",
		Role:         "OWNER",
		TokenID:      tokenID,
		IssuedAt:     time.Now(),
		LastActivity: time.Now(),
	}

	// SECURITY FIX: Operations now fail loudly when Redis is unavailable
	err := sm.SaveSession(ctx, tokenID, session)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "redis client not available")

	retrieved, err := sm.GetSession(ctx, session.UserID, tokenID)
	assert.NoError(t, err) // GetSession still returns nil, nil for nil client
	assert.Nil(t, retrieved)

	err = sm.UpdateLastActivity(ctx, session.UserID, tokenID)
	assert.Error(t, err) // UpdateLastActivity will fail because SaveSession fails
	assert.Contains(t, err.Error(), "redis client not available")

	err = sm.DeleteSession(ctx, session.UserID, tokenID)
	assert.NoError(t, err) // DeleteSession still succeeds silently for nil client

	err = sm.RevokeToken(ctx, tokenID, 1*time.Hour)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "redis client not available")

	revoked, err := sm.IsTokenRevoked(ctx, tokenID)
	assert.Error(t, err) // IsTokenRevoked now fails when Redis is unavailable
	assert.Contains(t, err.Error(), "redis client not available")
	assert.False(t, revoked)
}
