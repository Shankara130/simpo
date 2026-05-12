-- Migration: create_branches_table (rollback)
-- Description: Drops branches table and associated indexes

BEGIN;

DROP TRIGGER IF EXISTS trigger_branches_updated_at ON branches;
DROP FUNCTION IF EXISTS update_branches_updated_at();
DROP INDEX IF EXISTS idx_branches_email;
DROP INDEX IF EXISTS idx_branches_name;
DROP TABLE IF EXISTS branches;

COMMIT;
