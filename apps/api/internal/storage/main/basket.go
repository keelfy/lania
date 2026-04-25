package sql

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/domain"
)

const insertBasketItem = `
INSERT INTO basket_items (
	user_id,
	product_id,
	profile_id,
	quantity,
	created_at,
	created_by,
	updated_at,
	updated_by
) VALUES (
	?,
	?,
	?,
	?,
	now(),
	?,
	now(),
	?
)
`

type InsertBasketItemParams struct {
	UserID    uuid.UUID
	ProductID uuid.UUID
	ProfileID uuid.UUID
	Quantity  int
	CreatedBy uuid.UUID
}

func (q *queries) InsertBasketItem(ctx context.Context, arg InsertBasketItemParams) error {
	_, err := q.x.ExecContext(ctx, insertBasketItem,
		arg.UserID,
		arg.ProductID,
		arg.ProfileID,
		arg.Quantity,
		arg.CreatedBy,
		arg.CreatedBy,
	)
	return err
}

const findBasketItemsByUserID = `
SELECT 
	id,
	user_id,
	product_id,
	profile_id,
	quantity,
	created_at,
	created_by,
	updated_at,
	updated_by
FROM basket_items 
WHERE user_id = ?
`

func (q *queries) FindBasketItemsByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.BasketItem, error) {
	rows, err := q.x.QueryContext(ctx, findBasketItemsByUserID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	basketItems := make([]*domain.BasketItem, 0)
	for rows.Next() {
		var basketItem domain.BasketItem
		err := rows.Scan(
			&basketItem.ID,
			&basketItem.UserID,
			&basketItem.ProductID,
			&basketItem.ProfileID,
			&basketItem.Quantity,
			&basketItem.CreatedAt,
			&basketItem.CreatedBy,
			&basketItem.UpdatedAt,
			&basketItem.UpdatedBy,
		)
		if err != nil {
			return nil, err
		}
		basketItems = append(basketItems, &basketItem)
	}
	return basketItems, nil
}

const clearBasketItemsByUserID = `
DELETE FROM basket_items 
WHERE user_id = ?
`

func (q *queries) ClearBasketItemsByUserID(ctx context.Context, userID uuid.UUID) error {
	_, err := q.x.ExecContext(ctx, clearBasketItemsByUserID, userID)
	return err
}

const deleteBasketItemByIDs = `
DELETE FROM basket_items 
WHERE id IN ('%s')
`

func (q *queries) DeleteBasketItemByIDs(ctx context.Context, ids []uuid.UUID) error {
	idsStr := make([]string, len(ids))
	for i, id := range ids {
		idsStr[i] = id.String()
	}
	query := fmt.Sprintf(deleteBasketItemByIDs, strings.Join(idsStr, "','"))
	_, err := q.x.ExecContext(ctx, query)
	return err
}
