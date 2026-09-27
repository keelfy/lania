package sql

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/domain"
)

// A premium conflict is over after a rename: the profile no longer uses the nickname someone else bought.
const renameProfile = `
UPDATE profiles
SET mc_uuid = ?, mc_username = ?, premium_conflict = 0, updated_at = NOW(), updated_by = ?
WHERE id = ? AND mc_uuid = ?
`

// RenameProfile gives the profile a new nickname and UUID and ends its premium conflict. The mc_uuid foreign keys
// cascade, so playtime, accesses, violations and the Mojang lookup move along. It reports false when the profile is
// no longer at oldMcUUID.
func (q *queries) RenameProfile(ctx context.Context, profileID, oldMcUUID, newMcUUID uuid.UUID, newUsername string, updatedBy *uuid.UUID) (bool, error) {
	res, err := q.x.ExecContext(ctx, renameProfile, newMcUUID, newUsername, updatedBy, profileID, oldMcUUID)
	if err != nil {
		return false, err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return affected > 0, nil
}

const insertProfileUsernameChange = `
INSERT INTO profile_username_changes (
	profile_id,
	old_username,
	new_username,
	old_mc_uuid,
	new_mc_uuid,
	source,
	changed_by,
	created_at
) VALUES (?, ?, ?, ?, ?, ?, ?, now())
`

type InsertProfileUsernameChangeParams struct {
	ProfileID        uuid.UUID
	OldUsername      string
	NewUsername      string
	OldMinecraftUUID uuid.UUID
	NewMinecraftUUID uuid.UUID
	Source           domain.ProfileUsernameChangeSource
	ChangedBy        *uuid.UUID
}

func (q *queries) InsertProfileUsernameChange(ctx context.Context, arg InsertProfileUsernameChangeParams) error {
	_, err := q.x.ExecContext(ctx, insertProfileUsernameChange,
		arg.ProfileID,
		arg.OldUsername,
		arg.NewUsername,
		arg.OldMinecraftUUID,
		arg.NewMinecraftUUID,
		arg.Source,
		arg.ChangedBy,
	)
	return err
}

const findLastProfileUsernameChangeAt = `
SELECT MAX(created_at) FROM profile_username_changes WHERE profile_id = ? AND source = ?
`

// FindLastProfileUsernameChangeAt returns when the profile last changed its nickname from the source, nil if never.
func (q *queries) FindLastProfileUsernameChangeAt(ctx context.Context, profileID uuid.UUID, source domain.ProfileUsernameChangeSource) (*time.Time, error) {
	var changedAt *time.Time
	err := q.x.QueryRowContext(ctx, findLastProfileUsernameChangeAt, profileID, source).Scan(&changedAt)
	return changedAt, err
}

const findLastProfileUsernameChangeAtByProfileIDs = `
SELECT profile_id, MAX(created_at) FROM profile_username_changes
WHERE source = ? AND profile_id IN (%s)
GROUP BY profile_id
`

// FindLastProfileUsernameChangeAtByProfileIDs is FindLastProfileUsernameChangeAt for several profiles.
// Profiles that never changed their nickname from the source are absent from the result.
func (q *queries) FindLastProfileUsernameChangeAtByProfileIDs(ctx context.Context, profileIDs uuid.UUIDs, source domain.ProfileUsernameChangeSource) (map[uuid.UUID]time.Time, error) {
	res := make(map[uuid.UUID]time.Time, len(profileIDs))
	if len(profileIDs) == 0 {
		return res, nil
	}

	args := make([]any, 0, len(profileIDs)+1)
	args = append(args, source)
	for _, profileID := range profileIDs {
		args = append(args, profileID)
	}
	placeholders := strings.TrimSuffix(strings.Repeat("?,", len(profileIDs)), ",")
	rows, err := q.x.QueryContext(ctx, fmt.Sprintf(findLastProfileUsernameChangeAtByProfileIDs, placeholders), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var profileID uuid.UUID
		var changedAt time.Time
		if err := rows.Scan(&profileID, &changedAt); err != nil {
			return nil, err
		}
		res[profileID] = changedAt
	}
	return res, rows.Err()
}

const findProfileUsernameChanges = `
SELECT id, profile_id, old_username, new_username, old_mc_uuid, new_mc_uuid, source, changed_by, created_at
FROM profile_username_changes
WHERE profile_id = ?
ORDER BY created_at DESC
`

// FindProfileUsernameChanges returns every nickname change of the profile, newest first.
func (q *queries) FindProfileUsernameChanges(ctx context.Context, profileID uuid.UUID) ([]*domain.ProfileUsernameChange, error) {
	rows, err := q.x.QueryContext(ctx, findProfileUsernameChanges, profileID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	changes := make([]*domain.ProfileUsernameChange, 0)
	for rows.Next() {
		var change domain.ProfileUsernameChange
		err := rows.Scan(
			&change.ID,
			&change.ProfileID,
			&change.OldUsername,
			&change.NewUsername,
			&change.OldMinecraftUUID,
			&change.NewMinecraftUUID,
			&change.Source,
			&change.ChangedBy,
			&change.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		changes = append(changes, &change)
	}
	return changes, rows.Err()
}

// A former UUID moves to the profile that took it over last, e.g. the target of a later merge.
const insertProfileFormerUUID = `
INSERT INTO profile_former_uuids (mc_uuid, profile_id, created_at)
VALUES (?, ?, now())
ON DUPLICATE KEY UPDATE profile_id = VALUES(profile_id)
`

// InsertProfileFormerUUID keeps an in-game UUID the profile left, so the player sync still sums its playtime into
// the profile and nobody else can take it.
func (q *queries) InsertProfileFormerUUID(ctx context.Context, mcUUID, profileID uuid.UUID) error {
	_, err := q.x.ExecContext(ctx, insertProfileFormerUUID, mcUUID, profileID)
	return err
}

const deleteProfileFormerUUID = `
DELETE FROM profile_former_uuids WHERE mc_uuid = ? AND profile_id = ?
`

// DeleteProfileFormerUUID forgets a former UUID of the profile, when the profile takes that UUID back.
func (q *queries) DeleteProfileFormerUUID(ctx context.Context, mcUUID, profileID uuid.UUID) error {
	_, err := q.x.ExecContext(ctx, deleteProfileFormerUUID, mcUUID, profileID)
	return err
}

// A UUID is held by the profile that has it now, had it before a premium rekey, or left it by a merge or a rename.
const findProfileIDsHoldingMinecraftUUID = `
SELECT id FROM profiles WHERE mc_uuid = ? OR legacy_mc_uuid = ?
UNION
SELECT profile_id FROM profile_former_uuids WHERE mc_uuid = ?
`

// FindProfileIDsHoldingMinecraftUUID returns every profile whose playtime the UUID counts toward.
func (q *queries) FindProfileIDsHoldingMinecraftUUID(ctx context.Context, mcUUID uuid.UUID) (uuid.UUIDs, error) {
	rows, err := q.x.QueryContext(ctx, findProfileIDsHoldingMinecraftUUID, mcUUID, mcUUID, mcUUID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	profileIDs := make(uuid.UUIDs, 0)
	for rows.Next() {
		var profileID uuid.UUID
		if err := rows.Scan(&profileID); err != nil {
			return nil, err
		}
		profileIDs = append(profileIDs, profileID)
	}
	return profileIDs, rows.Err()
}
