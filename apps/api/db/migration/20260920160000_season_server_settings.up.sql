-- Technical connection settings are managed with the season metadata.
-- The password is never returned by the API.
ALTER TABLE seasons
    ADD COLUMN server_ip varchar(45),
    ADD COLUMN server_port smallint unsigned,
    ADD COLUMN is_active boolean NOT NULL DEFAULT false,
    ADD COLUMN rcon_password text;

-- Startup copies ACTIVE_SEASON_ID into this column once for existing deployments.
