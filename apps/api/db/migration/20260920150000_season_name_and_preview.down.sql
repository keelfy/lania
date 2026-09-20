-- The dates that the up migration replaced are not restored.
ALTER TABLE seasons
    DROP COLUMN IF EXISTS preview_image,
    DROP COLUMN IF EXISTS name;
