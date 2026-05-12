-- +migrate Down
-- Story 1.9: Rollback email_whitelist table

DROP TRIGGER IF EXISTS trigger_update_email_whitelist_updated_at ON email_whitelist;
DROP FUNCTION IF EXISTS update_email_whitelist_updated_at();
DROP INDEX IF EXISTS idx_email_whitelist_domain;
DROP TABLE IF EXISTS email_whitelist;
