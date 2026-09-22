-- game_version is what the client needs to join, e.g. "1.21.1" or "1.20.1 (Create 6)".
-- world_url is an absolute http(s) link to the world archive of a finished season.
ALTER TABLE seasons
    ADD COLUMN IF NOT EXISTS game_version varchar(64) NULL DEFAULT NULL,
    ADD COLUMN IF NOT EXISTS world_url varchar(512) NULL DEFAULT NULL;

-- A screenshot of a season, shown in the feed on the season page.
-- image is the location as the image proxy reads it, the same form as seasons.preview_image.
CREATE TABLE IF NOT EXISTS season_screenshots (
    id uuid NOT NULL DEFAULT UUID_v4(),
    season_id uuid NOT NULL,
    image varchar(512) NOT NULL,
    title varchar(255) NULL,
    position int NOT NULL DEFAULT 0,
    created_at timestamp NOT NULL DEFAULT now(),
    PRIMARY KEY (id),
    FOREIGN KEY (season_id) REFERENCES seasons(id)
);
CREATE INDEX IF NOT EXISTS idx_season_screenshots_season ON season_screenshots(season_id, position, created_at);

-- A screenshot may credit several players; a player may be credited on several screenshots.
-- position orders the credits shown under one screenshot ("keelfy, Foxizans" reads left to right).
CREATE TABLE IF NOT EXISTS season_screenshot_authors (
    screenshot_id uuid NOT NULL,
    profile_id uuid NOT NULL,
    position int NOT NULL DEFAULT 0,
    PRIMARY KEY (screenshot_id, profile_id),
    FOREIGN KEY (screenshot_id) REFERENCES season_screenshots(id),
    FOREIGN KEY (profile_id) REFERENCES profiles(id)
);
