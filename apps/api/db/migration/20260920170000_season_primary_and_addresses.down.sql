-- The previous application supports one active season and uses it as its primary context.
UPDATE seasons SET is_active = is_primary;

ALTER TABLE seasons
    ADD COLUMN season_number integer NOT NULL DEFAULT 0 AFTER id,
    CHANGE COLUMN public_address server_ip varchar(255),
    ADD COLUMN server_port smallint unsigned,
    DROP COLUMN system_address,
    DROP COLUMN rcon_port,
    DROP COLUMN is_primary,
    DROP COLUMN preregistration,
    DROP COLUMN free_registration;
