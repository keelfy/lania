package sql

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/domain"
)

// profileColumns are the columns scanProfileRows reads, of the profile joined as p.
const profileColumns = `
	p.id,
	p.mc_uuid,
	p.mc_username,
	p.owner_user_id,
	p.first_seen_at,
	p.last_seen_at,
	p.role,
	p.is_slim,
	p.created_at,
	p.updated_at,
	p.updated_by,
	p.legacy_mc_uuid,
	p.premium_conflict,
	IF(p.verified_mc_uuid <=> p.mc_uuid, p.verified_at, NULL)
`

func unmarshalStreamerChannels(raw string) ([]domain.StreamerChannel, error) {
	channels := make([]domain.StreamerChannel, 0)
	if err := json.Unmarshal([]byte(raw), &channels); err != nil {
		return nil, fmt.Errorf("failed to parse streamer channels: %w", err)
	}
	return channels, nil
}

const findStreamer = `
SELECT profile_id, channels, description, created_at, updated_at FROM streamers WHERE profile_id = ?
`

func (q *queries) FindStreamer(ctx context.Context, profileID uuid.UUID) (*domain.Streamer, error) {
	var streamer domain.Streamer
	var channels string
	err := q.x.QueryRowContext(ctx, findStreamer, profileID).
		Scan(&streamer.ProfileID, &channels, &streamer.Description, &streamer.CreatedAt, &streamer.UpdatedAt)
	if err != nil {
		return nil, err
	}
	if streamer.Channels, err = unmarshalStreamerChannels(channels); err != nil {
		return nil, err
	}
	return &streamer, nil
}

const findStreamers = `
SELECT ` + profileColumns + `, s.channels, s.description, s.created_at, s.updated_at
FROM streamers s
JOIN profiles p ON p.id = s.profile_id
ORDER BY s.created_at ASC
`

func (q *queries) FindStreamers(ctx context.Context) ([]*domain.Streamer, error) {
	rows, err := q.x.QueryContext(ctx, findStreamers)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	streamers := make([]*domain.Streamer, 0)
	for rows.Next() {
		var streamer domain.Streamer
		var channels string
		profile, err := scanProfileRows(rows, &channels, &streamer.Description, &streamer.CreatedAt, &streamer.UpdatedAt)
		if err != nil {
			return nil, err
		}
		if streamer.Channels, err = unmarshalStreamerChannels(channels); err != nil {
			return nil, err
		}
		streamer.ProfileID = profile.ID
		streamer.Profile = profile
		streamers = append(streamers, &streamer)
	}
	return streamers, rows.Err()
}

const findStreamerProfileIDs = `
SELECT profile_id FROM streamers WHERE profile_id IN (%s)
`

func (q *queries) FindStreamerProfileIDs(ctx context.Context, profileIDs uuid.UUIDs) (map[uuid.UUID]bool, error) {
	res := make(map[uuid.UUID]bool, len(profileIDs))
	if len(profileIDs) == 0 {
		return res, nil
	}

	args := make([]any, len(profileIDs))
	for i, profileID := range profileIDs {
		args[i] = profileID
	}
	placeholders := strings.TrimSuffix(strings.Repeat("?,", len(profileIDs)), ",")
	rows, err := q.x.QueryContext(ctx, fmt.Sprintf(findStreamerProfileIDs, placeholders), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var profileID uuid.UUID
		if err := rows.Scan(&profileID); err != nil {
			return nil, err
		}
		res[profileID] = true
	}
	return res, rows.Err()
}

const saveStreamer = `
INSERT INTO streamers (profile_id, channels, description, created_at, created_by, updated_at)
VALUES (?, ?, ?, now(), ?, now())
ON DUPLICATE KEY UPDATE channels = VALUES(channels), description = VALUES(description), updated_at = now()
`

type SaveStreamerParams struct {
	ProfileID   uuid.UUID
	Channels    []domain.StreamerChannel
	Description *string
	// CreatedBy is kept only when the profile becomes a streamer.
	CreatedBy *uuid.UUID
}

func (q *queries) SaveStreamer(ctx context.Context, arg SaveStreamerParams) error {
	channels, err := json.Marshal(arg.Channels)
	if err != nil {
		return err
	}
	_, err = q.x.ExecContext(ctx, saveStreamer, arg.ProfileID, string(channels), arg.Description, arg.CreatedBy)
	return err
}

const deleteStreamer = `
DELETE FROM streamers WHERE profile_id = ?
`

func (q *queries) DeleteStreamer(ctx context.Context, profileID uuid.UUID) (bool, error) {
	affected, err := execAffected(ctx, q.x, deleteStreamer, profileID)
	return affected > 0, err
}

const streamerApplicationColumns = `
	a.id, a.profile_id, a.user_id, a.channels, a.about, a.status, a.reject_reason, a.reviewed_by, a.reviewed_at, a.created_at
`

// scanStreamerApplication scans streamerApplicationColumns with scan.
func scanStreamerApplication(scan func(dest ...any) error) (*domain.StreamerApplication, error) {
	var application domain.StreamerApplication
	var channels string
	err := scan(
		&application.ID,
		&application.ProfileID,
		&application.UserID,
		&channels,
		&application.About,
		&application.Status,
		&application.RejectReason,
		&application.ReviewedBy,
		&application.ReviewedAt,
		&application.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	if application.Channels, err = unmarshalStreamerChannels(channels); err != nil {
		return nil, err
	}
	return &application, nil
}

const insertStreamerApplication = `
INSERT INTO streamer_applications (id, profile_id, user_id, channels, about, status, created_at)
VALUES (?, ?, ?, ?, ?, ?, now())
`

func (q *queries) InsertStreamerApplication(ctx context.Context, application *domain.StreamerApplication) error {
	channels, err := json.Marshal(application.Channels)
	if err != nil {
		return err
	}
	_, err = q.x.ExecContext(ctx, insertStreamerApplication,
		application.ID,
		application.ProfileID,
		application.UserID,
		string(channels),
		application.About,
		application.Status,
	)
	return err
}

const findLatestStreamerApplication = `
SELECT ` + streamerApplicationColumns + `
FROM streamer_applications a
WHERE a.profile_id = ?
ORDER BY a.created_at DESC
LIMIT 1
`

func (q *queries) FindLatestStreamerApplication(ctx context.Context, profileID uuid.UUID) (*domain.StreamerApplication, error) {
	return scanStreamerApplication(q.x.QueryRowContext(ctx, findLatestStreamerApplication, profileID).Scan)
}

const findStreamerApplicationByID = `
SELECT ` + streamerApplicationColumns + `
FROM streamer_applications a
WHERE a.id = ?
`

func (q *queries) FindStreamerApplicationByID(ctx context.Context, applicationID uuid.UUID) (*domain.StreamerApplication, error) {
	return scanStreamerApplication(q.x.QueryRowContext(ctx, findStreamerApplicationByID, applicationID).Scan)
}

const findStreamerApplications = `
SELECT ` + profileColumns + `, ` + streamerApplicationColumns + `
FROM streamer_applications a
JOIN profiles p ON p.id = a.profile_id
WHERE a.status = ?
ORDER BY a.created_at %s
LIMIT ? OFFSET ?
`

func (q *queries) FindStreamerApplications(ctx context.Context, status domain.StreamerApplicationStatus, size, from int) ([]*domain.StreamerApplication, error) {
	direction := "DESC"
	if status == domain.StreamerApplicationStatusPending {
		direction = "ASC"
	}
	rows, err := q.x.QueryContext(ctx, fmt.Sprintf(findStreamerApplications, direction), status, size, from)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applications := make([]*domain.StreamerApplication, 0)
	for rows.Next() {
		var profile *domain.Profile
		application, err := scanStreamerApplication(func(dest ...any) error {
			var scanErr error
			profile, scanErr = scanProfileRows(rows, dest...)
			return scanErr
		})
		if err != nil {
			return nil, err
		}
		application.Profile = profile
		applications = append(applications, application)
	}
	return applications, rows.Err()
}

const countStreamerApplications = `
SELECT COUNT(*) FROM streamer_applications WHERE status = ?
`

func (q *queries) CountStreamerApplications(ctx context.Context, status domain.StreamerApplicationStatus) (int64, error) {
	var count int64
	err := q.x.QueryRowContext(ctx, countStreamerApplications, status).Scan(&count)
	return count, err
}

const reviewStreamerApplication = `
UPDATE streamer_applications
SET status = ?, reject_reason = ?, reviewed_by = ?, reviewed_at = now()
WHERE id = ? AND status = 'pending'
`

func (q *queries) ReviewStreamerApplication(ctx context.Context, applicationID uuid.UUID, status domain.StreamerApplicationStatus, reason *string, reviewedBy uuid.UUID) (bool, error) {
	affected, err := execAffected(ctx, q.x, reviewStreamerApplication, status, reason, reviewedBy, applicationID)
	return affected > 0, err
}
