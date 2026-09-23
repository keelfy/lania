-- Restores the in-game prefix token for glyths whose metadata.prefix was
-- emptied by 20251003155523_name_prefix_fix, leaving only the website icon.
-- Values match the original tokens from 20250821190745_seeding.
UPDATE name_prefixes SET metadata = JSON_SET(metadata, '$.prefix', ':glyth_aga_adun:') WHERE name = 'AgaAdun';
UPDATE name_prefixes SET metadata = JSON_SET(metadata, '$.prefix', ':glyth_aska_cute:') WHERE name = 'AskaCute';
UPDATE name_prefixes SET metadata = JSON_SET(metadata, '$.prefix', ':glyth_boykisser:') WHERE name = 'Boykisser';
UPDATE name_prefixes SET metadata = JSON_SET(metadata, '$.prefix', ':glyth_boykisser_love:') WHERE name = 'BoykisserLove';
UPDATE name_prefixes SET metadata = JSON_SET(metadata, '$.prefix', ':glyth_hexbloom:') WHERE name = 'Hexbloom';
UPDATE name_prefixes SET metadata = JSON_SET(metadata, '$.prefix', ':glyth_netwatcher:') WHERE name = 'Netwatcher';
UPDATE name_prefixes SET metadata = JSON_SET(metadata, '$.prefix', ':glyth_fox_face:') WHERE name = 'FoxFace';
UPDATE name_prefixes SET metadata = JSON_SET(metadata, '$.prefix', ':glyth_arcanemist:') WHERE name = 'Arcanemist';
UPDATE name_prefixes SET metadata = JSON_SET(metadata, '$.prefix', ':glyth_jokerge:') WHERE name = 'Jokerge';
UPDATE name_prefixes SET metadata = JSON_SET(metadata, '$.prefix', ':glyth_energy_drink:') WHERE name = 'EnergyDrink';
UPDATE name_prefixes SET metadata = JSON_SET(metadata, '$.prefix', ':glyth_oh:') WHERE name = 'Oh';
UPDATE name_prefixes SET metadata = JSON_SET(metadata, '$.prefix', ':glyth_popcat:') WHERE name = 'Popcat';
UPDATE name_prefixes SET metadata = JSON_SET(metadata, '$.prefix', ':glyth_zipzap:') WHERE name = 'ZipZap';
UPDATE name_prefixes SET metadata = JSON_SET(metadata, '$.prefix', ':glyth_blaze_powder:') WHERE name = 'BlazePowder';
UPDATE name_prefixes SET metadata = JSON_SET(metadata, '$.prefix', ':glyth_bow_pulling:') WHERE name = 'BowPulling';
UPDATE name_prefixes SET metadata = JSON_SET(metadata, '$.prefix', ':glyth_bread:') WHERE name = 'Bread';
UPDATE name_prefixes SET metadata = JSON_SET(metadata, '$.prefix', ':glyth_cake:') WHERE name = 'Cake';
UPDATE name_prefixes SET metadata = JSON_SET(metadata, '$.prefix', ':glyth_cod:') WHERE name = 'Cod';
UPDATE name_prefixes SET metadata = JSON_SET(metadata, '$.prefix', ':glyth_cookie:') WHERE name = 'Cookie';
UPDATE name_prefixes SET metadata = JSON_SET(metadata, '$.prefix', ':glyth_enchanted_book:') WHERE name = 'EnchantedBook';
UPDATE name_prefixes SET metadata = JSON_SET(metadata, '$.prefix', ':glyth_end_crystal:') WHERE name = 'EndCrystal';
UPDATE name_prefixes SET metadata = JSON_SET(metadata, '$.prefix', ':glyth_ender_eye:') WHERE name = 'EnderEye';
UPDATE name_prefixes SET metadata = JSON_SET(metadata, '$.prefix', ':glyth_experience_bottle:') WHERE name = 'ExpBottle';
UPDATE name_prefixes SET metadata = JSON_SET(metadata, '$.prefix', ':glyth_ghast_tear:') WHERE name = 'GhastTear';
UPDATE name_prefixes SET metadata = JSON_SET(metadata, '$.prefix', ':glyth_music_disk:') WHERE name = 'MusicDisk';
UPDATE name_prefixes SET metadata = JSON_SET(metadata, '$.prefix', ':glyth_name_tag:') WHERE name = 'NameTag';
UPDATE name_prefixes SET metadata = JSON_SET(metadata, '$.prefix', ':glyth_nautilus_shell:') WHERE name = 'NautilusShell';
UPDATE name_prefixes SET metadata = JSON_SET(metadata, '$.prefix', ':glyth_nether_star:') WHERE name = 'NetherStar';
UPDATE name_prefixes SET metadata = JSON_SET(metadata, '$.prefix', ':glyth_pufferfish:') WHERE name = 'Pufferfish';
UPDATE name_prefixes SET metadata = JSON_SET(metadata, '$.prefix', ':glyth_totem_of_undying:') WHERE name = 'TotemOfUndying';
UPDATE name_prefixes SET metadata = JSON_SET(metadata, '$.prefix', ':glyth_trident:') WHERE name = 'Trident';
