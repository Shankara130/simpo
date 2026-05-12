-- Migration: create_transactions_table (rollback)
-- Description: Drops transactions table and associated indexes

BEGIN;

DROP TRIGGER IF EXISTS trigger_transactions_updated_at ON transactions;
DROP FUNCTION IF EXISTS update_transactions_updated_at();
DROP INDEX IF EXISTS idx_transactions_deleted_at;
DROP INDEX IF EXISTS idx_transactions_status;
DROP INDEX IF EXISTS idx_transactions_created_at;
DROP INDEX IF EXISTS idx_transactions_branch;
DROP INDEX IF EXISTS idx_transactions_cashier;
DROP INDEX IF EXISTS idx_transactions_number;
DROP TABLE IF EXISTS transactions;

COMMIT;
