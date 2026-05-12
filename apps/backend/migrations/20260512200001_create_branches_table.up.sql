-- Migration: create_branches_table
-- Description: Creates branches table for multi-location pharmacy support

BEGIN;

CREATE TABLE IF NOT EXISTS branches (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    address TEXT,
    phone VARCHAR(20),
    email VARCHAR(100),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by INTEGER,
    updated_by INTEGER,
    version INTEGER NOT NULL DEFAULT 1,
    CONSTRAINT check_name_not_empty CHECK (TRIM(name) <> ''),
    CONSTRAINT check_name_max_length CHECK (CHAR_LENGTH(name) <= 100),
    CONSTRAINT check_email_format CHECK (email IS NULL OR email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$'),
    CONSTRAINT check_phone_format CHECK (phone IS NULL OR phone ~* '^[0-9+()\- ]{10,20}$')
);

-- Unique index on branch name
CREATE UNIQUE INDEX IF NOT EXISTS idx_branches_name ON branches(name);

-- Index on email for lookups
CREATE INDEX IF NOT EXISTS idx_branches_email ON branches(email);

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_branches_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    NEW.version = OLD.version + 1;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger to automatically update updated_at and version
DROP TRIGGER IF EXISTS trigger_branches_updated_at ON branches;
CREATE TRIGGER trigger_branches_updated_at
    BEFORE UPDATE ON branches
    FOR EACH ROW
    EXECUTE FUNCTION update_branches_updated_at();

-- Documentation comments
COMMENT ON TABLE branches IS 'Pharmacy branch locations for multi-tenant support';
COMMENT ON COLUMN branches.id IS 'Primary key';
COMMENT ON COLUMN branches.name IS 'Branch name (unique across all branches)';
COMMENT ON COLUMN branches.address IS 'Physical address of the branch';
COMMENT ON COLUMN branches.phone IS 'Contact phone number';
COMMENT ON COLUMN branches.email IS 'Contact email address (validated format)';
COMMENT ON COLUMN branches.created_at IS 'Timestamp when branch was created';
COMMENT ON COLUMN branches.updated_at IS 'Timestamp when branch was last updated (auto-updated)';
COMMENT ON COLUMN branches.created_by IS 'User who created the branch (references users.id)';
COMMENT ON COLUMN branches.updated_by IS 'User who last updated the branch (references users.id)';
COMMENT ON COLUMN branches.version IS 'Optimistic locking version';

COMMIT;
