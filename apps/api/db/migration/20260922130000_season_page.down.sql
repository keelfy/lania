DROP TABLE IF EXISTS season_screenshot_authors;
DROP TABLE IF EXISTS season_screenshots;

ALTER TABLE seasons
    DROP COLUMN IF EXISTS game_version,
    DROP COLUMN IF EXISTS world_url;
