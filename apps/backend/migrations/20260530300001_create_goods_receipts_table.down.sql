-- +migrate Down
-- Story 10.3: Drop goods_receipts table (rollback migration)

-- Drop indexes first
DROP INDEX IF EXISTS idx_goods_receipts_received_by;
DROP INDEX IF EXISTS idx_goods_receipts_branch_id;
DROP INDEX IF EXISTS idx_goods_receipts_received_date;
DROP INDEX IF EXISTS idx_goods_receipts_purchase_invoice_id;

-- Drop the table
DROP TABLE IF EXISTS goods_receipts;
