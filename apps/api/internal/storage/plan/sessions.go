package plan

import (
	"context"
	stdsql "database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"
	plandomain "github.com/lania-smp/backend/internal/domain/plan"
)

const findPLANUserByMinecraftUUID = `
SELECT id, uuid
FROM plan_users
WHERE uuid IN ('%s')
`

func (q *queries) FindPLANUserByMinecraftUUID(ctx context.Context, mcUUIDs uuid.UUIDs) (map[uuid.UUID]*plandomain.User, error) {
	mcUUIDsStr := make([]string, len(mcUUIDs))
	for i, mcUUID := range mcUUIDs {
		mcUUIDsStr[i] = mcUUID.String()
	}
	query := fmt.Sprintf(findPLANUserByMinecraftUUID, strings.Join(mcUUIDsStr, "','"))
	rows, err := q.x.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	planUsers := make(map[uuid.UUID]*plandomain.User)
	for rows.Next() {
		var user plandomain.User
		err := rows.Scan(&user.ID, &user.UUID)
		if err != nil {
			return nil, err
		}
		planUsers[user.UUID] = &user
	}
	return planUsers, nil
}

const findPlaytimeByUserIDs = `
SELECT 
	s.user_id,
	SUM(s.session_end - s.session_start - s.afk_time) AS total_playtime,
	MIN(s.session_start) AS first_session_start,
	MAX(s.session_end) AS last_session_end
FROM plan_sessions s
WHERE s.user_id IN ('%s')
GROUP BY s.user_id
`

func (q *queries) FindPlaytimeByUserIDs(ctx context.Context, userIDs []int64) (map[int64]*plandomain.Playtime, error) {
	playtimes := make(map[int64]*plandomain.Playtime)
	userIDsStr := make([]string, len(userIDs))
	for i, userID := range userIDs {
		userIDsStr[i] = strconv.FormatInt(userID, 10)
	}
	query := fmt.Sprintf(findPlaytimeByUserIDs, strings.Join(userIDsStr, "','"))
	rows, err := q.x.QueryContext(ctx, query)
	if err != nil && err != stdsql.ErrNoRows {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var playtime plandomain.Playtime
		var userID int64
		err := rows.Scan(&userID, &playtime.TotalPlaytime, &playtime.FirstSessionStart, &playtime.LastSessionEnd)
		if err == stdsql.ErrNoRows {
			playtimes[userID] = &playtime
			continue
		} else if err != nil {
			return nil, err
		}
		playtimes[userID] = &playtime
	}
	return playtimes, nil
}
