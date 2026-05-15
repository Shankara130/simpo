# Database Migration Guide - Add Idempotency Key

**Migration:** `20260515120000_add_idempotency_key_to_transactions`
**Date:** 2026-05-15
**Purpose:** CRITICAL-003 Fix - Prevent duplicate transaction charges

---

## Overview

This migration adds an `idempotency_key` column to the `transactions` table to prevent duplicate charges when network retries occur.

---

## Migration Files

- **Up:** `20260515120000_add_idempotency_key_to_transactions.up.sql`
- **Down:** `20260515120000_add_idempotency_key_to_transactions.down.sql`

---

## How It Works

1. **Client generates UUID** - Mobile app generates UUID v4 for each payment attempt
2. **Backend checks for existing** - Before creating transaction, backend checks if key exists
3. **Idempotent response** - If key exists, return existing transaction instead of creating new one
4. **Database enforcement** - Unique index ensures no duplicate keys at database level

---

## Pre-Migration Checklist

- [ ] Review migration files (up/down SQL)
- [ ] Backup production database
- [ ] Schedule maintenance window (5-10 minutes expected downtime)
- [ ] Ensure backend code is deployed with idempotency support
- [ ] Ensure mobile app is updated to send idempotency_key

---

## Running the Migration

### Development

```bash
cd apps/backend

# Using golang-migrate
migrate -path migrations -database "postgresql://user:pass@localhost:5432/simpo?sslmode=disable" up 1

# Using migrate tool (if configured)
go run cmd/migrate/main.go up
```

### Using Go Migration Library

If the project uses a Go-based migration runner:

```bash
cd apps/backend
go run cmd/migrate/main.go
```

### Manual SQL Execution

For immediate execution:

```bash
psql -U your_user -d simpo -f migrations/20260515120000_add_idempotency_key_to_transactions.up.sql
```

---

## Migration Steps

The up migration performs these steps in order:

1. **Add nullable column** - `idempotency_key VARCHAR(255)` (nullable initially)
2. **Create unique index** - `idx_transactions_idempotency_key` (partial index, NULL excluded)
3. **Backfill existing data** - Set `idempotency_key = 'legacy-{id}'` for all existing rows
4. **Add documentation** - Comment on column explaining purpose
5. **Make NOT NULL** - Enforce NOT NULL constraint after backfill
6. **Recreate index** - Full unique index (no partial filter needed)

---

## Verification

After migration, verify:

```sql
-- Check column exists
SELECT column_name, data_type, is_nullable
FROM information_schema.columns
WHERE table_name = 'transactions'
  AND column_name = 'idempotency_key';

-- Expected result:
-- column_name      | data_type | is_nullable
-- -----------------+-----------+-------------
-- idempotency_key  | varchar   | NO

-- Check unique index exists
SELECT indexname, indexdef
FROM pg_indexes
WHERE tablename = 'transactions'
  AND indexname = 'idx_transactions_idempotency_key';

-- Expected result:
-- indexname                        | indexdef
-- ---------------------------------+----------
-- idx_transactions_idempotency_key | CREATE UNIQUE INDEX...

-- Verify all transactions have keys
SELECT COUNT(*) as without_key
FROM transactions
WHERE idempotency_key IS NULL;

-- Expected result: 0

-- Verify legacy format for old transactions
SELECT COUNT(*) as legacy_keys
FROM transactions
WHERE idempotency_key LIKE 'legacy-%';

-- Expected result: All transactions created before this migration
```

---

## Rollback Procedure

### If Issues Occur

1. **Stop the application** - Prevent new transactions while rolling back
2. **Check for recent transactions** - Verify no transactions were created during migration
3. **Run rollback migration:**

```bash
cd apps/backend

# Using golang-migrate
migrate -path migrations -database "postgresql://user:pass@localhost:5432/simpo?sslmode=disable" down 1

# Manual SQL
psql -U your_user -d simpo -f migrations/20260515120000_add_idempotency_key_to_transactions.down.sql
```

4. **Verify rollback** - Check column is removed:

```sql
-- Column should no longer exist
SELECT column_name
FROM information_schema.columns
WHERE table_name = 'transactions'
  AND column_name = 'idempotency_key';

-- Expected: 0 rows
```

5. **Restart application** - Application will work without idempotency (loses duplicate charge protection)

### Safer Rollback (Keep Column Optional)

If you prefer to keep the column but make it optional:

```sql
ALTER TABLE transactions
    ALTER COLUMN idempotency_key DROP NOT NULL;
```

This allows:
- Old mobile clients (without idempotency support) to continue working
- New mobile clients to benefit from idempotency
- Gradual rollout of the feature

---

## Post-Migration Tasks

1. **Update backend code** - Already done in CRITICAL-003 fix
2. **Update mobile app** - Already done in CRITICAL-003 fix
3. **Monitor for errors** - Check logs for unique constraint violations
4. **Verify idempotency works** - Test network retry scenario
5. **Update documentation** - Document idempotency behavior for developers

---

## Potential Issues and Solutions

### Issue 1: Unique Constraint Violation During Backfill

**Symptom:** Migration fails with "duplicate key value violates unique constraint"

**Cause:** Legacy transactions with same ID collision (unlikely)

**Solution:**
```sql
-- Find and fix duplicates
DELETE FROM transactions
WHERE id IN (
    SELECT MIN(id)
    FROM transactions
    GROUP BY id
    HAVING COUNT(*) > 1
);
```

### Issue 2: Mobile App Still Sending Requests Without Key

**Symptom:** Backend returns 400 error: "idempotency_key cannot be empty"

**Solution:** Make column temporarily nullable:
```sql
ALTER TABLE transactions ALTER COLUMN idempotency_key DROP NOT NULL;
```

Then gradually roll out mobile app update.

### Issue 3: Performance Impact

**Symptom:** Transaction creation slower than before

**Cause:** Additional index lookup for idempotency check

**Solution:** Index is on VARCHAR(255) - should be minimal impact. If issues:
- Monitor query performance with `EXPLAIN ANALYZE`
- Consider hash index instead of b-tree (PostgreSQL specific)

---

## Testing

### Before Production Deployment

Test on staging environment:

```sql
-- Test 1: Insert with idempotency key
INSERT INTO transactions (
    transaction_number, cashier_id, branch_id, total, subtotal,
    payment_method, idempotency_key
) VALUES (
    'TRX-20260515-TEST-001', 1, 1, 100.00, 100.00,
    'CASH', 'test-uuid-12345'
);

-- Test 2: Try duplicate (should fail)
INSERT INTO transactions (
    transaction_number, cashier_id, branch_id, total, subtotal,
    payment_method, idempotency_key
) VALUES (
    'TRX-20260515-TEST-002', 1, 1, 200.00, 200.00,
    'CASH', 'test-uuid-12345'
);
-- Expected: ERROR: duplicate key value violates unique constraint

-- Test 3: Backend idempotency check (via API)
-- First request: Creates transaction
-- Second request with same key: Returns existing transaction
```

---

## Monitoring

After deployment, monitor:

1. **Unique constraint violations** - Should be zero (mobile client generates unique UUIDs)
2. **Transaction creation time** - Should add <5ms per transaction
3. **Failed transactions** - Should not increase due to idempotency

Queries:
```sql
-- Check for constraint violations in logs
grep "duplicate key" /var/log/postgresql/postgresql.log

-- Monitor transaction performance
SELECT avg(execution_time) as avg_time
FROM pg_stat_statements
WHERE query LIKE '%INSERT INTO transactions%';
```

---

## Additional Resources

- **CRITICAL-003 Fix:** `patches/CRITICAL-003-add-idempotency.patch`
- **Applied Fixes:** `patches/APPLIED_FIXES_SUMMARY.md`
- **PostgreSQL Documentation:** https://www.postgresql.org/docs/current/ddl-constraints.html

---

## Contact

For questions about this migration:
- Review the patch documentation in `patches/CRITICAL-003-add-idempotency.patch`
- Check the story file: `_bmad-output/implementation-artifacts/3-6-implement-transaction-processing-30-seconds.md`
