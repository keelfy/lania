-- profile_merged_uuids keeps the in-game UUIDs of profiles merged into another one. Plan still has the playtime
-- played under them, and a running season keeps adding to it, so the player sync sums it into the profile.
CREATE TABLE IF NOT EXISTS profile_merged_uuids (
    mc_uuid uuid NOT NULL,
    profile_id uuid NOT NULL,
    created_at timestamp NOT NULL DEFAULT now(),
    PRIMARY KEY (mc_uuid),
    CONSTRAINT fk_profile_merged_uuids_profile_id FOREIGN KEY (profile_id) REFERENCES profiles(id) ON DELETE CASCADE
);
