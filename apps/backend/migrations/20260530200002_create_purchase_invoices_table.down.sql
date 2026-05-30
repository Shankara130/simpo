-- Migration: create_purchase_invoices_table
-- Description: Creates purchase_invoices table for recording supplier purchase invoices
-- Story: 10.2 - Implement Purchase Invoice Recording

BEGIN;

-- Drop triggers
DROP TRIGGER IF EXISTS trigger_increment_purchase_invoices_version ON purchase_invoices;
DROP TRIGGER IF EXISTS trigger_update_purchase_invoices_updated_at ON purchase_invoices;

-- Drop functions
DROP FUNCTION IF EXISTS increment_purchase_invoices_version();
DROP FUNCTION IF EXISTS update_purchase_invoices_updated_at();

-- Drop indexes
DROP INDEX IF EXISTS idx_purchase_invoices_deleted_at;
DROP INDEX IF EXISTS idx_purchase_invoices_branch_id;
DROP INDEX IF EXISTS idx_purchase_invoices_payment_status;
DROP INDEX IF EXISTS idx_purchase_invoices_invoice_date;
DROP INDEX IF EXISTS idx_purchase_invoices_supplier_id;
DROP INDEX IF EXISTS idx_purchase_invoices_invoice_number;

-- Drop table
DROP TABLE IF EXISTS purchase_invoices;

COMMIT;
