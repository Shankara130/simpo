-- Migration: create_suppliers_table (rollback)
-- Description: Drops suppliers table and associated indexes
-- Story: 10.1 - Implement Supplier Master Data Management

BEGIN;

DROP TRIGGER IF EXISTS trigger_suppliers_updated_at ON suppliers;
DROP FUNCTION IF EXISTS update_suppliers_updated_at();
DROP INDEX IF EXISTS idx_suppliers_deleted_at;
DROP INDEX IF EXISTS idx_suppliers_email;
DROP INDEX IF EXISTS idx_suppliers_phone;
DROP INDEX IF EXISTS idx_suppliers_name;
DROP TABLE IF EXISTS suppliers;

COMMIT;
