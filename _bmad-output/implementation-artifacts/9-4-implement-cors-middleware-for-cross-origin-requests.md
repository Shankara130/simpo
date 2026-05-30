# Story 9.4: Implement CORS Middleware for Cross-Origin Requests

**Status:** done

**Epic:** 9 - API Foundation & Core Services
**Priority:** CRITICAL (API Security)
**Story Type:** Middleware Implementation
**Story ID:** 9.4
**Story Key:** 9-4-implement-cors-middleware-for-cross-origin-requests

---

## Story

**As a** System (Development Team),
**I want** to implement CORS middleware so the web dashboard can call the API even if hosted on different domains,
**So that** we have deployment flexibility and don't force same-origin deployment.

---

## Acceptance Criteria

1. **AC1: Origin Validation**  
   Given the API is running and CORS middleware is configured,
   When cross-origin requests come from the web dashboard,
   Then the middleware validates the Origin header against allowed origins,
   And only requests from allowed origins proceed.

2. **AC2: CORS Headers for Allowed Origins**  
   When requests come from allowed origins,
   Then the system responds with appropriate CORS headers:
   - `Access-Control-Allow-Origin`: specific origin (not "*" for security)
   - `Access-Control-Allow-Methods`: GET, POST, PUT, DELETE, OPTIONS
   - `Access-Control-Allow-Headers`: Authorization, Content-Type, X-Requested-With
   And credentials are supported if needed.

3. **AC3: Pre-flight OPTIONS Requests**  
   When a pre-flight OPTIONS request is received,
   Then the system returns 200 OK with appropriate CORS headers,
   And no authentication is required for OPTIONS requests.

4. **AC4: Environment-Based Configuration**  
   CORS configuration is environment-based:
   - Development: localhost origins (3000, 19006, etc.)
   - Production: specific web dashboard domains
   And configuration is via environment variables or config files.

5. **AC5: Security - No Wildcard Origins**  
   The system never returns `Access-Control-Allow-Origin: "*"` when credentials are involved,
   And each allowed origin is explicitly configured.

---

## Tasks / Subtasks

- [x] **Task 1: Add CORS Configuration to Config Struct (AC: 4)**
  - [x] Add CORS configuration fields to `internal/config/config.go`
  - [x] Add AllowedOrigins []string field
  - [x] Add AllowCredentials bool field
  - [x] Add AllowedMethods []string field
  - [x] Add AllowedHeaders []string field
  - [x] Add MaxAge int field (for pre-flight cache)
  - [x] Add default values in config loading

- [x] **Task 2: Update .env.example with CORS Configuration (AC: 4)**
  - [x] Add CORS_ENABLED environment variable
  - [x] Add CORS_ALLOWED_ORIGINS environment variable (comma-separated)
  - [x] Add CORS_ALLOW_CREDENTIALS environment variable
  - [x] Add CORS_ALLOWED_METHODS environment variable
  - [x] Add CORS_ALLOWED_HEADERS environment variable
  - [x] Add CORS_MAX_AGE environment variable

- [x] **Task 3: Update config.yaml with CORS Configuration (AC: 4)**
  - [x] Add cors.enabled field
  - [x] Add cors.allowedOrigins array field
  - [x] Add cors.allowCredentials field
  - [x] Add cors.allowedMethods array field
  - [x] Add cors.allowedHeaders array field
  - [x] Add cors.maxAge field

- [x] **Task 4: Replace Insecure CORS Configuration in router.go (AC: 1, 2, 5)**
  - [x] Remove `corsConfig.AllowAllOrigins = true` (SECURITY ISSUE)
  - [x] Update CORS config to use values from cfg.Cors
  - [x] Set AllowOrigins from cfg.Cors.AllowedOrigins
  - [x] Set AllowCredentials from cfg.Cors.AllowCredentials
  - [x] Set AllowMethods from cfg.Cors.AllowedMethods
  - [x] Set AllowHeaders from cfg.Cors.AllowedHeaders
  - [x] Set MaxAge from cfg.Cors.MaxAge

- [x] **Task 5: Implement Origin Validation Logic (AC: 1)**
  - [x] Use cors.Config with specific origins (not AllowAllOrigins)
  - [x] Ensure AllowOrigins is set from config (not wildcard)
  - [x] Verify origin matching is case-sensitive
  - [x] Test that requests from disallowed origins are rejected

- [x] **Task 6: Verify Pre-flight OPTIONS Handling (AC: 3)**
  - [x] Confirm OPTIONS requests return 200 OK
  - [x] Verify CORS headers are present in OPTIONS response
  - [x] Ensure OPTIONS bypasses auth middleware
  - [x] Test pre-flight for various HTTP methods

- [x] **Task 7: Add CORS Middleware Tests (AC: 1, 2, 3, 5)**
  - [x] Create `internal/middleware/cors_test.go`
  - [x] Test allowed origins receive CORS headers
  - [x] Test disallowed origins are rejected
  - [x] Test pre-flight OPTIONS requests
  - [x] Test credentials are supported when configured
  - [x] Test specific origin is returned (not wildcard)

- [x] **Task 8: Add Integration Tests (AC: 1, 2)**
  - [x] Test cross-origin GET requests
  - [x] Test cross-origin POST requests with Authorization
  - [x] Test cross-origin PUT/DELETE requests
  - [x] Test requests from multiple allowed origins
  - [x] Test requests from disallowed origins fail

- [x] **Task 9: Update Documentation (AC: 4)**
  - [x] Document CORS configuration in docs/CORS.md
  - [x] Document how to configure allowed origins per environment
  - [x] Document security implications of CORS configuration
  - [x] Add CORS configuration examples for common scenarios

---

## Dev Notes

### Implementation Context

**CRITICAL SECURITY ISSUE:** The current CORS configuration in `router.go` (lines 46-49) uses:
```go
corsConfig.AllowAllOrigins = true
```

This is **INSECURE** for production deployments as it allows any origin to make requests with credentials. This story fixes this security vulnerability by implementing environment-based origin validation.

**Current Implementation Location:**
- `apps/backend/internal/server/router.go` (lines 46-49) - Insecure CORS config
- `apps/backend/internal/config/config.go` - Config struct to extend
- `apps/backend/.env.example` - Environment configuration
- `apps/backend/configs/config.yaml` - YAML configuration

**Required Enhancement:**
- Replace `AllowAllOrigins = true` with specific allowed origins
- Implement environment-based configuration (development vs production)
- Support pre-flight OPTIONS requests properly
- Ensure security: no wildcard origins with credentials

### Architecture Context

**From Architecture Decision 6 - API Security Strategy:**
> Defense in depth: HTTPS + rate limiting + CORS + input sanitization + API versioning

**CORS Implementation Requirements:**
> Restrict to known origins (mobile app, web dashboard). Environment-based configuration.

**Security Requirements (NFR-SEC-006):**
> TLS 1.3 enforcement - CORS works alongside HTTPS for secure cross-origin requests.

### Existing Middleware Pattern (from Story 9-3)

Middleware order in `router.go`:
1. Logger (line 42)
2. ErrorHandler (line 43)
3. Recovery (line 44)
4. **CORS (line 49)** ← This story's focus
5. Rate Limit (line 100)
6. Auth (SessionAuthMiddleware)
7. RBAC

**CORS must remain early in the chain** (before auth) so that OPTIONS requests work without authentication.

### Configuration Pattern

**From .env.example (existing patterns):**
```bash
RATELIMIT_ENABLED=true
RATELIMIT_REQUESTS=100
RATELIMIT_WINDOW=1m
```

**Proposed CORS configuration:**
```bash
CORS_ENABLED=true
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:19006,https://admin.simpo.com
CORS_ALLOW_CREDENTIALS=true
CORS_ALLOWED_METHODS=GET,POST,PUT,DELETE,OPTIONS
CORS_ALLOWED_HEADERS=Authorization,Content-Type,X-Requested-With
CORS_MAX_AGE=86400
```

**From config.yaml (existing patterns):**
```yaml
ratelimit:
  enabled: true
  requests: 100
  window: "1m"
```

**Proposed CORS YAML:**
```yaml
cors:
  enabled: true
  allowedOrigins:
    - "http://localhost:3000"
    - "http://localhost:19006"
    - "https://admin.simpo.com"
  allowCredentials: true
  allowedMethods:
    - "GET"
    - "POST"
    - "PUT"
    - "DELETE"
    - "OPTIONS"
  allowedHeaders:
    - "Authorization"
    - "Content-Type"
    - "X-Requested-With"
  maxAge: 86400  # 24 hours
```

### Previous Story Learnings

**From Story 9-3 (Rate Limiting):**
- Use environment-based configuration patterns
- Test middleware with and without authentication
- Integration tests cover multiple scenarios
- Configuration should work for both .env and config.yaml

**From Story 9-2 (Swagger):**
- Document security implications
- Use RFC 7807 for error responses

**From Story 9-1 (Health Check):**
- Environment-based behavior (development vs production)

### Code Changes Required

**Current Code (router.go lines 46-49):**
```go
corsConfig := cors.DefaultConfig()
corsConfig.AllowAllOrigins = true  // ← SECURITY ISSUE
corsConfig.AllowHeaders = append(corsConfig.AllowHeaders, "Authorization")
router.Use(cors.New(corsConfig))
```

**Required Code:**
```go
if cfg.Cors.Enabled {
    corsConfig := cors.Config{
        AllowOrigins:     cfg.Cors.AllowedOrigins,  // Specific origins from config
        AllowMethods:     cfg.Cors.AllowedMethods,
        AllowHeaders:     cfg.Cors.AllowedHeaders,
        AllowCredentials: cfg.Cors.AllowCredentials,
        MaxAge:           time.Duration(cfg.Cors.MaxAge) * time.Second,
        ExposeHeaders:    []string{"Content-Length"},
    }
    router.Use(cors.New(corsConfig))
}
```

### Testing Strategy

**Unit Tests (cors_test.go):**
1. Test allowed origin receives CORS headers
2. Test disallowed origin is rejected (no CORS headers)
3. Test pre-flight OPTIONS returns 200 OK
4. Test credentials cookie/header is supported
5. Test multiple allowed origins work correctly
6. Test case-sensitive origin matching

**Integration Tests:**
1. Cross-origin GET request with Authorization header
2. Cross-origin POST request with JSON body
3. Pre-flight OPTIONS for POST with custom headers
4. Request from disallowed origin returns error
5. Multiple origins in configuration all work

### Security Considerations

**CORS Security Best Practices:**
- Never use `AllowAllOrigins = true` in production
- Never use `Access-Control-Allow-Origin: "*"` with `AllowCredentials = true`
- Explicitly list all allowed origins
- Use HTTPS origins in production (not HTTP)
- Validate origins on server-side (never trust client)

**Common CORS Vulnerabilities:**
- ✗ Wildcard origins with credentials (currently present!)
- ✗ Reflecting Origin header without validation
- ✗ Overly permissive headers (e.g., `Access-Control-Allow-Headers: "*"`)
- ✗ Missing pre-flight handling

**This story fixes the wildcard origin vulnerability.**

### Performance Considerations

**CORS middleware impact:**
- O(1) origin check (string comparison)
- No significant performance overhead
- Pre-flight requests are cached by browser (MaxAge)

**No performance degradation expected** - CORS is a lightweight header check.

### Deployment Scenarios

**Development Environment:**
- Origins: `http://localhost:3000`, `http://localhost:19006` (Expo)
- Credentials: true (for auth tokens)
- Methods: GET, POST, PUT, DELETE, OPTIONS

**Production Environment:**
- Origins: `https://admin.simpo.com`, `https://simpo-pharmacy.com`
- Credentials: true (for auth tokens)
- Methods: GET, POST, PUT, DELETE, OPTIONS
- **No HTTP origins** (HTTPS only)

**Self-Hosted Deployment:**
- Customer configures their own web dashboard domain
- Customer adds domain to CORS_ALLOWED_ORIGINS
- Documentation should provide clear examples

### Integration Points

**After Story 9-4:**
- ✅ CORS middleware validates origins
- ✅ Pre-flight OPTIONS requests work
- ✅ Environment-based configuration
- ✅ Web dashboard can call API from different domain
- ✅ Security: no wildcard origins

**Depends On:**
- Story 9-1: API Health Check (router structure)
- Story 1.5: User Authentication with JWT (auth headers)

**Enables:**
- Story 6.x: Web Admin Dashboard (cross-origin API calls)
- Future: Multi-domain deployments

---

## References

- [Source: epics.md#Story-9.4] - Story 9.4 acceptance criteria
- [Source: architecture.md#Decision-6] - API security strategy (CORS as defense in depth)
- [Source: apps/backend/internal/server/router.go] - Current insecure CORS config (lines 46-49)
- [Source: apps/backend/internal/config/config.go] - Config struct pattern
- [Source: apps/backend/.env.example] - Environment configuration pattern
- [Source: apps/backend/configs/config.yaml] - YAML configuration pattern
- [Source: Story 9-3] - Middleware pattern and configuration approach

---

## Dev Agent Record

### Agent Model Used

claude-opus-4-6

### Debug Log References

None - Implementation completed successfully

### Completion Notes List

**Story Status:** done - All tasks complete, tests passing, documentation updated, code review approved

**Implementation Completed:**
- ✅ **CORS Configuration** added to Config struct with environment bindings and default values
- ✅ **Environment variables** configured in .env.example (CORS_ENABLED, CORS_ALLOWED_ORIGINS, etc.)
- ✅ **YAML configuration** added to config.yaml with development and production examples
- ✅ **SECURITY FIX:** Replaced insecure `AllowAllOrigins = true` with specific allowed origins from config
- ✅ **Origin validation** implemented via gin-contrib/cors with case-sensitive matching
- ✅ **Pre-flight OPTIONS** handling verified (returns 200 OK with CORS headers, bypasses auth)
- ✅ **Credentials support** configured (AllowCredentials: true for cookies and auth headers)
- ✅ **Unit tests** created (cors_test.go) - 7 test cases all passing
- ✅ **Integration tests** created (cors_integration_test.go) - 7 test cases all passing
- ✅ **Documentation** created (docs/CORS.md) with security guide and configuration examples

**Critical Security Fix:**
- **BEFORE:** `corsConfig.AllowAllOrigins = true` ❌ (allowed any origin)
- **AFTER:** `corsConfig.AllowOrigins = cfg.Cors.AllowedOrigins` ✅ (specific origins only)

**Configuration Added:**
- Development: `http://localhost:3000`, `http://localhost:19006`
- Production: `https://admin.simpo.com` (HTTPS only)
- Credentials: Supported (cookies, auth headers)
- Pre-flight cache: 24 hours (86400 seconds)

**Testing Completed:**
- ✅ Allowed origins receive CORS headers
- ✅ Disallowed origins are rejected (no CORS headers)
- ✅ Pre-flight OPTIONS returns 200 OK with proper headers
- ✅ Credentials are supported when configured
- ✅ Specific origin returned (not wildcard)
- ✅ Multiple allowed origins work correctly
- ✅ Case-sensitive origin matching verified
- ✅ Cross-origin GET/POST/PUT/DELETE all work
- ✅ Cross-origin requests with Authorization header work

**All Acceptance Criteria Met:**
- ✅ AC1: Origin validation implemented (gin-contrib/cors handles validation)
- ✅ AC2: CORS headers for allowed origins (specific origin, methods, headers, credentials)
- ✅ AC3: Pre-flight OPTIONS returns 200 OK with CORS headers
- ✅ AC4: Environment-based configuration (.env + config.yaml)
- ✅ AC5: Security - no wildcard origins with credentials

**Code Review Patches Applied (2026-05-30):**
- ✅ Patch 1: Nil slice validation in router.go for AllowedOrigins, AllowedMethods, AllowedHeaders
- ✅ Patch 2: Changed default origins from localhost to empty array (production-safe)
- ✅ Patch 3: Added validation to reject wildcard "*" when AllowCredentials is true
- ✅ Patch 4: Nil slice validation for AllowedMethods in router.go
- ✅ Patch 5: Nil slice validation for AllowedHeaders in router.go
- ✅ Patch 6: Deferred (handled by Viper's built-in CSV parsing)
- ✅ Patch 7: MaxAge bounds checking (0 to 9223372036 range)
- ✅ Patch 8: Origin format validation (must include http:// or https:// protocol)
- ✅ Patch 9: Redacted sensitive origins in LogSafeConfig output

### File List

**Files MODIFIED:**
- `apps/backend/internal/config/config.go` - Added CorsConfig struct and Cors field to Config
- `apps/backend/internal/config/config.go` - Added CORS environment variable bindings
- `apps/backend/internal/config/config.go` - Added CORS default values and logging
- `apps/backend/.env.example` - Added CORS configuration section
- `apps/backend/configs/config.yaml` - Added CORS configuration section
- `apps/backend/internal/server/router.go` - Replaced insecure CORS with secure config-based implementation

**Files CREATED:**
- `apps/backend/internal/middleware/cors_test.go` - Unit tests for CORS middleware (7 test cases)
- `apps/backend/internal/middleware/cors_integration_test.go` - Integration tests for CORS (7 test cases)
- `apps/backend/docs/CORS.md` - Comprehensive CORS documentation with security guide

---

### Senior Developer Review (AI)

**Review Date:** 2026-05-30
**Review Outcome:** Approved - All patches applied
**Total Action Items:** 9 patches, 4 deferred, 4 dismissed
**Resolution Date:** 2026-05-30

#### Action Items

**Patch Items (9) - Require Code Fixes:**

- [x] [Review][Patch] **Nil slice dereference in AllowedOrigins causes runtime panic** [config.go:50, router.go:50] — Add nil/empty validation before passing cfg.Cors.AllowedOrigins to cors.Config
- [x] [Review][Patch] **Development origins as defaults expose production to localhost** [config.go:137] — Add production-safe defaults or require explicit CORS_ALLOWED_ORIGINS override
- [x] [Review][Patch] **No validation prevents AllowCredentials with wildcard origins** [config.go:131] — Add validation to reject wildcard "*" when AllowCredentials is true
- [x] [Review][Patch] **Nil slice dereference in AllowedMethods breaks all HTTP methods** [config.go:51, router.go:51] — Add nil/empty validation for AllowedMethods slice
- [x] [Review][Patch] **Nil slice dereference in AllowedHeaders prevents authentication** [config.go:52, router.go:52] — Add nil/empty validation for AllowedHeaders slice
- [x] [Review][Patch] **Environment variable CSV parsing may produce invalid origins** [config.go:224] — Ensure CORS_ALLOWED_ORIGINS is properly parsed as comma-separated values (DEFERRED: handled by Viper's CSV parsing)
- [x] [Review][Patch] **MaxAge integer overflow possible** [router.go:54] — Add bounds checking for negative or extremely large MaxAge values
- [x] [Review][Patch] **Missing origin format validation** [config loading] — Add URL format validation for origins (must include protocol)
- [x] [Review][Patch] **Sensitive origins logged without sanitization** [config.go:293] — Redact or mask origins in LogSafeConfig output

**Deferred Items (4) - Pre-existing or Require Broader Discussion:**

- [x] [Review][Defer] **Production domain hardcoded in config files** [config.yaml:56, .env.example:91] — deferred, architecture decision (production domain naming)
- [x] [Review][Defer] **No explicit OPTIONS bypass visible** [router.go:48-58] — deferred, gin-contrib/cors handles OPTIONS internally
- [x] [Review][Defer] **No CORS rejection logging for security monitoring** [router.go] — deferred, monitoring enhancement (add CORS rejection logging)
- [x] [Review][Defer] **Missing validation of origin vs credentials combination** — deferred, spec enhancement (document browser behavior constraints)

**Dismissed Items (4) - Noise or Not Actionable:**

- [x] CORS MaxAge 24 hours - dismissed as reasonable default for pre-flight caching
- [x] Config documentation comments - dismissed as acceptable for security-sensitive settings
- [x] Sprint status implementation details - dismissed as appropriate status tracking
- [x] Config loading race condition - dismissed as extremely unlikely, not actionable
