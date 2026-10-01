package sql

import (
	"context"
	stdsql "database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/domain"
)

// The profile columns end with verified_at, which is NULL unless the owner verified the current mc_uuid.
func scanProfileRow(row *stdsql.Row) (*domain.Profile, error) {
	var profile domain.Profile
	err := row.Scan(
		&profile.ID,
		&profile.MinecraftUUID,
		&profile.MinecraftUsername,
		&profile.OwnerUserID,
		&profile.FirstSeenAt,
		&profile.LastSeenAt,
		&profile.Role,
		&profile.IsSlimModel,
		&profile.CreatedAt,
		&profile.UpdatedAt,
		&profile.UpdatedBy,
		&profile.LegacyMinecraftUUID,
		&profile.PremiumConflict,
		&profile.VerifiedAt,
	)
	if err != nil {
		return nil, err
	}
	return &profile, nil
}

// scanProfileRows scans the profile columns followed by extra destinations for columns selected after them.
func scanProfileRows(rows *stdsql.Rows, extra ...any) (*domain.Profile, error) {
	var profile domain.Profile
	err := rows.Scan(append([]any{
		&profile.ID,
		&profile.MinecraftUUID,
		&profile.MinecraftUsername,
		&profile.OwnerUserID,
		&profile.FirstSeenAt,
		&profile.LastSeenAt,
		&profile.Role,
		&profile.IsSlimModel,
		&profile.CreatedAt,
		&profile.UpdatedAt,
		&profile.UpdatedBy,
		&profile.LegacyMinecraftUUID,
		&profile.PremiumConflict,
		&profile.VerifiedAt,
	}, extra...)...)
	if err != nil {
		return nil, err
	}
	return &profile, nil
}

const getProfilesByOwnerUserID = `
SELECT 
	p.id,
	p.mc_uuid,
	p.mc_username,
	p.owner_user_id,
	p.first_seen_at,
	p.last_seen_at,
	p.role,
	p.is_slim,
	p.created_at,
	p.updated_at,
	p.updated_by,
	p.legacy_mc_uuid,
	p.premium_conflict,
	IF(p.verified_mc_uuid <=> p.mc_uuid, p.verified_at, NULL)
FROM profiles p
WHERE owner_user_id = ? 
ORDER BY created_at ASC
`

func (q *queries) GetProfilesByOwnerUserID(ctx context.Context, ownerUserID uuid.UUID) ([]*domain.Profile, error) {
	rows, err := q.x.QueryContext(ctx, getProfilesByOwnerUserID, ownerUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	profiles := make([]*domain.Profile, 0)
	for rows.Next() {
		profile, err := scanProfileRows(rows)
		if err != nil {
			return nil, err
		}
		profiles = append(profiles, profile)
	}
	return profiles, nil
}

const countPublicProfiles = `
SELECT COUNT(p.id) FROM profiles p
%s
`

func (q *queries) CountPublicProfiles(ctx context.Context, search string, only *uuid.UUIDs) (int64, error) {
	where, args := profileWhereClause(search, only)
	row := q.x.QueryRowContext(ctx, fmt.Sprintf(countPublicProfiles, where), args...)
	var count int64
	err := row.Scan(&count)
	return count, err
}

// The cutoff is computed by the database, so it uses the same clock as created_at.
const countRecentProfiles = `
SELECT COUNT(id) FROM profiles WHERE created_at >= DATE_SUB(NOW(), INTERVAL ? DAY)
`

func (q *queries) CountRecentProfiles(ctx context.Context, days int) (int64, error) {
	var count int64
	err := q.x.QueryRowContext(ctx, countRecentProfiles, days).Scan(&count)
	return count, err
}

const publicProfileColumns = `
	p.id,
	p.mc_uuid,
	p.mc_username,
	p.owner_user_id,
	p.first_seen_at,
	p.last_seen_at,
	p.role,
	p.is_slim,
	p.created_at,
	p.updated_at,
	p.updated_by,
	p.legacy_mc_uuid,
	p.premium_conflict,
	IF(p.verified_mc_uuid <=> p.mc_uuid, p.verified_at, NULL)
`

const findPublicProfiles = `
SELECT ` + publicProfileColumns + `
FROM profiles p
%s
%s
ORDER BY %s
LIMIT ? OFFSET ?
`

// The last column is the value the list is sorted by, as text. It goes into the cursor of the next page.
const findPublicProfilesPage = `
SELECT ` + publicProfileColumns + `,
	CAST(%s AS CHAR)
FROM profiles p
%s
%s
ORDER BY %s
LIMIT ?
`

// Playtime is summed over all seasons. The join is only added when the list is sorted by playtime.
const profilePlaytimeJoin = `
LEFT JOIN (
	SELECT mc_uuid, SUM(playtime) AS total_playtime
	FROM profile_season_stats
	GROUP BY mc_uuid
) pt ON pt.mc_uuid = p.mc_uuid
`

var profileSearchEscaper = strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)

// profileWhereClause matches usernames by prefix, so the mc_username index is used.
// A non-nil only keeps just those players, and an empty one matches nobody.
func profileWhereClause(search string, only *uuid.UUIDs) (string, []any) {
	conditions := make([]string, 0, 2)
	args := make([]any, 0)
	if search != "" {
		conditions = append(conditions, "p.mc_username LIKE ?")
		args = append(args, profileSearchEscaper.Replace(search)+"%")
	}
	if only != nil {
		if len(*only) == 0 {
			conditions = append(conditions, "1 = 0")
		} else {
			placeholders := make([]string, len(*only))
			for i, mcUUID := range *only {
				placeholders[i] = "?"
				args = append(args, mcUUID.String())
			}
			conditions = append(conditions, "p.mc_uuid IN ("+strings.Join(placeholders, ", ")+")")
		}
	}
	if len(conditions) == 0 {
		return "", nil
	}
	return "WHERE " + strings.Join(conditions, " AND "), args
}

func getProfileSortDirection(direction string) string {
	switch strings.ToLower(direction) {
	case "desc":
		return "DESC"
	default:
		return "ASC"
	}
}

// Last seen is per season: the list shows the date in the season the visitor looks at.
const profileSeasonLastSeenJoin = `
LEFT JOIN profile_season_stats lspt ON lspt.mc_uuid = p.mc_uuid AND lspt.season_id = ?
`

// profileSort is what a list of profiles is sorted by.
type profileSort struct {
	join     string
	joinArgs []any
	// key is the expression of the sorted value.
	key string
	// nullable keys are listed last, whatever the direction is.
	nullable bool
}

// profileSortOf returns the sort of the column. Last seen is sorted in seasonID, and by the latest date over
// every season when seasonID is uuid.Nil.
func profileSortOf(sortCol string, seasonID uuid.UUID) profileSort {
	switch sortCol {
	case "username":
		return profileSort{key: "p.mc_username"}
	case "first_seen_at":
		return profileSort{key: "p.first_seen_at", nullable: true}
	case "last_seen_at":
		if seasonID == uuid.Nil {
			return profileSort{key: "p.last_seen_at", nullable: true}
		}
		return profileSort{join: profileSeasonLastSeenJoin, joinArgs: []any{seasonID}, key: "lspt.last_seen_at", nullable: true}
	case "playtime":
		return profileSort{join: profilePlaytimeJoin, key: "COALESCE(pt.total_playtime, 0)"}
	default:
		return profileSort{key: "p.created_at"}
	}
}

// orderBy lists profiles without a value last, and id keeps the order stable between pages.
func (s profileSort) orderBy(direction string) string {
	dir := getProfileSortDirection(direction)
	if s.nullable {
		return fmt.Sprintf("%s IS NULL, %s %s, p.id", s.key, s.key, dir)
	}
	return fmt.Sprintf("%s %s, p.id", s.key, dir)
}

// after returns the condition that keeps the profiles listed after the cursor, and its arguments.
// It is spelled out instead of a row comparison, because id is ascending in every direction.
func (s profileSort) after(direction string, cursor *domain.ProfileCursor) (string, []any) {
	if cursor.Value == nil {
		// Only the profiles without a value are left, and they are ordered by id.
		return fmt.Sprintf("(%s IS NULL AND p.id > ?)", s.key), []any{cursor.ID.String()}
	}
	cmp := ">"
	if getProfileSortDirection(direction) == "DESC" {
		cmp = "<"
	}
	condition := fmt.Sprintf("(%[1]s %[2]s ? OR (%[1]s = ? AND p.id > ?))", s.key, cmp)
	if s.nullable {
		condition = fmt.Sprintf("(%s IS NULL OR %s)", s.key, condition)
	}
	return condition, []any{*cursor.Value, *cursor.Value, cursor.ID.String()}
}

// profileOrderBy returns the join needed by the sort column, the arguments of that join and the ORDER BY expression.
func profileOrderBy(sortCol, direction string, seasonID uuid.UUID) (join string, joinArgs []any, orderBy string) {
	sort := profileSortOf(sortCol, seasonID)
	return sort.join, sort.joinArgs, sort.orderBy(direction)
}

func (q *queries) FindPublicProfiles(ctx context.Context, search string, only *uuid.UUIDs, sortCol, direction string, seasonID uuid.UUID, size, from int) ([]*domain.Profile, error) {
	join, joinArgs, orderBy := profileOrderBy(sortCol, direction, seasonID)
	where, whereArgs := profileWhereClause(search, only)
	query := fmt.Sprintf(findPublicProfiles, join, where, orderBy)
	// The join comes before the WHERE clause in the query.
	args := append(joinArgs, whereArgs...)
	args = append(args, size, from)
	rows, err := q.x.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	profiles := make([]*domain.Profile, 0)
	for rows.Next() {
		profile, err := scanProfileRows(rows)
		if err != nil {
			return nil, err
		}
		profiles = append(profiles, profile)
	}
	return profiles, rows.Err()
}

// FindPublicProfilesPage returns up to limit profiles listed after the cursor, or from the start when it is nil.
// The cursors are of the returned profiles, in the same order.
func (q *queries) FindPublicProfilesPage(ctx context.Context, search string, only *uuid.UUIDs, sortCol, direction string, seasonID uuid.UUID, cursor *domain.ProfileCursor, limit int) ([]*domain.Profile, []*domain.ProfileCursor, error) {
	sort := profileSortOf(sortCol, seasonID)
	where, whereArgs := profileWhereClause(search, only)
	if cursor != nil {
		after, afterArgs := sort.after(direction, cursor)
		if where == "" {
			where = "WHERE " + after
		} else {
			where += " AND " + after
		}
		whereArgs = append(whereArgs, afterArgs...)
	}
	query := fmt.Sprintf(findPublicProfilesPage, sort.key, sort.join, where, sort.orderBy(direction))
	// The join comes before the WHERE clause in the query.
	args := append(sort.joinArgs, whereArgs...)
	args = append(args, limit)
	rows, err := q.x.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	profiles := make([]*domain.Profile, 0, limit)
	cursors := make([]*domain.ProfileCursor, 0, limit)
	for rows.Next() {
		var value stdsql.NullString
		profile, err := scanProfileRows(rows, &value)
		if err != nil {
			return nil, nil, err
		}
		next := &domain.ProfileCursor{ID: profile.ID}
		if value.Valid {
			next.Value = &value.String
		}
		profiles = append(profiles, profile)
		cursors = append(cursors, next)
	}
	return profiles, cursors, rows.Err()
}

const findTopProfilePlaytimes = `
SELECT
	p.id,
	p.mc_uuid,
	p.mc_username,
	p.owner_user_id,
	p.first_seen_at,
	p.last_seen_at,
	p.role,
	p.is_slim,
	p.created_at,
	p.updated_at,
	p.updated_by,
	p.legacy_mc_uuid,
	p.premium_conflict,
	IF(p.verified_mc_uuid <=> p.mc_uuid, p.verified_at, NULL),
	pt.playtime
FROM profile_season_stats pt
JOIN profiles p ON p.mc_uuid = pt.mc_uuid
WHERE pt.season_id = ? AND pt.playtime > 0
ORDER BY pt.playtime DESC, p.id
LIMIT ?
`

// FindTopProfilePlaytimes returns players with the most playtime in the season, each with its Profile.
func (q *queries) FindTopProfilePlaytimes(ctx context.Context, seasonID uuid.UUID, limit int) ([]*domain.ProfilePlaytime, error) {
	rows, err := q.x.QueryContext(ctx, findTopProfilePlaytimes, seasonID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	playtimes := make([]*domain.ProfilePlaytime, 0, limit)
	for rows.Next() {
		playtime := &domain.ProfilePlaytime{SeasonID: seasonID}
		profile, err := scanProfileRows(rows, &playtime.Playtime)
		if err != nil {
			return nil, err
		}
		playtime.MinecraftUUID = profile.MinecraftUUID
		playtime.Profile = profile
		playtimes = append(playtimes, playtime)
	}
	return playtimes, rows.Err()
}

const insertProfile = `
INSERT INTO profiles (
	id,
	mc_uuid,
	mc_username,
	owner_user_id,
	role,
	is_slim,
	created_at,
	updated_at,
	updated_by
) VALUES (
	?,
	?,
	?,
	?,
	?,
	?,
	now(),
	now(),
	?
) ON DUPLICATE KEY UPDATE
	mc_uuid = VALUES(mc_uuid),
  mc_username = VALUES(mc_username),
	owner_user_id = COALESCE(owner_user_id, VALUES(owner_user_id)),
	is_slim = VALUES(is_slim),
	updated_at = NOW(),
	updated_by = VALUES(updated_by);
`

type InsertProfileParams struct {
	ID                uuid.UUID
	MinecraftUUID     uuid.UUID
	MinecraftUsername string
	OwnerUserID       *uuid.UUID
	Role              string
	IsSlim            bool
	UpdatedBy         uuid.UUID
}

func (q *queries) InsertProfile(ctx context.Context, arg InsertProfileParams) error {
	_, err := q.x.ExecContext(ctx, insertProfile,
		arg.ID,
		arg.MinecraftUUID,
		arg.MinecraftUsername,
		arg.OwnerUserID,
		arg.Role,
		arg.IsSlim,
		arg.UpdatedBy,
	)
	return err
}

const claimProfile = `
UPDATE profiles
SET owner_user_id = ?, updated_at = NOW(), updated_by = ?
WHERE id = ? AND owner_user_id IS NULL
`

func (q *queries) ClaimProfile(ctx context.Context, profileID, ownerUserID uuid.UUID, updatedBy uuid.UUID) (bool, error) {
	res, err := q.x.ExecContext(ctx, claimProfile, ownerUserID, updatedBy, profileID)
	if err != nil {
		return false, err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return affected > 0, nil
}

// The license verification belongs to the owner who did it, so it is dropped when the owner changes.
const setProfileOwner = `
UPDATE profiles
SET verified_mc_uuid = IF(owner_user_id <=> ?, verified_mc_uuid, NULL),
	verified_at = IF(owner_user_id <=> ?, verified_at, NULL),
	owner_user_id = ?, updated_at = NOW(), updated_by = ?
WHERE id = ?
`

// SetProfileOwner gives the profile to the user, or leaves it without an owner when ownerUserID is nil.
// A new owner loses the license verification of the previous one.
func (q *queries) SetProfileOwner(ctx context.Context, profileID uuid.UUID, ownerUserID, updatedBy *uuid.UUID) error {
	_, err := q.x.ExecContext(ctx, setProfileOwner, ownerUserID, ownerUserID, ownerUserID, updatedBy, profileID)
	return err
}

const insertProfileAcquisition = `
INSERT INTO profile_acquisitions (user_id, profile_id) VALUES (?, ?)
`

func (q *queries) InsertProfileAcquisition(ctx context.Context, userID, profileID uuid.UUID) error {
	_, err := q.x.ExecContext(ctx, insertProfileAcquisition, userID, profileID)
	return err
}

const countRecentProfileAcquisitions = `
SELECT COUNT(id) FROM profile_acquisitions WHERE user_id = ? AND created_at >= DATE_SUB(NOW(), INTERVAL ? DAY)
`

func (q *queries) CountRecentProfileAcquisitions(ctx context.Context, userID uuid.UUID, days int) (int, error) {
	var count int
	err := q.x.QueryRowContext(ctx, countRecentProfileAcquisitions, userID, days).Scan(&count)
	return count, err
}

const findProfileByID = `
SELECT 
	p.id,
	p.mc_uuid,
	p.mc_username,
	p.owner_user_id,
	p.first_seen_at,
	p.last_seen_at,
	p.role,
	p.is_slim,
	p.created_at,
	p.updated_at,
	p.updated_by,
	p.legacy_mc_uuid,
	p.premium_conflict,
	IF(p.verified_mc_uuid <=> p.mc_uuid, p.verified_at, NULL)
FROM profiles p
WHERE p.id = ?
`

func (q *queries) FindProfileByID(ctx context.Context, profileID uuid.UUID) (*domain.Profile, error) {
	row := q.x.QueryRowContext(ctx, findProfileByID, profileID)
	profile, err := scanProfileRow(row)
	return profile, err
}

const findProfileByMinecraftUUID = `
SELECT 
	p.id,
	p.mc_uuid,
	p.mc_username,
	p.owner_user_id,
	p.first_seen_at,
	p.last_seen_at,
	p.role,
	p.is_slim,
	p.created_at,
	p.updated_at,
	p.updated_by,
	p.legacy_mc_uuid,
	p.premium_conflict,
	IF(p.verified_mc_uuid <=> p.mc_uuid, p.verified_at, NULL)
FROM profiles p
WHERE p.mc_uuid = ?
`

func (q *queries) FindProfileByMinecraftUUID(ctx context.Context, minecraftUUID uuid.UUID) (*domain.Profile, error) {
	row := q.x.QueryRowContext(ctx, findProfileByMinecraftUUID, minecraftUUID)
	profile, err := scanProfileRow(row)
	return profile, err
}

const setProfileRole = `
UPDATE profiles
SET role = ?, role_updated_at = NOW(), updated_at = NOW(), updated_by = ?
WHERE id = ?
`

// SetProfileRole stores the role and marks the moment, so role sync picks the change up.
func (q *queries) SetProfileRole(ctx context.Context, profileID uuid.UUID, role domain.Role, updatedBy *uuid.UUID) error {
	_, err := q.x.ExecContext(ctx, setProfileRole, role, updatedBy, profileID)
	return err
}

// The cutoff is computed by the database, so it uses the same clock as role_updated_at.
const findProfileRolesChangedSince = `
SELECT mc_uuid, role
FROM profiles
WHERE role_updated_at >= DATE_SUB(NOW(), INTERVAL ? SECOND)
`

func (q *queries) FindProfileRolesChangedSince(ctx context.Context, since time.Duration) ([]*domain.ProfileRole, error) {
	rows, err := q.x.QueryContext(ctx, findProfileRolesChangedSince, int64(since.Seconds()))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	roles := make([]*domain.ProfileRole, 0)
	for rows.Next() {
		var profileRole domain.ProfileRole
		if err := rows.Scan(&profileRole.MinecraftUUID, &profileRole.Role); err != nil {
			return nil, err
		}
		roles = append(roles, &profileRole)
	}
	return roles, rows.Err()
}

const findMinecraftUUIDsByRoles = `
SELECT mc_uuid FROM profiles WHERE role IN (%s)
`

func (q *queries) FindMinecraftUUIDsByRoles(ctx context.Context, roles []domain.Role) (uuid.UUIDs, error) {
	mcUUIDs := make(uuid.UUIDs, 0)
	if len(roles) == 0 {
		return mcUUIDs, nil
	}

	args := make([]any, len(roles))
	for i, role := range roles {
		args[i] = string(role)
	}
	placeholders := strings.TrimSuffix(strings.Repeat("?, ", len(roles)), ", ")
	rows, err := q.x.QueryContext(ctx, fmt.Sprintf(findMinecraftUUIDsByRoles, placeholders), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var mcUUID uuid.UUID
		if err := rows.Scan(&mcUUID); err != nil {
			return nil, err
		}
		mcUUIDs = append(mcUUIDs, mcUUID)
	}
	return mcUUIDs, rows.Err()
}

const findProfileByUsername = `
SELECT
	p.id,
	p.mc_uuid,
	p.mc_username,
	p.owner_user_id,
	p.first_seen_at,
	p.last_seen_at,
	p.role,
	p.is_slim,
	p.created_at,
	p.updated_at,
	p.updated_by,
	p.legacy_mc_uuid,
	p.premium_conflict,
	IF(p.verified_mc_uuid <=> p.mc_uuid, p.verified_at, NULL)
FROM profiles p
WHERE p.mc_username = ?
`

func (q *queries) FindProfileByUsername(ctx context.Context, username string) (*domain.Profile, error) {
	row := q.x.QueryRowContext(ctx, findProfileByUsername, username)
	return scanProfileRow(row)
}

// legacy_mc_uuid is assigned before mc_uuid: MariaDB reads the already updated value in later assignments.
const rekeyProfile = `
UPDATE profiles
SET legacy_mc_uuid = COALESCE(legacy_mc_uuid, mc_uuid), mc_uuid = ?, updated_at = NOW()
WHERE id = ? AND mc_uuid = ?
`

func (q *queries) RekeyProfile(ctx context.Context, profileID, oldMcUUID, newMcUUID uuid.UUID) (bool, error) {
	res, err := q.x.ExecContext(ctx, rekeyProfile, newMcUUID, profileID, oldMcUUID)
	if err != nil {
		return false, err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return affected > 0, nil
}

const setProfilePremiumConflict = `
UPDATE profiles SET premium_conflict = 1, updated_at = NOW() WHERE mc_uuid = ?
`

func (q *queries) SetProfilePremiumConflict(ctx context.Context, mcUUID uuid.UUID) error {
	_, err := q.x.ExecContext(ctx, setProfilePremiumConflict, mcUUID)
	return err
}

// Old UUIDs of a profile are its offline UUID before a premium rekey and the UUIDs it left by a merge or a rename.
// Every old UUID of a matched profile is returned, not only the matching one, so the sync can sum all of them.
const findLegacyMinecraftUUIDs = `
WITH old_uuids AS (
	SELECT legacy_mc_uuid AS old_mc_uuid, mc_uuid FROM profiles WHERE legacy_mc_uuid IS NOT NULL
	UNION
	SELECT f.mc_uuid, p.mc_uuid FROM profile_former_uuids f JOIN profiles p ON p.id = f.profile_id
)
SELECT old_mc_uuid, mc_uuid
FROM old_uuids
WHERE old_mc_uuid <> mc_uuid AND mc_uuid IN (
	SELECT mc_uuid FROM old_uuids WHERE old_mc_uuid IN (%[1]s) OR mc_uuid IN (%[1]s)
)
`

func (q *queries) FindLegacyMinecraftUUIDs(ctx context.Context, mcUUIDs uuid.UUIDs) (map[uuid.UUID]uuid.UUID, error) {
	res := make(map[uuid.UUID]uuid.UUID)
	if len(mcUUIDs) == 0 {
		return res, nil
	}

	placeholders := strings.TrimSuffix(strings.Repeat("?,", len(mcUUIDs)), ",")
	args := make([]any, 0, 2*len(mcUUIDs))
	for range 2 {
		for _, mcUUID := range mcUUIDs {
			args = append(args, mcUUID)
		}
	}
	rows, err := q.x.QueryContext(ctx, fmt.Sprintf(findLegacyMinecraftUUIDs, placeholders), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var legacyUUID, mcUUID uuid.UUID
		if err := rows.Scan(&legacyUUID, &mcUUID); err != nil {
			return nil, err
		}
		res[legacyUUID] = mcUUID
	}
	return res, rows.Err()
}
