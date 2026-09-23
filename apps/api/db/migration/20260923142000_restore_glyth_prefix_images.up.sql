-- 20251003155523_name_prefix_fix corrupted two image URLs: Popcat's got a
-- stray trailing "Y", and BoykisserLove's was overwritten with Hexbloom's.
-- Restores both to their original values from 20250821190745_seeding.
UPDATE name_prefixes SET metadata = JSON_SET(metadata, '$.image', 'https://czx1jtlf2o.ufs.sh/f/0UHiIrRo6i8c1xBuhOVoLKCaukWF3XtiD058vHEbz9U6OGjf') WHERE name = 'BoykisserLove';
UPDATE name_prefixes SET metadata = JSON_SET(metadata, '$.image', 'https://czx1jtlf2o.ufs.sh/f/0UHiIrRo6i8cVACXAuCTNxEb8dTXc9JAWqnpQ4zIvosSkuO5') WHERE name = 'Popcat';
