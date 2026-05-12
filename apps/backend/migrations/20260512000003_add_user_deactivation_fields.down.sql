-- Story 1.10: Remove user deactivation tracking fields (rollback)

-- Drop index first
DROP INDEX IF EXISTS idx_users_deactivated_at;

-- Drop foreign key constraint
ALTER TABLE users DROP CONSTRAINT IF EXISTS fk_users_deactivated_by;

-- Remove columns
ALTER TABLE users DROP COLUMN IF EXISTS deactivation_reason;
ALTER TABLE users DROP COLUMN IF EXISTS deactivated_by;
ALTER TABLE users DROP COLUMN IF EXISTS deactivated_at;
