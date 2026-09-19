-- Test data for the community page. Expects @owner_id (Kratos identity of the test user).
-- Every value is derived from the username, so reseeding gives the same data.

DELETE FROM profile_mojang_uuids;
DELETE FROM profile_playtimes;
DELETE FROM profile_accesses;
DELETE FROM profile_name_color_options;
DELETE FROM profile_prefixes;
DELETE FROM profile_name_prefix_options;
DELETE FROM basket_items;
DELETE FROM order_items;
DELETE FROM profiles;

CREATE TEMPORARY TABLE seed_profiles (name VARCHAR(16) NOT NULL, role VARCHAR(16) NOT NULL);
INSERT INTO seed_profiles (name, role) VALUES
    ('Keelfy', 'owner'),
    ('Admiral_Nemo', 'admin'),
    ('Moderator_Mia', 'mod'),
    ('Alex', 'player'),
    ('Alexander', 'player'),
    ('Alexandra', 'player'),
    ('alina_k', 'player'),
    ('Alien_Hunter', 'player'),
    ('Bob', 'player'),
    ('Bobby_Tables', 'player'),
    ('Creeper_Slayer', 'player'),
    ('Diamond_Dave', 'player'),
    ('Enderman_Eve', 'player'),
    ('Farmer_Fred', 'player'),
    ('Ghast_Hugger', 'player'),
    ('Herobrine_Fan', 'player'),
    ('Iron_Golem', 'player'),
    ('Jack_O_Lantern', 'player'),
    ('Kelp_Knight', 'player'),
    ('Lava_Larry', 'player'),
    ('Mushroom_Mike', 'player'),
    ('Netherite_Nina', 'player'),
    ('Obsidian_Oleg', 'player'),
    ('Pig_Pilot', 'player'),
    ('Quartz_Queen', 'player'),
    ('Redstone_Rita', 'player'),
    ('Steve', 'player'),
    ('Torch_Tim', 'player'),
    ('Under_Water', 'player'),
    ('Villager_Vic', 'player'),
    ('100%_Legit', 'player'),
    ('Wither_Wendy', 'player'),
    ('Zombie_Zed', 'player'),
    ('a_b', 'player'),
    ('a%b', 'player');

-- some profiles never joined the server, so their dates stay NULL
INSERT INTO profiles (mc_uuid, mc_username, owner_user_id, first_seen_at, last_seen_at, role, is_slim, created_at, updated_at)
SELECT
    CONCAT(SUBSTR(h, 1, 8), '-', SUBSTR(h, 9, 4), '-4', SUBSTR(h, 14, 3), '-8', SUBSTR(h, 18, 3), '-', SUBSTR(h, 21, 12)),
    name,
    CASE name WHEN 'Keelfy' THEN @owner_id WHEN 'Alex' THEN @owner_id END,
    CASE WHEN CRC32(name) % 6 = 0 THEN NULL ELSE NOW() - INTERVAL (200 + CRC32(name) % 400) DAY END,
    CASE WHEN CRC32(name) % 6 = 0 THEN NULL ELSE NOW() - INTERVAL (CRC32(name) % 200) DAY END,
    role,
    CRC32(name) % 2 = 1,
    NOW() - INTERVAL (100 + CRC32(name) % 500) DAY,
    NOW()
FROM (SELECT name, role, MD5(name) AS h FROM seed_profiles) p;

INSERT INTO profile_accesses (mc_uuid, season_id, source)
SELECT mc_uuid, '123e4567-e89b-12d3-a456-426614174003', 'free' FROM profiles;

-- playtime in ms over the last two seasons, profiles without a last_seen_at have none
INSERT INTO profile_playtimes (mc_uuid, season_id, playtime)
SELECT mc_uuid, '123e4567-e89b-12d3-a456-426614174003', (CRC32(mc_username) % 300) * 3600000 + 1234567
FROM profiles WHERE last_seen_at IS NOT NULL;
INSERT INTO profile_playtimes (mc_uuid, season_id, playtime)
SELECT mc_uuid, '123e4567-e89b-12d3-a456-426614174002', (CRC32(mc_username) % 150) * 3600000 + 7654321
FROM profiles WHERE last_seen_at IS NOT NULL AND CRC32(mc_username) % 3 = 0;

DROP TEMPORARY TABLE seed_profiles;
