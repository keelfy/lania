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
	// FindOnlineUUIDs returns every player that is online right now.
	FindOnlineUUIDs(ctx context.Context) (uuid.UUIDs, error)
}

type flectoneStorage struct {
	db *stdsql.DB
}

func NewFlectoneStorage(db *stdsql.DB) FlectoneStorage {
	return &flectoneStorage{db: db}
}

const findOnline = `
SELECT uuid, online
FROM %s
WHERE uuid IN (%s)
`

func (s *flectoneStorage) FindOnline(ctx context.Context, mcUUIDs uuid.UUIDs) (map[uuid.UUID]bool, error) {
	online := make(map[uuid.UUID]bool)
	if len(mcUUIDs) == 0 {
		return online, nil
	}

	placeholders, args := uuidArgs(mcUUIDs)
	rows, err := s.db.QueryContext(ctx, fmt.Sprintf(findOnline, config.GetFlectonePlayerTableName(), placeholders), args...)
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

const findOnlineUUIDs = `
SELECT uuid
FROM %s
WHERE online = 1
`

func (s *flectoneStorage) FindOnlineUUIDs(ctx context.Context) (uuid.UUIDs, error) {
	rows, err := s.db.QueryContext(ctx, fmt.Sprintf(findOnlineUUIDs, config.GetFlectonePlayerTableName()))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	mcUUIDs := make(uuid.UUIDs, 0)
	for rows.Next() {
		var mcUUID uuid.UUID
		if err := rows.Scan(&mcUUID); err != nil {
			return nil, err
		}
		mcUUIDs = append(mcUUIDs, mcUUID)
	}
	return mcUUIDs, rows.Err()
}
