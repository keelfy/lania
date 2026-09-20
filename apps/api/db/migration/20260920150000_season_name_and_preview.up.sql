-- A season gets a display name and a preview screenshot. The name is the same for every language.
-- preview_image is the location of the screenshot as the image proxy reads it. A season without one is not shown on the seasons page.
ALTER TABLE seasons
    ADD COLUMN IF NOT EXISTS name varchar(255) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS preview_image varchar(512) NULL DEFAULT NULL;

UPDATE seasons SET name = 'Lania I', preview_image = 's3://lania-web-134312503254-eu-central-1-an/2025-08-29_21.26.21.png' WHERE id = '123e4567-e89b-12d3-a456-426614174000';
UPDATE seasons SET name = 'Lania II', preview_image = 's3://lania-web-134312503254-eu-central-1-an/2025-08-29_21.23.23.png' WHERE id = '123e4567-e89b-12d3-a456-426614174001';
UPDATE seasons SET name = 'Lania III', preview_image = 's3://lania-web-134312503254-eu-central-1-an/2025-08-19_11.15.29.png' WHERE id = '123e4567-e89b-12d3-a456-426614174002';
UPDATE seasons SET name = 'Lania IV', preview_image = 's3://lania-web-134312503254-eu-central-1-an/lania-4-castle.png' WHERE id = '123e4567-e89b-12d3-a456-426614174003';
UPDATE seasons SET name = 'Lania V', preview_image = 's3://lania-web-134312503254-eu-central-1-an/lania-5-preview.jpg' WHERE id = '123e4567-e89b-12d3-a456-426614174004';
UPDATE seasons SET name = 'Lania Spinoff I', preview_image = 's3://lania-web-134312503254-eu-central-1-an/lania-spinoff-1-preview.png' WHERE id = '123e4567-e89b-12d3-a456-426614174010';
UPDATE seasons SET name = 'Glory Mine 7' WHERE id = '123e4567-e89b-12d3-a456-426614174099';
UPDATE seasons SET name = 'Glory Mine 6' WHERE id = '123e4567-e89b-12d3-a456-426614174098';

-- The seed gave every season the same placeholder dates. The seasons page showed the real ones, so they move here.
-- A season whose dates were already corrected by hand keeps them.
UPDATE seasons SET start_date = '2023-12-23', end_date = '2024-02-23' WHERE id = '123e4567-e89b-12d3-a456-426614174000' AND start_date = '2025-01-01';
UPDATE seasons SET start_date = '2024-07-12', end_date = '2024-09-03' WHERE id = '123e4567-e89b-12d3-a456-426614174001' AND start_date = '2025-01-01';
UPDATE seasons SET start_date = '2025-04-01', end_date = '2025-06-10' WHERE id = '123e4567-e89b-12d3-a456-426614174002' AND start_date = '2025-01-01';
UPDATE seasons SET start_date = '2025-09-26', end_date = '2025-12-10' WHERE id = '123e4567-e89b-12d3-a456-426614174003' AND start_date = '2025-01-01';
UPDATE seasons SET start_date = '2026-10-09', end_date = NULL WHERE id = '123e4567-e89b-12d3-a456-426614174004' AND start_date = '2025-01-01';
UPDATE seasons SET start_date = '2026-04-26', end_date = NULL WHERE id = '123e4567-e89b-12d3-a456-426614174010' AND start_date = '2025-01-01';

-- Any other season gets a name from its number instead of an empty one.
UPDATE seasons SET name = CONCAT('Season ', season_number) WHERE name = '';
