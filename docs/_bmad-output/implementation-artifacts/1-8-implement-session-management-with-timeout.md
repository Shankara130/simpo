# Story 1.8: Implement Session Management with Timeout

Status: ready-for-dev

**Epic:** 1 - Authentication & User Management
**Priority:** Foundation (Eighth Story)
**Story Type:** Core Feature

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

---

## Story

**As a** System,
**I want to** automatically terminate user sessions after 8 hours of inactivity for security,
**So that** unauthorized access is prevented if a device is left unattended.

---

## Acceptance Criteria

1. **AC1:** JWT Token Expiration
   - JWT tokens are generated with 8-hour expiration time (from Story 1.5)
   - Token expires automatically after 8 hours from issuance
   - Expired tokens return 401 Unauthorized when used

2. **AC2:** Session Activity Tracking
   - System tracks last activity timestamp for each active session
   - Last activity is updated on each authenticated API request
   - Activity tracking happens transparently in auth middleware

3. **AC3:** Automatic Session Invalidation
   - Sessions are automatically invalidated after 8 hours of inactivity
   - Inactivity is calculated from last activity timestamp
   - Invalidated sessions require user to re-authenticate

4. **AC4:** Token Refresh Capability
   - Active sessions can refresh tokens before expiration
   - Token refresh extends session for another 8 hours
   - Refresh endpoint validates current token before issuing new one

5. **AC5:** Logout Endpoint
   - POST /api/v1/auth/logout endpoint exists
   - Logout invalidates the current JWT token immediately
   - Token is added to blocklist or marked as revoked
   - Returns 200 OK on successful logout

6. **AC6:** Logout Audit Trail
   - Logout actions are logged in audit trail
   - Audit log includes: user_id, timestamp, ip_address, action (LOGOUT)
   - Audit trail is append-only per NFR-SEC-004

7. **AC7:** Error Response Format
   - Expired session returns 401 Unauthorized with RFC 7807 format
   - Error message indicates session has expired
   - Error includes type, title, status, detail, instance fields

8. **AC8:** Session Security
   - Tokens cannot be refreshed after expiration
   - Multiple concurrent sessions per user are allowed (for multi-device)
   - Logout invalidates only the specific session token

---

## Tasks / Subtasks

- [ ] **Task 1: Implement Session Activity Tracking** (AC: 2)
  - [ ] Add last_activity timestamp to JWT token claims
  - [ ] Create session tracking mechanism (Redis or in-memory)
  - [ ] Update last_activity on each authenticated request
  - [ ] Add unit tests for activity tracking

- [ ] **Task 2: Implement Session Invalidation Logic** (AC: 1, 3)
  - [ ] Add token expiration validation in auth middleware
  - [ ] Check token expiration time on each request
  - [ ] Return 401 for expired tokens
  - [ ] Add unit tests for expiration validation

- [ ] **Task 3: Implement Token Refresh Endpoint** (AC: 4)
  - [ ] POST /api/v1/auth/refresh endpoint
  - [ ] Validate current token (not expired)
  - [ ] Generate new token with updated expiration
  - [ ] Update session activity tracking
  - [ ] Return new token to client
  - [ ] Add integration tests for refresh flow

- [ ] **Task 4: Implement Logout Endpoint** (AC: 5, 6)
  - [ ] POST /api/v1/auth/logout endpoint
  - [ ] Invalidate the current JWT token
  - [ ] Log logout action to audit trail
  - [ ] Return 200 OK on success
  - [ ] Add integration tests for logout

- [ ] **Task 5: Implement Token Blocklist** (AC: 5, 8)
  - [ ] Create token blocklist mechanism (Redis recommended)
  - [ ] Add revoked tokens to blocklist on logout
  - [ ] Check blocklist during token validation
  - [ ] Implement blocklist expiration (TTL = token remaining lifetime)
  - [ ] Add unit tests for blocklist operations

- [ ] **Task 6: Update Auth Middleware** (AC: 1, 2, 3, 7)
  - [ ] Modify JWTAuthMiddleware to track session activity
  - [ ] Add token expiration validation
  - [ ] Check token blocklist
  - [ ] Return RFC 7807 formatted errors
  - [ ] Add unit tests for middleware

- [ ] **Task 7: API Documentation** (AC: all)
  - [ ] Add Swagger annotations to refresh endpoint
  - [ ] Add Swagger annotations to logout endpoint
  - [ ] Document request/response schemas
  - [ ] Document error responses (401, 403, 500)

- [ ] **Task 8: Integration Testing** (AC: all)
  - [ ] Test token expiration after 8 hours
  - [ ] Test token refresh before expiration
  - [ ] Test logout invalidates token
  - [ ] Test expired token returns 401
  - [ ] Test revoked token returns 401
  - [ ] Verify audit trail entries

---

## Dev Notes

### Context & Purpose

This is the **eighth foundational story** for simpo. It implements session management with timeout, building on the JWT authentication (Story 1.5) and RBAC (Story 1.6) foundations. This story enforces the 8-hour session timeout required by NFR-SEC-002 and provides logout capabilities for secure session termination.

**Business Context:**
- Pharmacy management requires secure session management (Badan POM compliance)
- 8-hour session timeout prevents unauthorized access from unattended devices (NFR-SEC-002)
- Logout capability allows users to explicitly terminate sessions
- Token refresh enables seamless user experience within active sessions
- Audit trail required for all session terminations

**Technical Context:**
- JWT tokens from Story 1.5 have 8-hour expiration (exp claim)
- Auth middleware from Story 1.6 validates tokens on each request
- Redis is available for session tracking and token blocklist
- AuditService from Story 1.5 logs session terminations
- RFC 7807 error response format is implemented

### Architecture Alignment

**[Source: docs/_bmad-output/planning-artifacts/architecture.md]**

**Session Management Requirements:**
- 8-hour session timeout (NFR-SEC-002)
- JWT token expiration validation
- Session activity tracking
- Token refresh capability
- Logout endpoint with token invalidation
- Token blocklist for revoked tokens

**Redis for Session Management:**
```
Session Storage (Redis):
- Key: "session:{user_id}:{token_id}"
- Value: JSON with user_id, last_activity, issued_at
- TTL: 8 hours (auto-cleanup)

Token Blocklist (Redis):
- Key: "revoked:{token_id}"
- Value: user_id, revoked_at
- TTL: remaining token lifetime (auto-cleanup)
```

**API Endpoints:**
```
POST /api/v1/auth/refresh
Authorization: Bearer <current_token>
Response: {
  "access_token": "<new_token>",
  "token_type": "Bearer",
  "expires_in": 28800
}

POST /api/v1/auth/logout
Authorization: Bearer <current_token>
Response: 200 OK
```

**API Response Format (RFC 7807):**
```json
// 401 Unauthorized (Expired Session)
{
  "type": "https://api.simpo.com/errors/session-expired",
  "title": "Session Expired",
  "status": 401,
  "detail": "Your session has expired. Please log in again.",
  "instance": "/api/v1/products"
}

// 401 Unauthorized (Revoked Token)
{
  "type": "https://api.simpo.com/errors/token-revoked",
  "title": "Token Revoked",
  "status": 401,
  "detail": "This token has been revoked. Please log in again.",
  "instance": "/api/v1/products"
}
```

### Previous Story Intelligence

**From Story 1.5 (Implement User Authentication with JWT):**

**Learnings to Apply:**
- JWT tokens are generated with 8-hour expiration (exp claim)
- JWT middleware validates tokens on each request
- Token structure includes user_id, username, email, role, branch_id, exp, iat
- AuditService is available for logging
- RFC 7807 error response format is implemented

**JWT Token Structure from Story 1.5:**
```go
type Claims struct {
    UserID    uint   `json:"user_id"`
    Username  string `json:"username"`
    Email     string `json:"email"`
    Role      string `json:"role"`
    BranchID  *uint  `json:"branch_id,omitempty"`
    jwt.RegisteredClaims // includes exp, iat
}
```

**JWT Configuration from Story 1.5:**
```bash
JWT_SECRET=simpo_jwt_secret_key_for_pharmacy_management_system_2026_secure_token
JWT_ACCESS_TOKEN_TTL=8h            # 8 hours per NFR-SEC-002
```

**What to Build:**
- Add token refresh endpoint to generate new tokens
- Add logout endpoint to invalidate tokens
- Implement token blocklist (Redis recommended)
- Track session activity in auth middleware
- Validate token expiration on each request
- Use existing AuditService for logging logout actions

**Common Issues from Story 1.5/1.6 Code Review:**
- Ensure audit logging actually writes logs (use slog.Info())
- Extract IP address from Gin context (c.ClientIP())
- Handle nil pointers gracefully (branch_id can be nil)
- Validate all inputs before processing
- Return appropriate error codes (401 for auth failures)

**From Story 1.6 (Implement Role-Based Access Control):**

**Learnings to Apply:**
- RBAC middleware runs after auth middleware
- User context is set by JWTAuthMiddleware
- GetUserID, GetUserRole, GetBranchID helper functions available
- AuditService integration pattern

**Auth Middleware Pattern from Story 1.6:**
```go
// internal/middleware/jwt_auth.go
func JWTAuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 1. Extract Bearer token
        // 2. Verify token signature
        // 3. Check expiration
        // 4. Set user context
        // 5. Call c.Next()
    }
}
```

### Implementation Pattern

**Session Tracking with Redis:**
```go
// Session tracking structure
type SessionInfo struct {
    UserID       uint      `json:"user_id"`
    Username     string    `json:"username"`
    Role         string    `json:"role"`
    BranchID     *uint     `json:"branch_id"`
    IssuedAt     time.Time `json:"issued_at"`
    LastActivity time.Time `json:"last_activity"`
    TokenID      string    `json:"token_id"` // JWT ID claim for tracking
}

// Store session in Redis
func SaveSession(ctx context.Context, tokenID string, session SessionInfo) error {
    key := fmt.Sprintf("session:%d:%s", session.UserID, tokenID)
    data, _ := json.Marshal(session)
    return redis.Set(ctx, key, data, 8*time.Hour)
}

// Update last activity
func UpdateLastActivity(ctx context.Context, tokenID string) error {
    // Get session, update last_activity, save back
}
```

**Token Blocklist with Redis:**
```go
// Add revoked token to blocklist
func RevokeToken(ctx context.Context, tokenID string, expiration time.Duration) error {
    key := fmt.Sprintf("revoked:%s", tokenID)
    return redis.Set(ctx, key, "revoked", expiration)
}

// Check if token is revoked
func IsTokenRevoked(ctx context.Context, tokenID string) (bool, error) {
    key := fmt.Sprintf("revoked:%s", tokenID)
    exists, _ := redis.Exists(ctx, key)
    return exists > 0, nil
}
```

**Auth Middleware with Session Tracking:**
```go
func JWTAuthMiddleware(redisClient *redis.Client) gin.HandlerFunc {
    return func(c *gin.Context) {
        // 1. Extract Bearer token
        tokenString := extractBearerToken(c)
        if tokenString == "" {
            c.JSON(401, gin.H{"type": "https://api.simpo.com/errors/missing-token", ...})
            c.Abort()
            return
        }

        // 2. Parse and validate token
        token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
            return []byte(jwtSecret), nil
        })
        if err != nil || !token.Valid {
            c.JSON(401, gin.H{"type": "https://api.simpo.com/errors/invalid-token", ...})
            c.Abort()
            return
        }

        // 3. Extract claims
        claims := token.Claims.(jwt.MapClaims)
        tokenID := claims["jti"].(string)  // JWT ID for tracking
        userID := claims["user_id"].(float64)

        // 4. Check if token is revoked
        revoked, _ := IsTokenRevoked(c.Request.Context(), tokenID)
        if revoked {
            c.JSON(401, gin.H{"type": "https://api.simpo.com/errors/token-revoked", ...})
            c.Abort()
            return
        }

        // 5. Check expiration (handled by jwt.Parse, but verify)
        if exp, ok := claims["exp"].(float64); ok {
            if time.Now().Unix() > int64(exp) {
                c.JSON(401, gin.H{"type": "https://api.simpo.com/errors/session-expired", ...})
                c.Abort()
                return
            }
        }

        // 6. Update last activity
        UpdateLastActivity(c.Request.Context(), tokenID)

        // 7. Set user context for downstream middleware
        c.Set("user_id", uint(userID))
        c.Set("username", claims["username"])
        c.Set("role", claims["role"])
        c.Set("token_id", tokenID)

        c.Next()
    }
}
```

**Token Refresh Endpoint:**
```go
// POST /api/v1/auth/refresh
func (h *AuthHandler) RefreshToken(c *gin.Context) {
    // 1. Extract current token from Authorization header
    tokenString := extractBearerToken(c)

    // 2. Validate current token (not expired, not revoked)
    token, err := jwt.Parse(tokenString, validationKey)
    if err != nil || !token.Valid {
        c.JSON(401, gin.H{"type": "https://api.simpo.com/errors/invalid-token", ...})
        return
    }

    claims := token.Claims.(jwt.MapClaims)
    tokenID := claims["jti"].(string)

    // 3. Check if current token is revoked
    revoked, _ := IsTokenRevoked(c.Request.Context(), tokenID)
    if revoked {
        c.JSON(401, gin.H{"type": "https://api.simpo.com/errors/token-revoked", ...})
        return
    }

    // 4. Generate new token with same user info
    newTokenID := uuid.New().String()
    newClaims := &Claims{
        UserID:   uint(claims["user_id"].(float64)),
        Username: claims["username"].(string),
        Email:    claims["email"].(string),
        Role:     claims["role"].(string),
        BranchID: getBranchID(claims),
        RegisteredClaims: jwt.RegisteredClaims{
            ID:        newTokenID,
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(8 * time.Hour)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            Issuer:    "simpo-api",
        },
    }

    newToken := jwt.NewWithClaims(jwt.SigningMethodHS256, newClaims)
    tokenString, err := newToken.SignedString([]byte(jwtSecret))
    if err != nil {
        c.JSON(500, gin.H{"type": "https://api.simpo.com/errors/internal-error", ...})
        return
    }

    // 5. Revoke old token
    oldTokenExpiry := time.Until(time.Unix(int64(claims["exp"].(float64)), 0))
    RevokeToken(c.Request.Context(), tokenID, oldTokenExpiry)

    // 6. Update session tracking
    SaveSession(c.Request.Context(), newTokenID, sessionInfo)

    // 7. Return new token
    c.JSON(200, gin.H{
        "access_token": tokenString,
        "token_type":   "Bearer",
        "expires_in":   28800,
    })
}
```

**Logout Endpoint:**
```go
// POST /api/v1/auth/logout
func (h *AuthHandler) Logout(c *gin.Context) {
    // 1. Extract user context from middleware
    userID, exists := c.Get("user_id")
    if !exists {
        c.JSON(401, gin.H{"type": "https://api.simpo.com/errors/unauthorized", ...})
        return
    }

    tokenID, exists := c.Get("token_id")
    if !exists {
        c.JSON(401, gin.H{"type": "https://api.simpo.com/errors/missing-token", ...})
        return
    }

    // 2. Get token expiration time
    claims := c.MustGet("claims").(jwt.MapClaims)
    exp := int64(claims["exp"].(float64))
    tokenExpiry := time.Until(time.Unix(exp, 0))

    // 3. Revoke token
    RevokeToken(c.Request.Context(), tokenID.(string), tokenExpiry)

    // 4. Log logout action
    ipAddress := c.ClientIP()
    _ = h.auditService.LogLogout(c.Request.Context(), audit.LogoutEntry{
        UserID:    userID.(uint),
        TokenID:   tokenID.(string),
        Action:    "LOGOUT",
        IPAddress: ipAddress,
        Timestamp: time.Now(),
    })

    // 5. Return success
    c.JSON(200, gin.H{
        "success": true,
        "message": "Logged out successfully",
    })
}
```

### File Structure Requirements

**Files to Create/Modify:**

1. **internal/middleware/jwt_auth.go** (MODIFY)
   - Add token blocklist checking
   - Add session activity tracking
   - Add token expiration validation
   - Return RFC 7807 formatted errors
   - Add token_id to user context

2. **internal/middleware/session.go** (NEW)
   - Session tracking functions
   - Token blocklist functions
   - Redis integration for session storage
   - Unit tests for session operations

3. **internal/dto/refresh_dto.go** (NEW)
   - RefreshRequest DTO (empty body, uses current token)
   - RefreshResponse DTO with new token

4. **internal/dto/logout_dto.go** (NEW or OPTIONAL)
   - LogoutRequest DTO (empty body, uses current token)
   - LogoutResponse DTO

5. **internal/handlers/auth_handler.go** (MODIFY)
   - Add RefreshToken handler
   - Add Logout handler
   - Swagger annotations
   - Error handling

6. **internal/services/audit_service.go** (MODIFY)
   - Add LogLogout method
   - Include user_id, token_id, action fields

7. **internal/server/router.go** (MODIFY)
   - Register POST /api/v1/auth/refresh route
   - Register POST /api/v1/auth/logout route
   - Apply auth middleware

8. **docs/swagger.yaml** (UPDATE via swaggo)
   - Document refresh endpoint
   - Document logout endpoint
   - Error responses

### Database Schema

**No database schema changes required** - Session management uses Redis for in-memory storage:
- Session storage: `session:{user_id}:{token_id}` (TTL: 8 hours)
- Token blocklist: `revoked:{token_id}` (TTL: remaining token lifetime)

### Testing Requirements

**Unit Tests:**
- Test session creation and retrieval
- Test token blocklist add/check operations
- Test token expiration validation
- Test token refresh logic
- Test logout token revocation
- Test RFC 7807 error responses

**Integration Tests:**
- Test successful token refresh
- Test refresh with expired token fails
- Test refresh with revoked token fails
- Test successful logout
- Test logout invalidates token
- Test expired token returns 401
- Test revoked token returns 401
- Test audit trail entries for logout

**Test Coverage Goal:** >80% for session management

### Environment Variables

**Required Variables (from Story 1.5):**
```bash
JWT_SECRET=simpo_jwt_secret_key_for_pharmacy_management_system_2026_secure_token
JWT_ACCESS_TOKEN_TTL=8h

# Redis for session tracking
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_DB=0
```

### API Contract

**Refresh Endpoint:**

**Endpoint:** POST /api/v1/auth/refresh
**Authentication:** Bearer JWT (any valid token)

**Success Response (200 OK):**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_in": 28800
}
```

**Error Responses:**
- 401 Unauthorized: Invalid or expired token
- 401 Unauthorized: Revoked token
- 500 Internal Server Error: Token generation failed

**Logout Endpoint:**

**Endpoint:** POST /api/v1/auth/logout
**Authentication:** Bearer JWT (any valid token)

**Success Response (200 OK):**
```json
{
  "success": true,
  "message": "Logged out successfully"
}
```

**Error Responses:**
- 401 Unauthorized: No token provided
- 401 Unauthorized: Invalid token
- 500 Internal Server Error: Logout failed

**401 Unauthorized Error Response (RFC 7807):**
```json
{
  "type": "https://api.simpo.com/errors/session-expired",
  "title": "Session Expired",
  "status": 401,
  "detail": "Your session has expired. Please log in again.",
  "instance": "/api/v1/products"
}
```

### Naming Conventions

**Follow Architecture Patterns:**
- Database: Redis key pattern `session:{user_id}:{token_id}` and `revoked:{token_id}`
- JSON API: camelCase (accessToken, tokenType, expiresIn)
- Go code: camelCase for vars/functions, PascalCase for types
- API endpoints: plural REST (/api/v1/auth/refresh, /api/v1/auth/logout)
- Error types: lowercase snake_case in URLs (RFC 7807)

### Security Considerations

**Token Security:**
- Tokens must be signed with JWT_SECRET
- Token expiration is enforced on every request
- Revoked tokens are immediately invalid
- Token ID (jti claim) enables per-token revocation

**Session Security:**
- Last activity tracking enables inactivity detection
- Sessions auto-expire after 8 hours of inactivity
- Logout immediately invalidates tokens
- Multiple concurrent sessions allowed (multi-device)

**Redis Security:**
- Redis keys should have appropriate TTL (auto-cleanup)
- Token blocklist TTL matches remaining token lifetime
- Session data should not contain sensitive information
- Consider Redis password protection in production

**Audit Trail:**
- All logout actions are logged
- Audit log includes user_id, token_id, timestamp, ip_address
- Append-only log per NFR-SEC-004

### References

- [Source: docs/_bmad-output/planning-artifacts/architecture.md] - Decision 5 (Password Hashing), Decision 6 (API Security)
- [Source: docs/_bmad-output/planning-artifacts/epics.md] - Epic 1, Story 1.8
- [Source: docs/_bmad-output/planning-artifacts/prd.md] - FR5 (Session Timeout), NFR-SEC-002 (8-hour timeout)
- [Source: Story 1.5 - JWT Auth] - JWT token structure, expiration, AuditService
- [Source: Story 1.6 - RBAC] - Auth middleware pattern, user context
- [Source: Story 1-5 - JWT Auth Story File] - Token generation, expiration logic
- [Source: Story 1-6 - RBAC Story File] - JWTAuthMiddleware implementation

---

## Dev Agent Record

### Agent Model Used

claude-opus-4-6 (Senior Software Engineer - Amelia)

### Debug Log References

_Story created via bmad-create-story workflow on 2026-05-11_

### Completion Notes List

_Story ready for implementation_

---

## Change Log

| Date | Change | Author |
|------|--------|--------|
| 2026-05-11 | Story created via create-story workflow with comprehensive session management context. Built on Story 1.5 (JWT Auth) and Story 1.6 (RBAC) foundations. | BMad System (Claude Opus 4.6) |
