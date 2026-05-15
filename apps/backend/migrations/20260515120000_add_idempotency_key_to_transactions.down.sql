-- Migration: add_idempotency_key_to_transactions (rollback)
-- Description: Removes idempotency_key column from transactions table
-- CRITICAL-003: Code Review Fix Rollback
-- Date: 2026-05-15

BEGIN;

-- Step 1: Drop the unique index
DROP INDEX IF EXISTS idx_transactions_idempotency_key;

-- Step 2: Make column nullable (in case NOT NULL constraint needs to be removed first)
ALTER TABLE transactions
    ALTER COLUMN idempotency_key DROP NOT NULL;

-- Step 3: Remove the column
ALTER TABLE transactions
    DROP COLUMN IF EXISTS idempotency_key;

COMMIT;

-- Notes:
-- This rollback removes the idempotency_key column entirely.
-- If you need a safer rollback that preserves the column but makes it optional,
-- stop after Step 2 in the up migration and do not run this down migration.
-- Alternatively, you can modify the rollback to keep the column as nullable:
--
--   -- Safe rollback: Keep column but make optional
--   ALTER TABLE transactions ALTER COLUMN idempotency_key DROP NOT NULL;
--
-- This allows the system to continue functioning even without idempotency,
-- though duplicate charge protection will be lost.
