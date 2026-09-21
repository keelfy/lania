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

const insertProfileNameColorOption = `
INSERT INTO profile_name_color_options (
	profile_id,
	name_color_id,
	for_season_id,
	order_item_id,
	created_by
) VALUES (?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE revoked_at = NULL, revoked_by = NULL
`

type InsertProfileNameColorOptionParams struct {
	ProfileID   uuid.UUID
	NameColorID uuid.UUID
	ForSeasonID *uuid.UUID
	OrderItemID *uuid.UUID
	CreatedBy   *uuid.UUID
}

func (q *queries) InsertProfileNameColorOption(ctx context.Context, arg InsertProfileNameColorOptionParams) error {
	_, err := q.x.ExecContext(ctx, insertProfileNameColorOption,
		arg.ProfileID,
		arg.NameColorID,
		arg.ForSeasonID,
		arg.OrderItemID,
		arg.CreatedBy,
	)
	return err
}

const insertProfileNamePrefixOption = `
INSERT INTO profile_name_prefix_options (
	profile_id,
	name_prefix_id,
	type,
	for_season_id,
	order_item_id,
	created_by
) VALUES (?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE type = IF(revoked_at IS NULL, type, VALUES(type)), revoked_at = NULL, revoked_by = NULL
`

type InsertProfileNamePrefixOptionParams struct {
	ProfileID    uuid.UUID
	NamePrefixID uuid.UUID
	Type         domain.ProfilePrefixType
	ForSeasonID  *uuid.UUID
	OrderItemID  *uuid.UUID
	CreatedBy    *uuid.UUID
}

func (q *queries) InsertProfileNamePrefixOption(ctx context.Context, arg InsertProfileNamePrefixOptionParams) error {
	_, err := q.x.ExecContext(ctx, insertProfileNamePrefixOption,
		arg.ProfileID,
		arg.NamePrefixID,
		arg.Type,
		arg.ForSeasonID,
		arg.OrderItemID,
		arg.CreatedBy,
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
WHERE pnc.profile_id = ? AND (pnc.for_season_id = ? OR pnc.for_season_id IS NULL) AND pnc.revoked_at IS NULL
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
WHERE pnc.profile_id = ? AND pnc.type = ? AND (pnc.for_season_id = ? OR pnc.for_season_id IS NULL) AND pnc.revoked_at IS NULL
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
WHERE pnc.id = ? AND pnc.profile_id = ? AND (pnc.for_season_id = ? OR pnc.for_season_id IS NULL) AND pnc.revoked_at IS NULL
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
WHERE pnp.id = ? AND pnp.profile_id = ? AND pnp.type = ? AND (pnp.for_season_id = ? OR pnp.for_season_id IS NULL) AND pnp.revoked_at IS NULL
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
WHERE pnc.profile_id IN (SELECT id FROM profiles WHERE owner_user_id = ?) AND (pnc.for_season_id = ? OR pnc.for_season_id IS NULL) AND pnc.revoked_at IS NULL
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
WHERE pnc.profile_id IN (SELECT id FROM profiles WHERE owner_user_id = ?) AND pnc.type = ? AND (pnc.for_season_id = ? OR pnc.for_season_id IS NULL) AND pnc.revoked_at IS NULL
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

// The default name color stands in for a NULL selection, so every profile has a color.
// A profile without a row for the season gets the default color and no prefixes.
const findProfilesSeasonCosmetics = `
SELECT
	p.id,
	COALESCE(nc.id, dnc.id),
	COALESCE(nc.name, dnc.name),
	COALESCE(nc.colors, dnc.colors),
	gp.id,
	gp.name,
	gp.metadata,
	sp.id,
	sp.name,
	sp.metadata
FROM profiles p
LEFT JOIN profile_season_cosmetics psc ON psc.profile_id = p.id AND psc.season_id = ?
LEFT JOIN name_colors nc ON nc.id = psc.name_color_id
LEFT JOIN name_colors dnc ON dnc.id = ?
LEFT JOIN name_prefixes gp ON gp.id = psc.glyth_prefix_id
LEFT JOIN name_prefixes sp ON sp.id = psc.special_prefix_id
WHERE p.id IN (%s)
`

// FindProfilesSeasonCosmetics returns what every profile shows in the season, keyed by profile ID.
func (q *queries) FindProfilesSeasonCosmetics(ctx context.Context, profileIDs uuid.UUIDs, seasonID, defaultNameColorID uuid.UUID) (map[uuid.UUID]*domain.ProfileCosmetics, error) {
	cosmetics := make(map[uuid.UUID]*domain.ProfileCosmetics, len(profileIDs))
	if len(profileIDs) == 0 {
		return cosmetics, nil
	}

	placeholders := make([]string, len(profileIDs))
	args := make([]any, 0, len(profileIDs)+2)
	args = append(args, seasonID, defaultNameColorID)
	for i, profileID := range profileIDs {
		placeholders[i] = "?"
		args = append(args, profileID)
	}
	rows, err := q.x.QueryContext(ctx, fmt.Sprintf(findProfilesSeasonCosmetics, strings.Join(placeholders, ", ")), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var profileID uuid.UUID
		var colorID, glythID, specialID uuid.NullUUID
		var colorName, glythName, specialName stdsql.NullString
		var colors, glythMetadata, specialMetadata json.RawMessage
		err := rows.Scan(
			&profileID,
			&colorID, &colorName, &colors,
			&glythID, &glythName, &glythMetadata,
			&specialID, &specialName, &specialMetadata,
		)
		if err != nil {
			return nil, err
		}

		selected := &domain.ProfileCosmetics{}
		if colorID.Valid {
			selected.NameColor = &domain.NameColor{ID: colorID.UUID, Name: colorName.String}
			if err := json.Unmarshal(colors, &selected.NameColor.Metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal name color metadata for name color %s: %w", colorID.UUID, err)
			}
		}
		if selected.Glyth, err = scanSelectedNamePrefix(glythID, glythName, glythMetadata); err != nil {
			return nil, err
		}
		if selected.Special, err = scanSelectedNamePrefix(specialID, specialName, specialMetadata); err != nil {
			return nil, err
		}
		cosmetics[profileID] = selected
	}
	return cosmetics, rows.Err()
}

// scanSelectedNamePrefix builds the name prefix of a joined row, nil when the profile selected none.
func scanSelectedNamePrefix(id uuid.NullUUID, name stdsql.NullString, rawMetadata json.RawMessage) (*domain.NamePrefix, error) {
	if !id.Valid {
		return nil, nil
	}
	prefix := &domain.NamePrefix{ID: id.UUID, Name: name.String}
	if err := json.Unmarshal(rawMetadata, &prefix.Metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal name prefix metadata for name prefix %s: %w", id.UUID, err)
	}
	return prefix, nil
}

const setProfileSeasonNameColor = `
INSERT INTO profile_season_cosmetics (profile_id, season_id, name_color_id)
VALUES (?, ?, ?)
ON DUPLICATE KEY UPDATE name_color_id = VALUES(name_color_id), updated_at = NOW()
`

// SetProfileSeasonNameColor selects the name color of the profile in the season.
func (q *queries) SetProfileSeasonNameColor(ctx context.Context, profileID, seasonID, nameColorID uuid.UUID) error {
	_, err := q.x.ExecContext(ctx, setProfileSeasonNameColor, profileID, seasonID, nameColorID)
	return err
}

// The column comes from this map, never from the caller.
var profileSeasonPrefixColumns = map[domain.ProfilePrefixType]string{
	domain.ProfilePrefixTypeGlyth:   "glyth_prefix_id",
	domain.ProfilePrefixTypeSpecial: "special_prefix_id",
}

// SetProfileSeasonPrefix selects the name prefix of the type for the profile in the season.
// A nil namePrefixID clears it.
func (q *queries) SetProfileSeasonPrefix(ctx context.Context, profileID, seasonID uuid.UUID, prefixType domain.ProfilePrefixType, namePrefixID *uuid.UUID) error {
	column, ok := profileSeasonPrefixColumns[prefixType]
	if !ok {
		return fmt.Errorf("unknown name prefix type %q", prefixType)
	}
	_, err := q.x.ExecContext(ctx, fmt.Sprintf(`
INSERT INTO profile_season_cosmetics (profile_id, season_id, %[1]s)
VALUES (?, ?, ?)
ON DUPLICATE KEY UPDATE %[1]s = VALUES(%[1]s), updated_at = NOW()
`, column), profileID, seasonID, namePrefixID)
	return err
}

// A selection stays valid while the profile keeps an unrevoked option for the item that is for the season or permanent.
const pruneProfileSeasonCosmetics = `
UPDATE profile_season_cosmetics psc
SET
	psc.name_color_id = IF(psc.name_color_id IS NOT NULL AND NOT EXISTS (
		SELECT 1 FROM profile_name_color_options o
		WHERE o.profile_id = psc.profile_id AND o.name_color_id = psc.name_color_id
			AND o.revoked_at IS NULL AND (o.for_season_id = psc.season_id OR o.for_season_id IS NULL)
	), NULL, psc.name_color_id),
	psc.glyth_prefix_id = IF(psc.glyth_prefix_id IS NOT NULL AND NOT EXISTS (
		SELECT 1 FROM profile_name_prefix_options o
		WHERE o.profile_id = psc.profile_id AND o.name_prefix_id = psc.glyth_prefix_id AND o.type = 'glyth'
			AND o.revoked_at IS NULL AND (o.for_season_id = psc.season_id OR o.for_season_id IS NULL)
	), NULL, psc.glyth_prefix_id),
	psc.special_prefix_id = IF(psc.special_prefix_id IS NOT NULL AND NOT EXISTS (
		SELECT 1 FROM profile_name_prefix_options o
		WHERE o.profile_id = psc.profile_id AND o.name_prefix_id = psc.special_prefix_id AND o.type = 'special'
			AND o.revoked_at IS NULL AND (o.for_season_id = psc.season_id OR o.for_season_id IS NULL)
	), NULL, psc.special_prefix_id)
WHERE psc.profile_id = ?
`

// PruneProfileSeasonCosmetics resets every selection of the profile that no unrevoked option covers any more,
// in every season. It leaves the other selections alone.
func (q *queries) PruneProfileSeasonCosmetics(ctx context.Context, profileID uuid.UUID) error {
	_, err := q.x.ExecContext(ctx, pruneProfileSeasonCosmetics, profileID)
	return err
}
