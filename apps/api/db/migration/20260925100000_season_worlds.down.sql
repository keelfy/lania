-- A season keeps one map again: the first world by position. Claims of its other worlds are deleted, they
-- would collide with the first world's claims on the old per-season index.
ALTER TABLE seasons
    ADD COLUMN IF NOT EXISTS map_url varchar(512) NULL DEFAULT NULL,
    ADD COLUMN IF NOT EXISTS claim_limit int NOT NULL DEFAULT 100;

CREATE TEMPORARY TABLE first_season_worlds AS
SELECT w.id, w.season_id, w.map_url, w.claim_limit
FROM season_worlds w
WHERE w.id = (
    SELECT f.id FROM season_worlds f WHERE f.season_id = w.season_id ORDER BY f.position, f.created_at, f.id LIMIT 1
);

UPDATE seasons s
JOIN first_season_worlds f ON f.season_id = s.id
SET s.map_url = f.map_url, s.claim_limit = f.claim_limit;

DELETE FROM chunk_claims WHERE world_id NOT IN (SELECT id FROM first_season_worlds);

ALTER TABLE chunk_claims ADD COLUMN IF NOT EXISTS season_id uuid NULL AFTER id;
UPDATE chunk_claims c JOIN first_season_worlds f ON f.id = c.world_id SET c.season_id = f.season_id;
ALTER TABLE chunk_claims
    MODIFY season_id uuid NOT NULL,
    CHANGE COLUMN dimension world varchar(64) NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_chunk_claims_active ON chunk_claims(season_id, world, chunk_x, chunk_z, active);
CREATE INDEX IF NOT EXISTS idx_chunk_claims_profile ON chunk_claims(profile_id, season_id, active);
ALTER TABLE chunk_claims ADD CONSTRAINT chunk_claims_ibfk_1 FOREIGN KEY (season_id) REFERENCES seasons(id);

ALTER TABLE chunk_claims DROP FOREIGN KEY IF EXISTS fk_chunk_claims_world;
DROP INDEX IF EXISTS idx_chunk_claims_world_active ON chunk_claims;
DROP INDEX IF EXISTS idx_chunk_claims_profile_world ON chunk_claims;
ALTER TABLE chunk_claims DROP COLUMN IF EXISTS world_id;

DROP TEMPORARY TABLE first_season_worlds;
DROP TABLE IF EXISTS season_worlds;
