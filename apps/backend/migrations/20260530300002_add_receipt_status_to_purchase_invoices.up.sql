-- +migrate Up
-- Story 10.3: Add receipt tracking columns to purchase_invoices table
-- This migration adds fields to track whether an invoice has been received

-- Add receipt_status column to track invoice receipt status
ALTER TABLE purchase_invoices
ADD COLUMN receipt_status VARCHAR(20) NOT NULL DEFAULT 'pending';

-- Add check constraint for valid receipt status values
ALTER TABLE purchase_invoices
ADD CONSTRAINT check_receipt_status
CHECK (receipt_status IN ('pending', 'received', 'partial'));

-- Add goods_receipt_id foreign key column (nullable, set when invoice is received)
ALTER TABLE purchase_invoices
ADD COLUMN goods_receipt_id INTEGER REFERENCES goods_receipts(id) ON DELETE SET NULL;

-- Create index for filtering by receipt status
CREATE INDEX idx_purchase_invoices_receipt_status ON purchase_invoices(receipt_status);

-- Create index for goods_receipt_id lookups
CREATE INDEX idx_purchase_invoices_goods_receipt_id ON purchase_invoices(goods_receipt_id);

-- Add comments for documentation
COMMENT ON COLUMN purchase_invoices.receipt_status IS 'Receipt status: pending (not received), received (fully received), partial (partially received)';
COMMENT ON COLUMN purchase_invoices.goods_receipt_id IS 'Foreign key to goods_receipts table when invoice has been received';
