UPDATE name_prefixes SET metadata = JSON_SET(metadata, '$.prefix', '') WHERE name IN (
  'Meowdy', 'Cat', 'HelloKittySkull', 'CatSus', 'Cute'
);
