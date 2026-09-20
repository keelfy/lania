-- A season has separate player-facing and internal connection settings.
-- The former active season becomes the primary season during the transition.
ALTER TABLE seasons
    DROP COLUMN season_number,
    CHANGE COLUMN server_ip public_address varchar(255),
    DROP COLUMN server_port,
    ADD COLUMN system_address varchar(255),
    ADD COLUMN rcon_port smallint unsigned,
    ADD COLUMN is_primary boolean NOT NULL DEFAULT false,
    ADD COLUMN preregistration boolean NOT NULL DEFAULT false,
    ADD COLUMN free_registration boolean NOT NULL DEFAULT false;

UPDATE seasons SET is_primary = is_active WHERE is_active = true;
