-- +migrate Up
-- Story 10.3: Create goods_receipts table for tracking supplier goods receipts
-- This table records when goods are received from suppliers, linking to purchase invoices

CREATE TABLE goods_receipts (
    id SERIAL PRIMARY KEY,
    purchase_invoice_id INTEGER NOT NULL UNIQUE REFERENCES purchase_invoices(id) ON DELETE RESTRICT,
    received_date DATE NOT NULL,
    received_by INTEGER NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    notes TEXT,
    branch_id INTEGER NOT NULL REFERENCES branches(id) ON DELETE RESTRICT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT check_one_receipt_per_invoice UNIQUE (purchase_invoice_id)
);

-- Index for efficient lookups by purchase invoice
CREATE INDEX idx_goods_receipts_purchase_invoice_id ON goods_receipts(purchase_invoice_id);

-- Index for date range queries (reporting and filtering)
CREATE INDEX idx_goods_receipts_received_date ON goods_receipts(received_date);

-- Index for branch-level filtering (multi-branch support)
CREATE INDEX idx_goods_receipts_branch_id ON goods_receipts(branch_id);

-- Index for received_by lookups (audit trails and user activity)
CREATE INDEX idx_goods_receipts_received_by ON goods_receipts(received_by);

-- Comment for documentation
COMMENT ON TABLE goods_receipts IS 'Records goods receipts from suppliers, linking to purchase invoices and updating stock quantities and cost prices';
COMMENT ON COLUMN goods_receipts.purchase_invoice_id IS 'Reference to the purchase invoice being received (one-to-one relationship)';
COMMENT ON COLUMN goods_receipts.received_date IS 'Date when goods were physically received from supplier';
COMMENT ON COLUMN goods_receipts.received_by IS 'User who processed the goods receipt';
COMMENT ON COLUMN goods_receipts.notes IS 'Optional notes about the goods receipt (quality issues, discrepancies, etc.)';
COMMENT ON COLUMN goods_receipts.branch_id IS 'Branch where goods were received (for multi-branch support)';
