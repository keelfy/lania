-- profile_playtimes is a many-to-many link between profiles and seasons.
-- playtime is stored in milliseconds (same unit as PLAN sessions).
-- Drop duplicated (mc_uuid, season_id) pairs, keeping the largest playtime.
DELETE p1 FROM profile_playtimes p1
JOIN profile_playtimes p2
  ON p1.mc_uuid = p2.mc_uuid
 AND p1.season_id = p2.season_id
 AND (p1.playtime < p2.playtime OR (p1.playtime = p2.playtime AND p1.id < p2.id));

ALTER TABLE profile_playtimes
    DROP PRIMARY KEY,
    DROP COLUMN id,
    MODIFY playtime bigint NOT NULL COMMENT 'milliseconds',
    ADD PRIMARY KEY (mc_uuid, season_id);
