package storage

import (
	"context"
	stdsql "database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/lania-smp/shell/internal/config"
)

// LuckpermsStorage reads and writes LuckPerms user permission nodes.
type LuckpermsStorage interface {
	// FindPermissionsWithPrefix returns permission nodes starting with nodePrefix, keyed by player.
	FindPermissionsWithPrefix(ctx context.Context, mcUUIDs uuid.UUIDs, nodePrefix string) (map[uuid.UUID][]string, error)
	// FindPlayersWithPermissions returns players that have at least one of the permission nodes.
	FindPlayersWithPermissions(ctx context.Context, permissions []string) (uuid.UUIDs, error)
	// ReplacePermissionsWithPrefix deletes the player's nodes starting with nodePrefix and inserts node.
	ReplacePermissionsWithPrefix(ctx context.Context, mcUUID uuid.UUID, nodePrefix string, node string) error
}

type luckpermsStorage struct {
	db *stdsql.DB
}

func NewLuckpermsStorage(ctx context.Context) (LuckpermsStorage, func(), error) {
	db, cleanup, err := newMySQLStorage(ctx, config.GetDatabaseLuckpermsName())
	if err != nil {
		return nil, nil, err
	}
	return &luckpermsStorage{db: db}, cleanup, nil
}

const findPermissionsWithPrefix = `
SELECT uuid, permission
FROM %s
WHERE uuid IN (%s) AND permission LIKE ?
`

func (s *luckpermsStorage) FindPermissionsWithPrefix(ctx context.Context, mcUUIDs uuid.UUIDs, nodePrefix string) (map[uuid.UUID][]string, error) {
	permissions := make(map[uuid.UUID][]string)
	if len(mcUUIDs) == 0 {
		return permissions, nil
	}

	placeholders, args := uuidArgs(mcUUIDs)
	args = append(args, nodePrefix+"%")
	query := fmt.Sprintf(findPermissionsWithPrefix, config.GetLuckpermsUserPermissionsTableName(), placeholders)
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var mcUUID uuid.UUID
		var permission string
		if err := rows.Scan(&mcUUID, &permission); err != nil {
			return nil, err
		}
		permissions[mcUUID] = append(permissions[mcUUID], permission)
	}
	return permissions, rows.Err()
}

const findPlayersWithPermissions = `
SELECT DISTINCT uuid
FROM %s
WHERE permission IN (%s)
`

func (s *luckpermsStorage) FindPlayersWithPermissions(ctx context.Context, permissions []string) (uuid.UUIDs, error) {
	mcUUIDs := make(uuid.UUIDs, 0)
	if len(permissions) == 0 {
		return mcUUIDs, nil
	}

	placeholders := make([]string, len(permissions))
	args := make([]any, len(permissions))
	for i, permission := range permissions {
		placeholders[i] = "?"
		args[i] = permission
	}
	query := fmt.Sprintf(findPlayersWithPermissions, config.GetLuckpermsUserPermissionsTableName(), strings.Join(placeholders, ", "))
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var mcUUID uuid.UUID
		if err := rows.Scan(&mcUUID); err != nil {
			return nil, err
		}
		mcUUIDs = append(mcUUIDs, mcUUID)
	}
	return mcUUIDs, rows.Err()
}

const deletePermissionsWithPrefix = `
DELETE FROM %s
WHERE uuid = ? AND permission LIKE ?
`

const insertPermission = `
INSERT INTO %s (uuid, permission, value, server, world, expiry, contexts)
VALUES (?, ?, 1, 'global', 'global', 0, '{}')
`

func (s *luckpermsStorage) ReplacePermissionsWithPrefix(ctx context.Context, mcUUID uuid.UUID, nodePrefix string, node string) error {
	tableName := config.GetLuckpermsUserPermissionsTableName()
	return beginTx(ctx, s.db, func(tx *stdsql.Tx) error {
		if _, err := tx.ExecContext(ctx, fmt.Sprintf(deletePermissionsWithPrefix, tableName), mcUUID.String(), nodePrefix+"%"); err != nil {
			return err
		}
		_, err := tx.ExecContext(ctx, fmt.Sprintf(insertPermission, tableName), mcUUID.String(), node)
		return err
	})
}
