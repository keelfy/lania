CREATE TABLE IF NOT EXISTS name_colors (
  id uuid NOT NULL DEFAULT UUID_v4(),
  colors json NOT NULL,
  name tinytext NOT NULL,
  PRIMARY KEY (id)
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_name_colors_name ON name_colors(name);

CREATE TABLE IF NOT EXISTS profile_name_color_options (
  id uuid NOT NULL DEFAULT UUID_v4(),
  profile_id uuid NOT NULL,
  name_color_id uuid NOT NULL,
  -- NULL for role color (owner, moderator, etc.)
  order_item_id bigint,
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

-- add season_id to order_items
ALTER TABLE order_items ADD COLUMN season_id uuid NOT NULL;
ALTER TABLE order_items ADD FOREIGN KEY (season_id) REFERENCES seasons(id);

-- name colors
INSERT INTO name_colors (id, name, colors) VALUES
    ('2628bf9d-5b7c-438b-900a-67753261a823', 'Default', '{"colors": []}'),
    ('a1ed40de-8a1b-4f49-b50d-3fa9cd92ae10', 'Owner', '{"colors": ["#F4320B", "#FF320B"]}'),
    ('81d3c858-1a2e-45a8-845b-4f803cdf62b5', 'Moderator', '{"colors": ["#00AA00", "#00B100"]}'),
    ('b148e30c-c346-42c0-8eb5-a2c3f50e7009', 'loonatya', '{"colors": ["#654ea3", "#eaafc8"]}'),
    ('276bb1e0-dbe2-49f6-883b-13b359c5ee46', 'Memariani', '{"colors": ["#aa4b6b", "#6b6b83", "#3b8d99"]}'),
    ('55861c6c-a32f-49ef-a05d-d976db0a0ca1', 'PacificDream', '{"colors": ["#34e89e", "#0f3443"]}'),
    ('c9d8dc62-e592-45b6-b892-528db887c81e', 'SteelGray', '{"colors": ["#1F1C2C", "#928DAB"]}'),
    ('089d09c2-9044-4846-b153-f43d93584929', 'Dracula', '{"colors": ["#DC2424", "#4A569D"]}'),
    ('6cdc5d63-3ed0-423c-ba2a-2e90f834ecff', 'Kyoto', '{"colors": ["#c21500", "#ffc500"]}'),
    ('7d1ac8ea-feb4-4a9a-888f-86826bf74c0c', 'WheatField', '{"colors": ["#004FF9", "#FFF94C"]}'),
    ('831303da-ee6e-4fd7-81c9-e971a1042f72', 'RedSquare', '{"colors": ["#FFFFFF", "#0033A0", "#DA291C"]}'),
    ('4e96a645-f978-4f34-a92b-4f7bdd5e8f89', 'RedMist', '{"colors": ["#000000", "#e74c3c"]}'),
    ('51a037b7-b56c-4fa6-be15-fa594e548a33', 'MiceElf', '{"colors": ["#948E99", "#2E1437"]}'),
    ('6296c8a8-d897-4aa6-bc24-ab3ad9d70ac2', 'Timber', '{"colors": ["#fc00ff", "#00dbde"]}'),
    ('c2ec77b0-1062-4e6c-8080-794c55f9b7c4', 'Lawrencium', '{"colors": ["#0f0c29", "#302b63", "#24243e"]}'),
    ('50dab533-8642-4ce0-83cb-3de1774353ed', 'Magic', '{"colors": ["#59C173", "#a17fe0", "#5D26C1"]}'),
    ('d55d6881-918b-4f5a-86be-f7b82db3473f', 'PinkFlavour', '{"colors": ["#800080", "#ffc0cb"]}'),
    ('69323f24-ecbe-4fea-a4e1-0ce8b1608d68', 'CocoaaIce', '{"colors": ["#c0c0aa", "#1cefff"]}'),
    ('a977667b-7272-406e-a9ec-7d0cf513ec67', 'Dusk', '{"colors": ["#2C3E50", "#FD746C"]}'),
    ('d21f754e-fa5b-4010-8889-2a98bc796a63', 'KyeMeh', '{"colors": ["#8360c3", "#2ebf91"]}'),
    ('35a55540-a996-4e4d-b3b8-85495b525799', 'VelvetSun', '{"colors": ["#e1eec3", "#f05053"]}'),
    ('61839696-171e-4488-9b19-903647b0525b', 'Telegram', '{"colors": ["#1c92d2", "#f2fcfe"]}'),
    ('672b7493-d05b-4171-874b-2312b08813a0', 'DirtyFog', '{"colors": ["#B993D6", "#8CA6DB"]}'),
    ('45690458-d16b-49b5-8813-654444b35343', 'Moonrise', '{"colors": ["#dae2f8", "#d6a4a4"]}'),
    ('f808aee9-a300-4efa-b098-2b41e48afff4', 'Frozen', '{"colors": ["#403B4A", "#E7E9BB"]}'),
    ('9a8b613c-f7bc-45c8-88fe-0c76f2f5c8d4', 'DeepSpace', '{"colors": ["#000000", "#434343"]}'),
    ('81f5a82d-1e4d-4749-a821-c4eda6a6656f', 'Mango', '{"colors": ["#ffe259", "#ffa751"]}');

-- add name color id to product metadata
UPDATE products
SET metadata = '{"colors": ["#aa4b6b", "#6b6b83", "#3b8d99"], "nameColorId": "276bb1e0-dbe2-49f6-883b-13b359c5ee46"}'
WHERE id = '21852047-a4e9-4f7d-acfd-9fce898bc03d';

-- Memariani
UPDATE products
SET metadata = '{"colors": ["#34e89e", "#0f3443"], "nameColorId": "55861c6c-a32f-49ef-a05d-d976db0a0ca1"}'
WHERE id = '6509dcd7-6177-4e39-8474-e0304870ff23';

-- SteelGray
UPDATE products
SET metadata = '{"colors": ["#1F1C2C", "#928DAB"], "nameColorId": "c9d8dc62-e592-45b6-b892-528db887c81e"}'
WHERE id = 'f1740c10-850a-41a6-bcd8-d3449846f284';

-- Dracula
UPDATE products
SET metadata = '{"colors": ["#DC2424", "#4A569D"], "nameColorId": "089d09c2-9044-4846-b153-f43d93584929"}'
WHERE id = '1cef33ed-08e5-49c5-9858-dac85872386e';

-- Kyoto
UPDATE products
SET metadata = '{"colors": ["#c21500", "#ffc500"], "nameColorId": "6cdc5d63-3ed0-423c-ba2a-2e90f834ecff"}'
WHERE id = '6a213c8c-4269-4329-9875-917cb3cfd033';

-- WheatField
UPDATE products
SET metadata = '{"colors": ["#004FF9", "#FFF94C"], "nameColorId": "7d1ac8ea-feb4-4a9a-888f-86826bf74c0c"}'
WHERE id = '068e6c5a-94e1-4927-9eed-f9b02beb9f64';

-- RedSquare
UPDATE products
SET metadata = '{"colors": ["#FFFFFF", "#0033A0", "#DA291C"], "nameColorId": "831303da-ee6e-4fd7-81c9-e971a1042f72"}'
WHERE id = '8bbf4210-24fa-4382-af93-4da30dfeae1f';

-- RedMist
UPDATE products
SET metadata = '{"colors": ["#000000", "#e74c3c"], "nameColorId": "4e96a645-f978-4f34-a92b-4f7bdd5e8f89"}'
WHERE id = '54075756-222b-4917-9924-696436654254';

-- MiceElf
UPDATE products
SET metadata = '{"colors": ["#948E99", "#2E1437"], "nameColorId": "51a037b7-b56c-4fa6-be15-fa594e548a33"}'
WHERE id = '72150255-2649-4339-895b-44142701117d';

-- Timber
UPDATE products
SET metadata = '{"colors": ["#fc00ff", "#00dbde"], "nameColorId": "6296c8a8-d897-4aa6-bc24-ab3ad9d70ac2"}'
WHERE id = 'f555a7d6-c52f-4f22-8183-cdfee77eb59e';

-- Lawrencium
UPDATE products
SET metadata = '{"colors": ["#0f0c29", "#302b63", "#24243e"], "nameColorId": "c2ec77b0-1062-4e6c-8080-794c55f9b7c4"}'
WHERE id = '56ef1088-6ce5-4cb6-9c3d-0c35070def83';

-- Magic
UPDATE products
SET metadata = '{"colors": ["#59C173", "#a17fe0", "#5D26C1"], "nameColorId": "50dab533-8642-4ce0-83cb-3de1774353ed"}'
WHERE id = '8f461be1-9320-4d17-96e7-e5215fc52522';

-- PinkFlavour
UPDATE products
SET metadata = '{"colors": ["#800080", "#ffc0cb"], "nameColorId": "d55d6881-918b-4f5a-86be-f7b82db3473f"}'
WHERE id = '383c4fc1-5f92-4bae-a4b4-b743922f1eba';

-- CocoaaIce
UPDATE products
SET metadata = '{"colors": ["#c0c0aa", "#1cefff"], "nameColorId": "69323f24-ecbe-4fea-a4e1-0ce8b1608d68"}'
WHERE id = 'bdd07172-ea5f-4a0a-b82a-16776a6b4b5c';

-- Dusk
UPDATE products
SET metadata = '{"colors": ["#2C3E50", "#FD746C"], "nameColorId": "a977667b-7272-406e-a9ec-7d0cf513ec67"}'
WHERE id = '9765c31a-c5d9-4d0a-9eb7-3c10aec6e66c';

-- KyeMeh
UPDATE products
SET metadata = '{"colors": ["#8360c3", "#2ebf91"], "nameColorId": "d21f754e-fa5b-4010-8889-2a98bc796a63"}'
WHERE id = '68a560a1-878b-4cee-b79d-465aea06862b';

-- VelvetSun
UPDATE products
SET metadata = '{"colors": ["#e1eec3", "#f05053"], "nameColorId": "35a55540-a996-4e4d-b3b8-85495b525799"}'
WHERE id = '69c9ef9e-f66a-4df7-9b3b-bdb2dbf97834';

-- Telegram
UPDATE products
SET metadata = '{"colors": ["#1c92d2", "#f2fcfe"], "nameColorId": "61839696-171e-4488-9b19-903647b0525b"}'
WHERE id = '66678f14-e431-41d1-a172-d301564deba0';

-- DirtyFog
UPDATE products
SET metadata = '{"colors": ["#B993D6", "#8CA6DB"], "nameColorId": "672b7493-d05b-4171-874b-2312b08813a0"}'
WHERE id = '3189113f-9253-4bef-9ce4-0374aff2df7c';

-- Moonrise
UPDATE products
SET metadata = '{"colors": ["#dae2f8", "#d6a4a4"], "nameColorId": "45690458-d16b-49b5-8813-654444b35343"}'
WHERE id = 'de397bbd-782c-4eb2-a59c-f955cc06340e';

-- Frozen
UPDATE products
SET metadata = '{"colors": ["#403B4A", "#E7E9BB"], "nameColorId": "f808aee9-a300-4efa-b098-2b41e48afff4"}'
WHERE id = '63406c8d-67c2-43c1-a8a7-e7e79b7ac2c3';

-- DeepSpace
UPDATE products
SET metadata = '{"colors": ["#000000", "#434343"], "nameColorId": "9a8b613c-f7bc-45c8-88fe-0c76f2f5c8d4"}'
WHERE id = 'd9607e73-ad81-46e6-b0af-ee33b5987e16';

-- Mango
UPDATE products
SET metadata = '{"colors": ["#ffe259", "#ffa751"], "nameColorId": "81f5a82d-1e4d-4749-a821-c4eda6a6656f"}'
WHERE id = '2a65d174-8896-411a-a6dd-8341b3620171';

-- add name color id to profiles
ALTER TABLE profiles
ADD COLUMN name_color_id uuid;

UPDATE profiles
SET name_color_id = '2628bf9d-5b7c-438b-900a-67753261a823';

ALTER TABLE profiles MODIFY COLUMN name_color_id uuid NOT NULL;
ALTER TABLE profiles ALTER COLUMN name_color_id SET DEFAULT '2628bf9d-5b7c-438b-900a-67753261a823';

-- add order item id to profile access
ALTER TABLE profile_accesses ADD COLUMN order_item_id bigint;
ALTER TABLE profile_accesses ADD FOREIGN KEY (order_item_id) REFERENCES order_items(id);


-- add all name colors for profile
-- INSERT INTO profile_name_color_options (profile_id, name_color_id)
-- SELECT DISTINCT 'ea48eec6-ed89-4437-b3b0-07355139bad0' as profile_id, nc.id
-- FROM name_colors nc
-- LEFT JOIN profile_name_color_options pnco 
--     ON pnco.name_color_id = nc.id 
--     AND pnco.profile_id <> 'ea48eec6-ed89-4437-b3b0-07355139bad0'
-- WHERE pnco.id IS NOT NULL;