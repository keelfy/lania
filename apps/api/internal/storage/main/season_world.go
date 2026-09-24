package sql

import (
	"context"
	stdsql "database/sql"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/domain"
)

const seasonWorldColumns = `
	id, season_id, slug, name, preview_image, map_url, claim_limit, claim_dimensions, position
FROM season_worlds`

func scanSeasonWorld(row interface{ Scan(...any) error }) (*domain.SeasonWorld, error) {
	var world domain.SeasonWorld
	var dimensions string
	err := row.Scan(
		&world.ID, &world.SeasonID, &world.Slug, &world.Name, &world.PreviewImage,
		&world.MapURL, &world.ClaimLimit, &dimensions, &world.Position,
	)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(dimensions), &world.ClaimDimensions); err != nil {
		return nil, err
	}
	return &world, nil
}

// FindSeasonWorlds returns the worlds of the season in display order.
func (q *queries) FindSeasonWorlds(ctx context.Context, seasonID uuid.UUID) ([]*domain.SeasonWorld, error) {
	rows, err := q.x.QueryContext(ctx, "SELECT"+seasonWorldColumns+" WHERE season_id = ? ORDER BY position, created_at, id", seasonID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	worlds := make([]*domain.SeasonWorld, 0)
	for rows.Next() {
		world, err := scanSeasonWorld(rows)
		if err != nil {
			return nil, err
		}
		worlds = append(worlds, world)
	}
	return worlds, rows.Err()
}

// FindSeasonWorldByID returns stdsql.ErrNoRows when the world does not exist.
func (q *queries) FindSeasonWorldByID(ctx context.Context, worldID uuid.UUID) (*domain.SeasonWorld, error) {
	return scanSeasonWorld(q.x.QueryRowContext(ctx, "SELECT"+seasonWorldColumns+" WHERE id = ?", worldID))
}

type SaveSeasonWorldParams struct {
	ID              uuid.UUID
	SeasonID        uuid.UUID
	Slug            string
	Name            string
	PreviewImage    *string
	MapURL          *string
	ClaimLimit      int
	ClaimDimensions []string
	Position        int
}

func claimDimensionsJSON(dimensions []string) (string, error) {
	if dimensions == nil {
		dimensions = []string{}
	}
	raw, err := json.Marshal(dimensions)
	return string(raw), err
}

const insertSeasonWorld = `
INSERT INTO season_worlds (id, season_id, slug, name, preview_image, map_url, claim_limit, claim_dimensions, position)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
`

func (q *queries) CreateSeasonWorld(ctx context.Context, arg SaveSeasonWorldParams) error {
	dimensions, err := claimDimensionsJSON(arg.ClaimDimensions)
	if err != nil {
		return err
	}
	_, err = q.x.ExecContext(ctx, insertSeasonWorld,
		arg.ID, arg.SeasonID, arg.Slug, arg.Name, arg.PreviewImage, arg.MapURL, arg.ClaimLimit, dimensions, arg.Position,
	)
	return err
}

const updateSeasonWorld = `
UPDATE season_worlds SET
	slug = ?, name = ?, preview_image = ?, map_url = ?, claim_limit = ?, claim_dimensions = ?, position = ?
WHERE id = ?
`

// UpdateSeasonWorld reports whether a world with the id existed.
func (q *queries) UpdateSeasonWorld(ctx context.Context, arg SaveSeasonWorldParams) (bool, error) {
	dimensions, err := claimDimensionsJSON(arg.ClaimDimensions)
	if err != nil {
		return false, err
	}
	result, err := q.x.ExecContext(ctx, updateSeasonWorld,
		arg.Slug, arg.Name, arg.PreviewImage, arg.MapURL, arg.ClaimLimit, dimensions, arg.Position, arg.ID,
	)
	if err != nil {
		return false, err
	}
	// MariaDB counts changed rows, so an update that changes nothing still has to find the world.
	count, err := result.RowsAffected()
	if err != nil || count > 0 {
		return count > 0, err
	}
	_, err = q.FindSeasonWorldByID(ctx, arg.ID)
	if errors.Is(err, stdsql.ErrNoRows) {
		return false, nil
	}
	return err == nil, err
}

const countChunkClaimsInWorld = `
SELECT COUNT(*) FROM chunk_claims WHERE world_id = ?
`

// CountChunkClaimsInWorld counts every claim the world ever had, released ones too.
func (q *queries) CountChunkClaimsInWorld(ctx context.Context, worldID uuid.UUID) (int, error) {
	var count int
	err := q.x.QueryRowContext(ctx, countChunkClaimsInWorld, worldID).Scan(&count)
	return count, err
}

// DeleteSeasonWorld reports whether the world existed.
func (q *queries) DeleteSeasonWorld(ctx context.Context, worldID uuid.UUID) (bool, error) {
	result, err := q.x.ExecContext(ctx, "DELETE FROM season_worlds WHERE id = ?", worldID)
	if err != nil {
		return false, err
	}
	count, err := result.RowsAffected()
	return count > 0, err
}
