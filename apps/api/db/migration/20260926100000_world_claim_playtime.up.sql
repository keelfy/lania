-- plan_server is the Plan name of the world server, where playtime for claims is counted; NULL counts the whole
-- season network.
-- claim_min_playtime_hours is how long a profile must have played on the world server before it can claim
-- chunks; 0 lets anyone with season access claim.
ALTER TABLE season_worlds
    ADD COLUMN IF NOT EXISTS plan_server varchar(100) NULL DEFAULT NULL AFTER claim_dimensions,
    ADD COLUMN IF NOT EXISTS claim_min_playtime_hours int NOT NULL DEFAULT 5 AFTER plan_server;
