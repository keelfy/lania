-- Rekeyed profiles keep their Mojang UUID: the old offline UUID is only in legacy_mc_uuid, which is dropped here.

SET @accesses_fk = (
    SELECT CONSTRAINT_NAME FROM information_schema.KEY_COLUMN_USAGE
    WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'profile_accesses' AND COLUMN_NAME = 'mc_uuid'
      AND REFERENCED_TABLE_NAME = 'profiles'
    LIMIT 1
);
SET @drop_accesses_fk = IF(@accesses_fk IS NULL, 'DO 0', CONCAT('ALTER TABLE profile_accesses DROP FOREIGN KEY `', @accesses_fk, '`'));
PREPARE drop_accesses_fk FROM @drop_accesses_fk;
EXECUTE drop_accesses_fk;
DEALLOCATE PREPARE drop_accesses_fk;
ALTER TABLE profile_accesses ADD CONSTRAINT fk_profile_accesses_mc_uuid FOREIGN KEY (mc_uuid) REFERENCES profiles(mc_uuid);

SET @playtimes_fk = (
    SELECT CONSTRAINT_NAME FROM information_schema.KEY_COLUMN_USAGE
    WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'profile_playtimes' AND COLUMN_NAME = 'mc_uuid'
      AND REFERENCED_TABLE_NAME = 'profiles'
    LIMIT 1
);
SET @drop_playtimes_fk = IF(@playtimes_fk IS NULL, 'DO 0', CONCAT('ALTER TABLE profile_playtimes DROP FOREIGN KEY `', @playtimes_fk, '`'));
PREPARE drop_playtimes_fk FROM @drop_playtimes_fk;
EXECUTE drop_playtimes_fk;
DEALLOCATE PREPARE drop_playtimes_fk;
ALTER TABLE profile_playtimes ADD CONSTRAINT fk_profile_playtimes_mc_uuid FOREIGN KEY (mc_uuid) REFERENCES profiles(mc_uuid);

SET @violations_fk = (
    SELECT CONSTRAINT_NAME FROM information_schema.KEY_COLUMN_USAGE
    WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'profile_violations' AND COLUMN_NAME = 'mc_uuid'
      AND REFERENCED_TABLE_NAME = 'profiles'
    LIMIT 1
);
SET @drop_violations_fk = IF(@violations_fk IS NULL, 'DO 0', CONCAT('ALTER TABLE profile_violations DROP FOREIGN KEY `', @violations_fk, '`'));
PREPARE drop_violations_fk FROM @drop_violations_fk;
EXECUTE drop_violations_fk;
DEALLOCATE PREPARE drop_violations_fk;
ALTER TABLE profile_violations ADD CONSTRAINT fk_profile_violations_mc_uuid FOREIGN KEY (mc_uuid) REFERENCES profiles(mc_uuid);

SET @mojang_uuids_fk = (
    SELECT CONSTRAINT_NAME FROM information_schema.KEY_COLUMN_USAGE
    WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'profile_mojang_uuids' AND COLUMN_NAME = 'mc_uuid'
      AND REFERENCED_TABLE_NAME = 'profiles'
    LIMIT 1
);
SET @drop_mojang_uuids_fk = IF(@mojang_uuids_fk IS NULL, 'DO 0', CONCAT('ALTER TABLE profile_mojang_uuids DROP FOREIGN KEY `', @mojang_uuids_fk, '`'));
PREPARE drop_mojang_uuids_fk FROM @drop_mojang_uuids_fk;
EXECUTE drop_mojang_uuids_fk;
DEALLOCATE PREPARE drop_mojang_uuids_fk;
ALTER TABLE profile_mojang_uuids ADD CONSTRAINT fk_profile_mojang_uuids_mc_uuid FOREIGN KEY (mc_uuid) REFERENCES profiles(mc_uuid);

ALTER TABLE profiles
    DROP COLUMN IF EXISTS premium_conflict,
    DROP COLUMN IF EXISTS legacy_mc_uuid;
