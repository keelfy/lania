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

const findSeasons = `
SELECT
	id,
	season_number,
	start_date,
	end_date
FROM seasons
ORDER BY season_number DESC
`

// FindSeasons returns every season, the newest number first.
func (q *queries) FindSeasons(ctx context.Context) ([]*domain.Season, error) {
	rows, err := q.x.QueryContext(ctx, findSeasons)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	seasons := make([]*domain.Season, 0)
	for rows.Next() {
		var season domain.Season
		if err := rows.Scan(&season.ID, &season.SeasonNumber, &season.StartDate, &season.EndDate); err != nil {
			return nil, err
		}
		seasons = append(seasons, &season)
	}
	return seasons, rows.Err()
}
