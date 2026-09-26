package storage

import (
	"context"
	stdsql "database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/lania-smp/shell/internal/config"
	"github.com/lania-smp/shell/internal/domain"
)

// PlanStorage reads data written by the Plan plugin.
type PlanStorage interface {
	// FindPlaytimes returns playtime only for players known to Plan. A server name counts sessions on that server
	// of the network only; nil counts every server.
	FindPlaytimes(ctx context.Context, mcUUIDs uuid.UUIDs, serverName *string) (map[uuid.UUID]*domain.Playtime, error)
	// FindPlaytimesChangedSince returns playtime of players whose last session ended at or after sinceMs.
	FindPlaytimesChangedSince(ctx context.Context, sinceMs int64) (map[uuid.UUID]*domain.Playtime, error)
}

type planStorage struct {
	db *stdsql.DB
}

func NewPlanStorage(db *stdsql.DB) PlanStorage {
	return &planStorage{db: db}
}

const findPlaytimes = `
SELECT
	u.uuid,
	CAST(COALESCE(SUM(s.session_end - s.session_start - s.afk_time), 0) AS SIGNED) AS total_playtime,
	MIN(s.session_start) AS first_session_start,
	MAX(s.session_end) AS last_session_end
FROM %[1]s u
LEFT JOIN %[2]s s ON s.user_id = u.id%[4]s
WHERE u.uuid IN (%[3]s)
GROUP BY u.uuid
`

// sessionsOnServer narrows the sessions joined in findPlaytimes to one server; it takes the server name as the
// first argument.
const sessionsOnServer = `
	AND s.server_id IN (SELECT id FROM %s WHERE name = ?)`

func (s *planStorage) FindPlaytimes(ctx context.Context, mcUUIDs uuid.UUIDs, serverName *string) (map[uuid.UUID]*domain.Playtime, error) {
	playtimes := make(map[uuid.UUID]*domain.Playtime)
	if len(mcUUIDs) == 0 {
		return playtimes, nil
	}

	placeholders, args := uuidArgs(mcUUIDs)
	serverFilter := ""
	if serverName != nil {
		serverFilter = fmt.Sprintf(sessionsOnServer, config.GetPlanServersTableName())
		args = append([]any{*serverName}, args...)
	}
	query := fmt.Sprintf(findPlaytimes, config.GetPlanUsersTableName(), config.GetPlanSessionsTableName(), placeholders, serverFilter)
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return scanPlaytimes(rows, playtimes)
}

const findPlaytimesChangedSince = `
SELECT
	u.uuid,
	CAST(SUM(s.session_end - s.session_start - s.afk_time) AS SIGNED) AS total_playtime,
	MIN(s.session_start) AS first_session_start,
	MAX(s.session_end) AS last_session_end
FROM %[1]s u
JOIN %[2]s s ON s.user_id = u.id
GROUP BY u.uuid
HAVING MAX(s.session_end) >= ?
`

func (s *planStorage) FindPlaytimesChangedSince(ctx context.Context, sinceMs int64) (map[uuid.UUID]*domain.Playtime, error) {
	rows, err := s.db.QueryContext(ctx, fmt.Sprintf(findPlaytimesChangedSince, config.GetPlanUsersTableName(), config.GetPlanSessionsTableName()), sinceMs)
	if err != nil {
		return nil, err
	}
	return scanPlaytimes(rows, make(map[uuid.UUID]*domain.Playtime))
}

func scanPlaytimes(rows *stdsql.Rows, playtimes map[uuid.UUID]*domain.Playtime) (map[uuid.UUID]*domain.Playtime, error) {
	defer rows.Close()

	for rows.Next() {
		var mcUUID uuid.UUID
		var playtime domain.Playtime
		if err := rows.Scan(&mcUUID, &playtime.TotalMs, &playtime.FirstSeenMs, &playtime.LastSeenMs); err != nil {
			return nil, err
		}
		playtimes[mcUUID] = &playtime
	}
	return playtimes, rows.Err()
}
