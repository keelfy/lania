ALTER TABLE profile_playtimes
    DROP PRIMARY KEY,
    MODIFY playtime bigint NOT NULL,
    ADD COLUMN id uuid NOT NULL DEFAULT UUID_v4() FIRST,
    ADD PRIMARY KEY (id);
