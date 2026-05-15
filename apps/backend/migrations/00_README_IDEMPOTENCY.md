# Database Migrations

This directory contains database migrations for the Simpo application.

## Latest Migration

### `20260515120000_add_idempotency_key_to_transactions`

**Date:** 2026-05-15
**Purpose:** CRITICAL-003 Fix - Add idempotency key to prevent duplicate transaction charges

#### What This Migration Does

1. Adds `idempotency_key` column to `transactions` table
2. Creates unique index to enforce uniqueness
3. Backfills existing transactions with `legacy-{id}` format
4. Enforces NOT NULL constraint for new transactions

#### Why This Is Needed

Without idempotency, network retries can create duplicate transactions, charging customers twice. This is a **critical financial bug**.

#### Running the Migration

**Option 1: Using the migration script**
```bash
cd apps/backend/scripts
./run_idempotency_migration.sh
```

**Option 2: Using psql directly**
```bash
cd apps/backend/migrations
psql -U postgres -d simpo -f 20260515120000_add_idempotency_key_to_transactions.up.sql
```

**Option 3: Dry run (see what will be executed)**
```bash
cd apps/backend/scripts
./run_idempotency_migration.sh --dry-run
```

#### Before Running

- ✅ Backend code updated with idempotency support (CRITICAL-003)
- ✅ Mobile app updated to generate UUID
- ⚠️  Backup production database

#### Verification

After migration, verify:
```sql
-- Check column exists
SELECT column_name FROM information_schema.columns
WHERE table_name = 'transactions' AND column_name = 'idempotency_key';

-- Should return: idempotency_key

-- Check no NULL values
SELECT COUNT(*) FROM transactions WHERE idempotency_key IS NULL;

-- Should return: 0
```

#### Rollback

If needed:
```bash
psql -U postgres -d simpo -f 20260515120000_add_idempotency_key_to_transactions.down.sql
```

---

## All Migrations

| Date | Migration | Description |
|------|-----------|-------------|
| 2025-10-25 | `create_users_table` | User accounts and authentication |
| 2025-10-28 | `create_refresh_tokens_table` | JWT refresh tokens |
| 2025-11-22 | `create_roles_table` | Role-based access control |
| 2025-11-22 | `create_user_roles_table` | User-role associations |
| 2026-05-12 | `create_email_whitelist_table` | Email whitelist for registration |
| 2026-05-12 | `create_email_verification_tokens_table` | Email verification |
| 2026-05-12 | `add_user_deactivation_fields` | User account deactivation |
| 2026-05-12 | `create_branches_table` | Branch/location management |
| 2026-05-12 | `create_products_table` | Product catalog |
| 2026-05-12 | `create_transactions_table` | Sales transactions |
| 2026-05-12 | `create_transaction_items_table` | Transaction line items |
| **2026-05-15** | **`add_idempotency_key_to_transactions`** | **Prevent duplicate charges** |

---

## Migration Naming Convention

`YYYYMMDDHHMMSS_description.up.sql` - Apply migration
`YYYYMMDDHHMMSS_description.down.sql` - Rollback migration

Example: `20260515120000_add_idempotency_key_to_transactions.up.sql`

---

## Documentation

For detailed migration information, see:
- `MIGRATION_GUIDE.md` - Complete guide for idempotency migration
- `patches/CRITICAL-003-add-idempotency.patch` - Code changes for idempotency
- `patches/APPLIED_FIXES_SUMMARY.md` - Summary of all critical fixes
