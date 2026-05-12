-- +migrate Down
-- Story 1.9: Rollback email_verification_tokens table

DROP INDEX IF EXISTS idx_email_verification_tokens_expires_at;
DROP INDEX IF EXISTS idx_email_verification_tokens_email;
DROP INDEX IF EXISTS idx_email_verification_tokens_token;
DROP TABLE IF EXISTS email_verification_tokens;
