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

## Deferred from: code review of 1-6-implement-role-based-access-control-rbac (2026-05-11)

### Code Quality & Standards

- **Test helper functions location**
  - File: `internal/middleware/auth_middleware_test.go:537-547`
  - Issue: `containsRFC7807Fields()` and `containsString()` defined in auth_middleware_test.go but used across multiple test files (rbac_test.go, integration_test.go)
  - Why deferred: Pre-existing pattern from Story 1.5/GRAB boilerplate, not introduced by this story
  - Recommendation: Create `internal/middleware/test_helpers.go` to share common test utilities

## Deferred from: code review of 2-2-create-initial-migration-with-golang-migrate (2026-05-12)

### Data Integrity & Audit

- **User ID deleted while referenced**
  - File: `apps/backend/migrations/20260512200001_create_branches_table.up.sql:14-15`
  - Issue: `created_by INTEGER` and `updated_by INTEGER` have no foreign key constraints to users table
  - Why deferred: Intentional design to preserve audit trail even when referenced users are deleted
  - Recommendation: Document this behavior in data dictionary and ensure application handles NULL values gracefully

- **NULL cost_price with profit calculation**
  - File: `apps/backend/migrations/20260512200002_create_products_table.up.sql:13`
  - Issue: `cost_price DECIMAL(15,2)` is nullable, will cause NULL in profit/loss calculations
  - Why deferred: Business decision - some products may not have cost price (e.g., donations, samples)
  - Recommendation: Application queries must use COALESCE or filter out NULL cost_price in profit calculations

- **Cashier user deletion while transactions exist**
  - File: `apps/backend/migrations/20260512200003_create_transactions_table.up.sql:33-35`
  - Issue: `cashier_id` has `ON DELETE RESTRICT` preventing cashier account deletion if transactions exist
  - Why deferred: Intentional design to preserve transaction audit trail
  - Recommendation: Application should implement user soft delete or reassign transactions before deleting cashier

### Schema Design

- **Version INTEGER overflow after many updates**
  - File: All migration files, `version INTEGER NOT NULL DEFAULT 1`
  - Issue: After 2 billion updates, version column will overflow
  - Why deferred: Theoretical concern - unlikely to happen in practice for pharmacy system
  - Recommendation: Monitor version values in production; consider BIGINT if approaching limit (extremely unlikely)

- **Soft delete cascade inconsistency**
  - File: `apps/backend/migrations/20260512200004_create_transaction_items_table.up.sql:25-27`
  - Issue: `transaction_id` has `ON DELETE CASCADE` but transaction_items has `deleted_at` for soft delete
  - Why deferred: Complex interaction - hard delete cascade for consistency with snapshot pattern
  - Recommendation: Application must handle soft delete of transactions explicitly (reassign or delete child items first)

## Deferred from: code review of story 2.4 (2026-05-12)

- **Inconsistent defaults between old and new functions** — `NewPostgresDB()` uses different defaults than `NewPostgresDBFromDatabaseConfig()`. Pre-existing code, not changed in this story.
- **Potential connection leak on validation failure** — Validator runs BEFORE connection is established, so no leak possible. Not applicable.
- **No handling of connection state transitions** — Long-running connection health management is operational concern, out of scope for this infrastructure story.

