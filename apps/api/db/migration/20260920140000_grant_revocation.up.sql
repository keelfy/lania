-- A revoked grant stays in its table as history. Everything that reads active grants must filter on revoked_at IS NULL.
ALTER TABLE profile_accesses
    ADD COLUMN IF NOT EXISTS revoked_at timestamp NULL DEFAULT NULL,
    ADD COLUMN IF NOT EXISTS revoked_by uuid NULL DEFAULT NULL;

-- created_by is the admin who gave the option. It stays empty for options that come from an order.
ALTER TABLE profile_name_color_options
    ADD COLUMN IF NOT EXISTS created_by uuid NULL DEFAULT NULL,
    ADD COLUMN IF NOT EXISTS revoked_at timestamp NULL DEFAULT NULL,
    ADD COLUMN IF NOT EXISTS revoked_by uuid NULL DEFAULT NULL;

ALTER TABLE profile_name_prefix_options
    ADD COLUMN IF NOT EXISTS created_by uuid NULL DEFAULT NULL,
    ADD COLUMN IF NOT EXISTS revoked_at timestamp NULL DEFAULT NULL,
    ADD COLUMN IF NOT EXISTS revoked_by uuid NULL DEFAULT NULL;
