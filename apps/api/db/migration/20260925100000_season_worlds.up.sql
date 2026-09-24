-- A world is one server of a season network: survival, farms, creative. Each has its own squaremap and its
-- own chunk claims.
-- slug is the key of the world page, /worlds/<slug>; unique within the season.
-- map_url is the squaremap of the world server; the world has no map page while it is empty.
-- claim_limit is how many chunks one profile may hold in the world, over its claim dimensions together.
-- claim_dimensions is a JSON array of squaremap world names where chunks can be claimed, e.g.
-- ["minecraft_overworld"]; an empty array makes the map view-only.
CREATE TABLE IF NOT EXISTS season_worlds (
    id uuid NOT NULL DEFAULT UUID_v4(),
    season_id uuid NOT NULL,
    slug varchar(32) NOT NULL,
    name varchar(64) NOT NULL,
    preview_image varchar(512) NULL DEFAULT NULL,
    map_url varchar(512) NULL DEFAULT NULL,
    claim_limit int NOT NULL DEFAULT 100,
    claim_dimensions longtext NOT NULL DEFAULT '[]' CHECK (json_valid(claim_dimensions)),
    position int NOT NULL DEFAULT 0,
    created_at timestamp NOT NULL DEFAULT now(),
    PRIMARY KEY (id),
    FOREIGN KEY (season_id) REFERENCES seasons(id)
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_season_worlds_slug ON season_worlds(season_id, slug);

-- The one map a season had becomes its survival world, claims in the overworld as before.
INSERT INTO season_worlds (season_id, slug, name, map_url, claim_limit, claim_dimensions)
SELECT s.id, 'survival', 'Выживание', s.map_url, s.claim_limit, '["minecraft_overworld"]'
FROM seasons s
WHERE (s.map_url IS NOT NULL OR EXISTS (SELECT 1 FROM chunk_claims c WHERE c.season_id = s.id))
  AND NOT EXISTS (SELECT 1 FROM season_worlds w WHERE w.season_id = s.id);

-- Claims move from the season to its world; the squaremap world name they kept is a dimension of that world.
ALTER TABLE chunk_claims ADD COLUMN IF NOT EXISTS world_id uuid NULL AFTER id;
UPDATE chunk_claims c
JOIN season_worlds w ON w.season_id = c.season_id AND w.slug = 'survival'
SET c.world_id = w.id
WHERE c.world_id IS NULL;
ALTER TABLE chunk_claims
    MODIFY world_id uuid NOT NULL,
    CHANGE COLUMN world dimension varchar(64) NOT NULL;

-- The new indexes come first so the profile foreign key always has an index to lean on.
CREATE UNIQUE INDEX IF NOT EXISTS idx_chunk_claims_world_active ON chunk_claims(world_id, dimension, chunk_x, chunk_z, active);
CREATE INDEX IF NOT EXISTS idx_chunk_claims_profile_world ON chunk_claims(profile_id, world_id, active);
ALTER TABLE chunk_claims ADD CONSTRAINT fk_chunk_claims_world FOREIGN KEY (world_id) REFERENCES season_worlds(id);

-- chunk_claims_ibfk_1 is the unnamed season_id foreign key of 20260924120000.
ALTER TABLE chunk_claims DROP FOREIGN KEY IF EXISTS chunk_claims_ibfk_1;
DROP INDEX IF EXISTS idx_chunk_claims_active ON chunk_claims;
DROP INDEX IF EXISTS idx_chunk_claims_profile ON chunk_claims;
ALTER TABLE chunk_claims DROP COLUMN IF EXISTS season_id;

ALTER TABLE seasons
    DROP COLUMN IF EXISTS map_url,
    DROP COLUMN IF EXISTS claim_limit;
