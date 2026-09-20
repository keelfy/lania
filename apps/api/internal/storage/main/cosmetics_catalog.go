package sql

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/lania-smp/backend/internal/domain"
)

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
