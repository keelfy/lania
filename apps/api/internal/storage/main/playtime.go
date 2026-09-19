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

// Profiles unknown to the API are skipped, so players outside of the API never break the sync.
// updated_at goes first because MySQL applies assignments left to right.
const upsertProfilePlaytime = `
INSERT INTO profile_playtimes (mc_uuid, season_id, playtime, updated_at)
SELECT mc_uuid, ?, ?, now()
FROM profiles
WHERE mc_uuid = ?
ON DUPLICATE KEY UPDATE
	updated_at = IF(playtime <> VALUES(playtime), VALUES(updated_at), updated_at),
	playtime = VALUES(playtime)
`

// UpsertProfilePlaytime stores playtime of the profile in the season, in milliseconds.
func (q *queries) UpsertProfilePlaytime(ctx context.Context, mcUUID, seasonID uuid.UUID, playtime int64) error {
	_, err := q.x.ExecContext(ctx, upsertProfilePlaytime, seasonID, playtime, mcUUID)
	return err
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
