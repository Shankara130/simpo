# Critical Fixes Integration Tests

**Story 3.6: Transaction Processing**
**Date:** 2026-05-15
**Purpose:** Integration tests for CRITICAL code review fixes

---

## Overview

These tests verify the three CRITICAL fixes from the adversarial code review:

1. **CRITICAL-001:** Database deadlock prevention via deterministic lock ordering
2. **CRITICAL-002:** Integer overflow protection in stock calculation  
3. **CRITICAL-003:** Idempotency to prevent duplicate charges

---

## Test Files

| File | Location | Purpose |
|------|----------|---------|
| Backend Integration Tests | `apps/backend/tests/critical_fixes_integration_test.go` | Go integration tests |
| Mobile Integration Tests | `apps/mobile/src/features/pos/services/CriticalFixesIntegration.test.ts` | Jest integration tests |

---

## Running Backend Tests

### Run All Critical Fixes Tests

```bash
cd apps/backend
go test -v ./tests/critical_fixes_integration_test.go
```

### Run Specific Test

```bash
# Deadlock prevention test
go test -v ./tests/... -run TestCriticalFix001_ConcurrentTransactionsNoDeadlock

# Overflow protection test
go test -v ./tests/... -run TestCriticalFix002

# Idempotency test
go test -v ./tests/... -run TestCriticalFix003
```

### Run Performance Tests

```bash
# Skip performance tests (default)
go test -v ./tests/... -short

# Include performance tests
go test -v ./tests/...
```

---

## Running Mobile Tests

```bash
cd apps/mobile
npm test -- CriticalFixesIntegration
```

---

## Test Coverage

### CRITICAL-001: Deadlock Prevention

**Test:** `TestCriticalFix001_ConcurrentTransactionsNoDeadlock`

- Creates two concurrent transactions with overlapping products in different order
- Cashier 1: Product A → Product B → Product C
- Cashier 2: Product C → Product B → Product A
- Verifies no deadlock occurs with 10-second timeout
- Validates stock is correctly deducted for both transactions

**Expected Result:** Both transactions complete successfully without deadlock

---

### CRITICAL-002: Integer Overflow Protection

**Tests:**
- `TestCriticalFix002_QuantityValidation` - Validates quantity input (1-10000 range)
- `TestCriticalFix002_StockUnderflowPrevention` - Detects and prevents stock underflow

**Scenarios:**
- Normal quantities (5, 10000) - should succeed
- Zero quantity - should fail
- Negative quantity - should fail
- Insufficient stock - should fail and rollback

**Expected Result:** Invalid quantities rejected, stock underflow prevented

---

### CRITICAL-003: Idempotency

**Tests:**
- `TestCriticalFix003_IdempotencyPreventsDuplicateCharges` - Same key returns same transaction
- `TestCriticalFix003_DifferentIdempotencyKeysCreateDifferentTransactions` - Different keys create new transactions

**Scenarios:**
- First request with unique key - creates transaction
- Second request with same key - returns existing transaction (no duplicate charge)
- Different keys - create different transactions

**Expected Result:** Duplicate charges prevented, unique transactions for unique keys

---

### Performance Tests

**Test:** `TestCriticalFixes_Performance_ConcurrentLoad`

- Simulates 5 cashiers making 10 concurrent transactions each (50 total)
- Verifies all transactions complete within 30 seconds
- No deadlocks, data corruption, or stock inconsistencies

**Expected Result:** All 50 transactions succeed without errors

---

## CI Integration

These tests should run in CI on every PR:

```yaml
# .github/workflows/ci.yml
- name: Run critical fixes tests
  run: |
    cd apps/backend
    go test -v ./tests/critical_fixes_integration_test.go -timeout 30s
```

---

## Troubleshooting

### Test Timeout

If tests timeout, it may indicate a deadlock:
1. Check database logs for lock wait timeouts
2. Verify lock ordering is sorted by product_id
3. Check for long-running transactions

### Stock Inconsistencies

If stock deductions are incorrect:
1. Verify int64 arithmetic is used
2. Check for underflow detection
3. Review transaction rollback logic

### Idempotency Failures

If idempotency doesn't work:
1. Verify idempotency_key field exists in database
2. Check unique index is created
3. Verify backend checks for existing keys before creating

---

## Test Data

### Products
| SKU | Name | Price | Initial Stock |
|-----|------|-------|---------------|
| PROD001 | Product A | 10000.00 | 50 |
| PROD002 | Product B | 20000.00 | 30 |
| PROD003 | Product C | 15000.00 | 20 |
| PROD004 | Product D | 50000.00 | 10 |
| PROD005 | Product E | 75000.00 | 5 |

### Branch
- ID: 1
- Name: Test Branch
- Address: 123 Test Street

---

## Passing Criteria

All tests must pass before merging to production:

- ✅ No deadlocks in concurrent transaction test
- ✅ All quantity validation tests pass
- ✅ Stock underflow prevented
- ✅ Idempotency prevents duplicate charges
- ✅ Performance test completes within timeout

---

## Documentation

- **Patches:** `patches/CRITICAL-001-*.patch`, `patches/CRITICAL-002-*.patch`, `patches/CRITICAL-003-*.patch`
- **Applied Fixes:** `patches/APPLIED_FIXES_SUMMARY.md`
- **Migration:** `migrations/MIGRATION_GUIDE.md`
