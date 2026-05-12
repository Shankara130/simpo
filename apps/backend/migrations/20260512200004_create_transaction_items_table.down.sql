-- Migration: create_transaction_items_table (rollback)
-- Description: Drops transaction_items table and associated indexes

BEGIN;

DROP TRIGGER IF EXISTS trigger_transaction_items_updated_at ON transaction_items;
DROP FUNCTION IF EXISTS update_transaction_items_updated_at();
DROP INDEX IF EXISTS idx_transaction_items_deleted_at;
DROP INDEX IF EXISTS idx_transaction_items_tx_product;
DROP INDEX IF EXISTS idx_transaction_items_product;
DROP INDEX IF EXISTS idx_transaction_items_transaction;
DROP TABLE IF EXISTS transaction_items;

COMMIT;
