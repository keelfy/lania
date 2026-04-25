package sql

import (
	"context"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/domain"
)

const findSeasonByID = `
SELECT 
	id,
	season_number,
	start_date,
	end_date
FROM seasons
WHERE id = ?
`

func (q *queries) FindSeasonByID(ctx context.Context, seasonID uuid.UUID) (*domain.Season, error) {
	row := q.x.QueryRowContext(ctx, findSeasonByID, seasonID)
	var season domain.Season
	err := row.Scan(
		&season.ID,
		&season.SeasonNumber,
		&season.StartDate,
		&season.EndDate,
	)
	return &season, err
}
