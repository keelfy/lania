-- streamers are the profiles with the streamer role. A row is the role: it puts the player into the streamer group
-- on the season servers, next to the staff group, and shows the profile in the public list of content creators.
-- channels is a JSON array of {"platform": "...", "url": "..."}.
CREATE TABLE IF NOT EXISTS streamers (
    profile_id uuid NOT NULL,
    channels longtext NOT NULL DEFAULT '[]' CHECK (json_valid(channels)),
    description varchar(500) NULL DEFAULT NULL,
    created_at timestamp NOT NULL DEFAULT now(),
    created_by uuid NULL DEFAULT NULL,
    updated_at timestamp NOT NULL DEFAULT now(),
    PRIMARY KEY (profile_id),
    CONSTRAINT fk_streamers_profile_id FOREIGN KEY (profile_id) REFERENCES profiles(id) ON DELETE CASCADE
);

-- streamer_applications are the requests of owners for the streamer role, reviewed by an admin.
-- A rejected application also holds the owner back from applying again for a while.
CREATE TABLE IF NOT EXISTS streamer_applications (
    id uuid NOT NULL DEFAULT UUID_v4(),
    profile_id uuid NOT NULL,
    user_id uuid NOT NULL,
    channels longtext NOT NULL DEFAULT '[]' CHECK (json_valid(channels)),
    about text NOT NULL,
    status varchar(16) NOT NULL DEFAULT 'pending',
    reject_reason varchar(500) NULL DEFAULT NULL,
    reviewed_by uuid NULL DEFAULT NULL,
    reviewed_at timestamp NULL DEFAULT NULL,
    created_at timestamp NOT NULL DEFAULT now(),
    PRIMARY KEY (id),
    CONSTRAINT fk_streamer_applications_profile_id FOREIGN KEY (profile_id) REFERENCES profiles(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_streamer_applications_status_created ON streamer_applications(status, created_at);
CREATE INDEX IF NOT EXISTS idx_streamer_applications_profile_created ON streamer_applications(profile_id, created_at);
