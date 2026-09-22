DROP INDEX IF EXISTS idx_products_active_category ON products;

ALTER TABLE products DROP COLUMN is_active;
