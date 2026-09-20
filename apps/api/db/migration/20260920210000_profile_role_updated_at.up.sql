-- The profile is the source of truth for a role, and role sync pushes recent changes to the season servers.
-- role_updated_at marks when a role last changed. It stays empty for a profile whose role never changed.
ALTER TABLE profiles
    ADD COLUMN role_updated_at timestamp NULL DEFAULT NULL,
    ADD INDEX idx_profiles_role_updated_at (role_updated_at);
