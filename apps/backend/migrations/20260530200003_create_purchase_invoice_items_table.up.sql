-- Story 10.2: Create purchase_invoice_items table for line items in purchase invoices
-- AC1: Line items with product, quantity, unit cost, subtotal

CREATE TABLE purchase_invoice_items (
    id SERIAL PRIMARY KEY,
    purchase_invoice_id INTEGER NOT NULL REFERENCES purchase_invoices(id) ON DELETE CASCADE,
    product_id INTEGER NOT NULL REFERENCES products(id) ON DELETE RESTRICT,
    quantity INTEGER NOT NULL,
    unit_cost DECIMAL(15,2) NOT NULL,
    subtotal DECIMAL(15,2) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Index on purchase_invoice_id for invoice lookups
CREATE INDEX idx_purchase_invoice_items_invoice_id ON purchase_invoice_items(purchase_invoice_id);

-- Index on product_id for product analysis queries
CREATE INDEX idx_purchase_invoice_items_product_id ON purchase_invoice_items(product_id);

-- Trigger to automatically update updated_at timestamp
CREATE OR REPLACE FUNCTION update_purchase_invoice_items_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$body$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_purchase_invoice_items_updated_at
    BEFORE UPDATE ON purchase_invoice_items
    FOR EACH ROW
    EXECUTE FUNCTION update_purchase_invoice_items_updated_at();

-- Add check constraint for quantity positive
ALTER TABLE purchase_invoice_items
    ADD CONSTRAINT chk_purchase_invoice_items_quantity_positive
    CHECK (quantity > 0);

-- Add check constraint for unit_cost non-negative
ALTER TABLE purchase_invoice_items
    ADD CONSTRAINT chk_purchase_invoice_items_unit_cost_positive
    CHECK (unit_cost >= 0);

-- Add check constraint for subtotal non-negative
ALTER TABLE purchase_invoice_items
    ADD CONSTRAINT chk_purchase_invoice_items_subtotal_positive
    CHECK (subtotal >= 0);

-- Documentation comments
COMMENT ON TABLE purchase_invoice_items IS 'Line items for purchase invoices';
COMMENT ON COLUMN purchase_invoice_items.id IS 'Primary key';
COMMENT ON COLUMN purchase_invoice_items.purchase_invoice_id IS 'Reference to purchase invoice (CASCADE delete)';
COMMENT ON COLUMN purchase_invoice_items.product_id IS 'Reference to product (RESTRICT delete)';
COMMENT ON COLUMN purchase_invoice_items.quantity IS 'Quantity purchased (must be > 0)';
COMMENT ON COLUMN purchase_invoice_items.unit_cost IS 'Unit cost per item (must be >= 0)';
COMMENT ON COLUMN purchase_invoice_items.subtotal IS 'Line item total (quantity * unit_cost)';
COMMENT ON COLUMN purchase_invoice_items.created_at IS 'Timestamp when line item was created';
COMMENT ON COLUMN purchase_invoice_items.updated_at IS 'Timestamp when line item was last updated (auto-updated)';
