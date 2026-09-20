package sql

import (
	"context"
	stdsql "database/sql"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/domain"
)

// Every branch of the union returns the same columns. Revoked grants are included on purpose.
const findProfileGrants = `
SELECT id, type, season_id, item_id, prefix_type, name, source, order_item_id, granted_by, created_at, revoked_at, revoked_by
FROM (
	SELECT
		pa.id AS id,
		'access' AS type,
		pa.season_id AS season_id,
		NULL AS item_id,
		NULL AS prefix_type,
		NULL AS name,
		pa.source AS source,
		NULL AS order_item_id,
		pa.updated_by AS granted_by,
		pa.created_at AS created_at,
		pa.revoked_at AS revoked_at,
		pa.revoked_by AS revoked_by
	FROM profile_accesses pa
	WHERE pa.mc_uuid = ?

	UNION ALL

	SELECT
		pnc.id,
		'name-color',
		pnc.for_season_id,
		pnc.name_color_id,
		NULL,
		nc.name,
		NULL,
		pnc.order_item_id,
		pnc.created_by,
		pnc.created_at,
		pnc.revoked_at,
		pnc.revoked_by
	FROM profile_name_color_options pnc
	LEFT JOIN name_colors nc ON nc.id = pnc.name_color_id
	WHERE pnc.profile_id = ?

	UNION ALL

	SELECT
		pnp.id,
		'name-prefix',
		pnp.for_season_id,
		pnp.name_prefix_id,
		pnp.type,
		np.name,
		NULL,
		pnp.order_item_id,
		pnp.created_by,
		pnp.created_at,
		pnp.revoked_at,
		pnp.revoked_by
	FROM profile_name_prefix_options pnp
	LEFT JOIN name_prefixes np ON np.id = pnp.name_prefix_id
	WHERE pnp.profile_id = ?
) grants
ORDER BY created_at DESC, id
`

// FindProfileGrants returns everything the profile was given, revoked grants included, newest first.
func (q *queries) FindProfileGrants(ctx context.Context, profileID, mcUUID uuid.UUID) ([]*domain.Grant, error) {
	rows, err := q.x.QueryContext(ctx, findProfileGrants, mcUUID, profileID, profileID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	grants := make([]*domain.Grant, 0)
	for rows.Next() {
		grant := &domain.Grant{ProfileID: profileID}
		var itemID *uuid.UUID
		var prefixType, name, source stdsql.NullString
		err := rows.Scan(
			&grant.ID,
			&grant.Type,
			&grant.SeasonID,
			&itemID,
			&prefixType,
			&name,
			&source,
			&grant.OrderItemID,
			&grant.GrantedBy,
			&grant.CreatedAt,
			&grant.RevokedAt,
			&grant.RevokedBy,
		)
		if err != nil {
			return nil, err
		}

		if itemID != nil {
			grant.ItemID = *itemID
		}
		grant.PrefixType = domain.ProfilePrefixType(prefixType.String)
		grant.Name = name.String
		grant.Source = domain.AccessSource(source.String)
		grants = append(grants, grant)
	}
	return grants, rows.Err()
}

const revokeProfileAccess = `
UPDATE profile_accesses
SET revoked_at = NOW(), revoked_by = ?
WHERE id = ? AND mc_uuid = ? AND revoked_at IS NULL
`

func (q *queries) RevokeProfileAccess(ctx context.Context, mcUUID, accessID uuid.UUID, revokedBy *uuid.UUID) error {
	_, err := q.x.ExecContext(ctx, revokeProfileAccess, revokedBy, accessID, mcUUID)
	return err
}

const revokeProfileNameColorOption = `
UPDATE profile_name_color_options
SET revoked_at = NOW(), revoked_by = ?
WHERE id = ? AND profile_id = ? AND revoked_at IS NULL
`

func (q *queries) RevokeProfileNameColorOption(ctx context.Context, profileID, optionID uuid.UUID, revokedBy *uuid.UUID) error {
	_, err := q.x.ExecContext(ctx, revokeProfileNameColorOption, revokedBy, optionID, profileID)
	return err
}

const revokeProfileNamePrefixOption = `
UPDATE profile_name_prefix_options
SET revoked_at = NOW(), revoked_by = ?
WHERE id = ? AND profile_id = ? AND revoked_at IS NULL
`

func (q *queries) RevokeProfileNamePrefixOption(ctx context.Context, profileID, optionID uuid.UUID, revokedBy *uuid.UUID) error {
	_, err := q.x.ExecContext(ctx, revokeProfileNamePrefixOption, revokedBy, optionID, profileID)
	return err
}
