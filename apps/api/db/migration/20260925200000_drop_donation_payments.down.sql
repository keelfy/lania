UPDATE profile_accesses SET source = 'freekassa' WHERE source = 'order';

CREATE TABLE IF NOT EXISTS oauth2_integrations (
    id uuid NOT NULL DEFAULT UUID_v4(),
    service_name text NOT NULL,
    access_token text NOT NULL,
    refresh_token text NOT NULL,
    updated_at timestamp NOT NULL DEFAULT now(),
    PRIMARY KEY (id)
);
INSERT INTO oauth2_integrations (service_name, access_token, refresh_token) VALUES ('donation_alerts', '', '');
