# Critical Fix Patches - Story 3.6: Transaction Processing

**Date:** 2026-05-15
**Review:** Adversarial Code Review
**Priority:** CRITICAL - Must fix before production deployment

---

## Overview

Three CRITICAL issues were identified during code review that pose significant production risks:

1. **Database Deadlock** - Can cause system-wide freeze under concurrent load
2. **Integer Overflow** - Can corrupt stock data
3. **Missing Idempotency** - Can charge customers twice

All three issues **MUST** be resolved before the story can be merged to production.

---

## Patch Files

| Patch | File | Issue | Severity |
|-------|------|-------|----------|
| `CRITICAL-001-fix-deadlock-lock-ordering.patch` | transaction_repository_impl.go | Database deadlock | CRITICAL |
| `CRITICAL-002-fix-integer-overflow.patch` | transaction_repository_impl.go, transaction_handler.go | Stock data corruption | CRITICAL |
| `CRITICAL-003-add-idempotency.patch` | Multiple files | Duplicate charges | CRITICAL |

---

## Application Instructions

### Option 1: Apply Patches Manually

Each patch file contains detailed diff-style changes. Apply by:
1. Reading the patch file to understand changes
2. Manually applying the changes to each file
3. Running tests to verify

### Option 2: Apply via Git (if patches are in git format)

```bash
cd /Volumes/RX7\ 128GB\ SATA/Project/simpo
git apply patches/CRITICAL-001-fix-deadlock-lock-ordering.patch
git apply patches/CRITICAL-002-fix-integer-overflow.patch
git apply patches/CRITICAL-003-add-idempotency.patch
```

### Option 3: Automated Fix (Recommended)

Run the automated fix script that applies all patches:

```bash
cd /Volumes/RX7\ 128GB\ SATA/Project/simpo
python3 _bmad/scripts/apply_critical_fixes.py
```

---

## Detailed Change Summary

### CRITICAL-001: Fix Database Deadlock

**Problem:** Non-deterministic lock ordering causes deadlocks when multiple cashiers sell overlapping products simultaneously.

**Solution:** Sort products by ID before locking to ensure all transactions acquire locks in the same order.

**Files Changed:**
- `apps/backend/internal/repositories/transaction_repository_impl.go`
  - Add `sort` import
  - Sort stockUpdates by product_id before locking loop

**Lines Changed:** ~20 lines

**Risk:** Low - pure refactor, no API changes

---

### CRITICAL-002: Fix Integer Overflow

**Problem:** Stock calculation can underflow, corrupting inventory data.

**Solution:**
1. Use int64 for all stock arithmetic
2. Check for negative result before assignment
3. Add defensive underflow detection
4. Add quantity validation at handler level

**Files Changed:**
- `apps/backend/internal/repositories/transaction_repository_impl.go`
  - Change stock validation to use int64
  - Add underflow detection
- `apps/backend/internal/handlers/transaction_handler.go`
  - Add quantity validation (1-10000 range)

**Lines Changed:** ~15 lines

**Risk:** Low - defensive checks only

---

### CRITICAL-003: Add Idempotency

**Problem:** Network retries can create duplicate transactions, charging customers twice.

**Solution:** Implement idempotency keys that uniquely identify each payment attempt.

**Files Changed:**

**Backend:**
- `apps/backend/internal/models/transaction.go` - Add IdempotencyKey field
- `apps/backend/internal/services/transaction_service.go` - Add to SaleRequest
- `apps/backend/internal/services/transaction_service_impl.go` - Check for existing transaction
- `apps/backend/internal/repositories/transaction_repository.go` - Add GetByIdempotencyKey to interface
- `apps/backend/internal/repositories/transaction_repository_impl.go` - Implement GetByIdempotencyKey

**Mobile:**
- `apps/mobile/src/features/pos/types/transaction.types.ts` - Add to SaleRequest
- `apps/mobile/src/features/pos/services/TransactionService.ts` - Generate UUID

**Infrastructure:**
- Database migration to add idempotency_key column with unique index
- Mobile dependency: `npm install uuid @types/uuid`

**Lines Changed:** ~100 lines across 8 files

**Risk:** Medium - requires database migration and dependency addition

---

## Testing Requirements

After applying patches, run:

### Backend Tests
```bash
cd apps/backend
go test ./internal/repositories/...
go test ./internal/handlers/...
go test ./internal/services/...
```

### Mobile Tests
```bash
cd apps/mobile
npm test
```

### Integration Tests
```bash
cd apps/backend
go test ./tests/integration/...
```

### Manual Verification
1. Test concurrent transactions (2+ cashiers selling same products)
2. Test network retry scenario (disconnect network during payment)
3. Test edge cases (max quantities, boundary values)

---

## Rollback Plan

If issues occur after applying patches:

### For CRITICAL-001 and CRITICAL-002
- Revert changes to transaction_repository_impl.go
- Revert changes to transaction_handler.go (CRITICAL-002 only)
- No data migration needed

### For CRITICAL-003
- Make idempotency_key column nullable in database
- Remove NOT NULL constraint temporarily
- Mobile clients can gracefully handle missing keys

---

## Verification Checklist

After applying all patches, verify:

- [ ] All unit tests pass
- [ ] All integration tests pass
- [ ] Manual test: Concurrent transactions don't deadlock
- [ ] Manual test: Stock values remain valid after failed transactions
- [ ] Manual test: Network retry doesn't create duplicate charges
- [ ] Database migration executed successfully
- [ ] Mobile uuid dependency installed
- [ ] Code review approved

---

## Additional Notes

### Performance Impact
- CRITICAL-001: O(n log n) sort where n = cart size (typically <20 items) - negligible
- CRITICAL-002: Minimal - same number of database queries
- CRITICAL-003: One additional indexed query per transaction (~5ms)

### Backwards Compatibility
- CRITICAL-001: Full compatibility
- CRITICAL-002: Full compatibility
- CRITICAL-003: Requires database migration - handle gracefully during deployment

### Monitoring
After deployment, monitor:
- Database deadlock rate (should be 0)
- Stock data anomalies (should be 0)
- Duplicate transactions (should be 0)

---

## Contact

If you have questions about these patches:
1. Review the detailed patch files in `/patches/` directory
2. Check the full review findings in `review-findings-consolidated.md`
3. Consult the story specification in `_bmad-output/implementation-artifacts/3-6-implement-transaction-processing-30-seconds.md`
