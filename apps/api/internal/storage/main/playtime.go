package sql

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/domain"
)

const findProfilePlaytimesByMinecraftUUIDs = `
SELECT 
	mc_uuid,
	season_id,
	playtime,
	updated_at
FROM profile_playtimes
WHERE mc_uuid IN ('%s')
`

func (q *queries) FindProfilePlaytimesByMinecraftUUIDs(ctx context.Context, mcUUIDs uuid.UUIDs) ([]*domain.ProfilePlaytime, error) {

	mcUUIDsStr := make([]string, len(mcUUIDs))
	for i, mcUUID := range mcUUIDs {
		mcUUIDsStr[i] = mcUUID.String()
	}
	query := fmt.Sprintf(findProfilePlaytimesByMinecraftUUIDs, strings.Join(mcUUIDsStr, "','"))
	rows, err := q.x.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	playtimes := make([]*domain.ProfilePlaytime, 0)
	for rows.Next() {
		var playtime domain.ProfilePlaytime
		err := rows.Scan(
			&playtime.MinecraftUUID,
			&playtime.SeasonID,
			&playtime.Playtime,
			&playtime.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		playtimes = append(playtimes, &playtime)
	}
	return playtimes, nil
}

const sumProfilePlaytimesByMinecraftUUIDs = `
SELECT
	mc_uuid,
	CAST(SUM(playtime) AS SIGNED) AS total_playtime
FROM profile_playtimes
WHERE mc_uuid IN ('%s')
GROUP BY mc_uuid
`

func (q *queries) SumProfilePlaytimesByMinecraftUUIDs(ctx context.Context, mcUUIDs uuid.UUIDs) (map[uuid.UUID]int64, error) {
	mcUUIDsStr := make([]string, len(mcUUIDs))
	for i, mcUUID := range mcUUIDs {
		mcUUIDsStr[i] = mcUUID.String()
	}
	query := fmt.Sprintf(sumProfilePlaytimesByMinecraftUUIDs, strings.Join(mcUUIDsStr, "','"))
	rows, err := q.x.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	totals := make(map[uuid.UUID]int64)
	for rows.Next() {
		var mcUUID uuid.UUID
		var total int64
		if err := rows.Scan(&mcUUID, &total); err != nil {
			return nil, err
		}
		totals[mcUUID] = total
	}
	return totals, rows.Err()
}

const findProfileSeasonStats = `
SELECT
	s.id,
	s.name,
	s.start_date,
	s.end_date,
	s.is_active,
	s.is_primary,
	pt.playtime
FROM profile_playtimes pt
JOIN seasons s ON s.id = pt.season_id
WHERE pt.mc_uuid = ? AND pt.playtime > 0
ORDER BY s.start_date DESC, s.name ASC
`

// FindProfileSeasonStats returns the stats of the profile in every season it played in, the newest season first.
func (q *queries) FindProfileSeasonStats(ctx context.Context, mcUUID uuid.UUID) ([]*domain.ProfileSeasonStats, error) {
	rows, err := q.x.QueryContext(ctx, findProfileSeasonStats, mcUUID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats := make([]*domain.ProfileSeasonStats, 0)
	for rows.Next() {
		var seasonStats domain.ProfileSeasonStats
		if err := rows.Scan(
			&seasonStats.SeasonID,
			&seasonStats.SeasonName,
			&seasonStats.StartDate,
			&seasonStats.EndDate,
			&seasonStats.IsActive,
			&seasonStats.IsPrimary,
			&seasonStats.Playtime,
		); err != nil {
			return nil, err
		}
		stats = append(stats, &seasonStats)
	}
	return stats, rows.Err()
}

// Profiles unknown to the API are skipped, so players outside of the API never break the sync.
// updated_at goes first because MySQL applies assignments left to right.
// Columns are qualified, because profiles has updated_at too and the SELECT makes a bare name ambiguous.
// last_seen_at never moves back, and a NULL date keeps the stored one.
const upsertProfilePlaytime = `
INSERT INTO profile_playtimes (mc_uuid, season_id, playtime, updated_at, last_seen_at)
SELECT mc_uuid, ?, ?, now(), ?
FROM profiles
WHERE mc_uuid = ?
ON DUPLICATE KEY UPDATE
	profile_playtimes.updated_at = IF(profile_playtimes.playtime <> VALUES(playtime), VALUES(updated_at), profile_playtimes.updated_at),
	profile_playtimes.playtime = VALUES(playtime),
	profile_playtimes.last_seen_at = GREATEST(
		COALESCE(profile_playtimes.last_seen_at, VALUES(last_seen_at)),
		COALESCE(VALUES(last_seen_at), profile_playtimes.last_seen_at)
	)
`

// UpsertProfilePlaytime stores playtime of the profile in the season, in milliseconds,
// and moves the last seen date of the profile in the season forward.
func (q *queries) UpsertProfilePlaytime(ctx context.Context, mcUUID, seasonID uuid.UUID, playtime int64, lastSeenAt *time.Time) error {
	_, err := q.x.ExecContext(ctx, upsertProfilePlaytime, seasonID, playtime, lastSeenAt, mcUUID)
	return err
}

const findProfilesLastSeenInSeason = `
SELECT mc_uuid, last_seen_at
FROM profile_playtimes
WHERE season_id = ? AND last_seen_at IS NOT NULL AND mc_uuid IN (%s)
`

func (q *queries) FindProfilesLastSeenInSeason(ctx context.Context, mcUUIDs uuid.UUIDs, seasonID uuid.UUID) (map[uuid.UUID]time.Time, error) {
	lastSeen := make(map[uuid.UUID]time.Time, len(mcUUIDs))
	if len(mcUUIDs) == 0 {
		return lastSeen, nil
	}

	placeholders := make([]string, len(mcUUIDs))
	args := make([]any, 0, len(mcUUIDs)+1)
	args = append(args, seasonID)
	for i, mcUUID := range mcUUIDs {
		placeholders[i] = "?"
		args = append(args, mcUUID)
	}
	rows, err := q.x.QueryContext(ctx, fmt.Sprintf(findProfilesLastSeenInSeason, strings.Join(placeholders, ", ")), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var mcUUID uuid.UUID
		var seenAt time.Time
		if err := rows.Scan(&mcUUID, &seenAt); err != nil {
			return nil, err
		}
		lastSeen[mcUUID] = seenAt
	}
	return lastSeen, rows.Err()
}

// first_seen_at is only filled when empty, so dates imported from earlier seasons are kept.
// last_seen_at never moves back.
const updateProfileSeenAt = `
UPDATE profiles
SET
	first_seen_at = COALESCE(first_seen_at, ?),
	last_seen_at = GREATEST(COALESCE(last_seen_at, ?), COALESCE(?, last_seen_at))
WHERE mc_uuid = ?
`

// UpdateProfileSeenAt merges seen dates from the Minecraft server into the profile. Nil dates are ignored.
func (q *queries) UpdateProfileSeenAt(ctx context.Context, mcUUID uuid.UUID, firstSeenAt, lastSeenAt *time.Time) error {
	_, err := q.x.ExecContext(ctx, updateProfileSeenAt, firstSeenAt, lastSeenAt, lastSeenAt, mcUUID)
	return err
}
