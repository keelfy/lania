package sql

import (
	"context"
	stdsql "database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/domain"
)

func scanProfileRow(row *stdsql.Row) (*domain.Profile, error) {
	var profile domain.Profile
	var nameColor domain.NameColor
	var colors json.RawMessage
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
		&profile.NameColorID,
		&colors,
	)
	if err != nil {
		return nil, err
	}

	profile.NameColor = &nameColor
	metadata := domain.NameColorMetadata{}
	err = json.Unmarshal(colors, &metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal name color metadata for profile %s: %w", profile.ID, err)
	}
	nameColor.ID = profile.NameColorID
	nameColor.Metadata = metadata
	return &profile, err
}

func scanProfileRows(rows *stdsql.Rows) (*domain.Profile, error) {
	var profile domain.Profile
	var nameColor domain.NameColor
	var colors json.RawMessage
	err := rows.Scan(
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
		&profile.NameColorID,
		&colors,
	)
	if err != nil {
		return nil, err
	}

	profile.NameColor = &nameColor
	metadata := domain.NameColorMetadata{}
	err = json.Unmarshal(colors, &metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal name color metadata for profile %s: %w", profile.ID, err)
	}
	nameColor.ID = profile.NameColorID
	nameColor.Metadata = metadata
	return &profile, err
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
	p.name_color_id,
	nc.colors AS name_colors
FROM profiles p
LEFT JOIN name_colors nc ON p.name_color_id = nc.id
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

func (q *queries) CountPublicProfiles(ctx context.Context, search string) (int64, error) {
	where, args := profileSearchClause(search)
	row := q.x.QueryRowContext(ctx, fmt.Sprintf(countPublicProfiles, where), args...)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const findPublicProfiles = `
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
	p.name_color_id,
	nc.colors AS name_colors
FROM profiles p
LEFT JOIN name_colors nc ON p.name_color_id = nc.id
%s
%s
ORDER BY %s
LIMIT ? OFFSET ?
`

// Playtime is summed over all seasons. The join is only added when the list is sorted by playtime.
const profilePlaytimeJoin = `
LEFT JOIN (
	SELECT mc_uuid, SUM(playtime) AS total_playtime
	FROM profile_playtimes
	GROUP BY mc_uuid
) pt ON pt.mc_uuid = p.mc_uuid
`

var profileSearchEscaper = strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)

// profileSearchClause matches usernames by prefix, so the mc_username index is used.
func profileSearchClause(search string) (string, []any) {
	if search == "" {
		return "", nil
	}
	return "WHERE p.mc_username LIKE ?", []any{profileSearchEscaper.Replace(search) + "%"}
}

func getProfileSortDirection(direction string) string {
	switch strings.ToLower(direction) {
	case "desc":
		return "DESC"
	default:
		return "ASC"
	}
}

// profileOrderBy returns the join needed by the sort column and the ORDER BY expression.
// Profiles without a value are always listed last, and id keeps the order stable between pages.
func profileOrderBy(sortCol, direction string) (join string, orderBy string) {
	dir := getProfileSortDirection(direction)
	switch sortCol {
	case "username":
		return "", fmt.Sprintf("p.mc_username %s, p.id", dir)
	case "first_seen_at":
		return "", fmt.Sprintf("p.first_seen_at IS NULL, p.first_seen_at %s, p.id", dir)
	case "last_seen_at":
		return "", fmt.Sprintf("p.last_seen_at IS NULL, p.last_seen_at %s, p.id", dir)
	case "playtime":
		return profilePlaytimeJoin, fmt.Sprintf("COALESCE(pt.total_playtime, 0) %s, p.id", dir)
	default:
		return "", fmt.Sprintf("p.created_at %s, p.id", dir)
	}
}

func (q *queries) FindPublicProfiles(ctx context.Context, search, sortCol, direction string, size, from int) ([]*domain.Profile, error) {
	join, orderBy := profileOrderBy(sortCol, direction)
	where, args := profileSearchClause(search)
	query := fmt.Sprintf(findPublicProfiles, join, where, orderBy)
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

const insertProfile = `
INSERT INTO profiles (
	id,
	mc_uuid,
	mc_username,
	owner_user_id,
	role,
	is_slim,
	name_color_id,
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
	?,
	now(),
	now(),
	?
) ON DUPLICATE KEY UPDATE
	mc_uuid = VALUES(mc_uuid),
  mc_username = VALUES(mc_username),
	owner_user_id = COALESCE(owner_user_id, VALUES(owner_user_id)),
	role = VALUES(role),
	is_slim = VALUES(is_slim),
	name_color_id = VALUES(name_color_id),
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
	NameColorID       uuid.UUID
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
		arg.NameColorID,
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
	p.name_color_id,
	nc.colors AS name_colors
FROM profiles p
LEFT JOIN name_colors nc ON p.name_color_id = nc.id
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
	p.name_color_id,
	nc.colors AS name_colors
FROM profiles p
LEFT JOIN name_colors nc ON p.name_color_id = nc.id
WHERE p.mc_uuid = ?
`

func (q *queries) FindProfileByMinecraftUUID(ctx context.Context, minecraftUUID uuid.UUID) (*domain.Profile, error) {
	row := q.x.QueryRowContext(ctx, findProfileByMinecraftUUID, minecraftUUID)
	profile, err := scanProfileRow(row)
	return profile, err
}
