package sql

import (
	"context"
	stdsql "database/sql"
	"strings"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/domain"
)

func (q *queries) InsertPrivilege(ctx context.Context, id uuid.UUID, name string, names domain.CosmeticNames, permission string) error {
	namesPayload, err := marshalNames(names)
	if err != nil {
		return err
	}
	_, err = q.x.ExecContext(ctx, "INSERT INTO privileges (id, name, names, permission) VALUES (?, ?, ?, ?)", id, name, namesPayload, permission)
	return err
}

func (q *queries) UpdatePrivilege(ctx context.Context, id uuid.UUID, name string, names domain.CosmeticNames, permission string) error {
	namesPayload, err := marshalNames(names)
	if err != nil {
		return err
	}
	_, err = q.x.ExecContext(ctx, "UPDATE privileges SET name = ?, names = ?, permission = ? WHERE id = ?", name, namesPayload, permission, id)
	return err
}

const findPrivileges = `
SELECT id, name, names, permission, created_at
FROM privileges
%s
ORDER BY name
`

// The tariffs of privileges are read apart from the privileges: a name built from the id in SQL
// could clash with the collation of product_prices.name.
const findPrivilegePrices = `
SELECT name, currency, amount
FROM product_prices
WHERE name LIKE ?
ORDER BY currency
`

func (q *queries) findPrivileges(ctx context.Context, where string, args ...any) ([]*domain.Privilege, error) {
	rows, err := q.x.QueryContext(ctx, strings.Replace(findPrivileges, "%s", where, 1), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	privileges := make([]*domain.Privilege, 0)
	for rows.Next() {
		var privilege domain.Privilege
		var names []byte
		if err := rows.Scan(&privilege.ID, &privilege.Name, &names, &privilege.Permission, &privilege.CreatedAt); err != nil {
			return nil, err
		}
		if privilege.Names, err = unmarshalNames(names); err != nil {
			return nil, err
		}
		privileges = append(privileges, &privilege)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	rows.Close()

	priceRows, err := q.x.QueryContext(ctx, findPrivilegePrices, domain.PrivilegePriceNamePrefix+"%")
	if err != nil {
		return nil, err
	}
	defer priceRows.Close()

	pricesByName := make(map[domain.ProductPriceName][]*domain.ProductPrice)
	for priceRows.Next() {
		var price domain.ProductPrice
		if err := priceRows.Scan(&price.Name, &price.Currency, &price.Amount); err != nil {
			return nil, err
		}
		pricesByName[price.Name] = append(pricesByName[price.Name], &price)
	}
	if err := priceRows.Err(); err != nil {
		return nil, err
	}
	for _, privilege := range privileges {
		privilege.Prices = pricesByName[domain.PrivilegePriceName(privilege.ID)]
	}
	return privileges, nil
}

func (q *queries) FindPrivileges(ctx context.Context) ([]*domain.Privilege, error) {
	return q.findPrivileges(ctx, "")
}

// FindPrivilegeByID returns stdsql.ErrNoRows when the privilege does not exist.
func (q *queries) FindPrivilegeByID(ctx context.Context, id uuid.UUID) (*domain.Privilege, error) {
	privileges, err := q.findPrivileges(ctx, "WHERE id = ?", id)
	if err != nil {
		return nil, err
	}
	if len(privileges) == 0 {
		return nil, stdsql.ErrNoRows
	}
	return privileges[0], nil
}

// CountPrivilegeOwners returns how many profiles have the privilege and it is not revoked.
func (q *queries) CountPrivilegeOwners(ctx context.Context, id uuid.UUID) (int, error) {
	var count int
	err := q.x.QueryRowContext(ctx, "SELECT COUNT(DISTINCT profile_id) FROM profile_privileges WHERE privilege_id = ? AND revoked_at IS NULL", id).Scan(&count)
	return count, err
}

// DeletePrivilege removes the privilege with its revoked grants and its tariff.
// A grant that is not revoked keeps the foreign key, so the delete fails with 1451.
func (q *queries) DeletePrivilege(ctx context.Context, id uuid.UUID) error {
	if err := q.execAll(ctx, id,
		"DELETE FROM profile_privileges WHERE privilege_id = ? AND revoked_at IS NOT NULL",
		"DELETE FROM privileges WHERE id = ?",
	); err != nil {
		return err
	}
	return q.DeleteProductPrices(ctx, domain.PrivilegePriceName(id))
}

const insertProfilePrivilege = `
INSERT INTO profile_privileges (
	profile_id,
	privilege_id,
	for_season_id,
	order_item_id,
	created_by
) VALUES (?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE revoked_at = NULL, revoked_by = NULL
`

type InsertProfilePrivilegeParams struct {
	ProfileID   uuid.UUID
	PrivilegeID uuid.UUID
	SeasonID    uuid.UUID
	OrderItemID *uuid.UUID
	CreatedBy   *uuid.UUID
}

func (q *queries) InsertProfilePrivilege(ctx context.Context, arg InsertProfilePrivilegeParams) error {
	_, err := q.x.ExecContext(ctx, insertProfilePrivilege, arg.ProfileID, arg.PrivilegeID, arg.SeasonID, arg.OrderItemID, arg.CreatedBy)
	return err
}

const findProfilePrivileges = `
SELECT pp.id, pp.profile_id, pp.privilege_id, pp.for_season_id, p.permission
FROM profile_privileges pp
JOIN privileges p ON p.id = pp.privilege_id
WHERE %s AND pp.for_season_id = ? AND pp.revoked_at IS NULL
ORDER BY pp.created_at DESC
`

func (q *queries) findProfilePrivileges(ctx context.Context, owner string, ownerID, seasonID uuid.UUID) ([]*domain.ProfilePrivilege, error) {
	rows, err := q.x.QueryContext(ctx, strings.Replace(findProfilePrivileges, "%s", owner, 1), ownerID, seasonID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	privileges := make([]*domain.ProfilePrivilege, 0)
	for rows.Next() {
		var privilege domain.ProfilePrivilege
		if err := rows.Scan(&privilege.ID, &privilege.ProfileID, &privilege.PrivilegeID, &privilege.SeasonID, &privilege.Permission); err != nil {
			return nil, err
		}
		privileges = append(privileges, &privilege)
	}
	return privileges, rows.Err()
}

func (q *queries) FindProfilePrivileges(ctx context.Context, profileID, seasonID uuid.UUID) ([]*domain.ProfilePrivilege, error) {
	return q.findProfilePrivileges(ctx, "pp.profile_id = ?", profileID, seasonID)
}

func (q *queries) FindProfilePrivilegesByOwnerUserID(ctx context.Context, ownerUserID, seasonID uuid.UUID) ([]*domain.ProfilePrivilege, error) {
	return q.findProfilePrivileges(ctx, "pp.profile_id IN (SELECT id FROM profiles WHERE owner_user_id = ?)", ownerUserID, seasonID)
}

const revokeProfilePrivilege = `
UPDATE profile_privileges
SET revoked_at = NOW(), revoked_by = ?
WHERE id = ? AND profile_id = ? AND revoked_at IS NULL
`

func (q *queries) RevokeProfilePrivilege(ctx context.Context, profileID, privilegeGrantID uuid.UUID, revokedBy *uuid.UUID) error {
	_, err := q.x.ExecContext(ctx, revokeProfilePrivilege, revokedBy, privilegeGrantID, profileID)
	return err
}

const revokeProfilePrivileges = `
UPDATE profile_privileges
SET revoked_at = NOW(), revoked_by = ?
WHERE profile_id = ? AND revoked_at IS NULL
`

// RevokeProfilePrivileges revokes every privilege the profile still has, in every season.
func (q *queries) RevokeProfilePrivileges(ctx context.Context, profileID uuid.UUID, revokedBy *uuid.UUID) error {
	_, err := q.x.ExecContext(ctx, revokeProfilePrivileges, revokedBy, profileID)
	return err
}
