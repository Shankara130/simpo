# Deferred Work Items

This file tracks work items that were identified during reviews but deferred to later stories or infrastructure work.

## Deferred from: code review of 1-5-implement-user-authentication-with-jwt (2026-05-10)

### Infrastructure & Config

- **Hardcoded JWT secret in .env.example**
  - File: `.env.example:42`
  - Issue: Contains `JWT_SECRET=simpo_jwt_secret_key_min_32_chars_for_production_please_change`
  - Why deferred: Pre-existed before Story 1.5, already documented as placeholder for production
  - Recommendation: Add validation in CI/CD to detect usage of example secrets in production builds

- **Missing request ID generation**
  - File: `internal/errors/response.go:29`
  - Issue: `RequestID` field exists but is never populated
  - Why deferred: GRAB boilerplate infrastructure issue, requires middleware changes across all handlers
  - Recommendation: Implement as infrastructure improvement in separate story

- **Hardcoded error type URI**
  - File: `internal/errors/middleware.go:90`
  - Issue: `baseURL := "https://api.simpo.com/errors"` is hardcoded
  - Why deferred: Infrastructure-level concern, acceptable for MVP
  - Recommendation: Make configurable via environment variable for production deployments

### Code Quality & Standards

- **Bcrypt cost not configurable**
  - File: `internal/user/service.go:17-18`
  - Issue: `BcryptCost = 12` is hardcoded constant
  - Why deferred: Per Architecture Decision 5, cost factor 12 is specified. Making it configurable would violate the architecture decision.
  - Recommendation: Keep as-is per Decision 5. Only reconsider if hardware constraints arise.

### Testing

- **Missing integration tests**
  - Issue: No end-to-end tests verify full login flow with database
  - Why deferred: Out of scope for current story focus on unit tests
  - Recommendation: Add integration test suite in future testing-focused story
