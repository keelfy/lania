package sql

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/domain"
)

const seasonColumns = `
	id,
	season_number,
	name,
	preview_image,
	start_date,
	end_date,
	server_ip,
	server_port,
	is_active,
	rcon_password`

func scanSeason(row interface{ Scan(...any) error }) (*domain.Season, error) {
	var season domain.Season
	err := row.Scan(
		&season.ID,
		&season.SeasonNumber,
		&season.Name,
		&season.PreviewImage,
		&season.StartDate,
		&season.EndDate,
		&season.ServerIP,
		&season.ServerPort,
		&season.IsActive,
		&season.RCONPassword,
	)
	return &season, err
}

func (q *queries) FindSeasonByID(ctx context.Context, seasonID uuid.UUID) (*domain.Season, error) {
	return scanSeason(q.x.QueryRowContext(ctx, "SELECT"+seasonColumns+" FROM seasons WHERE id = ?", seasonID))
}

func (q *queries) FindPublicSeasons(ctx context.Context) ([]*domain.Season, error) {
	rows, err := q.x.QueryContext(ctx, `
SELECT id, name, preview_image, start_date, end_date, is_active
FROM seasons
ORDER BY start_date DESC, season_number DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	seasons := make([]*domain.Season, 0)
	for rows.Next() {
		var season domain.Season
		if err := rows.Scan(
			&season.ID, &season.Name, &season.PreviewImage,
			&season.StartDate, &season.EndDate, &season.IsActive,
		); err != nil {
			return nil, err
		}
		seasons = append(seasons, &season)
	}
	return seasons, rows.Err()
}

// FindSeasons returns every season, the newest start first.
func (q *queries) FindSeasons(ctx context.Context) ([]*domain.Season, error) {
	rows, err := q.x.QueryContext(ctx, "SELECT"+seasonColumns+" FROM seasons ORDER BY start_date DESC, season_number DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	seasons := make([]*domain.Season, 0)
	for rows.Next() {
		season, err := scanSeason(rows)
		if err != nil {
			return nil, err
		}
		seasons = append(seasons, season)
	}
	return seasons, rows.Err()
}

type InsertSeasonParams struct {
	ID           uuid.UUID
	SeasonNumber int
	Name         string
	PreviewImage *string
	StartDate    time.Time
	EndDate      *time.Time
	ServerIP     *string
	ServerPort   *uint16
	RCONPassword *string
}

func (q *queries) InsertSeason(ctx context.Context, arg InsertSeasonParams) error {
	_, err := q.x.ExecContext(ctx, `
INSERT INTO seasons (
	id, season_number, name, preview_image, start_date, end_date,
	server_ip, server_port, rcon_password
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		arg.ID, arg.SeasonNumber, arg.Name, arg.PreviewImage, arg.StartDate, arg.EndDate,
		arg.ServerIP, arg.ServerPort, arg.RCONPassword,
	)
	return err
}

type UpdateSeasonParams struct {
	ID              uuid.UUID
	SeasonNumber    int
	Name            string
	PreviewImage    *string
	StartDate       time.Time
	EndDate         *time.Time
	ServerIP        *string
	ServerPort      *uint16
	RCONPassword    *string
	SetRCONPassword bool
}

func (q *queries) UpdateSeason(ctx context.Context, arg UpdateSeasonParams) error {
	_, err := q.x.ExecContext(ctx, `
UPDATE seasons SET
	season_number = ?, name = ?, preview_image = ?, start_date = ?, end_date = ?,
	server_ip = ?, server_port = ?, 
	rcon_password = CASE WHEN ? THEN ? ELSE rcon_password END
WHERE id = ?`,
		arg.SeasonNumber, arg.Name, arg.PreviewImage, arg.StartDate, arg.EndDate,
		arg.ServerIP, arg.ServerPort,
		arg.SetRCONPassword, arg.RCONPassword, arg.ID,
	)
	return err
}

func (q *queries) DeleteSeason(ctx context.Context, seasonID uuid.UUID) (bool, error) {
	result, err := q.x.ExecContext(ctx, "DELETE FROM seasons WHERE id = ? AND is_active = false", seasonID)
	if err != nil {
		return false, err
	}
	count, err := result.RowsAffected()
	return count > 0, err
}

func (q *queries) ClearActiveSeasons(ctx context.Context) error {
	_, err := q.x.ExecContext(ctx, "UPDATE seasons SET is_active = false WHERE is_active = true")
	return err
}

func (q *queries) SetSeasonActive(ctx context.Context, seasonID uuid.UUID) (bool, error) {
	result, err := q.x.ExecContext(ctx, "UPDATE seasons SET is_active = true WHERE id = ?", seasonID)
	if err != nil {
		return false, err
	}
	count, err := result.RowsAffected()
	return count > 0, err
}
