-- Migration: create_products_table (rollback)
-- Description: Drops products table and associated indexes

BEGIN;

DROP TRIGGER IF EXISTS trigger_products_updated_at ON products;
DROP FUNCTION IF EXISTS update_products_updated_at();
DROP INDEX IF EXISTS idx_products_deleted_at;
DROP INDEX IF EXISTS idx_products_branch_id;
DROP INDEX IF EXISTS idx_products_category;
DROP INDEX IF EXISTS idx_products_expiry;
DROP INDEX IF EXISTS idx_products_branch_sku;
DROP TABLE IF EXISTS products;

COMMIT;
