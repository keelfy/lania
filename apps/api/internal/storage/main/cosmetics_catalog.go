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
