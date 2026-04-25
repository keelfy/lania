package sql

import (
	"context"
	"fmt"
	"strings"

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
