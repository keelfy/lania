package storage

import (
	"context"
	stdsql "database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/lania-smp/shell/internal/config"
)

// FlectoneStorage reads data written by the FlectonePulse plugin.
type FlectoneStorage interface {
	// FindOnline returns online flags only for players known to Flectone.
	FindOnline(ctx context.Context, mcUUIDs uuid.UUIDs) (map[uuid.UUID]bool, error)
}

type flectoneStorage struct {
	db *stdsql.DB
}

func NewFlectoneStorage(ctx context.Context) (FlectoneStorage, func(), error) {
	db, cleanup, err := newMySQLStorage(ctx, config.GetDatabaseFlectoneName())
	if err != nil {
		return nil, nil, err
	}
	return &flectoneStorage{db: db}, cleanup, nil
}

const findOnline = `
SELECT uuid, online
FROM player
WHERE uuid IN (%s)
`

func (s *flectoneStorage) FindOnline(ctx context.Context, mcUUIDs uuid.UUIDs) (map[uuid.UUID]bool, error) {
	online := make(map[uuid.UUID]bool)
	if len(mcUUIDs) == 0 {
		return online, nil
	}

	placeholders, args := uuidArgs(mcUUIDs)
	rows, err := s.db.QueryContext(ctx, fmt.Sprintf(findOnline, placeholders), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var mcUUID uuid.UUID
		var isOnline bool
		if err := rows.Scan(&mcUUID, &isOnline); err != nil {
			return nil, err
		}
		online[mcUUID] = isOnline
	}
	return online, rows.Err()
}
