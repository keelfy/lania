CREATE TABLE IF NOT EXISTS seasons (
    id uuid NOT NULL DEFAULT UUID_v4(),
    season_number integer NOT NULL,
    start_date timestamp NOT NULL,
    end_date timestamp,
    PRIMARY KEY(id)
);

CREATE TABLE IF NOT EXISTS name_colors (
  id uuid NOT NULL DEFAULT UUID_v4(),
  colors json NOT NULL,
  name tinytext NOT NULL,
  PRIMARY KEY (id)
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_name_colors_name ON name_colors(name);

-- default name color
INSERT INTO name_colors (id, name, colors) VALUES ('2628bf9d-5b7c-438b-900a-67753261a823', 'Default', '{"colors": []}');

CREATE TABLE IF NOT EXISTS profiles (
    id uuid NOT NULL DEFAULT UUID_v4(),
    mc_uuid uuid NOT NULL,
    mc_username tinytext NOT NULL,
    owner_user_id uuid, 
    first_seen_at timestamp,
    last_seen_at timestamp,
    role tinytext NOT NULL,
    is_slim boolean NOT NULL,
    name_color_id uuid NOT NULL DEFAULT '2628bf9d-5b7c-438b-900a-67753261a823',
    created_at timestamp NOT NULL DEFAULT now(),
    updated_at timestamp NOT NULL DEFAULT now(),
    updated_by uuid,
    PRIMARY KEY(id),
    FOREIGN KEY (name_color_id) REFERENCES name_colors(id)
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_profiles_mc_uuid ON profiles(mc_uuid);
CREATE UNIQUE INDEX IF NOT EXISTS idx_profiles_mc_username ON profiles(mc_username);
CREATE INDEX IF NOT EXISTS idx_profiles_owner_user_id ON profiles(owner_user_id);

CREATE TABLE IF NOT EXISTS profile_accesses (
    id uuid NOT NULL DEFAULT UUID_v4(),
    mc_uuid uuid NOT NULL,
    season_id uuid NOT NULL,
    source tinytext NOT NULL,
    created_at timestamp NOT NULL DEFAULT now(),
    updated_at timestamp NOT NULL DEFAULT now(),
    updated_by uuid,
    PRIMARY KEY (id),
    FOREIGN KEY (mc_uuid) REFERENCES profiles(mc_uuid),
    FOREIGN KEY (season_id) REFERENCES seasons(id)
);

CREATE TABLE IF NOT EXISTS profile_playtimes (
    id uuid NOT NULL DEFAULT UUID_v4(),
    mc_uuid uuid NOT NULL,
    season_id uuid NOT NULL,
    playtime bigint NOT NULL,
    updated_at timestamp NOT NULL DEFAULT now(),
    PRIMARY KEY (id),
    FOREIGN KEY (mc_uuid) REFERENCES profiles(mc_uuid),
    FOREIGN KEY (season_id) REFERENCES seasons(id)
);

CREATE TABLE IF NOT EXISTS profile_violations (
    id uuid NOT NULL DEFAULT UUID_v4(),
    mc_uuid uuid NOT NULL,
    season_id uuid NOT NULL,
    violation text NOT NULL,
    updated_at timestamp NOT NULL DEFAULT now(),
    PRIMARY KEY (id),
    FOREIGN KEY (mc_uuid) REFERENCES profiles(mc_uuid),
    FOREIGN KEY (season_id) REFERENCES seasons(id)
);

CREATE TABLE IF NOT EXISTS product_prices (
    name tinytext NOT NULL, 
    currency tinytext NOT NULL,
    amount decimal(10, 2) NOT NULL,
    PRIMARY KEY (name(255), currency(3))
);

CREATE TABLE IF NOT EXISTS products (
    id uuid NOT NULL DEFAULT UUID_v4(),
    category tinytext NOT NULL,
    metadata json NOT NULL,
    sold_count integer NOT NULL DEFAULT 0,
    price_name tinytext NOT NULL,
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

CREATE TABLE IF NOT EXISTS orders (
    id uuid NOT NULL DEFAULT UUID_v4(),
    user_id uuid NOT NULL,
    amounts json NOT NULL,
    status tinytext NOT NULL,
    external_id tinytext,
    created_at timestamp NOT NULL DEFAULT now(),
    created_by uuid,
    updated_at timestamp NOT NULL DEFAULT now(),
    updated_by uuid,
    PRIMARY KEY (id)
);
CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders(user_id);

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

CREATE TABLE IF NOT EXISTS profile_name_color_options (
  id uuid NOT NULL DEFAULT UUID_v4(),
  profile_id uuid NOT NULL,
  name_color_id uuid NOT NULL,
  -- NULL for role color (owner, moderator, etc.)
  order_item_id uuid,
  -- NULL for permanent color
  for_season_id uuid,
  created_at timestamp NOT NULL DEFAULT now(),
  PRIMARY KEY (id),
  FOREIGN KEY (profile_id) REFERENCES profiles(id),
  FOREIGN KEY (name_color_id) REFERENCES name_colors(id),
  FOREIGN KEY (order_item_id) REFERENCES order_items(id),
  FOREIGN KEY (for_season_id) REFERENCES seasons(id)
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_pnco_profile_id_name_color_id_for_season_id ON profile_name_color_options(profile_id, name_color_id, for_season_id);

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

CREATE TABLE IF NOT EXISTS oauth2_integrations (
    id uuid NOT NULL DEFAULT UUID_v4(),
    service_name text NOT NULL,
    access_token text NOT NULL,
    refresh_token text NOT NULL,
    updated_at timestamp NOT NULL DEFAULT now(),
    PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS ed_products (
    product_id uuid NOT NULL,
    ed_product_id bigint NOT NULL,
    PRIMARY KEY (product_id)
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_ed_products_ed_product_id ON ed_products(ed_product_id);

CREATE TABLE IF NOT EXISTS name_prefixes (
  id uuid NOT NULL DEFAULT UUID_v4(),
  name tinytext NOT NULL,
  metadata json NOT NULL,
  PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS profile_name_prefix_options (
  id uuid NOT NULL DEFAULT UUID_v4(),
  profile_id uuid NOT NULL,
  name_prefix_id uuid NOT NULL,
  -- NULL for role prefix (owner, moderator, etc.)
  order_item_id uuid,
  -- NULL for permanent prefix
  for_season_id uuid,
  -- glyth or special
  type tinytext NOT NULL,
  created_at timestamp NOT NULL DEFAULT now(),
  PRIMARY KEY (id),
  FOREIGN KEY (profile_id) REFERENCES profiles(id),
  FOREIGN KEY (name_prefix_id) REFERENCES name_prefixes(id),
  FOREIGN KEY (order_item_id) REFERENCES order_items(id),
  FOREIGN KEY (for_season_id) REFERENCES seasons(id)
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_pnco_profile_id_name_prefix_id_for_season_id ON profile_name_prefix_options(profile_id, name_prefix_id, for_season_id);
CREATE INDEX IF NOT EXISTS idx_pnco_profile_id_type_for_season_id ON profile_name_prefix_options(profile_id, type(50), for_season_id);

CREATE TABLE IF NOT EXISTS profile_prefixes (
  profile_id uuid NOT NULL,
  type tinytext NOT NULL,
  name_prefix_id uuid NOT NULL,
  created_at timestamp NOT NULL DEFAULT now(),
  created_by uuid,
  PRIMARY KEY (profile_id, type(50)),
  FOREIGN KEY (profile_id) REFERENCES profiles(id),
  FOREIGN KEY (name_prefix_id) REFERENCES name_prefixes(id)
);
