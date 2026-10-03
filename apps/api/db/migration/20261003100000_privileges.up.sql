-- A privilege is one LuckPerms permission node a player buys for a season, e.g. homes.commands.*.
-- name is the unique main name, shown in English and wherever a translation is missing.
-- names holds its translations by locale, for now Russian only, e.g. {"ru": "Дом"}.
-- The price lives in product_prices under the tariff name "privilege:<id>".
CREATE TABLE IF NOT EXISTS privileges (
    id uuid NOT NULL DEFAULT UUID_v4(),
    name varchar(255) NOT NULL,
    names json NOT NULL DEFAULT '{}',
    permission varchar(200) NOT NULL,
    created_at timestamp NOT NULL DEFAULT now(),
    updated_at timestamp NOT NULL DEFAULT now() ON UPDATE now(),
    PRIMARY KEY (id)
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_privileges_name ON privileges(name);
CREATE UNIQUE INDEX IF NOT EXISTS idx_privileges_permission ON privileges(permission);

-- A grant is for one season. A revoked grant stays as history, everything that reads active grants must filter on
-- revoked_at IS NULL. created_by is the admin who gave the privilege, empty for one that comes from an order.
CREATE TABLE IF NOT EXISTS profile_privileges (
    id uuid NOT NULL DEFAULT UUID_v4(),
    profile_id uuid NOT NULL,
    privilege_id uuid NOT NULL,
    for_season_id uuid NOT NULL,
    order_item_id uuid NULL DEFAULT NULL,
    created_by uuid NULL DEFAULT NULL,
    created_at timestamp NOT NULL DEFAULT now(),
    revoked_at timestamp NULL DEFAULT NULL,
    revoked_by uuid NULL DEFAULT NULL,
    PRIMARY KEY (id),
    FOREIGN KEY (profile_id) REFERENCES profiles(id),
    FOREIGN KEY (privilege_id) REFERENCES privileges(id),
    FOREIGN KEY (for_season_id) REFERENCES seasons(id),
    FOREIGN KEY (order_item_id) REFERENCES order_items(id)
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_profile_privileges_profile_privilege_season ON profile_privileges(profile_id, privilege_id, for_season_id);
