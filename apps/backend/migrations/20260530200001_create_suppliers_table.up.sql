-- Migration: create_suppliers_table
-- Description: Creates suppliers table for supplier master data management
-- Story: 10.1 - Implement Supplier Master Data Management

BEGIN;

CREATE TABLE IF NOT EXISTS suppliers (
    id SERIAL PRIMARY KEY,
    name VARCHAR(200) NOT NULL,
    contact_person VARCHAR(100),
    phone VARCHAR(20) NOT NULL,
    email VARCHAR(100),
    address VARCHAR(500),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by INTEGER,
    updated_by INTEGER,
    deleted_by INTEGER,
    version INTEGER NOT NULL DEFAULT 1,
    deleted_at TIMESTAMP WITH TIME ZONE,
    CONSTRAINT check_supplier_name_not_empty CHECK (TRIM(name) <> ''),
    CONSTRAINT check_supplier_name_max_length CHECK (CHAR_LENGTH(name) <= 200),
    CONSTRAINT check_supplier_phone_not_empty CHECK (TRIM(phone) <> ''),
    CONSTRAINT check_supplier_phone_format CHECK (phone ~* '^[0-9][0-9+()\- ]{9,19}$'),
    CONSTRAINT check_supplier_email_format CHECK (email IS NULL OR email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$'),
    CONSTRAINT fk_suppliers_created_by FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL,
    CONSTRAINT fk_suppliers_updated_by FOREIGN KEY (updated_by) REFERENCES users(id) ON DELETE SET NULL,
    CONSTRAINT fk_suppliers_deleted_by FOREIGN KEY (deleted_by) REFERENCES users(id) ON DELETE SET NULL
);

-- Unique index on supplier name
CREATE UNIQUE INDEX IF NOT EXISTS idx_suppliers_name ON suppliers(name) WHERE deleted_at IS NULL;

-- Index on phone for search performance
CREATE INDEX IF NOT EXISTS idx_suppliers_phone ON suppliers(phone) WHERE deleted_at IS NULL;

-- Index on email for lookups
CREATE INDEX IF NOT EXISTS idx_suppliers_email ON suppliers(email) WHERE deleted_at IS NULL;

-- Index on deleted_at for soft delete queries
CREATE INDEX IF NOT EXISTS idx_suppliers_deleted_at ON suppliers(deleted_at);

-- Function to update updated_at timestamp and version
CREATE OR REPLACE FUNCTION update_suppliers_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    NEW.version = OLD.version + 1;
    RETURN NEW;
END;
$body$ LANGUAGE plpgsql;

-- Trigger to automatically update updated_at and version
DROP TRIGGER IF EXISTS trigger_suppliers_updated_at ON suppliers;
CREATE TRIGGER trigger_suppliers_updated_at
    BEFORE UPDATE ON suppliers
    FOR EACH ROW
    EXECUTE FUNCTION update_suppliers_updated_at();

-- Documentation comments
COMMENT ON TABLE suppliers IS 'Supplier master data for purchase management';
COMMENT ON COLUMN suppliers.id IS 'Primary key';
COMMENT ON COLUMN suppliers.name IS 'Supplier name (unique across all suppliers)';
COMMENT ON COLUMN suppliers.contact_person IS 'Primary contact person name';
COMMENT ON COLUMN suppliers.phone IS 'Contact phone number (required)';
COMMENT ON COLUMN suppliers.email IS 'Contact email address (validated format)';
COMMENT ON COLUMN suppliers.address IS 'Physical address of the supplier (max 500 characters)';
COMMENT ON COLUMN suppliers.deleted_by IS 'User who deactivated the supplier (references users.id)';
COMMENT ON COLUMN suppliers.is_active IS 'Active status for filtering (soft delete via deleted_at)';
COMMENT ON COLUMN suppliers.created_at IS 'Timestamp when supplier was created';
COMMENT ON COLUMN suppliers.updated_at IS 'Timestamp when supplier was last updated (auto-updated)';
COMMENT ON COLUMN suppliers.created_by IS 'User who created the supplier (references users.id)';
COMMENT ON COLUMN suppliers.updated_by IS 'User who last updated the supplier (references users.id)';
COMMENT ON COLUMN suppliers.version IS 'Optimistic locking version';
COMMENT ON COLUMN suppliers.deleted_at IS 'Soft delete timestamp (NULL for active suppliers)';

COMMIT;
