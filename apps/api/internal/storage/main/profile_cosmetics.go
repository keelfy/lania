package sql

import (
	"context"
	stdsql "database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/domain"
)

const insertProfileNameColorOption = `
INSERT INTO profile_name_color_options (
	profile_id,
	name_color_id,
	for_season_id,
	order_item_id
) VALUES (?, ?, ?, ?)
`

type InsertProfileNameColorOptionParams struct {
	ProfileID   uuid.UUID
	NameColorID uuid.UUID
	ForSeasonID *uuid.UUID
	OrderItemID *uuid.UUID
}

func (q *queries) InsertProfileNameColorOption(ctx context.Context, arg InsertProfileNameColorOptionParams) error {
	_, err := q.x.ExecContext(ctx, insertProfileNameColorOption,
		arg.ProfileID,
		arg.NameColorID,
		arg.ForSeasonID,
		arg.OrderItemID,
	)
	return err
}

const insertProfileNamePrefixOption = `
INSERT INTO profile_name_prefix_options (
	profile_id,
	name_prefix_id,
	type,
	for_season_id,
	order_item_id
) VALUES (?, ?, ?, ?, ?)
`

type InsertProfileNamePrefixOptionParams struct {
	ProfileID    uuid.UUID
	NamePrefixID uuid.UUID
	Type         domain.ProfilePrefixType
	ForSeasonID  *uuid.UUID
	OrderItemID  *uuid.UUID
}

func (q *queries) InsertProfileNamePrefixOption(ctx context.Context, arg InsertProfileNamePrefixOptionParams) error {
	_, err := q.x.ExecContext(ctx, insertProfileNamePrefixOption,
		arg.ProfileID,
		arg.NamePrefixID,
		arg.Type,
		arg.ForSeasonID,
		arg.OrderItemID,
	)
	return err
}

const findProfileNameColorOptionsByProfileID = `
SELECT 
	pnc.id,
	pnc.profile_id,
	pnc.name_color_id,
	pnc.for_season_id,
	nc.colors AS name_colors,
	nc.name AS name_color_name
FROM profile_name_color_options pnc
LEFT JOIN name_colors nc ON pnc.name_color_id = nc.id
WHERE pnc.profile_id = ? AND (pnc.for_season_id = ? OR pnc.for_season_id IS NULL)
ORDER BY pnc.created_at DESC
`

func (q *queries) FindProfileNameColorOptionsByProfileID(ctx context.Context, profileID uuid.UUID, seasonID *uuid.UUID) ([]*domain.ProfileNameColorOption, error) {
	rows, err := q.x.QueryContext(ctx, findProfileNameColorOptionsByProfileID, profileID, seasonID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	options := make([]*domain.ProfileNameColorOption, 0)
	for rows.Next() {
		var option domain.ProfileNameColorOption
		var nameColor domain.NameColor
		var colors json.RawMessage
		err := rows.Scan(
			&option.ID,
			&option.ProfileID,
			&option.NameColorID,
			&option.ForSeasonID,
			&colors,
			&nameColor.Name,
		)
		if err != nil {
			return nil, err
		}
		metadata := domain.NameColorMetadata{}
		err = json.Unmarshal(colors, &metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal name color metadata for profile name color %s: %w", option.ID, err)
		}
		nameColor.Metadata = metadata
		option.NameColor = &nameColor
		options = append(options, &option)
	}
	return options, nil
}

const findProfileNamePrefixOptionsByProfileIDAndType = `
SELECT 
	pnc.id,
	pnc.profile_id,
	pnc.name_prefix_id,
	pnc.type,
	pnc.for_season_id,
	pnc.order_item_id,
	pnc.created_at,
	np.name AS name_prefix_name,
	np.metadata AS name_prefix_metadata
FROM profile_name_prefix_options pnc
LEFT JOIN name_prefixes np ON pnc.name_prefix_id = np.id
WHERE pnc.profile_id = ? AND pnc.type = ? AND (pnc.for_season_id = ? OR pnc.for_season_id IS NULL)
ORDER BY pnc.created_at DESC
`

func (q *queries) FindProfileNamePrefixOptionsByProfileIDAndType(ctx context.Context, profileID uuid.UUID, prefixType domain.ProfilePrefixType, seasonID *uuid.UUID) ([]*domain.ProfileNamePrefixOption, error) {
	rows, err := q.x.QueryContext(ctx, findProfileNamePrefixOptionsByProfileIDAndType, profileID, prefixType, seasonID)
	if err == stdsql.ErrNoRows {
		return []*domain.ProfileNamePrefixOption{}, nil
	} else if err != nil {
		return nil, err
	}
	defer rows.Close()

	options := make([]*domain.ProfileNamePrefixOption, 0)
	for rows.Next() {
		var option domain.ProfileNamePrefixOption
		var namePrefix domain.NamePrefix
		var rawMetadata json.RawMessage
		err := rows.Scan(
			&option.ID,
			&option.ProfileID,
			&option.NamePrefixID,
			&option.Type,
			&option.ForSeasonID,
			&option.OrderItemID,
			&option.CreatedAt,
			&namePrefix.Name,
			&rawMetadata,
		)
		if err != nil {
			return nil, err
		}
		metadata := domain.NamePrefixMetadata{}
		err = json.Unmarshal(rawMetadata, &metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal name prefix metadata for profile name prefix %s: %w", option.ID, err)
		}
		namePrefix.ID = option.NamePrefixID
		namePrefix.Metadata = metadata
		option.NamePrefix = &namePrefix
		options = append(options, &option)
	}
	return options, nil
}

const findProfileNameColorOptionByIDAndProfileID = `
SELECT 
	pnc.id,
	pnc.profile_id,
	pnc.name_color_id,
	pnc.for_season_id,
	nc.colors AS name_colors
FROM profile_name_color_options pnc
LEFT JOIN name_colors nc ON pnc.name_color_id = nc.id
WHERE pnc.id = ? AND pnc.profile_id = ? AND (pnc.for_season_id = ? OR pnc.for_season_id IS NULL)
`

func (q *queries) FindProfileNameColorOptionByIDAndProfileID(ctx context.Context, optionID uuid.UUID, profileID uuid.UUID, seasonID *uuid.UUID) (*domain.ProfileNameColorOption, error) {
	row := q.x.QueryRowContext(ctx, findProfileNameColorOptionByIDAndProfileID, optionID, profileID, seasonID)
	var option domain.ProfileNameColorOption
	var nameColor domain.NameColor
	var colors json.RawMessage
	err := row.Scan(
		&option.ID,
		&option.ProfileID,
		&option.NameColorID,
		&option.ForSeasonID,
		&colors,
	)
	if err != nil {
		return nil, err
	}
	metadata := domain.NameColorMetadata{}
	err = json.Unmarshal(colors, &metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal name color metadata for profile name color %s: %w", option.ID, err)
	}
	option.NameColor = &nameColor
	nameColor.Metadata = metadata
	return &option, nil
}

const findProfileNamePrefixOptionByIDAndProfileIDAndType = `
SELECT 
	pnp.id,
	pnp.profile_id,
	pnp.name_prefix_id,
	pnp.type,
	pnp.for_season_id,
	pnp.order_item_id,
	pnp.created_at,
	np.name AS name_prefix_name,
	np.metadata AS name_prefix_metadata
FROM profile_name_prefix_options pnp
LEFT JOIN name_prefixes np ON pnp.name_prefix_id = np.id
WHERE pnp.id = ? AND pnp.profile_id = ? AND pnp.type = ? AND (pnp.for_season_id = ? OR pnp.for_season_id IS NULL)
`

func (q *queries) FindProfileNamePrefixOptionByIDAndProfileIDAndType(ctx context.Context, optionID uuid.UUID, profileID uuid.UUID, prefixType domain.ProfilePrefixType, seasonID *uuid.UUID) (*domain.ProfileNamePrefixOption, error) {
	row := q.x.QueryRowContext(ctx, findProfileNamePrefixOptionByIDAndProfileIDAndType, optionID, profileID, prefixType, seasonID)
	var option domain.ProfileNamePrefixOption
	var namePrefix domain.NamePrefix
	var rawMetadata json.RawMessage
	err := row.Scan(
		&option.ID,
		&option.ProfileID,
		&option.NamePrefixID,
		&option.Type,
		&option.ForSeasonID,
		&option.OrderItemID,
		&option.CreatedAt,
		&namePrefix.Name,
		&rawMetadata,
	)
	if err != nil {
		return nil, err
	}
	metadata := domain.NamePrefixMetadata{}
	err = json.Unmarshal(rawMetadata, &metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal name prefix metadata for profile name prefix %s: %w", option.ID, err)
	}
	namePrefix.Metadata = metadata
	namePrefix.ID = option.NamePrefixID
	option.NamePrefix = &namePrefix
	return &option, nil
}

const findProfileNameColorOptionsByProfileOwnerUserID = `
SELECT 
	pnc.id,
	pnc.profile_id,
	pnc.name_color_id,
	pnc.for_season_id,
	nc.colors AS name_colors,
	nc.name AS name_color_name
FROM profile_name_color_options pnc
LEFT JOIN name_colors nc ON pnc.name_color_id = nc.id
WHERE pnc.profile_id IN (SELECT id FROM profiles WHERE owner_user_id = ?) AND (pnc.for_season_id = ? OR pnc.for_season_id IS NULL)
ORDER BY pnc.created_at DESC
`

func (q *queries) FindProfileNameColorOptionsByProfileOwnerUserID(ctx context.Context, ownerUserID uuid.UUID, seasonID *uuid.UUID) ([]*domain.ProfileNameColorOption, error) {
	rows, err := q.x.QueryContext(ctx, findProfileNameColorOptionsByProfileOwnerUserID, ownerUserID, seasonID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	options := make([]*domain.ProfileNameColorOption, 0)
	for rows.Next() {
		var option domain.ProfileNameColorOption
		var nameColor domain.NameColor
		var colors json.RawMessage
		err := rows.Scan(
			&option.ID,
			&option.ProfileID,
			&option.NameColorID,
			&option.ForSeasonID,
			&colors,
			&nameColor.Name,
		)
		if err != nil {
			return nil, err
		}
		metadata := domain.NameColorMetadata{}
		err = json.Unmarshal(colors, &metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal name color metadata for profile name color %s: %w", option.ID, err)
		}
		nameColor.Metadata = metadata
		option.NameColor = &nameColor
		options = append(options, &option)
	}
	return options, nil
}

const findProfileNamePrefixOptionsByProfileOwnerUserIDAndType = `
SELECT 
	pnc.id,
	pnc.profile_id,
	pnc.name_prefix_id,
	pnc.type,
	pnc.for_season_id,
	pnc.order_item_id,
	pnc.created_at,
	np.name AS name_prefix_name,
	np.metadata AS name_prefix_metadata
FROM profile_name_prefix_options pnc
LEFT JOIN name_prefixes np ON pnc.name_prefix_id = np.id
WHERE pnc.profile_id IN (SELECT id FROM profiles WHERE owner_user_id = ?) AND pnc.type = ? AND (pnc.for_season_id = ? OR pnc.for_season_id IS NULL)
ORDER BY pnc.created_at DESC
`

func (q *queries) FindProfileNamePrefixOptionsByProfileOwnerUserIDAndType(ctx context.Context, ownerUserID uuid.UUID, prefixType domain.ProfilePrefixType, seasonID *uuid.UUID) ([]*domain.ProfileNamePrefixOption, error) {
	rows, err := q.x.QueryContext(ctx, findProfileNamePrefixOptionsByProfileOwnerUserIDAndType, ownerUserID, prefixType, seasonID)
	if err == stdsql.ErrNoRows {
		return []*domain.ProfileNamePrefixOption{}, nil
	} else if err != nil {
		return nil, err
	}
	defer rows.Close()

	options := make([]*domain.ProfileNamePrefixOption, 0)
	for rows.Next() {
		var option domain.ProfileNamePrefixOption
		var namePrefix domain.NamePrefix
		var rawMetadata json.RawMessage
		err := rows.Scan(
			&option.ID,
			&option.ProfileID,
			&option.NamePrefixID,
			&option.Type,
			&option.ForSeasonID,
			&option.OrderItemID,
			&option.CreatedAt,
			&namePrefix.Name,
			&rawMetadata,
		)
		if err != nil {
			return nil, err
		}
		metadata := domain.NamePrefixMetadata{}
		err = json.Unmarshal(rawMetadata, &metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal name prefix metadata for profile name prefix %s: %w", option.ID, err)
		}
		namePrefix.Metadata = metadata
		option.NamePrefix = &namePrefix
		options = append(options, &option)
	}
	return options, nil
}

const updateProfileNameColorByID = `
UPDATE profiles SET name_color_id = ? WHERE id = ?
`

func (q *queries) UpdateProfileNameColorByID(ctx context.Context, profileID uuid.UUID, nameColorID uuid.UUID) error {
	_, err := q.x.ExecContext(ctx, updateProfileNameColorByID, nameColorID, profileID)
	return err
}

const findProfilePrefixesByProfileID = `
SELECT 
	pp.profile_id,
	pp.name_prefix_id,
	pp.type,
	np.name AS name_prefix_name,
	np.metadata AS name_prefix_metadata
FROM profile_prefixes pp
LEFT JOIN name_prefixes np ON pp.name_prefix_id = np.id
WHERE pp.profile_id = ?
`

func (q *queries) FindProfilePrefixesByProfileID(ctx context.Context, profileID uuid.UUID) ([]*domain.ProfilePrefix, error) {
	rows, err := q.x.QueryContext(ctx, findProfilePrefixesByProfileID, profileID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	prefixes := make([]*domain.ProfilePrefix, 0)
	for rows.Next() {
		var prefix domain.ProfilePrefix
		var namePrefix domain.NamePrefix
		var rawMetadata json.RawMessage
		err := rows.Scan(
			&prefix.ProfileID,
			&prefix.NamePrefixID,
			&prefix.Type,
			&namePrefix.Name,
			&rawMetadata,
		)
		if err != nil {
			return nil, err
		}
		metadata := domain.NamePrefixMetadata{}
		err = json.Unmarshal(rawMetadata, &metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal name prefix metadata for profile prefix %s: %w", prefix.NamePrefixID, err)
		}
		namePrefix.Metadata = metadata
		namePrefix.ID = prefix.NamePrefixID
		prefix.NamePrefix = &namePrefix
		prefixes = append(prefixes, &prefix)
	}
	return prefixes, nil
}

const insertProfilePrefix = `
INSERT INTO profile_prefixes (
	profile_id,
	name_prefix_id,
	type
) VALUES (?, ?, ?)
`

type InsertProfilePrefixParams struct {
	ProfileID    uuid.UUID
	NamePrefixID uuid.UUID
	Type         domain.ProfilePrefixType
}

func (q *queries) InsertProfilePrefix(ctx context.Context, arg InsertProfilePrefixParams) error {
	_, err := q.x.ExecContext(ctx, insertProfilePrefix, arg.ProfileID, arg.NamePrefixID, arg.Type)
	return err
}

const updateProfileNamePrefixByProfileIDAndType = `
UPDATE profile_prefixes SET name_prefix_id = ? WHERE profile_id = ? AND type = ?
`

func (q *queries) UpdateProfileNamePrefixByProfileIDAndType(ctx context.Context, profileID uuid.UUID, namePrefixID uuid.UUID, prefixType domain.ProfilePrefixType) error {
	_, err := q.x.ExecContext(ctx, updateProfileNamePrefixByProfileIDAndType, namePrefixID, profileID, prefixType)
	return err
}

const deleteProfilePrefixByProfileIDAndType = `
DELETE FROM profile_prefixes WHERE profile_id = ? AND type = ?
`

func (q *queries) DeleteProfilePrefixByProfileIDAndType(ctx context.Context, profileID uuid.UUID, prefixType domain.ProfilePrefixType) error {
	_, err := q.x.ExecContext(ctx, deleteProfilePrefixByProfileIDAndType, profileID, prefixType)
	if err == stdsql.ErrNoRows {
		return nil
	}
	return err
}
