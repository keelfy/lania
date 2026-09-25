DROP TABLE IF EXISTS profile_verifications;

ALTER TABLE profiles
    DROP COLUMN IF EXISTS verified_at,
    DROP COLUMN IF EXISTS verified_mc_uuid;
