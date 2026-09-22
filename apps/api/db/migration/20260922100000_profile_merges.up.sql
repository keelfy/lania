-- profile_merges records a profile an admin merged into another one and then deleted, for support history.
-- There is no foreign key to source_profile_id: the profile no longer exists once the merge is done.
CREATE TABLE IF NOT EXISTS profile_merges (
    id uuid NOT NULL DEFAULT UUID_v4(),
    source_profile_id uuid NOT NULL,
    source_mc_uuid uuid NOT NULL,
    source_mc_username tinytext NOT NULL,
    target_profile_id uuid NOT NULL,
    merged_by uuid NOT NULL,
    summary json NOT NULL,
    created_at timestamp NOT NULL DEFAULT now(),
    PRIMARY KEY (id),
    FOREIGN KEY (target_profile_id) REFERENCES profiles(id)
);
CREATE INDEX IF NOT EXISTS idx_profile_merges_target_profile_id ON profile_merges(target_profile_id);
