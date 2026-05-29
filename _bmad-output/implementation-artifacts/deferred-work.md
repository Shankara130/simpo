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

## Deferred from: Code Review of Story 9.6 (2026-05-13)

### Architecture & Design Decisions

- **Missing Documentation** — Pre-existing pattern. No godoc comments on existing services (AuthService, AuditService).
  - Recommendation: Defer to tech writer for consistent documentation pass.

- **No Metrics/Instrumentation** — Observability not implemented yet.
  - Recommendation: Defer to dedicated monitoring/instrumentation story.

- **Hardcoded Pagination Limits** — Existing pattern from Epic 2.
  - Recommendation: Defer to configuration story for consistent limits.

- **Panic in Constructors** — Existing pattern from AuthService.
  - Recommendation: Defer to architecture decision for consistent error handling pattern.

- **Context Timeout Propagation** — Design-level decision.
  - Recommendation: Handlers should set timeouts, not services. Defer to architecture decision.

### Business Logic

- **UpdateProduct SKU Change Prevention** — Could validate new SKU uniqueness, but repository likely enforces this at DB level.
  - Recommendation: Defer to architecture decision.

- **CheckAvailability Nil ExpiryDate** — Business rule decision: products without expiry dates are valid.
  - Recommendation: Defer to product owner.

### Security & Validation

- **sanitizeSearchInput Limited Scope** — GORM uses parameterized queries, so this is defense-in-depth.
  - Recommendation: Current sanitization sufficient for Epic 2 requirements.


## Deferred from: code review of Story 3.1 (2026-05-13)

- **Barcode scanning functionality** - Barcode scan input prominent and accessible (AC1) deferred to Story 3.2 (Barcode Scanner Integration) per user decision.

## Deferred from: code review of Story 4.1 (2026-05-18)

### Code Quality & Standards

- **MAGIC NUMBERS: Pagination limits hardcode** — `apps/backend/internal/handlers/product_handler.go:70-76`
  - Angka hardcode (20 default limit, 1000 max) adalah pola yang sudah ada di codebase sejak Epic 2.
  - Recommendation: Defer to configuration story untuk consistent pagination limits across all endpoints.

- **CONSTRUCTOR PANIC: Nil check dengan panic** — `apps/backend/internal/handlers/product_handler.go:36-39`
  - Panic pada nil service adalah design choice untuk fail-fast pattern, konsisten dengan AuthService constructor.
  - Recommendation: Defer to architecture decision untuk consistent error handling pattern across constructors.

### Business Logic

- **EMPTY SEARCH: Search string kosong return semua produk** — `apps/backend/internal/handlers/product_handler.go:115-119`
  - Perlu keputusan desain: apakah search kosong harus return empty list atau semua produk.
  - Recommendation: Defer to product owner untuk clarify expected behavior.

### Edge Cases

- **INTEGER OVERFLOW: Pagination calculation edge case** — `apps/backend/internal/handlers/product_handler.go:150-154`
  - Memerlukan 10K+ produk dengan limit kecil untuk trigger integer overflow di `totalPages` calculation.
  - Recommendation: Monitor production data; jika produk mendekati limit, implementasikan safe math division.

## Deferred from: code review of story 4-2-implement-real-time-stock-visibility (2026-05-19)

### Security & Authentication

- **JWT Authentication Implementation pada WebSocket Handler** — \`product_handler.go:218-235\`
  Issue: Kode memiliki placeholder logic yang mengecek \`userRole\` dan \`userBranchID\` dari Gin context, tetapi untuk WebSocket connections (HTTP upgrade), middleware tidak berjalan. Token dari query parameter tidak pernah di-decode atau divalidasi secara proper.
  Why deferred: Authentication logic requires architectural decision about JWT validation for WebSocket upgrades (different from regular HTTP requests)
  Recommendation: Implement proper JWT token decoding and validation from query parameter before accepting WebSocket connection

### Reliability & Performance

- **Redis Reconnection Handling Not Implemented** — \`stock_event_service.go\`
  Issue: Story specification explicitly notes "⚠️ Subtask 5.5: Redis reconnection handling NOT YET IMPLEMENTED". If Redis connection drops, the broadcaster stops working until server restart.
  Why deferred: Acknowledged in story as incomplete subtask
  Recommendation: Implement automatic reconnection to Redis with exponential backoff when connection is lost

### Infrastructure & Security

- **CORS Configuration TODO for Production** — \`product_handler.go:190\`
  Issue: Code contains \`return true // TODO: Configure CORS properly for production\` which allows all origins
  Why deferred: Security configuration for production deployment
  Recommendation: Implement proper CORS validation before production deployment

- **No Connection Limits on WebSocket** — \`product_handler.go:248-350\`
  Issue: No limit on number of WebSocket connections per user or globally
  Why deferred: DoS vulnerability mitigation requires infrastructure-level rate limiting
  Recommendation: Implement connection limits per user and globally to prevent resource exhaustion

### Monitoring & Validation

- **Missing Stock Reconciliation Accuracy Validation** — \`stock_metrics_service.go\`
  Issue: Framework for tracking accuracy is available, but there's no automated validation that proves the system achieves 99% accuracy as required by AC6
  Why deferred: Measurement gap, not implementation bug - requires additional monitoring/verification work
  Recommendation: Implement periodic background jobs that compare Redis cache values against database values to calculate actual reconciliation accuracy

## Deferred from: code review of story 4-5-implement-expiry-date-alerts (2026-05-21)

### Architecture & Design

- **Timezone Inconsistency (UTC vs local)** — `expiry_check_service.go:46`
  Issue: Server uses UTC but users in different timezones (e.g., UTC+7 for Indonesia) may see incorrect expiry dates
  Why deferred: Pre-existing architectural decision; should be addressed at system level with configurable business timezone
  Recommendation: Implement system-wide timezone configuration that respects user/branch local time for expiry calculations

- **Debounce Uses Key Instead of Sorted Set** — `expiry_check_service.go:127-143`
  Issue: Spec requires Redis Sorted Set with timestamp as score, but implementation uses simple key existence with TTL
  Why deferred: Functional equivalent achieves same result; spec could be clarified to allow simpler implementation
  Recommendation: Accept current implementation or update spec to allow key-based debounce with TTL

### Infrastructure & Performance

- **Performance Testing Not Verified** — N/A
  Issue: No performance tests to verify UI response < 500ms requirement (NFR-PERF-006)
  Why deferred: Testing infrastructure gap; requires performance monitoring and load testing setup
  Recommendation: Implement performance monitoring and load tests to verify UI response times meet requirement

- **Missing Pagination** — `product_repository_impl.go:253`
  Issue: No LIMIT clause in GetExpiringProducts query; large datasets could cause OOM
  Why deferred: Pre-existing pagination gap in repository pattern; should be addressed at system level
  Recommendation: Implement consistent pagination pattern across all repository queries

- **Unbounded Date Range Query Performance** — `product_repository_impl.go:253`
  Issue: Large date ranges could return thousands of products without limits
  Why deferred: Related to pagination gap; needs system-level approach with configurable limits
  Recommendation: Add configurable max result limits and enforce across all queries

### Configuration Management

- **Magic Numbers in Alert Thresholds** — `expiry_check_service.go:69-76`
  Issue: Thresholds (7, 14, 30) are hard-coded and not configurable
  Why deferred: Should be configurable via environment variables or database settings; defer for configuration management task
  Recommendation: Implement alert threshold configuration system to allow business-specific customization

## Deferred from: code review of 5-3-implement-report-export-functionality (2026-05-26)

- Logo support not implemented - Feature enhancement beyond current story scope
- Missing lock in job status updates - Pre-existing pattern in codebase
- Inconsistent error types - Existing codebase pattern, not introduced by this change
- Missing documentation - Code quality item, not functional issue
- Hardcoded company information - Product decision needed for config system architecture
- Inconsistent timezone handling - Assumes WIB by design for Indonesian pharmacy system
- System config integration - Architecture decision needed for multi-tenant support
- Missing rate limiting - Applies globally to all endpoints, not just this change
- Async export not fully implemented - Task 5 explicitly deferred for MVP scope


- Company branding hardcoded values — Deferred for MVP; requires dedicated story for config system architecture. Current hardcoded values acceptable for initial release.


## Deferred from: code review of 5-3-implement-report-export-functionality (2026-05-26 - Round 5)

- Hardcoded company information — Requires dedicated config system architecture story
- Logo support not implemented — Feature enhancement beyond current story scope
- Incomplete async export implementation — Task 5 explicitly deferred for MVP
- Missing file cleanup (actual deletion) — MVP placeholder; needs FileStorageService implementation
- Timezone inconsistency — Assumes WIB by design for Indonesian pharmacy system
- Code duplication — Existing pattern in handlers; refactoring deferred
- Inconsistent error handling patterns — Existing codebase pattern
- Unused validation function — validateAndParseDate exists but export handlers use different path
- Missing audit logging on failure paths — Intentional design to avoid blocking operations

## Deferred from: code review of 5-3-implement-report-export-functionality (2026-05-26 - Round 6)

- SQL injection risk via breakdown_by — Service layer should handle; whitelist validation already in place
- Missing rate limiting on exports — Applies globally to all endpoints, requires architecture decision
- Memory leak in job storage — Commented as TODO for future story with database persistence
- Audit log failures don't block operations — Intentional design choice to avoid blocking user operations
- Hardcoded company information — Product decision needed for config system architecture
- Missing file size validation in handler — Already validated in service layer (50MB limit)
- Code duplication in export handlers — Pre-existing pattern, acceptable for MVP
- Inconsistent date validation usage — Already noted; can be refactored later
- Missing input sanitization for user role — Minor issue; audit logs are internal-only
- Missing Content-Length header — Browser handles this automatically for HTTP responses
- Timezone inconsistency — Assumes WIB by design for Indonesian pharmacy system

## Deferred from: code review of 6-3-implement-automated-daily-backups (2026-05-27)

### Edge Cases & Validation

- **Empty path defaults to '/' without validation** — `disk_checker.go:19-24`
  - Issue: Path parameter defaults to "/" without explicit validation
  - Why deferred: Design choice - default path "/" is valid for Unix systems; not a bug introduced by this story
  - Recommendation: Keep as-is unless there's specific requirement for path validation

- **syscall.Statfs path not found handling** — `disk_checker.go:36`
  - Issue: No explicit check for path existence before Statfs call
  - Why deferred: Pre-existing error handling pattern - syscall already returns clear errors for non-existent paths
  - Recommendation: Current error handling is adequate; no additional validation needed

- **Negative freePercentage threshold validation** — `disk_checker.go:63-71`
  - Issue: No validation for negative freePercentage values
  - Why deferred: Extremely unlikely edge case (filesystem corruption) - defensive programming extreme case
  - Recommendation: Monitor production; add validation only if negative values observed in practice

- **Negative errorCount/totalRequests validation** — `metrics_collector.go:63-68`
  - Issue: No validation for negative errorCount or totalRequests values
  - Why deferred: Database constraints should prevent this at source - not an application layer concern
  - Recommendation: Verify database-level constraints are adequate; rely on data integrity

### Code Quality & Patterns

- **ClientIP empty string handling** — `system_settings_handler.go:181`
  - Issue: ClientIP() could return empty string or invalid IP format
  - Why deferred: Pre-existing audit logging issue - not introduced by this story
  - Recommendation: Address as part of broader audit logging improvement initiative

- **InvalidInputError type assertion** — `system_settings_handler.go:197-201`
  - Issue: Type assertion could fail if error type changes
  - Why deferred: Consistent pattern across entire codebase - requires architecture-level decision
  - Recommendation: Defer to architecture decision for consistent error handling patterns

### Authentication & Middleware

- **User role type assertion in auth middleware** — `system_settings_handler.go:65-71`
  - Issue: Type assertion assumes userRole is string type
  - Why deferred: Standard Gin framework pattern - tested and proven across codebase
  - Recommendation: Current pattern is acceptable; consider stronger typing in future API redesign

- **UserID type assertion** — `system_settings_handler.go:157-163`
  - Issue: Type assertion assumes userID is uint type
  - Why deferred: Same pattern as user role assertion - consistent with Gin middleware design
  - Recommendation: Keep as-is for consistency with existing authentication patterns

## Deferred from: code review of Story 7-4 (2026-05-28)

### Hardware Integration - Mobile

- **AC1 Enhancement - Actual Drawer Connection Detection**
  - Story: 7-4-implement-cash-drawer-control-via-printer-kick
  - File: PrinterManager.ts:438
  - Issue: Current implementation infers drawer connection from printer status (`get isDrawerConnected(): boolean { return this.currentStatus === PrinterStatus.CONNECTED; }`), not actual drawer detection
  - Impact: System cannot detect if drawer is actually connected to printer via RJ-12 or if drawer is disconnected/faulty
  - Why deferred: Requires hardware research and implementation of actual drawer detection mechanism via printer feedback, circuit detection, or drawer status polling
  - Recommendation: Implement as enhancement in future story with proper hardware testing and validation
  - Priority: HIGH for compliance and operational reliability


## Deferred from: Code Review of Story 8.3 (2026-05-29 - Chunk 1: Core Services)

### UI Components

- **Missing: UI component implementation for visual indicators**
  Story: 8-3-implement-bidirectional-data-synchronization
  File: No files found
  Issue: Hook implementation complete but UI display components not yet implemented
  Why deferred: Story 8.3 Task 6 explicitly deferred in original story ("admin intervention UI for permanently failed syncs" not implemented)
  Recommendation: Implement UI components in dedicated UI story or as part of Story 8-4 (Visual Sync Status Indicators)

### Code Quality & Standards

- **Hardcoded API URLs in constructor**
  Story: 8-3-implement-bidirectional-data-synchronization
  File: ProductSyncService.ts:34-36, UserSyncService.ts:32-34
  Issue: API URLs hardcoded as `http://localhost:8080/api/v1` in dev and `https://api.simpo.pharmacy/api/v1` in production
  Why deferred: Pre-existing pattern from other services; requires centralized config architecture to fix properly
  Recommendation: Extract to config system when implementing story for configuration management

- **Magic number timeout hardcoded (30000)**
  Story: 8-3-implement-bidirectional-data-synchronization
  File: ProductSyncService.ts:103, 176; UserSyncService.ts:799
  Issue: 30-second timeout hardcoded in multiple locations as `30000`
  Why deferred: Code quality issue that should be addressed consistently across codebase
  Recommendation: Extract to named constant (e.g., `SYNC_REQUEST_TIMEOUT_MS = 30000`) in future code quality improvement
