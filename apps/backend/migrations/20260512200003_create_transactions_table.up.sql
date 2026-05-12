-- Migration: create_transactions_table
-- Description: Creates transactions table for sales tracking

BEGIN;

CREATE TABLE IF NOT EXISTS transactions (
    id SERIAL PRIMARY KEY,
    transaction_number VARCHAR(50) NOT NULL,
    cashier_id INTEGER NOT NULL,
    branch_id INTEGER NOT NULL,
    total DECIMAL(12,2) NOT NULL,
    subtotal DECIMAL(12,2) NOT NULL,
    tax DECIMAL(12,2) DEFAULT 0,
    discount DECIMAL(12,2) DEFAULT 0,
    payment_method VARCHAR(20) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'COMPLETED',
    customer_name VARCHAR(100),
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    created_by INTEGER,
    updated_by INTEGER,
    version INTEGER NOT NULL DEFAULT 1,
    CONSTRAINT check_total_positive CHECK (total >= 0),
    CONSTRAINT check_subtotal_positive CHECK (subtotal >= 0),
    CONSTRAINT check_tax_non_negative CHECK (tax >= 0),
    CONSTRAINT check_discount_non_negative CHECK (discount >= 0),
    CONSTRAINT check_transaction_number_format CHECK (transaction_number ~* '^TRX-[0-9]{8}-[0-9]+-[0-9]{4}$'),
    CONSTRAINT check_payment_method_enum CHECK (payment_method IN ('CASH', 'TRANSFER', 'E_WALLET', 'CARD', 'QRIS')),
    CONSTRAINT check_status_enum CHECK (status IN ('PENDING', 'COMPLETED', 'CANCELLED', 'REFUNDED')),
    CONSTRAINT check_customer_name_trimmed CHECK (customer_name IS NULL OR TRIM(customer_name) = customer_name),
    CONSTRAINT transactions_cashier_id_fkey FOREIGN KEY (cashier_id)
        REFERENCES users(id)
        ON DELETE RESTRICT,
    CONSTRAINT transactions_branch_id_fkey FOREIGN KEY (branch_id)
        REFERENCES branches(id)
        ON DELETE CASCADE
);

-- Unique index on transaction number
CREATE UNIQUE INDEX IF NOT EXISTS idx_transactions_number ON transactions(transaction_number);

-- Index on cashier_id for cashier history
CREATE INDEX IF NOT EXISTS idx_transactions_cashier ON transactions(cashier_id);

-- Index on branch_id for branch reporting
CREATE INDEX IF NOT EXISTS idx_transactions_branch ON transactions(branch_id);

-- Index on created_at for date-based reporting
CREATE INDEX IF NOT EXISTS idx_transactions_created_at ON transactions(created_at);

-- Index on status for filtering
CREATE INDEX IF NOT EXISTS idx_transactions_status ON transactions(status);

-- Index on deleted_at for soft deletes
CREATE INDEX IF NOT EXISTS idx_transactions_deleted_at ON transactions(deleted_at);

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_transactions_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    NEW.version = OLD.version + 1;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger to automatically update updated_at and version
DROP TRIGGER IF EXISTS trigger_transactions_updated_at ON transactions;
CREATE TRIGGER trigger_transactions_updated_at
    BEFORE UPDATE ON transactions
    FOR EACH ROW
    EXECUTE FUNCTION update_transactions_updated_at();

-- Documentation comments
COMMENT ON TABLE transactions IS 'Sales transactions with payment and status tracking';
COMMENT ON COLUMN transactions.id IS 'Primary key';
COMMENT ON COLUMN transactions.transaction_number IS 'Human-readable transaction ID (unique)';
COMMENT ON COLUMN transactions.cashier_id IS 'User who processed the sale (foreign key to users)';
COMMENT ON COLUMN transactions.branch_id IS 'Branch location (foreign key to branches)';
COMMENT ON COLUMN transactions.total IS 'Final amount paid (subtotal + tax - discount)';
COMMENT ON COLUMN transactions.subtotal IS 'Amount before tax and discount';
COMMENT ON COLUMN transactions.tax IS 'Tax amount';
COMMENT ON COLUMN transactions.discount IS 'Discount amount';
COMMENT ON COLUMN transactions.payment_method IS 'Payment method: CASH, TRANSFER, E_WALLET, CARD, QRIS (enforced)';
COMMENT ON COLUMN transactions.status IS 'Transaction status: PENDING, COMPLETED, CANCELLED, REFUNDED (enforced)';
COMMENT ON COLUMN transactions.customer_name IS 'Optional customer information (trimmed)';
COMMENT ON COLUMN transactions.notes IS 'Transaction notes';
COMMENT ON COLUMN transactions.deleted_at IS 'Soft delete timestamp (NULL if active)';
COMMENT ON COLUMN transactions.created_by IS 'User who created the transaction (references users.id)';
COMMENT ON COLUMN transactions.updated_by IS 'User who last updated the transaction (references users.id)';
COMMENT ON COLUMN transactions.version IS 'Optimistic locking version';

COMMIT;
