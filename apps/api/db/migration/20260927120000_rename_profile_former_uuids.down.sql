ALTER TABLE profile_former_uuids DROP FOREIGN KEY IF EXISTS fk_profile_former_uuids_profile_id;
ALTER TABLE profile_former_uuids DROP INDEX IF EXISTS fk_profile_former_uuids_profile_id;
ALTER TABLE profile_former_uuids
    ADD CONSTRAINT fk_profile_merged_uuids_profile_id FOREIGN KEY IF NOT EXISTS (profile_id) REFERENCES profiles(id) ON DELETE CASCADE;

RENAME TABLE IF EXISTS profile_former_uuids TO profile_merged_uuids;
