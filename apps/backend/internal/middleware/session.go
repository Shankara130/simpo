package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	// SessionTTL is the time-to-live for session data (8 hours per NFR-SEC-002)
	SessionTTL = 8 * time.Hour
)

// SessionInfo represents session data stored in Redis
// Story 1.8, AC2: Track last activity timestamp for each active session
type SessionInfo struct {
	UserID       uint      `json:"user_id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	Role         string    `json:"role"`
	BranchID     *uint     `json:"branch_id,omitempty"`
	TokenID      string    `json:"token_id"`      // JWT ID claim for tracking
	IssuedAt     time.Time `json:"issued_at"`
	LastActivity time.Time `json:"last_activity"`
}

// SessionManager handles session tracking and token blocklist operations
// Story 1.8, Task 1: Create session tracking mechanism (Redis)
type SessionManager struct {
	redisClient *redis.Client
}

// NewSessionManager creates a new session manager with Redis client
func NewSessionManager(redisClient *redis.Client) *SessionManager {
	return &SessionManager{
		redisClient: redisClient,
	}
}

// SaveSession stores session information in Redis
// Story 1.8, Task 1: Store session data with 8-hour TTL
func (sm *SessionManager) SaveSession(ctx context.Context, tokenID string, session SessionInfo) error {
	if sm.redisClient == nil {
		return fmt.Errorf("redis client not available - session tracking requires Redis")
	}

	key := fmt.Sprintf("session:%d:%s", session.UserID, tokenID)
	data, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("failed to marshal session data: %w", err)
	}

	// Store with TTL for automatic cleanup
	return sm.redisClient.Set(ctx, key, data, SessionTTL).Err()
}

// GetSession retrieves session information from Redis
func (sm *SessionManager) GetSession(ctx context.Context, userID uint, tokenID string) (*SessionInfo, error) {
	if sm.redisClient == nil {
		return nil, nil // Redis not available
	}

	key := fmt.Sprintf("session:%d:%s", userID, tokenID)
	data, err := sm.redisClient.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Session not found
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	var session SessionInfo
	if err := json.Unmarshal([]byte(data), &session); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session data: %w", err)
	}

	return &session, nil
}

// UpdateLastActivity updates the last activity timestamp for a session
// Story 1.8, AC2: Last activity is updated on each authenticated API request
// P2 FIX: Use Lua script for atomic read-modify-write operation
func (sm *SessionManager) UpdateLastActivity(ctx context.Context, userID uint, tokenID string) error {
	if sm.redisClient == nil {
		return fmt.Errorf("redis client not available - session tracking requires Redis")
	}

	key := fmt.Sprintf("session:%d:%s", userID, tokenID)
	// Format as RFC3339 for proper JSON unmarshaling of time.Time
	newActivity := time.Now().Format(time.RFC3339)

	// Lua script for atomic update: reads session, updates last_activity, writes back
	// This prevents race conditions where multiple requests could overwrite updates
	luaScript := `
		local data = redis.call("GET", KEYS[1])
		if not data then
			return 0  -- Session doesn't exist
		end

		-- Parse JSON and update last_activity field
		local session = cjson.decode(data)
		session.last_activity = ARGV[1]

		-- Save back with TTL
		local ttl = redis.call("TTL", KEYS[1])
		if ttl > 0 then
			redis.call("SET", KEYS[1], cjson.encode(session), "EX", ttl)
		else
			redis.call("SET", KEYS[1], cjson.encode(session))
		end

		return 1
	`

	// Execute Lua script atomically
	err := sm.redisClient.Eval(ctx, luaScript, []string{key}, newActivity).Err()
	if err != nil {
		return fmt.Errorf("failed to atomically update session activity: %w", err)
	}

	return nil
}

// DeleteSession removes a session from Redis (used for logout)
// Story 1.8, AC5: Logout invalidates the current JWT token immediately
func (sm *SessionManager) DeleteSession(ctx context.Context, userID uint, tokenID string) error {
	if sm.redisClient == nil {
		return nil // Redis not available
	}

	key := fmt.Sprintf("session:%d:%s", userID, tokenID)
	return sm.redisClient.Del(ctx, key).Err()
}

// RevokeToken adds a token to the blocklist
// Story 1.8, Task 5: Add revoked tokens to blocklist on logout
func (sm *SessionManager) RevokeToken(ctx context.Context, tokenID string, ttl time.Duration) error {
	if sm.redisClient == nil {
		return fmt.Errorf("redis client not available - token revocation requires Redis")
	}

	key := fmt.Sprintf("revoked:%s", tokenID)
	// Store with TTL matching remaining token lifetime (auto-cleanup)
	return sm.redisClient.Set(ctx, key, "revoked", ttl).Err()
}

// IsTokenRevoked checks if a token is in the blocklist
// Story 1.8, Task 6: Check token blocklist during token validation
// SECURITY: Returns error when Redis is unavailable - fail closed for security
func (sm *SessionManager) IsTokenRevoked(ctx context.Context, tokenID string) (bool, error) {
	if sm.redisClient == nil {
		// SECURITY: Fail closed - if Redis is unavailable, assume token might be revoked
		// This prevents bypassing revocation during Redis outages
		return false, fmt.Errorf("redis client not available - cannot verify token revocation status")
	}

	key := fmt.Sprintf("revoked:%s", tokenID)
	exists, err := sm.redisClient.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check token revocation: %w", err)
	}

	return exists > 0, nil
}

// GetAllUserSessions retrieves all sessions for a specific user
// Useful for debugging and session management
func (sm *SessionManager) GetAllUserSessions(ctx context.Context, userID uint) ([]SessionInfo, error) {
	if sm.redisClient == nil {
		return nil, nil // Redis not available
	}

	pattern := fmt.Sprintf("session:%d:*", userID)
	keys, err := sm.redisClient.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to find user sessions: %w", err)
	}

	var sessions []SessionInfo
	for _, key := range keys {
		data, err := sm.redisClient.Get(ctx, key).Result()
		if err != nil {
			continue // Skip invalid sessions
		}

		var session SessionInfo
		if err := json.Unmarshal([]byte(data), &session); err != nil {
			continue // Skip unmarshallable sessions
		}

		sessions = append(sessions, session)
	}

	return sessions, nil
}

// RevokeAllUserSessions revokes all sessions for a specific user
// Useful for forced logout (e.g., password change, account suspension)
func (sm *SessionManager) RevokeAllUserSessions(ctx context.Context, userID uint) error {
	if sm.redisClient == nil {
		return nil // Redis not available
	}

	// Get all user sessions
	sessions, err := sm.GetAllUserSessions(ctx, userID)
	if err != nil {
		return err
	}

	// Revoke each token
	for _, session := range sessions {
		// Calculate remaining TTL for this token
		ttl := time.Until(session.IssuedAt.Add(SessionTTL))
		if ttl > 0 {
			if err := sm.RevokeToken(ctx, session.TokenID, ttl); err != nil {
				return err
			}
		}
		// Delete session data
		if err := sm.DeleteSession(ctx, userID, session.TokenID); err != nil {
			return err
		}
	}

	return nil
}

// RevokeAllUserTokens revokes all tokens for a specific user
// Story 1.10: User deactivation - revoke all active tokens
// This is an alias for RevokeAllUserSessions for consistency with service layer naming
func (sm *SessionManager) RevokeAllUserTokens(ctx context.Context, userID uint) error {
	return sm.RevokeAllUserSessions(ctx, userID)
}
