-- name is the unique main name, shown in English and wherever a translation is missing.
-- names holds its translations by locale, for now Russian only, e.g. {"ru": "Закат"}.
ALTER TABLE name_colors
    ADD COLUMN IF NOT EXISTS names json NOT NULL DEFAULT '{}' AFTER name;
ALTER TABLE name_prefixes
    ADD COLUMN IF NOT EXISTS names json NOT NULL DEFAULT '{}' AFTER name;
