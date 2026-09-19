package storage

import (
	"context"
	stdsql "database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/lania-smp/shell/internal/config"
	"github.com/lania-smp/shell/internal/logger"
)

// WhitelistStorage manages the table read by the VelocityWhitelist plugin.
type WhitelistStorage interface {
	Upsert(ctx context.Context, mcUUID uuid.UUID, username string) error
	Delete(ctx context.Context, mcUUID uuid.UUID) error
}

type whitelistStorage struct {
	db *stdsql.DB
}

// createWhitelistTable matches the schema VelocityWhitelist creates itself.
const createWhitelistTable = `
CREATE TABLE IF NOT EXISTS %s (
	mc_uuid varchar(36) PRIMARY KEY,
	username varchar(100) NOT NULL
)
`

func NewWhitelistStorage(ctx context.Context) (WhitelistStorage, func(), error) {
	db, cleanup, err := newMySQLStorage(ctx, config.GetDatabaseWhitelistName())
	if err != nil {
		return nil, nil, err
	}
	if _, err := db.ExecContext(ctx, fmt.Sprintf(createWhitelistTable, config.GetWhitelistTableName())); err != nil {
		logger.Errorf(ctx, "failed to create whitelist table: %v", err)
	}
	return &whitelistStorage{db: db}, cleanup, nil
}

const upsertWhitelist = `
INSERT INTO %s (mc_uuid, username)
VALUES (?, ?)
ON DUPLICATE KEY UPDATE username = VALUES(username)
`

func (s *whitelistStorage) Upsert(ctx context.Context, mcUUID uuid.UUID, username string) error {
	_, err := s.db.ExecContext(ctx, fmt.Sprintf(upsertWhitelist, config.GetWhitelistTableName()), mcUUID.String(), username)
	return err
}

const deleteWhitelist = `
DELETE FROM %s
WHERE mc_uuid = ?
`

func (s *whitelistStorage) Delete(ctx context.Context, mcUUID uuid.UUID) error {
	_, err := s.db.ExecContext(ctx, fmt.Sprintf(deleteWhitelist, config.GetWhitelistTableName()), mcUUID.String())
	return err
}
