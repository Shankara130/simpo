-- +migrate Up
-- Story 1.9: Create email_whitelist table for staff registration via whitelist

CREATE TABLE email_whitelist (
    id BIGSERIAL PRIMARY KEY,
    domain VARCHAR(255) NOT NULL UNIQUE,
    default_role VARCHAR(50) NOT NULL DEFAULT 'CASHIER',
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_default_role CHECK (default_role IN ('SYSTEM_ADMIN', 'OWNER', 'CASHIER'))
);

-- Index for faster domain lookups
CREATE INDEX idx_email_whitelist_domain ON email_whitelist(domain);

-- Trigger to update updated_at timestamp (PostgreSQL specific)
CREATE OR REPLACE FUNCTION update_email_whitelist_updated_at()
RETURNS TRIGGER AS $body$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$body$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_email_whitelist_updated_at
BEFORE UPDATE ON email_whitelist
FOR EACH ROW
EXECUTE FUNCTION update_email_whitelist_updated_at();
