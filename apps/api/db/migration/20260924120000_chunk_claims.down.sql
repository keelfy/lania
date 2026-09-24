DROP TABLE IF EXISTS chunk_claims;

ALTER TABLE seasons
    DROP COLUMN IF EXISTS map_url,
    DROP COLUMN IF EXISTS claim_limit;
