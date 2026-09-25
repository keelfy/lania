package sql

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/domain"
)

// execAffected runs an exec statement of a profile merge and returns how many rows it touched.
func execAffected(ctx context.Context, x queryable, query string, args ...any) (int64, error) {
	res, err := x.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

const lockProfilesForMerge = `
SELECT id FROM profiles WHERE id IN (?, ?) ORDER BY id FOR UPDATE
`

// LockProfilesForMerge locks both profile rows, in the id order MySQL scans them in, so two merges racing on
// the same two profiles never deadlock on each other regardless of which one is passed as the source.
func (q *queries) LockProfilesForMerge(ctx context.Context, firstProfileID, secondProfileID uuid.UUID) error {
	rows, err := q.x.QueryContext(ctx, lockProfilesForMerge, firstProfileID, secondProfileID)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return err
		}
	}
	return rows.Err()
}

const findLiveSyncedSeasonNamesWithPlaytime = `
SELECT s.name
FROM profile_playtimes pt
JOIN seasons s ON s.id = pt.season_id
WHERE pt.mc_uuid = ? AND pt.playtime > 0 AND s.is_active = 1 AND s.shell_address IS NOT NULL
ORDER BY s.name
`

// FindLiveSyncedSeasonNamesWithPlaytime returns the seasons where the profile has playtime and a shell still
// syncs it every minute. A merge in one of these seasons would have its summed playtime overwritten by the sync.
func (q *queries) FindLiveSyncedSeasonNamesWithPlaytime(ctx context.Context, mcUUID uuid.UUID) ([]string, error) {
	rows, err := q.x.QueryContext(ctx, findLiveSyncedSeasonNamesWithPlaytime, mcUUID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	names := make([]string, 0)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		names = append(names, name)
	}
	return names, rows.Err()
}

// Profile playtime: seasons the target does not have move over, seasons both have get summed into the target.
const moveNonOverlappingPlaytime = `
UPDATE profile_playtimes
SET mc_uuid = ?
WHERE mc_uuid = ? AND season_id NOT IN (
	SELECT season_id FROM (SELECT season_id FROM profile_playtimes WHERE mc_uuid = ?) existing
)
`

const sumOverlappingPlaytime = `
UPDATE profile_playtimes tgt
JOIN profile_playtimes src ON src.season_id = tgt.season_id AND src.mc_uuid = ?
SET
	tgt.playtime = tgt.playtime + src.playtime,
	tgt.last_seen_at = GREATEST(COALESCE(tgt.last_seen_at, src.last_seen_at), COALESCE(src.last_seen_at, tgt.last_seen_at)),
	tgt.updated_at = NOW()
WHERE tgt.mc_uuid = ?
`

const deleteRemainingPlaytime = `
DELETE FROM profile_playtimes WHERE mc_uuid = ?
`

// Profile accesses have no unique key, so a season can end up with two non-revoked rows after a plain move.
// The source's active access is dropped first for every season the target already has one, everything else moves,
// including revoked rows kept as history.
const dropDuplicateActiveAccesses = `
DELETE FROM profile_accesses
WHERE mc_uuid = ? AND revoked_at IS NULL AND season_id IN (
	SELECT season_id FROM (SELECT season_id FROM profile_accesses WHERE mc_uuid = ? AND revoked_at IS NULL) existing
)
`

const moveProfileAccesses = `
UPDATE profile_accesses SET mc_uuid = ? WHERE mc_uuid = ?
`

const moveProfileViolations = `
UPDATE profile_violations SET mc_uuid = ? WHERE mc_uuid = ?
`

const deleteProfileMojangUUID = `
DELETE FROM profile_mojang_uuids WHERE mc_uuid = ?
`

// Name color and name prefix options are unique on (profile_id, item_id, for_season_id) regardless of revoked_at.
// A clashing target row that is revoked gets un-revoked when the source's matching row is active, then the
// source's clashing row is dropped and whatever is left (no clash) moves over.
const unrevokeTargetNameColorOptions = `
UPDATE profile_name_color_options tgt
JOIN profile_name_color_options src
	ON src.profile_id = ? AND src.name_color_id = tgt.name_color_id AND src.for_season_id <=> tgt.for_season_id
SET tgt.revoked_at = NULL, tgt.revoked_by = NULL
WHERE tgt.profile_id = ? AND tgt.revoked_at IS NOT NULL AND src.revoked_at IS NULL
`

const dropDuplicateNameColorOptions = `
DELETE src FROM profile_name_color_options src
JOIN profile_name_color_options tgt
	ON tgt.profile_id = ? AND tgt.name_color_id = src.name_color_id AND tgt.for_season_id <=> src.for_season_id
WHERE src.profile_id = ?
`

const moveNameColorOptions = `
UPDATE profile_name_color_options SET profile_id = ? WHERE profile_id = ?
`

const unrevokeTargetNamePrefixOptions = `
UPDATE profile_name_prefix_options tgt
JOIN profile_name_prefix_options src
	ON src.profile_id = ? AND src.name_prefix_id = tgt.name_prefix_id AND src.for_season_id <=> tgt.for_season_id
SET tgt.revoked_at = NULL, tgt.revoked_by = NULL
WHERE tgt.profile_id = ? AND tgt.revoked_at IS NOT NULL AND src.revoked_at IS NULL
`

const dropDuplicateNamePrefixOptions = `
DELETE src FROM profile_name_prefix_options src
JOIN profile_name_prefix_options tgt
	ON tgt.profile_id = ? AND tgt.name_prefix_id = src.name_prefix_id AND tgt.for_season_id <=> src.for_season_id
WHERE src.profile_id = ?
`

const moveNamePrefixOptions = `
UPDATE profile_name_prefix_options SET profile_id = ? WHERE profile_id = ?
`

// Season cosmetics and legacy prefixes key on (profile_id, season_id) / (profile_id, type): the target's own
// selection wins on a clash, the source's row is dropped, and a season/type the target lacks moves over.
const dropDuplicateSeasonCosmetics = `
DELETE FROM profile_season_cosmetics
WHERE profile_id = ? AND season_id IN (
	SELECT season_id FROM (SELECT season_id FROM profile_season_cosmetics WHERE profile_id = ?) existing
)
`

const moveSeasonCosmetics = `
UPDATE profile_season_cosmetics SET profile_id = ? WHERE profile_id = ?
`

const dropDuplicateProfilePrefixes = `
DELETE FROM profile_prefixes
WHERE profile_id = ? AND type IN (
	SELECT type FROM (SELECT type FROM profile_prefixes WHERE profile_id = ?) existing
)
`

const moveProfilePrefixes = `
UPDATE profile_prefixes SET profile_id = ? WHERE profile_id = ?
`

const moveOrderItems = `
UPDATE order_items SET profile_id = ? WHERE profile_id = ?
`

// basket_items is unique on (user_id, product_id, profile_id, season_id); NULL season_id never clashes in MySQL,
// which matches dropping only the rows the unique index would actually reject.
const dropDuplicateBasketItems = `
DELETE FROM basket_items
WHERE profile_id = ? AND (user_id, product_id, season_id) IN (
	SELECT user_id, product_id, season_id FROM (
		SELECT user_id, product_id, season_id FROM basket_items WHERE profile_id = ?
	) existing
)
`

const moveBasketItems = `
UPDATE basket_items SET profile_id = ? WHERE profile_id = ?
`

// season_screenshot_authors is unique on (screenshot_id, profile_id): a screenshot already crediting the
// target keeps that credit, the source's clashing credit is dropped, everything else moves.
const dropDuplicateScreenshotAuthors = `
DELETE FROM season_screenshot_authors
WHERE profile_id = ? AND screenshot_id IN (
	SELECT screenshot_id FROM (SELECT screenshot_id FROM season_screenshot_authors WHERE profile_id = ?) existing
)
`

const moveScreenshotAuthors = `
UPDATE season_screenshot_authors SET profile_id = ? WHERE profile_id = ?
`

// chunk_claims has no uniqueness per profile: every claim, active or released, moves to the target.
const moveChunkClaims = `
UPDATE chunk_claims SET profile_id = ? WHERE profile_id = ?
`

// notifications keep the profile id inside their JSON payload, not as a column, so an old bell menu link
// still resolves after the merge.
const repointNotificationProfileID = `
UPDATE notifications
SET payload = JSON_SET(payload, '$.profileId', ?)
WHERE JSON_UNQUOTE(JSON_EXTRACT(payload, '$.profileId')) = ?
`

// An open verification request of the source is dropped: it proves nothing for the target's UUID.
const deleteSourceVerification = `
DELETE FROM profile_verifications WHERE profile_id = ?
`

// MergeProfileData moves every table that references the source profile into the target profile and reports
// how many rows each table moved, summed or dropped as a duplicate. It must run inside a transaction that also
// locked both profiles with LockProfilesForMerge; it does not touch the profiles row itself.
func (q *queries) MergeProfileData(ctx context.Context, sourceProfileID, sourceMcUUID, targetProfileID, targetMcUUID uuid.UUID) (*domain.ProfileMergeCounts, error) {
	counts := &domain.ProfileMergeCounts{}
	x := q.x

	var err error
	if counts.PlaytimeMoved, err = execAffected(ctx, x, moveNonOverlappingPlaytime, targetMcUUID, sourceMcUUID, targetMcUUID); err != nil {
		return nil, err
	}
	if counts.PlaytimeSummed, err = execAffected(ctx, x, sumOverlappingPlaytime, sourceMcUUID, targetMcUUID); err != nil {
		return nil, err
	}
	if _, err = execAffected(ctx, x, deleteRemainingPlaytime, sourceMcUUID); err != nil {
		return nil, err
	}

	if counts.AccessesDropped, err = execAffected(ctx, x, dropDuplicateActiveAccesses, sourceMcUUID, targetMcUUID); err != nil {
		return nil, err
	}
	if counts.AccessesMoved, err = execAffected(ctx, x, moveProfileAccesses, targetMcUUID, sourceMcUUID); err != nil {
		return nil, err
	}

	if counts.ViolationsMoved, err = execAffected(ctx, x, moveProfileViolations, targetMcUUID, sourceMcUUID); err != nil {
		return nil, err
	}

	if _, err = execAffected(ctx, x, deleteProfileMojangUUID, sourceMcUUID); err != nil {
		return nil, err
	}

	if _, err = execAffected(ctx, x, unrevokeTargetNameColorOptions, targetProfileID, targetProfileID); err != nil {
		return nil, err
	}
	if counts.NameColorOptionsDropped, err = execAffected(ctx, x, dropDuplicateNameColorOptions, targetProfileID, sourceProfileID); err != nil {
		return nil, err
	}
	if counts.NameColorOptionsMoved, err = execAffected(ctx, x, moveNameColorOptions, targetProfileID, sourceProfileID); err != nil {
		return nil, err
	}

	if _, err = execAffected(ctx, x, unrevokeTargetNamePrefixOptions, targetProfileID, targetProfileID); err != nil {
		return nil, err
	}
	if counts.NamePrefixOptionsDropped, err = execAffected(ctx, x, dropDuplicateNamePrefixOptions, targetProfileID, sourceProfileID); err != nil {
		return nil, err
	}
	if counts.NamePrefixOptionsMoved, err = execAffected(ctx, x, moveNamePrefixOptions, targetProfileID, sourceProfileID); err != nil {
		return nil, err
	}

	if counts.SeasonCosmeticsDropped, err = execAffected(ctx, x, dropDuplicateSeasonCosmetics, sourceProfileID, targetProfileID); err != nil {
		return nil, err
	}
	if counts.SeasonCosmeticsMoved, err = execAffected(ctx, x, moveSeasonCosmetics, targetProfileID, sourceProfileID); err != nil {
		return nil, err
	}

	if counts.PrefixesDropped, err = execAffected(ctx, x, dropDuplicateProfilePrefixes, sourceProfileID, targetProfileID); err != nil {
		return nil, err
	}
	if counts.PrefixesMoved, err = execAffected(ctx, x, moveProfilePrefixes, targetProfileID, sourceProfileID); err != nil {
		return nil, err
	}

	if counts.OrderItemsMoved, err = execAffected(ctx, x, moveOrderItems, targetProfileID, sourceProfileID); err != nil {
		return nil, err
	}

	if counts.BasketItemsDropped, err = execAffected(ctx, x, dropDuplicateBasketItems, sourceProfileID, targetProfileID); err != nil {
		return nil, err
	}
	if counts.BasketItemsMoved, err = execAffected(ctx, x, moveBasketItems, targetProfileID, sourceProfileID); err != nil {
		return nil, err
	}

	if counts.NotificationsRepointed, err = execAffected(ctx, x, repointNotificationProfileID, targetProfileID.String(), sourceProfileID.String()); err != nil {
		return nil, err
	}

	if counts.ScreenshotAuthorsDropped, err = execAffected(ctx, x, dropDuplicateScreenshotAuthors, sourceProfileID, targetProfileID); err != nil {
		return nil, err
	}
	if counts.ScreenshotAuthorsMoved, err = execAffected(ctx, x, moveScreenshotAuthors, targetProfileID, sourceProfileID); err != nil {
		return nil, err
	}

	if counts.ChunkClaimsMoved, err = execAffected(ctx, x, moveChunkClaims, targetProfileID, sourceProfileID); err != nil {
		return nil, err
	}

	if _, err = execAffected(ctx, x, deleteSourceVerification, sourceProfileID); err != nil {
		return nil, err
	}

	return counts, nil
}

// A license verification of the target stays only while its owner stays, like in SetProfileOwner.
const updateProfileAfterMerge = `
UPDATE profiles
SET verified_mc_uuid = IF(owner_user_id <=> ?, verified_mc_uuid, NULL),
	verified_at = IF(owner_user_id <=> ?, verified_at, NULL),
	owner_user_id = ?, first_seen_at = ?, last_seen_at = ?, updated_at = NOW(), updated_by = ?
WHERE id = ?
`

// UpdateProfileAfterMerge writes the merged owner and seen dates onto the target profile.
func (q *queries) UpdateProfileAfterMerge(ctx context.Context, targetProfileID uuid.UUID, ownerUserID *uuid.UUID, firstSeenAt, lastSeenAt *time.Time, updatedBy uuid.UUID) error {
	_, err := q.x.ExecContext(ctx, updateProfileAfterMerge, ownerUserID, ownerUserID, ownerUserID, firstSeenAt, lastSeenAt, updatedBy, targetProfileID)
	return err
}

const deleteProfile = `
DELETE FROM profiles WHERE id = ?
`

// DeleteProfile removes the profile row. Every table that referenced it by mc_uuid or id must be cleared first.
func (q *queries) DeleteProfile(ctx context.Context, profileID uuid.UUID) error {
	_, err := q.x.ExecContext(ctx, deleteProfile, profileID)
	return err
}

const insertProfileMerge = `
INSERT INTO profile_merges (
	source_profile_id,
	source_mc_uuid,
	source_mc_username,
	target_profile_id,
	merged_by,
	summary,
	created_at
) VALUES (?, ?, ?, ?, ?, ?, now())
`

type InsertProfileMergeParams struct {
	SourceProfileID     uuid.UUID
	SourceMinecraftUUID uuid.UUID
	SourceUsername      string
	TargetProfileID     uuid.UUID
	MergedBy            uuid.UUID
	Summary             domain.ProfileMergeCounts
}

func (q *queries) InsertProfileMerge(ctx context.Context, arg InsertProfileMergeParams) error {
	summary, err := json.Marshal(arg.Summary)
	if err != nil {
		return err
	}
	_, err = q.x.ExecContext(ctx, insertProfileMerge,
		arg.SourceProfileID,
		arg.SourceMinecraftUUID,
		arg.SourceUsername,
		arg.TargetProfileID,
		arg.MergedBy,
		[]byte(summary),
	)
	return err
}

const findProfileMergesByTargetProfileID = `
SELECT id, source_profile_id, source_mc_uuid, source_mc_username, target_profile_id, merged_by, summary, created_at
FROM profile_merges
WHERE target_profile_id = ?
ORDER BY created_at DESC
`

// FindProfileMergesByTargetProfileID returns every profile merged into the profile, newest first.
func (q *queries) FindProfileMergesByTargetProfileID(ctx context.Context, targetProfileID uuid.UUID) ([]*domain.ProfileMerge, error) {
	rows, err := q.x.QueryContext(ctx, findProfileMergesByTargetProfileID, targetProfileID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	merges := make([]*domain.ProfileMerge, 0)
	for rows.Next() {
		var merge domain.ProfileMerge
		var summary json.RawMessage
		err := rows.Scan(
			&merge.ID,
			&merge.SourceProfileID,
			&merge.SourceMinecraftUUID,
			&merge.SourceUsername,
			&merge.TargetProfileID,
			&merge.MergedBy,
			&summary,
			&merge.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(summary, &merge.Summary); err != nil {
			return nil, err
		}
		merges = append(merges, &merge)
	}
	return merges, rows.Err()
}
