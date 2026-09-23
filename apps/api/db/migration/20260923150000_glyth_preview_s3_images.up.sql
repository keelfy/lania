-- Moves every cosmetic preview icon to the project's own S3 bucket, so the
-- website serves them through imgproxy instead of linking the old uploadthing
-- host. The key is the snake_case form already used by the in-game token:
-- :glyth_popcat: becomes glyth_preview/popcat.png.
--
-- MusicDisk keeps its uploadthing URL: glyth_preview/music_disk.png is not in
-- the bucket yet. Add it and a follow-up migration once the file is uploaded.

UPDATE name_prefixes SET metadata = JSON_SET(metadata, '$.image', 's3://lania-web-134312503254-eu-central-1-an/glyth_preview/aga_adun.png') WHERE name = 'AgaAdun';
UPDATE name_prefixes SET metadata = JSON_SET(metadata, '$.image', 's3://lania-web-134312503254-eu-central-1-an/glyth_preview/aska_cute.png') WHERE name = 'AskaCute';
UPDATE name_prefixes SET metadata = JSON_SET(metadata, '$.image', 's3://lania-web-134312503254-eu-central-1-an/glyth_preview/boykisser.png') WHERE name = 'Boykisser';
UPDATE name_prefixes SET metadata = JSON_SET(metadata, '$.image', 's3://lania-web-134312503254-eu-central-1-an/glyth_preview/boykisser_love.png') WHERE name = 'BoykisserLove';
UPDATE name_prefixes SET metadata = JSON_SET(metadata, '$.image', 's3://lania-web-134312503254-eu-central-1-an/glyth_preview/hexbloom.png') WHERE name = 'Hexbloom';
UPDATE name_prefixes SET metadata = JSON_SET(metadata, '$.image', 's3://lania-web-134312503254-eu-central-1-an/glyth_preview/netwatcher.png') WHERE name = 'Netwatcher';
UPDATE name_prefixes SET metadata = JSON_SET(metadata, '$.image', 's3://lania-web-134312503254-eu-central-1-an/glyth_preview/fox_face.png') WHERE name = 'FoxFace';
UPDATE name_prefixes SET metadata = JSON_SET(metadata, '$.image', 's3://lania-web-134312503254-eu-central-1-an/glyth_preview/arcanemist.png') WHERE name = 'Arcanemist';
UPDATE name_prefixes SET metadata = JSON_SET(metadata, '$.image', 's3://lania-web-134312503254-eu-central-1-an/glyth_preview/jokerge.png') WHERE name = 'Jokerge';
UPDATE name_prefixes SET metadata = JSON_SET(metadata, '$.image', 's3://lania-web-134312503254-eu-central-1-an/glyth_preview/energy_drink.png') WHERE name = 'EnergyDrink';
UPDATE name_prefixes SET metadata = JSON_SET(metadata, '$.image', 's3://lania-web-134312503254-eu-central-1-an/glyth_preview/oh.png') WHERE name = 'Oh';
UPDATE name_prefixes SET metadata = JSON_SET(metadata, '$.image', 's3://lania-web-134312503254-eu-central-1-an/glyth_preview/popcat.png') WHERE name = 'Popcat';
UPDATE name_prefixes SET metadata = JSON_SET(metadata, '$.image', 's3://lania-web-134312503254-eu-central-1-an/glyth_preview/zipzap.png') WHERE name = 'ZipZap';
UPDATE name_prefixes SET metadata = JSON_SET(metadata, '$.image', 's3://lania-web-134312503254-eu-central-1-an/glyth_preview/blaze_powder.png') WHERE name = 'BlazePowder';
UPDATE name_prefixes SET metadata = JSON_SET(metadata, '$.image', 's3://lania-web-134312503254-eu-central-1-an/glyth_preview/bow_pulling.png') WHERE name = 'BowPulling';
UPDATE name_prefixes SET metadata = JSON_SET(metadata, '$.image', 's3://lania-web-134312503254-eu-central-1-an/glyth_preview/bread.png') WHERE name = 'Bread';
UPDATE name_prefixes SET metadata = JSON_SET(metadata, '$.image', 's3://lania-web-134312503254-eu-central-1-an/glyth_preview/cake.png') WHERE name = 'Cake';
UPDATE name_prefixes SET metadata = JSON_SET(metadata, '$.image', 's3://lania-web-134312503254-eu-central-1-an/glyth_preview/cod.png') WHERE name = 'Cod';
UPDATE name_prefixes SET metadata = JSON_SET(metadata, '$.image', 's3://lania-web-134312503254-eu-central-1-an/glyth_preview/cookie.png') WHERE name = 'Cookie';
UPDATE name_prefixes SET metadata = JSON_SET(metadata, '$.image', 's3://lania-web-134312503254-eu-central-1-an/glyth_preview/enchanted_book.png') WHERE name = 'EnchantedBook';
UPDATE name_prefixes SET metadata = JSON_SET(metadata, '$.image', 's3://lania-web-134312503254-eu-central-1-an/glyth_preview/end_crystal.png') WHERE name = 'EndCrystal';
UPDATE name_prefixes SET metadata = JSON_SET(metadata, '$.image', 's3://lania-web-134312503254-eu-central-1-an/glyth_preview/ender_eye.png') WHERE name = 'EnderEye';
UPDATE name_prefixes SET metadata = JSON_SET(metadata, '$.image', 's3://lania-web-134312503254-eu-central-1-an/glyth_preview/experience_bottle.png') WHERE name = 'ExpBottle';
UPDATE name_prefixes SET metadata = JSON_SET(metadata, '$.image', 's3://lania-web-134312503254-eu-central-1-an/glyth_preview/ghast_tear.png') WHERE name = 'GhastTear';
UPDATE name_prefixes SET metadata = JSON_SET(metadata, '$.image', 's3://lania-web-134312503254-eu-central-1-an/glyth_preview/name_tag.png') WHERE name = 'NameTag';
UPDATE name_prefixes SET metadata = JSON_SET(metadata, '$.image', 's3://lania-web-134312503254-eu-central-1-an/glyth_preview/nautilus_shell.png') WHERE name = 'NautilusShell';
UPDATE name_prefixes SET metadata = JSON_SET(metadata, '$.image', 's3://lania-web-134312503254-eu-central-1-an/glyth_preview/nether_star.png') WHERE name = 'NetherStar';
UPDATE name_prefixes SET metadata = JSON_SET(metadata, '$.image', 's3://lania-web-134312503254-eu-central-1-an/glyth_preview/pufferfish.png') WHERE name = 'Pufferfish';
UPDATE name_prefixes SET metadata = JSON_SET(metadata, '$.image', 's3://lania-web-134312503254-eu-central-1-an/glyth_preview/totem_of_undying.png') WHERE name = 'TotemOfUndying';
UPDATE name_prefixes SET metadata = JSON_SET(metadata, '$.image', 's3://lania-web-134312503254-eu-central-1-an/glyth_preview/trident.png') WHERE name = 'Trident';
UPDATE name_prefixes SET metadata = JSON_SET(metadata, '$.image', 's3://lania-web-134312503254-eu-central-1-an/glyth_preview/meowdy.png') WHERE name = 'Meowdy';
UPDATE name_prefixes SET metadata = JSON_SET(metadata, '$.image', 's3://lania-web-134312503254-eu-central-1-an/glyth_preview/cat.png') WHERE name = 'Cat';
UPDATE name_prefixes SET metadata = JSON_SET(metadata, '$.image', 's3://lania-web-134312503254-eu-central-1-an/glyth_preview/hello_kitty_skull.png') WHERE name = 'HelloKittySkull';
UPDATE name_prefixes SET metadata = JSON_SET(metadata, '$.image', 's3://lania-web-134312503254-eu-central-1-an/glyth_preview/cat_sus.png') WHERE name = 'CatSus';
UPDATE name_prefixes SET metadata = JSON_SET(metadata, '$.image', 's3://lania-web-134312503254-eu-central-1-an/glyth_preview/cute.png') WHERE name = 'Cute';
UPDATE name_prefixes SET metadata = JSON_SET(metadata, '$.image', 's3://lania-web-134312503254-eu-central-1-an/glyth_preview/yobshestvo.png') WHERE name = 'Yobshestvo';
