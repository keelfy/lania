-- profile_username_changes records every nickname change of a profile: by its owner on the site, or picked up from
-- Mojang for a licensed account. It is the audit, the admin history and the source of the owner's cooldown.
-- old_mc_uuid equals new_mc_uuid for a licensed account: Mojang keeps the UUID on a rename.
CREATE TABLE IF NOT EXISTS profile_username_changes (
    id uuid NOT NULL DEFAULT UUID_v4(),
    profile_id uuid NOT NULL,
    old_username tinytext NOT NULL,
    new_username tinytext NOT NULL,
    old_mc_uuid uuid NOT NULL,
    new_mc_uuid uuid NOT NULL,
    source varchar(16) NOT NULL,
    changed_by uuid NULL DEFAULT NULL,
    created_at timestamp NOT NULL DEFAULT now(),
    PRIMARY KEY (id),
    CONSTRAINT fk_profile_username_changes_profile_id FOREIGN KEY (profile_id) REFERENCES profiles(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_profile_username_changes_profile_created ON profile_username_changes(profile_id, created_at);
