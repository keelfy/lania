ALTER TABLE profile_season_stats
    DROP COLUMN mob_kills,
    DROP COLUMN deaths;

RENAME TABLE profile_season_stats TO profile_playtimes;
