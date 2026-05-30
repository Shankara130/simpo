-- Story 10.2: Create purchase_invoices table for recording supplier purchase invoices
-- AC1: Invoice recording with supplier, date, items, total amount, payment status

CREATE TABLE purchase_invoices (
    id SERIAL PRIMARY KEY,
    invoice_number VARCHAR(100) NOT NULL UNIQUE,
    invoice_date DATE NOT NULL,
    supplier_id INTEGER NOT NULL REFERENCES suppliers(id) ON DELETE RESTRICT,
    total_amount DECIMAL(15,2) NOT NULL DEFAULT 0.00,
    payment_status VARCHAR(20) NOT NULL DEFAULT 'unpaid',
    notes TEXT,
    document_url VARCHAR(255),
    branch_id INTEGER NOT NULL REFERENCES branches(id) ON DELETE RESTRICT,
    created_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
    updated_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
    version INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- Unique index on invoice_number for duplicate prevention
CREATE UNIQUE INDEX idx_purchase_invoices_invoice_number ON purchase_invoices(invoice_number);

-- Index on supplier_id for filtering by supplier
CREATE INDEX idx_purchase_invoices_supplier_id ON purchase_invoices(supplier_id);

-- Index on invoice_date for date range queries
CREATE INDEX idx_purchase_invoices_invoice_date ON purchase_invoices(invoice_date);

-- Index on payment_status for filtering by payment status
CREATE INDEX idx_purchase_invoices_payment_status ON purchase_invoices(payment_status);

-- Index on branch_id for multi-branch support
CREATE INDEX idx_purchase_invoices_branch_id ON purchase_invoices(branch_id);

-- Index on deleted_at for soft delete queries
CREATE INDEX idx_purchase_invoices_deleted_at ON purchase_invoices(deleted_at);

-- Trigger to automatically update updated_at timestamp
CREATE OR REPLACE FUNCTION update_purchase_invoices_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_purchase_invoices_updated_at
    BEFORE UPDATE ON purchase_invoices
    FOR EACH ROW
    EXECUTE FUNCTION update_purchase_invoices_updated_at();

-- Trigger to automatically increment version on update
CREATE OR REPLACE FUNCTION increment_purchase_invoices_version()
RETURNS TRIGGER AS $$
BEGIN
    NEW.version = OLD.version + 1;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_increment_purchase_invoices_version
    BEFORE UPDATE ON purchase_invoices
    FOR EACH ROW
    WHEN (OLD.version IS NOT NULL)
    EXECUTE FUNCTION increment_purchase_invoices_version();

-- Add check constraint for payment_status values
ALTER TABLE purchase_invoices
    ADD CONSTRAINT chk_purchase_invoices_payment_status
    CHECK (payment_status IN ('unpaid', 'partial', 'paid'));

-- Add check constraint for total_amount non-negative
ALTER TABLE purchase_invoices
    ADD CONSTRAINT chk_purchase_invoices_total_amount_positive
    CHECK (total_amount >= 0);

-- Add check constraint for invoice_date not in future
ALTER TABLE purchase_invoices
    ADD CONSTRAINT chk_purchase_invoices_invoice_date_not_future
    CHECK (invoice_date <= CURRENT_DATE);
