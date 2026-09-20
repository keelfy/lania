ALTER TABLE seasons
    ADD COLUMN system_address varchar(255),
    ADD COLUMN rcon_port smallint unsigned,
    ADD COLUMN rcon_password text,
    DROP COLUMN shell_address;
