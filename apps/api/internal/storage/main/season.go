package sql

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/domain"
)

const seasonColumns = `
	id,
	name,
	preview_image,
	start_date,
	end_date,
	public_address,
	shell_address,
	plan_url,
	is_active,
	is_primary,
	preregistration,
	free_registration,
	game_version,
	world_url,
	map_url,
	claim_limit`

func scanSeason(row interface{ Scan(...any) error }) (*domain.Season, error) {
	var season domain.Season
	err := row.Scan(
		&season.ID,
		&season.Name,
		&season.PreviewImage,
		&season.StartDate,
		&season.EndDate,
		&season.PublicAddress,
		&season.ShellAddress,
		&season.PlanURL,
		&season.IsActive,
		&season.IsPrimary,
		&season.Preregistration,
		&season.FreeRegistration,
		&season.GameVersion,
		&season.WorldURL,
		&season.MapURL,
		&season.ClaimLimit,
	)
	season.HasServer = season.ShellAddress != nil
	return &season, err
}

func (q *queries) FindSeasonByID(ctx context.Context, seasonID uuid.UUID) (*domain.Season, error) {
	return scanSeason(q.x.QueryRowContext(ctx, "SELECT"+seasonColumns+" FROM seasons WHERE id = ?", seasonID))
}

func (q *queries) FindPrimarySeasonID(ctx context.Context) (uuid.UUID, error) {
	var id uuid.UUID
	err := q.x.QueryRowContext(ctx, "SELECT id FROM seasons WHERE is_primary = true LIMIT 1").Scan(&id)
	return id, err
}

func (q *queries) FindPublicSeasons(ctx context.Context) ([]*domain.Season, error) {
	rows, err := q.x.QueryContext(ctx, `
SELECT id, name, preview_image, start_date, end_date,
       public_address, shell_address IS NOT NULL, is_active, is_primary, preregistration, free_registration,
       game_version, world_url, map_url, claim_limit
FROM seasons
ORDER BY start_date DESC, name ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	seasons := make([]*domain.Season, 0)
	for rows.Next() {
		var season domain.Season
		if err := rows.Scan(
			&season.ID, &season.Name, &season.PreviewImage,
			&season.StartDate, &season.EndDate,
			&season.PublicAddress,
			&season.HasServer,
			&season.IsActive, &season.IsPrimary,
			&season.Preregistration, &season.FreeRegistration,
			&season.GameVersion, &season.WorldURL,
			&season.MapURL, &season.ClaimLimit,
		); err != nil {
			return nil, err
		}
		seasons = append(seasons, &season)
	}
	return seasons, rows.Err()
}

// FindSeasons returns every season, the newest start first.
func (q *queries) FindSeasons(ctx context.Context) ([]*domain.Season, error) {
	rows, err := q.x.QueryContext(ctx, "SELECT"+seasonColumns+" FROM seasons ORDER BY start_date DESC, name ASC")
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
	ID               uuid.UUID
	Name             string
	PreviewImage     *string
	StartDate        time.Time
	EndDate          *time.Time
	PublicAddress    *string
	ShellAddress     *string
	PlanURL          *string
	IsActive         bool
	Preregistration  bool
	FreeRegistration bool
	GameVersion      *string
	WorldURL         *string
	MapURL           *string
	ClaimLimit       int
}

func (q *queries) InsertSeason(ctx context.Context, arg InsertSeasonParams) error {
	_, err := q.x.ExecContext(ctx, `
INSERT INTO seasons (
	id, name, preview_image, start_date, end_date,
	public_address, shell_address, plan_url, is_active, preregistration, free_registration,
	game_version, world_url, map_url, claim_limit
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		arg.ID, arg.Name, arg.PreviewImage, arg.StartDate, arg.EndDate,
		arg.PublicAddress,
		arg.ShellAddress,
		arg.PlanURL,
		arg.IsActive,
		arg.Preregistration,
		arg.FreeRegistration,
		arg.GameVersion,
		arg.WorldURL,
		arg.MapURL,
		arg.ClaimLimit,
	)
	return err
}

type UpdateSeasonParams struct {
	ID               uuid.UUID
	Name             string
	PreviewImage     *string
	StartDate        time.Time
	EndDate          *time.Time
	PublicAddress    *string
	ShellAddress     *string
	PlanURL          *string
	IsActive         bool
	Preregistration  bool
	FreeRegistration bool
	GameVersion      *string
	WorldURL         *string
	MapURL           *string
	ClaimLimit       int
}

func (q *queries) UpdateSeason(ctx context.Context, arg UpdateSeasonParams) error {
	_, err := q.x.ExecContext(ctx, `
UPDATE seasons SET
	name = ?, preview_image = ?, start_date = ?, end_date = ?,
	public_address = ?, shell_address = ?, plan_url = ?, is_active = ?,
	preregistration = ?, free_registration = ?, game_version = ?, world_url = ?,
	map_url = ?, claim_limit = ?
WHERE id = ?`,
		arg.Name, arg.PreviewImage, arg.StartDate, arg.EndDate,
		arg.PublicAddress,
		arg.ShellAddress,
		arg.PlanURL,
		arg.IsActive,
		arg.Preregistration,
		arg.FreeRegistration,
		arg.GameVersion,
		arg.WorldURL,
		arg.MapURL,
		arg.ClaimLimit,
		arg.ID,
	)
	return err
}

func (q *queries) DeleteSeason(ctx context.Context, seasonID uuid.UUID) (bool, error) {
	result, err := q.x.ExecContext(
		ctx,
		"DELETE FROM seasons WHERE id = ? AND is_active = false AND is_primary = false",
		seasonID,
	)
	if err != nil {
		return false, err
	}
	count, err := result.RowsAffected()
	return count > 0, err
}

func (q *queries) ClearPrimarySeasons(ctx context.Context) error {
	_, err := q.x.ExecContext(ctx, "UPDATE seasons SET is_primary = false WHERE is_primary = true")
	return err
}

func (q *queries) SetSeasonPrimary(ctx context.Context, seasonID uuid.UUID) (bool, error) {
	result, err := q.x.ExecContext(ctx, "UPDATE seasons SET is_primary = true WHERE id = ?", seasonID)
	if err != nil {
		return false, err
	}
	count, err := result.RowsAffected()
	return count > 0, err
}

func (q *queries) SetSeasonShellAddress(ctx context.Context, seasonID uuid.UUID, address string) error {
	_, err := q.x.ExecContext(ctx, "UPDATE seasons SET shell_address = ? WHERE id = ?", address, seasonID)
	return err
}
