-- Migration: create_products_table
-- Description: Creates products table for inventory management

BEGIN;

CREATE TABLE IF NOT EXISTS products (
    id SERIAL PRIMARY KEY,
    sku VARCHAR(50) NOT NULL,
    name VARCHAR(200) NOT NULL,
    description TEXT,
    stock_qty BIGINT NOT NULL DEFAULT 0,
    price DECIMAL(15,2) NOT NULL,
    cost_price DECIMAL(15,2),
    expiry_date DATE,
    branch_id INTEGER NOT NULL,
    reorder_threshold INTEGER DEFAULT 10,
    category VARCHAR(50),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    created_by INTEGER,
    updated_by INTEGER,
    version INTEGER NOT NULL DEFAULT 1,
    CONSTRAINT check_stock_not_negative CHECK (stock_qty >= 0),
    CONSTRAINT check_price_positive CHECK (price > 0),
    CONSTRAINT check_sku_not_empty CHECK (TRIM(sku) <> ''),
    CONSTRAINT check_name_not_empty CHECK (TRIM(name) <> ''),
    CONSTRAINT check_reorder_threshold_positive CHECK (reorder_threshold >= 0),
    CONSTRAINT check_category_not_empty_if_set CHECK (category IS NULL OR TRIM(category) <> ''),
    CONSTRAINT products_branch_id_fkey FOREIGN KEY (branch_id)
        REFERENCES branches(id)
        ON DELETE CASCADE
);

-- Compound unique index: SKU is unique per branch
CREATE UNIQUE INDEX IF NOT EXISTS idx_products_branch_sku ON products(branch_id, sku);

-- Index on expiry date for alert queries
CREATE INDEX IF NOT EXISTS idx_products_expiry ON products(expiry_date);

-- Index on category for filtering
CREATE INDEX IF NOT EXISTS idx_products_category ON products(category);

-- Index on branch_id for branch queries
CREATE INDEX IF NOT EXISTS idx_products_branch_id ON products(branch_id);

-- Index on deleted_at for soft deletes
CREATE INDEX IF NOT EXISTS idx_products_deleted_at ON products(deleted_at);

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_products_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    NEW.version = OLD.version + 1;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger to automatically update updated_at and version
DROP TRIGGER IF EXISTS trigger_products_updated_at ON products;
CREATE TRIGGER trigger_products_updated_at
    BEFORE UPDATE ON products
    FOR EACH ROW
    EXECUTE FUNCTION update_products_updated_at();

-- Documentation comments
COMMENT ON TABLE products IS 'Product inventory with stock levels, pricing, and expiry tracking';
COMMENT ON COLUMN products.id IS 'Primary key';
COMMENT ON COLUMN products.sku IS 'Stock keeping unit (not unique globally, unique per branch)';
COMMENT ON COLUMN products.name IS 'Product name';
COMMENT ON COLUMN products.description IS 'Product description/details';
COMMENT ON COLUMN products.stock_qty IS 'Current stock quantity (cannot be negative, BIGINT for bulk items)';
COMMENT ON COLUMN products.price IS 'Selling price (must be positive, DECIMAL(15,2) for expensive medications)';
COMMENT ON COLUMN products.cost_price IS 'Purchase cost price for profit/loss calculations';
COMMENT ON COLUMN products.expiry_date IS 'Expiration date for medication tracking';
COMMENT ON COLUMN products.branch_id IS 'Branch location (foreign key to branches)';
COMMENT ON COLUMN products.reorder_threshold IS 'Stock level triggering low stock alert (must be non-negative)';
COMMENT ON COLUMN products.category IS 'Product category for classification';
COMMENT ON COLUMN products.deleted_at IS 'Soft delete timestamp (NULL if active)';
COMMENT ON COLUMN products.created_by IS 'User who created the product (references users.id)';
COMMENT ON COLUMN products.updated_by IS 'User who last updated the product (references users.id)';
COMMENT ON COLUMN products.version IS 'Optimistic locking version for concurrent stock updates';

COMMIT;
