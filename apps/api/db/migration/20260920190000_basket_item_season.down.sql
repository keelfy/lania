-- The old unique index cannot hold the same product for two seasons, so keep one item per product and profile.
DELETE bi FROM basket_items bi
    JOIN basket_items other
        ON other.user_id = bi.user_id AND other.product_id = bi.product_id AND other.profile_id = bi.profile_id
        AND other.id < bi.id;

ALTER TABLE basket_items
    DROP INDEX idx_basket_items_user_id_product_id_profile_id_season_id,
    ADD UNIQUE INDEX idx_basket_items_user_id_product_id_profile_id (user_id, product_id, profile_id),
    DROP FOREIGN KEY fk_basket_items_season_id,
    DROP COLUMN season_id;
