-- Migration: create_purchase_invoice_items_table
-- Description: Creates purchase_invoice_items table for line items in purchase invoices
-- Story: 10.2 - Implement Purchase Invoice Recording

BEGIN;

-- Drop triggers
DROP TRIGGER IF EXISTS trigger_update_purchase_invoice_items_updated_at ON purchase_invoice_items;

-- Drop functions
DROP FUNCTION IF EXISTS update_purchase_invoice_items_updated_at();

-- Drop table (will also drop the check constraints automatically)
DROP TABLE IF EXISTS purchase_invoice_items;

COMMIT;
