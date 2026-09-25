-- verified_mc_uuid is the licensed account the owner proved in game; the checkmark shows only while it equals
-- mc_uuid, so a rekey or any other UUID change drops it by itself.
ALTER TABLE profiles
    ADD COLUMN IF NOT EXISTS verified_mc_uuid uuid NULL DEFAULT NULL AFTER premium_conflict,
    ADD COLUMN IF NOT EXISTS verified_at timestamp NULL DEFAULT NULL AFTER verified_mc_uuid;

-- profile_verifications is the open verification request of a profile: the proxy plugin issues the code to the
-- licensed player on join, the owner types it on the site. One request per profile, replaced on each start.
CREATE TABLE IF NOT EXISTS profile_verifications (
    profile_id uuid NOT NULL,
    code varchar(8) NULL DEFAULT NULL,
    attempts int NOT NULL DEFAULT 0,
    expires_at timestamp NOT NULL,
    created_at timestamp NOT NULL DEFAULT now(),
    PRIMARY KEY (profile_id),
    CONSTRAINT fk_profile_verifications_profile_id FOREIGN KEY (profile_id) REFERENCES profiles(id)
);
