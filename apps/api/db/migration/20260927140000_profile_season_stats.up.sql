-- profile_playtimes becomes profile_season_stats: one row per profile and season with a typed column per metric.
-- deaths and mob_kills are summed over every Plan session of the season, on all servers of its network.
RENAME TABLE profile_playtimes TO profile_season_stats;

ALTER TABLE profile_season_stats
    ADD COLUMN deaths bigint NOT NULL DEFAULT 0 AFTER playtime,
    ADD COLUMN mob_kills bigint NOT NULL DEFAULT 0 AFTER deaths;
