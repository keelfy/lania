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
	// FindPlaytimes returns playtime only for players known to Plan.
	FindPlaytimes(ctx context.Context, mcUUIDs uuid.UUIDs) (map[uuid.UUID]*domain.Playtime, error)
}

type planStorage struct {
	db *stdsql.DB
}

func NewPlanStorage(ctx context.Context) (PlanStorage, func(), error) {
	db, cleanup, err := newMySQLStorage(ctx, config.GetDatabasePlanName())
	if err != nil {
		return nil, nil, err
	}
	return &planStorage{db: db}, cleanup, nil
}

const findPlaytimes = `
SELECT
	u.uuid,
	CAST(COALESCE(SUM(s.session_end - s.session_start - s.afk_time), 0) AS SIGNED) AS total_playtime,
	MIN(s.session_start) AS first_session_start,
	MAX(s.session_end) AS last_session_end
FROM plan_users u
LEFT JOIN plan_sessions s ON s.user_id = u.id
WHERE u.uuid IN (%s)
GROUP BY u.uuid
`

func (s *planStorage) FindPlaytimes(ctx context.Context, mcUUIDs uuid.UUIDs) (map[uuid.UUID]*domain.Playtime, error) {
	playtimes := make(map[uuid.UUID]*domain.Playtime)
	if len(mcUUIDs) == 0 {
		return playtimes, nil
	}

	placeholders, args := uuidArgs(mcUUIDs)
	rows, err := s.db.QueryContext(ctx, fmt.Sprintf(findPlaytimes, placeholders), args...)
	if err != nil {
		return nil, err
	}
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
