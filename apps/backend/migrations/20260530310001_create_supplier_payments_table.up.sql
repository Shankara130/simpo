-- +migrate Up
-- Create supplier_payments table for tracking payments to suppliers

CREATE TABLE supplier_payments (
    id SERIAL PRIMARY KEY,
    purchase_invoice_id INTEGER NOT NULL REFERENCES purchase_invoices(id),
    payment_date DATE NOT NULL,
    payment_amount DECIMAL(15,2) NOT NULL,
    payment_method VARCHAR(50) NOT NULL,
    notes TEXT,
    reference_number VARCHAR(100),
    branch_id INTEGER NOT NULL REFERENCES branches(id),
    created_by INTEGER NOT NULL REFERENCES users(id),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for performance
CREATE INDEX idx_supplier_payments_invoice_id ON supplier_payments(purchase_invoice_id);
CREATE INDEX idx_supplier_payments_payment_date ON supplier_payments(payment_date);
CREATE INDEX idx_supplier_payments_branch_id ON supplier_payments(branch_id);

-- Add check constraint for payment amount
ALTER TABLE supplier_payments ADD CONSTRAINT chk_supplier_payments_payment_amount_positive CHECK (payment_amount > 0);

-- Add check constraint for payment method
ALTER TABLE supplier_payments ADD CONSTRAINT chk_supplier_payments_payment_method_valid CHECK (payment_method IN ('cash', 'transfer', 'e-wallet', 'check', 'other'));
