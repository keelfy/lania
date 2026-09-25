package sql

import (
	"context"
	stdsql "database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/domain"
)

const seasonScreenshotJoinColumns = `
	ss.id, ss.season_id, ss.image, ss.title, ss.position, ss.created_at,
	p.id, p.mc_username
FROM season_screenshots ss
LEFT JOIN season_screenshot_authors ssa ON ssa.screenshot_id = ss.id
LEFT JOIN profiles p ON p.id = ssa.profile_id
`

const findSeasonScreenshots = `SELECT` + seasonScreenshotJoinColumns + `
WHERE ss.season_id = ?
ORDER BY ss.position, ss.created_at, ss.id, ssa.position
`

const findSeasonScreenshotByID = `SELECT` + seasonScreenshotJoinColumns + `
WHERE ss.id = ? AND ss.season_id = ?
ORDER BY ssa.position
`

// scanSeasonScreenshotRows groups the joined rows into one entry per screenshot, each with its credited profiles
// in credit order. A screenshot with no matching author row (none credited yet) still comes back with Authors empty.
func scanSeasonScreenshotRows(rows *stdsql.Rows) ([]*domain.SeasonScreenshot, error) {
	screenshots := make([]*domain.SeasonScreenshot, 0)
	byID := make(map[uuid.UUID]*domain.SeasonScreenshot)
	for rows.Next() {
		var screenshot domain.SeasonScreenshot
		var authorID *uuid.UUID
		var authorUsername *string
		err := rows.Scan(
			&screenshot.ID, &screenshot.SeasonID, &screenshot.Image, &screenshot.Title,
			&screenshot.Position, &screenshot.CreatedAt,
			&authorID, &authorUsername,
		)
		if err != nil {
			return nil, err
		}

		existing, ok := byID[screenshot.ID]
		if !ok {
			screenshot.Authors = make([]*domain.Profile, 0)
			existing = &screenshot
			byID[screenshot.ID] = existing
			screenshots = append(screenshots, existing)
		}
		if authorID != nil {
			existing.Authors = append(existing.Authors, &domain.Profile{
				ID: *authorID, MinecraftUsername: *authorUsername,
			})
		}
	}
	return screenshots, rows.Err()
}

// FindSeasonScreenshots returns every screenshot of the season, in feed order.
func (q *queries) FindSeasonScreenshots(ctx context.Context, seasonID uuid.UUID) ([]*domain.SeasonScreenshot, error) {
	rows, err := q.x.QueryContext(ctx, findSeasonScreenshots, seasonID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanSeasonScreenshotRows(rows)
}

// FindSeasonScreenshotByID returns the screenshot, or stdsql.ErrNoRows when it does not exist in the season.
func (q *queries) FindSeasonScreenshotByID(ctx context.Context, screenshotID, seasonID uuid.UUID) (*domain.SeasonScreenshot, error) {
	rows, err := q.x.QueryContext(ctx, findSeasonScreenshotByID, screenshotID, seasonID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	screenshots, err := scanSeasonScreenshotRows(rows)
	if err != nil {
		return nil, err
	}
	if len(screenshots) == 0 {
		return nil, stdsql.ErrNoRows
	}
	return screenshots[0], nil
}

type SaveSeasonScreenshotParams struct {
	ID       uuid.UUID
	SeasonID uuid.UUID
	Image    string
	Title    *string
	Position int
}

const insertSeasonScreenshot = `
INSERT INTO season_screenshots (id, season_id, image, title, position)
VALUES (?, ?, ?, ?, ?)
`

func (q *queries) CreateSeasonScreenshot(ctx context.Context, arg SaveSeasonScreenshotParams) error {
	_, err := q.x.ExecContext(ctx, insertSeasonScreenshot, arg.ID, arg.SeasonID, arg.Image, arg.Title, arg.Position)
	return err
}

const updateSeasonScreenshot = `
UPDATE season_screenshots SET image = ?, title = ?, position = ? WHERE id = ? AND season_id = ?
`

// UpdateSeasonScreenshot reports whether a screenshot with the id existed in the season.
func (q *queries) UpdateSeasonScreenshot(ctx context.Context, arg SaveSeasonScreenshotParams) (bool, error) {
	result, err := q.x.ExecContext(ctx, updateSeasonScreenshot, arg.Image, arg.Title, arg.Position, arg.ID, arg.SeasonID)
	if err != nil {
		return false, err
	}
	count, err := result.RowsAffected()
	return count > 0, err
}

const deleteSeasonScreenshotAuthors = `
DELETE FROM season_screenshot_authors WHERE screenshot_id = ?
`

const deleteSeasonScreenshot = `
DELETE FROM season_screenshots WHERE id = ? AND season_id = ?
`

// DeleteSeasonScreenshot removes the screenshot and its author credits, reporting whether it existed in the season.
func (q *queries) DeleteSeasonScreenshot(ctx context.Context, screenshotID, seasonID uuid.UUID) (bool, error) {
	if _, err := q.x.ExecContext(ctx, deleteSeasonScreenshotAuthors, screenshotID); err != nil {
		return false, err
	}
	result, err := q.x.ExecContext(ctx, deleteSeasonScreenshot, screenshotID, seasonID)
	if err != nil {
		return false, err
	}
	count, err := result.RowsAffected()
	return count > 0, err
}

const insertSeasonScreenshotAuthor = `
INSERT INTO season_screenshot_authors (screenshot_id, profile_id, position) VALUES (?, ?, ?)
`

// SetSeasonScreenshotAuthors replaces every credit of the screenshot with profileIDs, in the given order.
// It deletes then inserts rather than diffing: the admin form always submits the full author list, never a delta.
func (q *queries) SetSeasonScreenshotAuthors(ctx context.Context, screenshotID uuid.UUID, profileIDs uuid.UUIDs) error {
	if _, err := q.x.ExecContext(ctx, deleteSeasonScreenshotAuthors, screenshotID); err != nil {
		return err
	}
	for position, profileID := range profileIDs {
		if _, err := q.x.ExecContext(ctx, insertSeasonScreenshotAuthor, screenshotID, profileID, position); err != nil {
			return err
		}
	}
	return nil
}

const findProfilesByIDs = `
SELECT
	id, mc_uuid, mc_username, owner_user_id, first_seen_at, last_seen_at,
	role, is_slim, created_at, updated_at, updated_by, legacy_mc_uuid, premium_conflict,
	IF(verified_mc_uuid <=> mc_uuid, verified_at, NULL)
FROM profiles
WHERE id IN ('%s')
`

// FindProfilesByIDs returns the profiles that exist among ids, in no particular order. A missing id is simply
// absent from the result, the caller compares lengths to know which ids do not exist.
func (q *queries) FindProfilesByIDs(ctx context.Context, ids uuid.UUIDs) ([]*domain.Profile, error) {
	if len(ids) == 0 {
		return []*domain.Profile{}, nil
	}
	idsStr := make([]string, len(ids))
	for i, id := range ids {
		idsStr[i] = id.String()
	}
	query := fmt.Sprintf(findProfilesByIDs, strings.Join(idsStr, "','"))
	rows, err := q.x.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	profiles := make([]*domain.Profile, 0, len(ids))
	for rows.Next() {
		profile, err := scanProfileRows(rows)
		if err != nil {
			return nil, err
		}
		profiles = append(profiles, profile)
	}
	return profiles, rows.Err()
}
