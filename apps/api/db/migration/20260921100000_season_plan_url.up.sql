-- Each season server runs its own Plan (player analytics) web interface.
-- Only admins see the link, so it stays out of the public season response.
ALTER TABLE seasons
    ADD COLUMN plan_url varchar(2048);
