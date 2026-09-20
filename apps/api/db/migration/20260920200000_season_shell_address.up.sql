-- The API talks to a season server only through its shell service.
-- Shell owns the RCON credentials, so the season keeps just the shell address (host:port).
-- Startup copies SHELL_ADDRESS into the primary season once for existing deployments.
ALTER TABLE seasons
    ADD COLUMN shell_address varchar(255),
    DROP COLUMN system_address,
    DROP COLUMN rcon_port,
    DROP COLUMN rcon_password;
