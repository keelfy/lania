-- A basket item belongs to one season, so the same product can wait in the basket for several seasons.
-- Existing items move to the primary season, the only one they could be bought for so far.
ALTER TABLE basket_items ADD COLUMN season_id uuid NULL AFTER profile_id;

UPDATE basket_items SET season_id = (SELECT id FROM seasons WHERE is_primary = true LIMIT 1);

ALTER TABLE basket_items
    MODIFY COLUMN season_id uuid NOT NULL,
    ADD CONSTRAINT fk_basket_items_season_id FOREIGN KEY (season_id) REFERENCES seasons(id),
    DROP INDEX idx_basket_items_user_id_product_id_profile_id,
    ADD UNIQUE INDEX idx_basket_items_user_id_product_id_profile_id_season_id (user_id, product_id, profile_id, season_id);
