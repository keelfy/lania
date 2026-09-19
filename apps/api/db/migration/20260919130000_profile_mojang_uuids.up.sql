-- profile_mojang_uuids caches the Mojang account UUID of a profile.
-- mojang_uuid is NULL when Mojang has no account with that username (checked_at tells when we asked last).
-- Rows are filled by the background sync so requests never depend on the Mojang rate limit.
CREATE TABLE IF NOT EXISTS profile_mojang_uuids (
    mc_uuid uuid NOT NULL,
    mojang_uuid uuid,
    checked_at timestamp NOT NULL DEFAULT now(),
    PRIMARY KEY (mc_uuid),
    FOREIGN KEY (mc_uuid) REFERENCES profiles(mc_uuid)
);
