CREATE TABLE IF NOT EXISTS orders (
    id BIGINT NOT NULL AUTO_INCREMENT,
    user_id uuid NOT NULL,
    amount integer NOT NULL,
    status tinytext NOT NULL,
    created_at timestamp NOT NULL DEFAULT now(),
    created_by uuid,
    updated_at timestamp NOT NULL DEFAULT now(),
    updated_by uuid,
    PRIMARY KEY (id)
);
CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders(user_id);
CREATE TABLE IF NOT EXISTS order_items (
    id BIGINT NOT NULL AUTO_INCREMENT,
    order_id BIGINT NOT NULL,
    product_id uuid NOT NULL,
    profile_id uuid NOT NULL,
    price integer NOT NULL,
    quantity integer NOT NULL,
    PRIMARY KEY (id),
    FOREIGN KEY (order_id) REFERENCES orders(id),
    FOREIGN KEY (product_id) REFERENCES products(id),
    FOREIGN KEY (profile_id) REFERENCES profiles(id)
);
CREATE INDEX IF NOT EXISTS idx_order_items_order_id ON order_items(order_id);