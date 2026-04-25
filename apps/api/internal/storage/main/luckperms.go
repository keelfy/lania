package sql

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/config"
	"github.com/lania-smp/backend/internal/domain"
)

const findLuckpermsPermissionLikeByMinecraftUUID = `
SELECT 
	id, 
	uuid, 
	permission, 
	value,
	server,
	world,
	expiry,
	contexts
FROM %s 
WHERE uuid = ? AND permission LIKE ?
`

func (q *queries) FindLuckpermsPermissionLikeByMinecraftUUID(ctx context.Context, minecraftUUID uuid.UUID, like string) ([]*domain.LuckpermsUserPermission, error) {
	like = "%" + like + "%"
	query := fmt.Sprintf(findLuckpermsPermissionLikeByMinecraftUUID, config.GetLuckpermsUserPermissionsTableName())
	rows, err := q.x.QueryContext(ctx, query, minecraftUUID, like)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	permissions := make([]*domain.LuckpermsUserPermission, 0)
	for rows.Next() {
		var permission domain.LuckpermsUserPermission
		err := rows.Scan(
			&permission.ID,
			&permission.UUID,
			&permission.Permission,
			&permission.Value,
			&permission.Server,
			&permission.World,
			&permission.Expiry,
			&permission.Contexts,
		)
		if err != nil {
			return nil, err
		}
		permissions = append(permissions, &permission)
	}
	return permissions, nil
}

const deleteLuckpermsPermissionByIDs = `
DELETE FROM %s 
WHERE id IN ('%s')
`

func (q *queries) DeleteLuckpermsPermissionByIDs(ctx context.Context, ids []int64) error {
	idsStr := make([]string, len(ids))
	for i, id := range ids {
		idsStr[i] = strconv.FormatInt(id, 10)
	}
	query := fmt.Sprintf(deleteLuckpermsPermissionByIDs, config.GetLuckpermsUserPermissionsTableName(), strings.Join(idsStr, "','"))
	_, err := q.x.ExecContext(ctx, query)
	if err != nil {
		return err
	}
	return err
}

const insertLuckpermsPermission = `
INSERT INTO %s (
	uuid,
	permission,
	value,
	server,
	world,
	expiry,
	contexts
) VALUES (
	?,
	?,
	?,
	?,
	?,
	?,
	'{}'
)
`

type InsertLuckpermsPermissionParams struct {
	UUID       uuid.UUID
	Permission string
	Value      domain.LuckpermsPermissionValue
	Server     domain.LuckpermsPermissionServer
	World      domain.LuckpermsPermissionWorld
	Expiry     int64
}

func (q *queries) InsertLuckpermsPermission(ctx context.Context, arg InsertLuckpermsPermissionParams) error {
	query := fmt.Sprintf(insertLuckpermsPermission, config.GetLuckpermsUserPermissionsTableName())
	_, err := q.x.ExecContext(ctx, query,
		arg.UUID,
		arg.Permission,
		arg.Value,
		arg.Server,
		arg.World,
		arg.Expiry,
	)
	return err
}
