-- Migration: create_transaction_items_table
-- Description: Creates transaction_items table for line items

BEGIN;

CREATE TABLE IF NOT EXISTS transaction_items (
    id SERIAL PRIMARY KEY,
    transaction_id INTEGER NOT NULL,
    product_id INTEGER NOT NULL,
    quantity BIGINT NOT NULL,
    unit_price DECIMAL(12,2) NOT NULL,
    subtotal DECIMAL(12,2) NOT NULL,
    cost_price DECIMAL(12,2),
    product_name VARCHAR(200) NOT NULL,
    product_sku VARCHAR(50) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    created_by INTEGER,
    version INTEGER NOT NULL DEFAULT 1,
    CONSTRAINT check_quantity_positive CHECK (quantity > 0),
    CONSTRAINT check_unit_price_positive CHECK (unit_price >= 0),
    CONSTRAINT check_subtotal_matches_calculation CHECK (ABS(subtotal - ROUND((quantity * unit_price)::numeric, 2)) < 0.01),
    CONSTRAINT check_product_name_not_empty CHECK (TRIM(product_name) <> ''),
    CONSTRAINT check_product_sku_not_empty CHECK (TRIM(product_sku) <> ''),
    CONSTRAINT transaction_items_transaction_id_fkey FOREIGN KEY (transaction_id)
        REFERENCES transactions(id)
        ON DELETE CASCADE,
    CONSTRAINT transaction_items_product_id_fkey FOREIGN KEY (product_id)
        REFERENCES products(id)
        ON DELETE RESTRICT
);

-- Index on transaction_id for finding items in a transaction
CREATE INDEX IF NOT EXISTS idx_transaction_items_transaction ON transaction_items(transaction_id);

-- Index on product_id for product sales history
CREATE INDEX IF NOT EXISTS idx_transaction_items_product ON transaction_items(product_id);

-- Unique index on (transaction_id, product_id) to prevent duplicate products in same transaction
CREATE UNIQUE INDEX IF NOT EXISTS idx_transaction_items_tx_product ON transaction_items(transaction_id, product_id);

-- Index on deleted_at for soft deletes (consistent with transactions table)
CREATE INDEX IF NOT EXISTS idx_transaction_items_deleted_at ON transaction_items(deleted_at);

-- Function to update updated_at timestamp and version
CREATE OR REPLACE FUNCTION update_transaction_items_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    NEW.version = OLD.version + 1;
    RETURN NEW;
END;
$body$ LANGUAGE plpgsql;

-- Trigger to automatically update updated_at and version
DROP TRIGGER IF EXISTS trigger_transaction_items_updated_at ON transaction_items;
CREATE TRIGGER trigger_transaction_items_updated_at
    BEFORE UPDATE ON transaction_items
    FOR EACH ROW
    EXECUTE FUNCTION update_transaction_items_updated_at();

-- Documentation comments
COMMENT ON TABLE transaction_items IS 'Line items within a transaction (snapshot of product data at sale time)';
COMMENT ON COLUMN transaction_items.id IS 'Primary key';
COMMENT ON COLUMN transaction_items.transaction_id IS 'Parent transaction (foreign key to transactions)';
COMMENT ON COLUMN transaction_items.product_id IS 'Product reference (foreign key to products)';
COMMENT ON COLUMN transaction_items.quantity IS 'Quantity sold (must be positive, BIGINT for bulk items)';
COMMENT ON COLUMN transaction_items.unit_price IS 'Price at time of sale (snapshot, DECIMAL(12,2) for consistency)';
COMMENT ON COLUMN transaction_items.subtotal IS 'Line item total (quantity * unit_price, validated)';
COMMENT ON COLUMN transaction_items.cost_price IS 'Cost price at time of sale for profit calculation';
COMMENT ON COLUMN transaction_items.product_name IS 'Snapshot: product name at time of sale';
COMMENT ON COLUMN transaction_items.product_sku IS 'Snapshot: product SKU at time of sale';
COMMENT ON COLUMN transaction_items.updated_at IS 'Timestamp when item was last updated (auto-updated)';
COMMENT ON COLUMN transaction_items.deleted_at IS 'Soft delete timestamp (NULL if active, consistent with transactions)';
COMMENT ON COLUMN transaction_items.created_by IS 'User who created the transaction item (references users.id)';
COMMENT ON COLUMN transaction_items.version IS 'Optimistic locking version';

COMMIT;
