-- profile_merged_uuids becomes profile_former_uuids: it keeps the old in-game UUIDs of a profile, left by a merge or
-- by a nickname change, and the player sync keeps summing their Plan playtime into the profile. Safe to repeat.
RENAME TABLE IF EXISTS profile_merged_uuids TO profile_former_uuids;

ALTER TABLE profile_former_uuids DROP FOREIGN KEY IF EXISTS fk_profile_merged_uuids_profile_id;
ALTER TABLE profile_former_uuids DROP INDEX IF EXISTS fk_profile_merged_uuids_profile_id;
ALTER TABLE profile_former_uuids
    ADD CONSTRAINT fk_profile_former_uuids_profile_id FOREIGN KEY IF NOT EXISTS (profile_id) REFERENCES profiles(id) ON DELETE CASCADE;
