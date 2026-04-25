package sql

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/domain"
)

const insertProfileAccess = `
INSERT INTO profile_accesses (
	mc_uuid,
	season_id,
	source,
	order_item_id,
	created_at,
	updated_at,
	updated_by
) VALUES (
	?,
	?,
	?,
	?,
	now(),
	now(),
	?
) ON DUPLICATE KEY UPDATE mc_uuid = mc_uuid;
`

type InsertProfileAccessParams struct {
	MinecraftUUID uuid.UUID
	SeasonID      uuid.UUID
	Source        string
	OrderItemID   *uuid.UUID
	UpdatedBy     *uuid.UUID
}

func (q *queries) InsertProfileAccess(ctx context.Context, arg InsertProfileAccessParams) error {
	_, err := q.x.ExecContext(ctx, insertProfileAccess,
		arg.MinecraftUUID,
		arg.SeasonID,
		arg.Source,
		arg.OrderItemID,
		arg.UpdatedBy,
	)
	return err
}

const checkIfProfileHasAccessBySeasonIDAndMinecraftUUID = `
SELECT EXISTS(
	SELECT 1
	FROM profile_accesses
	WHERE mc_uuid = ? AND season_id = ?
)
`

func (q *queries) CheckIfProfileHasAccessBySeasonIDAndMinecraftUUID(ctx context.Context, mcUUID uuid.UUID, seasonID uuid.UUID) (bool, error) {
	var exists bool
	err := q.x.QueryRowContext(ctx, checkIfProfileHasAccessBySeasonIDAndMinecraftUUID, mcUUID, seasonID).Scan(&exists)
	return exists, err
}

const getProfileAccessesBySeasonIDAndOwnerUserID = `
	SELECT profiles.id
	FROM profile_accesses
	LEFT JOIN profiles ON profile_accesses.mc_uuid = profiles.mc_uuid
	WHERE season_id = ? AND owner_user_id = ?
`

func (q *queries) GetProfileAccessesBySeasonIDAndOwnerUserID(ctx context.Context, seasonID uuid.UUID, ownerUserID uuid.UUID) (uuid.UUIDs, error) {
	rows, err := q.x.QueryContext(ctx, getProfileAccessesBySeasonIDAndOwnerUserID, seasonID, ownerUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	profileIDs := make(uuid.UUIDs, 0)
	for rows.Next() {
		var access uuid.UUID
		err := rows.Scan(&access)
		if err != nil {
			return nil, err
		}
		profileIDs = append(profileIDs, access)
	}
	return profileIDs, err
}

const findProfileAccessesByMinecraftUUIDs = `
SELECT 
	pa.mc_uuid,
	pa.season_id,
	pa.source,
	pa.created_at,
	pa.updated_at,
	pa.updated_by
FROM profile_accesses pa
WHERE mc_uuid IN ('%s')
`

func (q *queries) FindProfileAccessesByMinecraftUUIDs(ctx context.Context, mcUUIDs uuid.UUIDs) ([]*domain.ProfileAccess, error) {
	mcUUIDsStr := make([]string, len(mcUUIDs))
	for i, mcUUID := range mcUUIDs {
		mcUUIDsStr[i] = mcUUID.String()
	}
	query := fmt.Sprintf(findProfileAccessesByMinecraftUUIDs, strings.Join(mcUUIDsStr, "','"))
	rows, err := q.x.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	accesses := make([]*domain.ProfileAccess, 0)
	for rows.Next() {
		var access domain.ProfileAccess
		err := rows.Scan(
			&access.MinecraftUUID,
			&access.SeasonID,
			&access.Source,
			&access.CreatedAt,
			&access.UpdatedAt,
			&access.UpdatedBy,
		)
		if err != nil {
			return nil, err
		}
		accesses = append(accesses, &access)
	}
	return accesses, nil
}
