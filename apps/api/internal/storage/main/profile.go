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
SELECT COUNT(id) FROM profiles
`

func (q *queries) CountPublicProfiles(ctx context.Context) (int64, error) {
	row := q.x.QueryRowContext(ctx, countPublicProfiles)
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
	p.role,
	p.is_slim,
	p.created_at,
	p.updated_at,
	p.updated_by,
	p.name_color_id,
	nc.colors AS name_colors
FROM profiles p
LEFT JOIN name_colors nc ON p.name_color_id = nc.id
ORDER BY %s %s
LIMIT ? OFFSET ?
`

func (q *queries) getProfileSortColumn(sortCol string) string {
	switch sortCol {
	case "username":
		return "p.mc_username"
	default:
		return "p.created_at"
	}
}

func (q *queries) getProfileSortDirection(direction string) string {
	switch strings.ToLower(direction) {
	case "desc":
		return "DESC"
	default:
		return "ASC"
	}
}

func (q *queries) FindPublicProfiles(ctx context.Context, sortCol, direction string, size, from int) ([]*domain.Profile, error) {
	sortColumn := q.getProfileSortColumn(sortCol)
	sortDirection := q.getProfileSortDirection(direction)
	query := fmt.Sprintf(findPublicProfiles, sortColumn, sortDirection)
	rows, err := q.x.QueryContext(ctx, query, size, from)
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
	owner_user_id = VALUES(owner_user_id),
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
	OwnerUserID       uuid.UUID
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

const findProfileByID = `
SELECT 
	p.id,
	p.mc_uuid,
	p.mc_username,
	p.owner_user_id,
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
