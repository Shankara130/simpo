-- +migrate Down
-- Drop supplier_product_catalogs table and related indexes

-- Drop the table (indexes will be dropped automatically)
DROP TABLE IF EXISTS supplier_product_catalogs;
