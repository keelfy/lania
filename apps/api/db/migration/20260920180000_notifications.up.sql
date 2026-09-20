-- notifications holds what the site tells a user about, newest first.
-- user_id is an Ory identity, so there is no foreign key.
-- payload keeps the data of the event, the text is built by the frontend from type and payload.
CREATE TABLE IF NOT EXISTS notifications (
    id uuid NOT NULL DEFAULT UUID_v4(),
    user_id uuid NOT NULL,
    type tinytext NOT NULL,
    payload json NOT NULL,
    read_at timestamp NULL DEFAULT NULL,
    created_at timestamp NOT NULL DEFAULT now(),
    PRIMARY KEY (id)
);
CREATE INDEX IF NOT EXISTS idx_notifications_user_id_created_at ON notifications(user_id, created_at);
