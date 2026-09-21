package storage

import (
	"context"
	stdsql "database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/lania-smp/shell/internal/config"
)

// PermissionReplacement swaps permission nodes of one player.
type PermissionReplacement struct {
	MinecraftUUID uuid.UUID
	// Remove lists the nodes deleted from the player.
	Remove []string
	// Add lists the nodes the player has after the swap.
	Add []string
}

// LuckpermsStorage reads and writes LuckPerms user permission nodes.
type LuckpermsStorage interface {
	// FindPermissionsWithPrefix returns permission nodes starting with nodePrefix, keyed by player.
	FindPermissionsWithPrefix(ctx context.Context, mcUUIDs uuid.UUIDs, nodePrefix string) (map[uuid.UUID][]string, error)
	// FindPlayersWithPermissions returns players that have at least one of the permission nodes.
	FindPlayersWithPermissions(ctx context.Context, permissions []string) (uuid.UUIDs, error)
	// ReplacePermissions applies every replacement in one transaction. A node in both lists stays once.
	ReplacePermissions(ctx context.Context, replacements []PermissionReplacement) error
}

type luckpermsStorage struct {
	db *stdsql.DB
}

func NewLuckpermsStorage(db *stdsql.DB) LuckpermsStorage {
	return &luckpermsStorage{db: db}
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

const insertPermission = `
INSERT INTO %s (uuid, permission, value, server, world, expiry, contexts)
VALUES (?, ?, 1, 'global', 'global', 0, '{}')
`

const deletePermissions = `
DELETE FROM %s
WHERE uuid = ? AND permission IN (%s)
`

func (s *luckpermsStorage) ReplacePermissions(ctx context.Context, replacements []PermissionReplacement) error {
	tableName := config.GetLuckpermsUserPermissionsTableName()
	return beginTx(ctx, s.db, func(tx *stdsql.Tx) error {
		for _, replacement := range replacements {
			// Nodes to add are deleted too, so applying a replacement twice leaves one row per node.
			remove := append(append([]string{}, replacement.Remove...), replacement.Add...)
			if len(remove) > 0 {
				placeholders := make([]string, len(remove))
				args := make([]any, 0, len(remove)+1)
				args = append(args, replacement.MinecraftUUID.String())
				for i, node := range remove {
					placeholders[i] = "?"
					args = append(args, node)
				}
				query := fmt.Sprintf(deletePermissions, tableName, strings.Join(placeholders, ", "))
				if _, err := tx.ExecContext(ctx, query, args...); err != nil {
					return err
				}
			}
			for _, node := range replacement.Add {
				if _, err := tx.ExecContext(ctx, fmt.Sprintf(insertPermission, tableName), replacement.MinecraftUUID.String(), node); err != nil {
					return err
				}
			}
		}
		return nil
	})
}
