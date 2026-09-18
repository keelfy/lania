-- oauth2 integrations table
CREATE TABLE IF NOT EXISTS oauth2_integrations (
    id uuid NOT NULL DEFAULT UUID_v4(),
    service_name text NOT NULL,
    access_token text NOT NULL,
    refresh_token text NOT NULL,
    updated_at timestamp NOT NULL DEFAULT now(),
    PRIMARY KEY (id)
);

INSERT INTO oauth2_integrations (service_name, access_token, refresh_token) VALUES ('donation_alerts', '', '');

-- product_prices table
CREATE TABLE IF NOT EXISTS product_prices (
    name tinytext NOT NULL, 
    currency tinytext NOT NULL,
    amount decimal(10, 2) NOT NULL,
    PRIMARY KEY (name(255), currency(3))
);

-- create product_prices for each currency
INSERT INTO product_prices (name, currency, amount) VALUES ('season_access', 'RUB', 49.00);
INSERT INTO product_prices (name, currency, amount) VALUES ('name_color', 'RUB', 199.00);
INSERT INTO product_prices (name, currency, amount) VALUES ('name_prefix', 'RUB', 199.00);

INSERT INTO product_prices (name, currency, amount) VALUES ('season_access', 'USD', 0.99);
INSERT INTO product_prices (name, currency, amount) VALUES ('name_color', 'USD', 2.99);
INSERT INTO product_prices (name, currency, amount) VALUES ('name_prefix', 'USD', 2.99);

INSERT INTO product_prices (name, currency, amount) VALUES ('season_access', 'EUR', 0.99);
INSERT INTO product_prices (name, currency, amount) VALUES ('name_color', 'EUR', 2.99);
INSERT INTO product_prices (name, currency, amount) VALUES ('name_prefix', 'EUR', 2.99);

INSERT INTO product_prices (name, currency, amount) VALUES ('season_access', 'BRL', 3.49);
INSERT INTO product_prices (name, currency, amount) VALUES ('name_color', 'BRL', 12.99);
INSERT INTO product_prices (name, currency, amount) VALUES ('name_prefix', 'BRL', 12.99);

INSERT INTO product_prices (name, currency, amount) VALUES ('season_access', 'TRY', 24.99);
INSERT INTO product_prices (name, currency, amount) VALUES ('name_color', 'TRY', 99.99);
INSERT INTO product_prices (name, currency, amount) VALUES ('name_prefix', 'TRY', 99.99);

INSERT INTO product_prices (name, currency, amount) VALUES ('season_access', 'PLN', 2.99);
INSERT INTO product_prices (name, currency, amount) VALUES ('name_color', 'PLN', 9.99);
INSERT INTO product_prices (name, currency, amount) VALUES ('name_prefix', 'PLN', 9.99);

-- change static price to price reference
ALTER TABLE products DROP COLUMN price;
ALTER TABLE products ADD COLUMN price_name tinytext;

-- fill price_name for each existing product
UPDATE products SET price_name = 'season_access' WHERE id = '4086268a-34fe-47ec-afab-a014a2f02688';
UPDATE products SET price_name = 'name_color' WHERE id = '21852047-a4e9-4f7d-acfd-9fce898bc03d';
UPDATE products SET price_name = 'name_color' WHERE id = '6509dcd7-6177-4e39-8474-e0304870ff23';
UPDATE products SET price_name = 'name_color' WHERE id = 'f1740c10-850a-41a6-bcd8-d3449846f284';
UPDATE products SET price_name = 'name_color' WHERE id = '1cef33ed-08e5-49c5-9858-dac85872386e';
UPDATE products SET price_name = 'name_color' WHERE id = '6a213c8c-4269-4329-9875-917cb3cfd033';
UPDATE products SET price_name = 'name_color' WHERE id = '068e6c5a-94e1-4927-9eed-f9b02beb9f64';
UPDATE products SET price_name = 'name_color' WHERE id = '8bbf4210-24fa-4382-af93-4da30dfeae1f';
UPDATE products SET price_name = 'name_color' WHERE id = '54075756-222b-4917-9924-696436654254';
UPDATE products SET price_name = 'name_color' WHERE id = '72150255-2649-4339-895b-44142701117d';
UPDATE products SET price_name = 'name_color' WHERE id = 'f555a7d6-c52f-4f22-8183-cdfee77eb59e';
UPDATE products SET price_name = 'name_color' WHERE id = '56ef1088-6ce5-4cb6-9c3d-0c35070def83';
UPDATE products SET price_name = 'name_color' WHERE id = '8f461be1-9320-4d17-96e7-e5215fc52522';
UPDATE products SET price_name = 'name_color' WHERE id = '383c4fc1-5f92-4bae-a4b4-b743922f1eba';
UPDATE products SET price_name = 'name_color' WHERE id = 'bdd07172-ea5f-4a0a-b82a-16776a6b4b5c';
UPDATE products SET price_name = 'name_color' WHERE id = '9765c31a-c5d9-4d0a-9eb7-3c10aec6e66c';
UPDATE products SET price_name = 'name_color' WHERE id = '68a560a1-878b-4cee-b79d-465aea06862b';
UPDATE products SET price_name = 'name_color' WHERE id = '69c9ef9e-f66a-4df7-9b3b-bdb2dbf97834';
UPDATE products SET price_name = 'name_color' WHERE id = '66678f14-e431-41d1-a172-d301564deba0';
UPDATE products SET price_name = 'name_color' WHERE id = '3189113f-9253-4bef-9ce4-0374aff2df7c';
UPDATE products SET price_name = 'name_color' WHERE id = 'de397bbd-782c-4eb2-a59c-f955cc06340e';
UPDATE products SET price_name = 'name_color' WHERE id = '63406c8d-67c2-43c1-a8a7-e7e79b7ac2c3';
UPDATE products SET price_name = 'name_color' WHERE id = 'd9607e73-ad81-46e6-b0af-ee33b5987e16';
UPDATE products SET price_name = 'name_color' WHERE id = '2a65d174-8896-411a-a6dd-8341b3620171';

-- make price_name not nullable
ALTER TABLE products MODIFY COLUMN price_name tinytext NOT NULL;

-- temporary drop order_item_id from profile_name_color_options
ALTER TABLE profile_name_color_options DROP FOREIGN KEY profile_name_color_options_ibfk_3;
ALTER TABLE profile_name_color_options DROP COLUMN order_item_id;

-- drop old order items and orders tables
DROP TABLE IF EXISTS order_items CASCADE;
DROP TABLE IF EXISTS orders CASCADE;

-- create new order tables (bigint -> uuid identity)
CREATE TABLE IF NOT EXISTS orders (
    id uuid NOT NULL DEFAULT UUID_v4(),
    user_id uuid NOT NULL,
    amounts json NOT NULL,
    status tinytext NOT NULL,
    created_at timestamp NOT NULL DEFAULT now(),
    created_by uuid,
    updated_at timestamp NOT NULL DEFAULT now(),
    updated_by uuid,
    PRIMARY KEY (id)
);
CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders(user_id);

-- create new order items table (bigint -> uuid identity)
CREATE TABLE IF NOT EXISTS order_items (
    id uuid NOT NULL DEFAULT UUID_v4(),
    order_id uuid NOT NULL,
    product_id uuid NOT NULL,
    profile_id uuid NOT NULL,
    season_id uuid NOT NULL,
    amounts json NOT NULL,
    quantity integer NOT NULL,
    PRIMARY KEY (id),
    FOREIGN KEY (order_id) REFERENCES orders(id),
    FOREIGN KEY (product_id) REFERENCES products(id),
    FOREIGN KEY (profile_id) REFERENCES profiles(id),
    FOREIGN KEY (season_id) REFERENCES seasons(id)
);
CREATE INDEX IF NOT EXISTS idx_order_items_order_id ON order_items(order_id);

-- add order_item_id to profile_name_color_options
ALTER TABLE profile_name_color_options ADD COLUMN order_item_id uuid;
ALTER TABLE profile_name_color_options ADD FOREIGN KEY (order_item_id) REFERENCES order_items(id);