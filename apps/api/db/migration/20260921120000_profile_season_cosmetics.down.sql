ALTER TABLE profile_playtimes
    DROP COLUMN IF EXISTS last_seen_at;

DROP TABLE IF EXISTS profile_season_cosmetics;
