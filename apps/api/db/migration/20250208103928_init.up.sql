CREATE TABLE IF NOT EXISTS seasons (
    id uuid NOT NULL DEFAULT UUID_v4(),
    season_number integer NOT NULL,
    start_date timestamp NOT NULL,
    end_date timestamp,
    PRIMARY KEY(id)
);
CREATE TABLE IF NOT EXISTS profiles (
    id uuid NOT NULL DEFAULT UUID_v4(),
    mc_uuid uuid NOT NULL,
    mc_username tinytext NOT NULL,
    owner_user_id uuid NOT NULL, 
    first_seen_at timestamp,
    last_seen_at timestamp,
    role tinytext NOT NULL,
    is_slim boolean NOT NULL,
    created_at timestamp NOT NULL DEFAULT now(),
    updated_at timestamp NOT NULL DEFAULT now(),
    updated_by uuid,
    PRIMARY KEY(id)
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_profiles_mc_uuid ON profiles(mc_uuid);
CREATE UNIQUE INDEX IF NOT EXISTS idx_profiles_mc_username ON profiles(mc_username);
CREATE INDEX IF NOT EXISTS idx_profiles_owner_user_id ON profiles(owner_user_id);

CREATE TABLE IF NOT EXISTS profile_accesses (
    id uuid NOT NULL DEFAULT UUID_v4(),
    mc_uuid uuid NOT NULL,
    season_id uuid NOT NULL,
    source tinytext NOT NULL,
    created_at timestamp NOT NULL DEFAULT now(),
    updated_at timestamp NOT NULL DEFAULT now(),
    updated_by uuid,
    PRIMARY KEY (id),
    FOREIGN KEY (mc_uuid) REFERENCES profiles(mc_uuid),
    FOREIGN KEY (season_id) REFERENCES seasons(id)
);
CREATE TABLE IF NOT EXISTS profile_playtimes (
    id uuid NOT NULL DEFAULT UUID_v4(),
    mc_uuid uuid NOT NULL,
    season_id uuid NOT NULL,
    playtime bigint NOT NULL,
    updated_at timestamp NOT NULL DEFAULT now(),
    PRIMARY KEY (id),
    FOREIGN KEY (mc_uuid) REFERENCES profiles(mc_uuid),
    FOREIGN KEY (season_id) REFERENCES seasons(id)
);
CREATE TABLE IF NOT EXISTS profile_violations (
    id uuid NOT NULL DEFAULT UUID_v4(),
    mc_uuid uuid NOT NULL,
    season_id uuid NOT NULL,
    violation text NOT NULL,
    updated_at timestamp NOT NULL DEFAULT now(),
    PRIMARY KEY (id),
    FOREIGN KEY (mc_uuid) REFERENCES profiles(mc_uuid),
    FOREIGN KEY (season_id) REFERENCES seasons(id)
);