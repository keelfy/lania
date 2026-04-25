CREATE TABLE IF NOT EXISTS basket_items (
    id uuid NOT NULL DEFAULT UUID_v4(),
    user_id uuid NOT NULL,
    product_id uuid NOT NULL,
    profile_id uuid NOT NULL,
    quantity integer NOT NULL,
    created_at timestamp NOT NULL DEFAULT now(),
    created_by uuid NOT NULL,
    updated_at timestamp NOT NULL DEFAULT now(),
    updated_by uuid NOT NULL,
    PRIMARY KEY (id),
    FOREIGN KEY (product_id) REFERENCES products(id),
    FOREIGN KEY (profile_id) REFERENCES profiles(id)
);
CREATE INDEX IF NOT EXISTS idx_basket_items_user_id ON basket_items(user_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_basket_items_user_id_product_id_profile_id ON basket_items(user_id, product_id, profile_id);