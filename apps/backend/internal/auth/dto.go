package auth

import "time"

// Claims represents JWT token claims
// Story 1.8, AC1: Token includes ID (jti claim) for session tracking and revocation
type Claims struct {
	UserID    uint      `json:"user_id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Roles     []string  `json:"roles"`
	TokenID   string    `json:"token_id,omitempty"` // Story 1.8: JWT ID for tracking
	ExpiresAt time.Time `json:"exp,omitempty"`      // Expiration time for TTL calculation (P1 patch)
	IssuedAt  time.Time `json:"iat,omitempty"`      // Issued at time for session tracking
}

// TokenResponse represents token response (deprecated: use TokenPairResponse)
type TokenResponse struct {
	Token string `json:"token"`
}

// TokenPairResponse represents access and refresh token pair response
type TokenPairResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
}

// RefreshTokenRequest represents refresh token request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}
