ALTER TABLE products
  ADD COLUMN is_active boolean NOT NULL DEFAULT true AFTER price_name;

CREATE INDEX idx_products_active_category ON products(is_active, category);
