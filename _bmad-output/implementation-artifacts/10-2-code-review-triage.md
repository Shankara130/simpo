# Code Review Triage Results - Story 10-2: Purchase Invoice Recording

**Review Date:** 2026-05-30
**Layers:** Blind Hunter, Edge Case Hunter, Acceptance Auditor
**Total Raw Findings:** 56 (after deduplication)

---

## DECISION_NEEDED (8)

### DN-001: Missing PurchaseInvoiceItem persistence logic
**Source:** edge+auditor
**Location:** `purchase_invoice_service_impl.go:42-135`
**Detail:** CRITICAL - Service creates invoice but never persists line items. No repository method for creating items, no transaction wrapping invoice+items. Items array is passed to service but there's no code that creates line item records.
**Decision Required:** Should line items be created in the same transaction as invoice? What's the intended data flow?

### DN-002: Missing supplier contact information in response
**Source:** auditor
**Location:** `purchase_invoice_dto.go:107`
**Detail:** AC3 violation - Response only includes SupplierName, missing ContactPerson, Phone, Email, Address fields.
**Decision Required:** Should these fields be added to response DTO? Are they in Supplier model?

### DN-003: Missing payment history tracking
**Source:** auditor
**Location:** `purchase_invoice.go` model
**Detail:** AC3 violation - No payment history relationship or fields in model. Cannot display payment status history.
**Decision Required:** Is payment history part of this story or future story? What structure needed?

### DN-004: Missing branch information in response
**Source:** auditor
**Location:** `purchase_invoice_dto.go`
**Detail:** AC3 partial - Branch information not included in response DTO despite being in model.
**Decision Required:** Should BranchName be included in invoice response?

### DN-005: Line items relationship not handled
**Source:** edge+auditor
**Location:** Multiple files
**Detail:** Orphaned line items risk - no cascade handling for soft delete, items not created with invoice.
**Decision Required:** Should items use foreign key with cascade? How to handle soft delete?

### DN-006: Payment status transition validation
**Source:** edge
**Location:** `purchase_invoice_service_impl.go` Update
**Detail:** No state machine validation - can reverse "paid" to "unpaid".
**Decision Required:** What are valid payment status transitions? Should state machine be enforced?

### DN-007: Duplicate product IDs in items
**Source:** edge
**Location:** `purchase_invoice_service_impl.go:77-99`
**Detail:** No check for duplicate ProductID in items array.
**Decision Required:** Should same product appear multiple times? Should be aggregated or rejected?

### DN-008: Date format inconsistency
**Source:** blind
**Location:** Handler vs Service
**Detail:** Handler expects YYYY-MM-DD, service expects RFC3339.
**Decision Required:** Which format should be standardized? Is this intentional?

---

## PATCH (22)

### PATCH-001: SQL injection via sort order string concatenation
**Source:** blind
**Severity:** HIGH
**Location:** `purchase_invoice_repository_impl.go:3170`
**Detail:** `query = query.Order(sortBy + " " + sortOrder)` - despite whitelists, string concatenation is fragile. Use GORM's parameterized queries instead.

### PATCH-002: Race condition in invoice number validation
**Source:** blind+edge
**Severity:** HIGH
**Location:** `purchase_invoice_repository_impl.go:2914-2926`, `purchase_invoice_service_impl.go:106-109`
**Detail:** TOCTOU vulnerability - gap between duplicate check and insert. Database constraint catches but error handling inconsistent.

### PATCH-003: Missing branch access authorization
**Source:** blind
**Severity:** HIGH
**Location:** `purchase_invoice_handler.go` (all handlers)
**Detail:** Handlers never validate user's branch_id matches invoice's branch_id. RBAC only checks role, not branch membership.

### PATCH-004: Float overflow in total amount calculation
**Source:** edge
**Severity:** CRITICAL
**Location:** `purchase_invoice_service_impl.go:98`
**Detail:** `subtotal := float64(item.Quantity) * item.UnitCost` - large values can overflow. No bounds checking.

### PATCH-005: Date range inversion not validated
**Source:** edge
**Severity:** CRITICAL
**Location:** `purchase_invoice_repository_impl.go:251-256`
**Detail:** No validation that StartDate <= EndDate. Inverted ranges return empty results silently.

### PATCH-006: Zero SupplierID not validated in Update
**Source:** edge
**Severity:** HIGH
**Location:** `purchase_invoice_service_impl.go:186-294`
**Detail:** SupplierID=0 passes validation but causes FK violation.

### PATCH-007: Empty invoice number after normalization
**Source:** edge
**Severity:** HIGH
**Location:** `purchase_invoice_service_impl.go:202`
**Detail:** Whitespace-only invoice number becomes empty after normalization, causing constraint violations.

### PATCH-008: Payment status enum not validated in Update
**Source:** edge
**Severity:** HIGH
**Location:** `purchase_invoice_service_impl.go:267-272`
**Detail:** Can set payment_status to any string, violating database constraint.

### PATCH-009: Negative total amount not validated
**Source:** edge
**Severity:** HIGH
**Location:** `purchase_invoice_service_impl.go:102`
**Detail:** No validation that totalAmount >= 0 after calculation.

### PATCH-010: Malformed date string injection
**Source:** edge
**Severity:** HIGH
**Location:** `purchase_invoice_repository_impl.go:251-256`
**Detail:** Invalid date formats cause SQL errors instead of validation errors.

### PATCH-011: Missing Document URL format validation
**Source:** blind
**Severity:** MEDIUM
**Location:** `purchase_invoice_dto.go:38-39`
**Detail:** Only validates length, not URL format. Could allow `javascript:`, `data:`, `file:` protocols.

### PATCH-012: Integer overflow in pagination offset
**Source:** blind
**Severity:** MEDIUM
**Location:** `purchase_invoice_repository_impl.go:3138-3142`
**Detail:** `(page - 1) * limit` can overflow if limit is large.

### PATCH-013: Information leakage in error messages
**Source:** blind
**Severity:** MEDIUM
**Location:** `purchase_invoice_service_impl.go:3788-3790`
**Detail:** Error includes invoice number, enabling enumeration attacks.

### PATCH-014: Audit log failures silently ignored
**Source:** blind+edge
**Severity:** MEDIUM
**Location:** All service methods
**Detail:** Audit logging failures caught but ignored, no monitoring.

### PATCH-015: Large items array no limit
**Source:** edge
**Severity:** MEDIUM
**Location:** `purchase_invoice_service_impl.go` Create
**Detail:** No limit on items array size, can cause memory issues.

### PATCH-016: Nil pointer dereference in response conversion
**Source:** edge
**Severity:** MEDIUM
**Location:** `purchase_invoice_handler.go:653-655`
**Detail:** Checks `invoice.Supplier.ID != 0` but should check `invoice.Supplier != nil` first.

### PATCH-017: Unicode normalization inconsistency
**Source:** edge
**Severity:** MEDIUM
**Location:** Multiple layers
**Detail:** Handler doesn't normalize, service normalizes in Create not Update, repository normalizes differently.

### PATCH-018: Missing timezone handling
**Source:** edge
**Severity:** MEDIUM
**Location:** `purchase_invoice_service_impl.go:64-66`
**Detail:** Date comparison mixes UTC and local timezone.

### PATCH-019: Empty pagination result handling
**Source:** edge
**Severity:** MEDIUM
**Location:** `purchase_invoice_handler.go:687-690`
**Detail:** Total=0 returns 0 pages but user requested page > 0.

### PATCH-020: UnitCost precision loss
**Source:** edge
**Severity:** MEDIUM
**Location:** Total calculation
**Detail:** Float64 can't represent all decimal values precisely. Financial calculations should use decimal.Decimal.

### PATCH-021: Search query not sanitized at handler
**Source:** edge
**Severity:** LOW
**Location:** `purchase_invoice_handler.go:263`
**Detail:** Search passed directly to service without handler-level validation.

### PATCH-022: Reason field silently truncated
**Source:** edge
**Severity:** LOW
**Location:** `purchase_invoice_handler.go:712-714`
**Detail:** 500 char limit with no indication to user.

---

## DEFER (8)

### DEFER-001: Weak Password Requirements
**Source:** blind
**Severity:** LOW
**Detail:** Not in scope for this diff. Pre-existing user management issue.

### DEFER-002: Insufficient Rate Limiting
**Source:** blind
**Severity:** MEDIUM
**Detail:** Inherited from JWT middleware. Should add invoice-specific limits but not blocking.

### DEFER-003: Missing validation for negative limit
**Source:** edge
**Severity:** LOW
**Detail:** Silently converts to default, hiding client bugs. Pre-existing pagination pattern.

### DEFER-004: Soft-deleted invoice counting
**Source:** edge
**Severity:** LOW
**Detail:** GORM implicit soft-delete filtering. Pre-existing pattern across codebase.

### DEFER-005: Unused invoice number in response
**Source:** edge
**Severity:** LOW
**Detail**: If auto-generation added later, returned invoice differs from requested. Future consideration.

### DEFER-006: Missing validation for inactive supplier in Update
**Source:** edge
**Severity:** LOW
**Detail:** Supplier validation exists but race condition possible. Pre-existing supplier management pattern.

### DEFER-007: XSS in invoice number display
**Source:** blind
**Severity:** MEDIUM
**Detail:** API returns JSON without HTML escaping. Web frontend responsibility. Not backend issue.

### DEFER-008: Missing audit log failure monitoring
**Source:** edge
**Severity:** LOW
**Detail:** Audit failures logged but not monitored. Infrastructure/observability concern.

---

## DISMISS (2)

### DISMISS-001: BranchID=0 validation inconsistent
**Source:** edge
**Severity:** MEDIUM
**Location:** `purchase_invoice_repository_impl.go:34-105`
**Reason:** Repository validates BranchID=0. Service validation would be redundant. Pattern consistent with other entities.

### DISMISS-002: Missing invoice auto-generation
**Source:** edge
**Severity:** LOW
**Reason:** Not in requirements. User provides invoice number. No auto-generation needed.

---

## SUMMARY

| Category | Count | Percentage |
|----------|-------|------------|
| **Decision Needed** | 8 | 22% |
| **Patch** | 22 | 61% |
| **Defer** | 8 | 22% |
| **Dismiss** | 2 | 6% |
| **Total** | 40 | 100% |

### Severity Breakdown (Patch + Decision Needed)

| Severity | Count |
|----------|-------|
| CRITICAL | 3 |
| HIGH | 10 |
| MEDIUM | 13 |
| LOW | 4 |

### Key Issues

**CRITICAL:**
1. Line items not persisted to database
2. Float overflow in total calculation
3. Date range inversion not validated

**HIGH:**
1. SQL injection via sort order
2. Race condition in duplicate check
3. Missing branch access authorization
4. Zero SupplierID validation
5. Empty invoice number after normalization
6. Payment status enum validation
7. Negative total amount validation
8. Malformed date string injection
9. Float overflow in pagination
10. Information leakage in errors

**ACCEPTANCE CRITERIA GAPS:**
1. Missing supplier contact information (AC3)
2. Missing payment history (AC3)
3. Missing branch information in response (AC3)

---

## NEXT STEPS

1. **Address Decision Needed items** - User must clarify intent for 8 architectural decisions
2. **Patch critical issues** - 3 CRITICAL and 10 HIGH priority issues require immediate fixes
3. **Validate acceptance criteria** - Ensure AC3 is fully satisfied
4. **Address medium/low issues** - 17 remaining patch items for code quality
