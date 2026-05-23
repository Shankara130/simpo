-- Migration: add_report_performance_indexes (rollback)
-- Description: Remove composite indexes added for report performance

BEGIN;

-- Remove composite index on (created_at, branch_id)
DROP INDEX IF EXISTS idx_transactions_created_branch;

COMMIT;
