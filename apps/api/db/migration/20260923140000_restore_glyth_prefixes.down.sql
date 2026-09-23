UPDATE name_prefixes SET metadata = JSON_SET(metadata, '$.prefix', '') WHERE name IN (
  'AgaAdun', 'AskaCute', 'Boykisser', 'BoykisserLove', 'Hexbloom', 'Netwatcher',
  'FoxFace', 'Arcanemist', 'Jokerge', 'EnergyDrink', 'Oh', 'Popcat', 'ZipZap',
  'BlazePowder', 'BowPulling', 'Bread', 'Cake', 'Cod', 'Cookie', 'EnchantedBook',
  'EndCrystal', 'EnderEye', 'ExpBottle', 'GhastTear', 'MusicDisk', 'NameTag',
  'NautilusShell', 'NetherStar', 'Pufferfish', 'TotemOfUndying', 'Trident'
);
