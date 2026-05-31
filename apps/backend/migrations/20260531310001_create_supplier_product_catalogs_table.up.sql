-- +migrate Up
-- Create supplier_product_catalogs table for tracking supplier product associations and purchase prices
-- This table stores both current and historical price data using effective date ranges

CREATE TABLE supplier_product_catalogs (
    id SERIAL PRIMARY KEY,
    supplier_id INTEGER NOT NULL REFERENCES suppliers(id),
    product_id INTEGER NOT NULL REFERENCES products(id),
    purchase_price DECIMAL(15,2) NOT NULL,
    is_preferred BOOLEAN NOT NULL DEFAULT false,
    sku_code VARCHAR(50),
    minimum_order_quantity INTEGER NOT NULL DEFAULT 1,
    lead_time_days INTEGER,
    branch_id INTEGER NOT NULL REFERENCES branches(id),
    created_by INTEGER NOT NULL REFERENCES users(id),
    updated_by INTEGER REFERENCES users(id),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    price_effective_from DATE NOT NULL DEFAULT CURRENT_DATE,
    price_effective_to DATE
);

-- Create unique partial index for current price (one current price per supplier-product per branch)
-- This ensures only one active price entry per supplier-product combination
CREATE UNIQUE INDEX idx_supplier_product_catalog_current ON supplier_product_catalogs(supplier_id, product_id, price_effective_from)
WHERE price_effective_to IS NULL;

-- Create indexes for performance
CREATE INDEX idx_supplier_product_catalog_product ON supplier_product_catalogs(product_id);
CREATE INDEX idx_supplier_product_catalog_branch ON supplier_product_catalogs(branch_id);
CREATE INDEX idx_supplier_product_catalog_dates ON supplier_product_catalogs(price_effective_from, price_effective_to);

-- Add check constraint for purchase price (must be positive)
ALTER TABLE supplier_product_catalogs ADD CONSTRAINT chk_supplier_product_catalogs_price_positive CHECK (purchase_price > 0);

-- Add check constraint for minimum order quantity (must be at least 1)
ALTER TABLE supplier_product_catalogs ADD CONSTRAINT chk_supplier_product_catalogs_min_qty_positive CHECK (minimum_order_quantity >= 1);

-- Add check constraint for lead time days (must be non-negative if provided)
ALTER TABLE supplier_product_catalogs ADD CONSTRAINT chk_supplier_product_catalogs_lead_time_nonneg CHECK (lead_time_days IS NULL OR lead_time_days >= 0);
