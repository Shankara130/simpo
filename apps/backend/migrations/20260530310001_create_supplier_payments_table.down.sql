-- +migrate Down
-- Drop supplier_payments table and its indexes

DROP INDEX IF EXISTS idx_supplier_payments_branch_id;
DROP INDEX IF EXISTS idx_supplier_payments_payment_date;
DROP INDEX IF EXISTS idx_supplier_payments_invoice_id;
ALTER TABLE supplier_payments DROP CONSTRAINT IF EXISTS chk_supplier_payments_payment_method_valid;
ALTER TABLE supplier_payments DROP CONSTRAINT IF EXISTS chk_supplier_payments_payment_amount_positive;
DROP TABLE IF EXISTS supplier_payments;
