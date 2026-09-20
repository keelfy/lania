ALTER TABLE profiles
    DROP INDEX idx_profiles_role_updated_at,
    DROP COLUMN role_updated_at;
