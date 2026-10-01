-- profile_acquisitions records every profile a user added on the site: a new profile or a claim of one nobody owns.
-- It is the source of the limit on new profiles per month, which a release followed by a new profile cannot get around.
-- An admin transfer is not recorded.
CREATE TABLE IF NOT EXISTS profile_acquisitions (
    id uuid NOT NULL DEFAULT UUID_v4(),
    user_id uuid NOT NULL,
    profile_id uuid NOT NULL,
    created_at timestamp NOT NULL DEFAULT now(),
    PRIMARY KEY (id),
    CONSTRAINT fk_profile_acquisitions_profile_id FOREIGN KEY (profile_id) REFERENCES profiles(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_profile_acquisitions_user_created ON profile_acquisitions(user_id, created_at);
