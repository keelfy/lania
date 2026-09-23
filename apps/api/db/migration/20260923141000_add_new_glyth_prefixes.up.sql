-- These glyths were added by 20251005113019_name_prefix_update1 with an
-- empty prefix (icon-only). Backfills the in-game token, following the
-- :glyth_<snake_case_name>: convention used by the rest of the catalog.
UPDATE name_prefixes SET metadata = JSON_SET(metadata, '$.prefix', ':glyth_meowdy:') WHERE name = 'Meowdy';
UPDATE name_prefixes SET metadata = JSON_SET(metadata, '$.prefix', ':glyth_cat:') WHERE name = 'Cat';
UPDATE name_prefixes SET metadata = JSON_SET(metadata, '$.prefix', ':glyth_hello_kitty_skull:') WHERE name = 'HelloKittySkull';
UPDATE name_prefixes SET metadata = JSON_SET(metadata, '$.prefix', ':glyth_cat_sus:') WHERE name = 'CatSus';
UPDATE name_prefixes SET metadata = JSON_SET(metadata, '$.prefix', ':glyth_cute:') WHERE name = 'Cute';
