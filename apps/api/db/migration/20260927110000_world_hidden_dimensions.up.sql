-- hidden_dimensions is a JSON array of squaremap world names the site does not show on the world map, e.g.
-- ["minecraft_the_end"]; an empty array shows every dimension of the map.
ALTER TABLE season_worlds
    ADD COLUMN IF NOT EXISTS hidden_dimensions longtext NOT NULL DEFAULT '[]'
        CHECK (json_valid(hidden_dimensions)) AFTER claim_dimensions;
