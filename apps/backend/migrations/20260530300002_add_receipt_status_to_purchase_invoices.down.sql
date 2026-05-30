-- +migrate Down
-- Story 10.3: Remove receipt tracking columns from purchase_invoices table (rollback)

-- Drop indexes first
DROP INDEX IF EXISTS idx_purchase_invoices_goods_receipt_id;
DROP INDEX IF EXISTS idx_purchase_invoices_receipt_status;

-- Drop constraints
ALTER TABLE purchase_invoices
DROP CONSTRAINT IF EXISTS check_receipt_status;

-- Drop columns
ALTER TABLE purchase_invoices
DROP COLUMN IF EXISTS goods_receipt_id;

ALTER TABLE purchase_invoices
DROP COLUMN IF EXISTS receipt_status;
