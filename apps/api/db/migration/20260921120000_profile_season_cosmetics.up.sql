-- A player picks cosmetics per season, because every season server keeps its own chat prefix.
-- NULL name_color_id means the default color. A NULL prefix means none of that type.
-- profiles.name_color_id and profile_prefixes stay until the contract migration, so old code keeps working.
CREATE TABLE IF NOT EXISTS profile_season_cosmetics (
    profile_id uuid NOT NULL,
    season_id uuid NOT NULL,
    name_color_id uuid NULL,
    glyth_prefix_id uuid NULL,
    special_prefix_id uuid NULL,
    updated_at timestamp NOT NULL DEFAULT now(),
    PRIMARY KEY (profile_id, season_id),
    FOREIGN KEY (profile_id) REFERENCES profiles(id),
    FOREIGN KEY (season_id) REFERENCES seasons(id),
    FOREIGN KEY (name_color_id) REFERENCES name_colors(id),
    FOREIGN KEY (glyth_prefix_id) REFERENCES name_prefixes(id),
    FOREIGN KEY (special_prefix_id) REFERENCES name_prefixes(id)
);

-- Copy the current global selection into every season where the profile owns the item.
-- An option owns the item when it is not revoked and is for this season or permanent (for_season_id IS NULL).
INSERT INTO profile_season_cosmetics (profile_id, season_id, name_color_id, glyth_prefix_id, special_prefix_id)
SELECT profile_id, season_id, name_color_id, glyth_prefix_id, special_prefix_id
FROM (
    SELECT
        p.id AS profile_id,
        s.id AS season_id,
        CASE
            WHEN p.name_color_id <> '2628bf9d-5b7c-438b-900a-67753261a823' AND EXISTS (
                SELECT 1 FROM profile_name_color_options o
                WHERE o.profile_id = p.id AND o.name_color_id = p.name_color_id
                  AND o.revoked_at IS NULL AND (o.for_season_id = s.id OR o.for_season_id IS NULL)
            ) THEN p.name_color_id
        END AS name_color_id,
        (
            SELECT pp.name_prefix_id FROM profile_prefixes pp
            WHERE pp.profile_id = p.id AND pp.type = 'glyth' AND EXISTS (
                SELECT 1 FROM profile_name_prefix_options o
                WHERE o.profile_id = p.id AND o.name_prefix_id = pp.name_prefix_id AND o.type = 'glyth'
                  AND o.revoked_at IS NULL AND (o.for_season_id = s.id OR o.for_season_id IS NULL)
            )
        ) AS glyth_prefix_id,
        (
            SELECT pp.name_prefix_id FROM profile_prefixes pp
            WHERE pp.profile_id = p.id AND pp.type = 'special' AND EXISTS (
                SELECT 1 FROM profile_name_prefix_options o
                WHERE o.profile_id = p.id AND o.name_prefix_id = pp.name_prefix_id AND o.type = 'special'
                  AND o.revoked_at IS NULL AND (o.for_season_id = s.id OR o.for_season_id IS NULL)
            )
        ) AS special_prefix_id
    FROM profiles p
    CROSS JOIN seasons s
) AS selection
WHERE name_color_id IS NOT NULL OR glyth_prefix_id IS NOT NULL OR special_prefix_id IS NOT NULL;

-- The last seen date is per season too. History of old seasons is unknown, so it stays NULL there.
-- The primary season gets the global date as the best guess, and the next player sync corrects it.
ALTER TABLE profile_playtimes
    ADD COLUMN IF NOT EXISTS last_seen_at timestamp NULL DEFAULT NULL;

UPDATE profile_playtimes pt
JOIN seasons s ON s.id = pt.season_id AND s.is_primary = true
JOIN profiles p ON p.mc_uuid = pt.mc_uuid
SET pt.last_seen_at = p.last_seen_at;
