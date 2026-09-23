-- The "DirtyFog" name_colors row backs the product localized everywhere as
-- "SiriusTamed" (seeding.up.sql: product 3189113f-... -> nameColorId
-- 672b7493-... "DirtyFog", same colors). The admin grant-cosmetic catalog
-- reads name_colors.name directly, so it still shows "DirtyFog" instead of
-- the product-facing name. Renaming the row aligns it with the product.
UPDATE name_colors SET name = 'SiriusTamed' WHERE name = 'DirtyFog';
