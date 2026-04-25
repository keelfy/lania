CREATE TABLE IF NOT EXISTS products (
    id uuid NOT NULL DEFAULT UUID_v4(),
    price integer NOT NULL,
    category tinytext NOT NULL,
    metadata json NOT NULL,
    sold_count integer NOT NULL DEFAULT 0,
    created_at timestamp NOT NULL DEFAULT now(),
    updated_at timestamp NOT NULL DEFAULT now(),
    updated_by uuid,
    PRIMARY KEY (id)
);
CREATE INDEX IF NOT EXISTS idx_products_category ON products(category);
CREATE TABLE IF NOT EXISTS product_localizations (
    product_id uuid NOT NULL,
    locale tinytext NOT NULL,
    name tinytext NOT NULL,
    description text NOT NULL,
    PRIMARY KEY (product_id, locale(5)),
    FOREIGN KEY (product_id) REFERENCES products(id)
);
CREATE INDEX IF NOT EXISTS idx_product_localizations_locale ON product_localizations(locale(5));