DROP TABLE IF EXISTS profile_privileges;
DROP TABLE IF EXISTS privileges;
-- Tariffs of the privileges.
DELETE FROM product_prices WHERE name LIKE 'privilege:%';
