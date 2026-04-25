package flectonesql

import (
	"context"
	stdsql "database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

const findOnlineByMinecraftUUIDs = `
SELECT uuid, online
FROM player
WHERE uuid IN ('%s')
`

func (q *queries) FindOnlineByMinecraftUUIDs(ctx context.Context, mcUUIDs uuid.UUIDs) (map[uuid.UUID]bool, error) {
	onlineMap := make(map[uuid.UUID]bool)
	mcUUIDsStr := make([]string, len(mcUUIDs))
	for i, mcUUID := range mcUUIDs {
		mcUUIDsStr[i] = mcUUID.String()
	}
	query := fmt.Sprintf(findOnlineByMinecraftUUIDs, strings.Join(mcUUIDsStr, "','"))
	rows, err := q.x.QueryContext(ctx, query)
	if err != nil && err != stdsql.ErrNoRows {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var online bool
		var minecraftUUID uuid.UUID
		err := rows.Scan(&minecraftUUID, &online)
		if err == stdsql.ErrNoRows {
			onlineMap[minecraftUUID] = false
			continue
		} else if err != nil {
			return nil, err
		}
		onlineMap[minecraftUUID] = online
	}
	return onlineMap, nil
}
