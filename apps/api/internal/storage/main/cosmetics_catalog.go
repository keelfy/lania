package sql

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/domain"
)

func (q *queries) InsertNameColor(ctx context.Context, id uuid.UUID, name string, colors []string) error {
	payload, err := json.Marshal(domain.NameColorMetadata{Colors: colors})
	if err != nil {
		return err
	}
	_, err = q.x.ExecContext(ctx, "INSERT INTO name_colors (id, name, colors) VALUES (?, ?, ?)", id, name, payload)
	return err
}

func (q *queries) UpdateNameColor(ctx context.Context, id uuid.UUID, name string, colors []string) error {
	payload, err := json.Marshal(domain.NameColorMetadata{Colors: colors})
	if err != nil {
		return err
	}
	_, err = q.x.ExecContext(ctx, "UPDATE name_colors SET name = ?, colors = ? WHERE id = ?", name, payload, id)
	return err
}

func (q *queries) InsertNamePrefix(ctx context.Context, id uuid.UUID, name string, metadata domain.NamePrefixMetadata) error {
	payload, err := json.Marshal(metadata)
	if err != nil {
		return err
	}
	_, err = q.x.ExecContext(ctx, "INSERT INTO name_prefixes (id, name, metadata) VALUES (?, ?, ?)", id, name, payload)
	return err
}

func (q *queries) UpdateNamePrefix(ctx context.Context, id uuid.UUID, name string, metadata domain.NamePrefixMetadata) error {
	payload, err := json.Marshal(metadata)
	if err != nil {
		return err
	}
	_, err = q.x.ExecContext(ctx, "UPDATE name_prefixes SET name = ?, metadata = ? WHERE id = ?", name, payload, id)
	return err
}

const findNameColors = `
SELECT id, name, colors
FROM name_colors
ORDER BY name
`

func (q *queries) FindNameColors(ctx context.Context) ([]*domain.NameColor, error) {
	rows, err := q.x.QueryContext(ctx, findNameColors)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	nameColors := make([]*domain.NameColor, 0)
	for rows.Next() {
		var nameColor domain.NameColor
		var colors json.RawMessage
		if err := rows.Scan(&nameColor.ID, &nameColor.Name, &colors); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(colors, &nameColor.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal name color metadata for name color %s: %w", nameColor.ID, err)
		}
		nameColors = append(nameColors, &nameColor)
	}
	return nameColors, rows.Err()
}

const findNamePrefixes = `
SELECT id, name, metadata
FROM name_prefixes
ORDER BY name
`

func (q *queries) FindNamePrefixes(ctx context.Context) ([]*domain.NamePrefix, error) {
	rows, err := q.x.QueryContext(ctx, findNamePrefixes)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	namePrefixes := make([]*domain.NamePrefix, 0)
	for rows.Next() {
		var namePrefix domain.NamePrefix
		var metadata json.RawMessage
		if err := rows.Scan(&namePrefix.ID, &namePrefix.Name, &metadata); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(metadata, &namePrefix.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal name prefix metadata for name prefix %s: %w", namePrefix.ID, err)
		}
		namePrefixes = append(namePrefixes, &namePrefix)
	}
	return namePrefixes, rows.Err()
}

const findNameColorByID = `
SELECT id, name, colors
FROM name_colors
WHERE id = ?
`

func (q *queries) FindNameColorByID(ctx context.Context, nameColorID uuid.UUID) (*domain.NameColor, error) {
	var nameColor domain.NameColor
	var colors json.RawMessage
	err := q.x.QueryRowContext(ctx, findNameColorByID, nameColorID).Scan(&nameColor.ID, &nameColor.Name, &colors)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(colors, &nameColor.Metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal name color metadata for name color %s: %w", nameColor.ID, err)
	}
	return &nameColor, nil
}

const findNamePrefixByID = `
SELECT id, name, metadata
FROM name_prefixes
WHERE id = ?
`

func (q *queries) FindNamePrefixByID(ctx context.Context, namePrefixID uuid.UUID) (*domain.NamePrefix, error) {
	var namePrefix domain.NamePrefix
	var metadata json.RawMessage
	err := q.x.QueryRowContext(ctx, findNamePrefixByID, namePrefixID).Scan(&namePrefix.ID, &namePrefix.Name, &metadata)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(metadata, &namePrefix.Metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal name prefix metadata for name prefix %s: %w", namePrefix.ID, err)
	}
	return &namePrefix, nil
}

// CountNameColorOwners returns how many profiles have the name color and it is not revoked.
func (q *queries) CountNameColorOwners(ctx context.Context, id uuid.UUID) (int, error) {
	var count int
	err := q.x.QueryRowContext(ctx, "SELECT COUNT(DISTINCT profile_id) FROM profile_name_color_options WHERE name_color_id = ? AND revoked_at IS NULL", id).Scan(&count)
	return count, err
}

// CountNamePrefixOwners returns how many profiles have the name prefix and it is not revoked.
func (q *queries) CountNamePrefixOwners(ctx context.Context, id uuid.UUID) (int, error) {
	var count int
	err := q.x.QueryRowContext(ctx, "SELECT COUNT(DISTINCT profile_id) FROM profile_name_prefix_options WHERE name_prefix_id = ? AND revoked_at IS NULL", id).Scan(&count)
	return count, err
}

func (q *queries) execAll(ctx context.Context, id uuid.UUID, statements ...string) error {
	for _, statement := range statements {
		if _, err := q.x.ExecContext(ctx, statement, id); err != nil {
			return err
		}
	}
	return nil
}

// DeleteNameColor removes the name color with its revoked grants and the selections left from them.
// A grant that is not revoked keeps the foreign key, so the delete fails with 1451.
// Profiles that still show it in the legacy column fall back to defaultNameColorID.
func (q *queries) DeleteNameColor(ctx context.Context, id, defaultNameColorID uuid.UUID) error {
	if _, err := q.x.ExecContext(ctx, "UPDATE profiles SET name_color_id = ? WHERE name_color_id = ?", defaultNameColorID, id); err != nil {
		return err
	}
	return q.execAll(ctx, id,
		"UPDATE profile_season_cosmetics SET name_color_id = NULL WHERE name_color_id = ?",
		"DELETE FROM profile_name_color_options WHERE name_color_id = ? AND revoked_at IS NOT NULL",
		"DELETE FROM name_colors WHERE id = ?",
	)
}

// DeleteNamePrefix removes the name prefix with its revoked grants and the selections left from them.
// A grant that is not revoked keeps the foreign key, so the delete fails with 1451.
func (q *queries) DeleteNamePrefix(ctx context.Context, id uuid.UUID) error {
	return q.execAll(ctx, id,
		"UPDATE profile_season_cosmetics SET glyth_prefix_id = NULL WHERE glyth_prefix_id = ?",
		"UPDATE profile_season_cosmetics SET special_prefix_id = NULL WHERE special_prefix_id = ?",
		"DELETE FROM profile_prefixes WHERE name_prefix_id = ?",
		"DELETE FROM profile_name_prefix_options WHERE name_prefix_id = ? AND revoked_at IS NOT NULL",
		"DELETE FROM name_prefixes WHERE id = ?",
	)
}
