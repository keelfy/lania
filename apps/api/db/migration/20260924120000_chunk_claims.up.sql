-- map_url is the squaremap of the season server; the claims page is off while it is empty.
-- claim_limit is how many chunks one profile may hold in the season, over every world together.
ALTER TABLE seasons
    ADD COLUMN IF NOT EXISTS map_url varchar(512) NULL DEFAULT NULL,
    ADD COLUMN IF NOT EXISTS claim_limit int NOT NULL DEFAULT 100;

-- A chunk a player reserved on the site map. Nothing is protected in game, the row only records who
-- reserved the chunk and since when. Releasing keeps the row, so the history of a chunk survives.
-- world is the squaremap world name, e.g. minecraft_overworld.
-- active is 1 while the claim holds and NULL after release: the unique index then allows one active claim
-- per chunk and any number of released ones, MariaDB having no partial indexes.
CREATE TABLE IF NOT EXISTS chunk_claims (
    id uuid NOT NULL DEFAULT UUID_v4(),
    season_id uuid NOT NULL,
    world varchar(64) NOT NULL,
    chunk_x int NOT NULL,
    chunk_z int NOT NULL,
    profile_id uuid NOT NULL,
    claimed_at timestamp NOT NULL DEFAULT now(),
    released_at timestamp NULL DEFAULT NULL,
    released_by uuid NULL DEFAULT NULL,
    active tinyint AS (IF(released_at IS NULL, 1, NULL)) PERSISTENT,
    PRIMARY KEY (id),
    FOREIGN KEY (season_id) REFERENCES seasons(id),
    FOREIGN KEY (profile_id) REFERENCES profiles(id)
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_chunk_claims_active ON chunk_claims(season_id, world, chunk_x, chunk_z, active);
CREATE INDEX IF NOT EXISTS idx_chunk_claims_profile ON chunk_claims(profile_id, season_id, active);
