package sql

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/domain"
)

const insertOrder = `
INSERT INTO orders (
	id,
	user_id,
	amounts,
	status,
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
)`

type InsertOrderParams struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Amounts   []*domain.OrderAmounts
	Status    domain.OrderStatus
	CreatedBy *uuid.UUID
	UpdatedBy *uuid.UUID
}

func (q *queries) InsertOrder(ctx context.Context, arg InsertOrderParams) (uuid.UUID, error) {
	amounts, err := json.Marshal(arg.Amounts)
	if err != nil {
		return uuid.Nil, err
	}

	_, err = q.x.ExecContext(ctx, insertOrder,
		arg.ID,
		arg.UserID,
		amounts,
		string(arg.Status),
		arg.CreatedBy,
		arg.UpdatedBy,
	)
	if err != nil {
		return uuid.Nil, err
	}
	return arg.ID, nil
}

const insertOrderItem = `
INSERT INTO order_items (
	order_id,
	product_id,
	profile_id,
	season_id,
	amounts,
	quantity
) VALUES (
	?,
	?,
	?,
	?,
	?,
	?
)
`

type InsertOrderItemParams struct {
	OrderID   uuid.UUID
	ProductID uuid.UUID
	ProfileID uuid.UUID
	SeasonID  uuid.UUID
	Amounts   []*domain.OrderAmounts
	Quantity  int
}

func (q *queries) InsertOrderItem(ctx context.Context, arg InsertOrderItemParams) error {
	amounts, err := json.Marshal(arg.Amounts)
	if err != nil {
		return err
	}

	_, err = q.x.ExecContext(ctx, insertOrderItem,
		arg.OrderID,
		arg.ProductID,
		arg.ProfileID,
		arg.SeasonID,
		amounts,
		arg.Quantity,
	)
	return err
}

const findOrderByID = `
SELECT 
	id,
	user_id,
	amounts,
	status,
	created_at,
	created_by,
	updated_at,
	updated_by
FROM orders 
WHERE id = ?
`

func (q *queries) FindOrderByID(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
	row := q.x.QueryRowContext(ctx, findOrderByID, id)
	var order domain.Order
	var amounts json.RawMessage
	err := row.Scan(
		&order.ID,
		&order.UserID,
		&amounts,
		&order.Status,
		&order.CreatedAt,
		&order.CreatedBy,
		&order.UpdatedAt,
		&order.UpdatedBy,
	)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(amounts, &order.Amounts)
	if err != nil {
		return nil, err
	}
	return &order, nil
}

const updateOrderStatusByID = `
UPDATE orders 
SET status = ?, 
	updated_at = now(), 
	updated_by = ? 
WHERE id = ?
`

func (q *queries) UpdateOrderStatusByID(ctx context.Context, id uuid.UUID, status domain.OrderStatus, updatedBy *uuid.UUID) error {
	_, err := q.x.ExecContext(ctx, updateOrderStatusByID, status, updatedBy, id)
	if err != nil {
		return err
	}
	return err
}

const findItemsByOrderID = `
SELECT 
	id,
	order_id,
	product_id,
	profile_id,
	season_id,
	amounts,
	quantity
FROM order_items 
WHERE order_id = ?
`

func (q *queries) FindItemsByOrderID(ctx context.Context, orderID uuid.UUID) ([]*domain.OrderItem, error) {
	rows, err := q.x.QueryContext(ctx, findItemsByOrderID, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]*domain.OrderItem, 0)
	for rows.Next() {
		var item domain.OrderItem
		var amounts json.RawMessage
		err := rows.Scan(
			&item.ID,
			&item.OrderID,
			&item.ProductID,
			&item.ProfileID,
			&item.SeasonID,
			&amounts,
			&item.Quantity,
		)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(amounts, &item.Amounts)
		if err != nil {
			return nil, err
		}
		items = append(items, &item)
	}
	return items, nil
}

const findOrdersByUserID = `
SELECT 
	id,
	user_id,
	amounts,
	status,
	created_at,
	created_by,
	updated_at,
	updated_by
FROM orders 
WHERE user_id = ?
ORDER BY created_at DESC
`

func (q *queries) FindOrdersByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Order, error) {
	rows, err := q.x.QueryContext(ctx, findOrdersByUserID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders := make([]*domain.Order, 0)
	for rows.Next() {
		var order domain.Order
		var amounts json.RawMessage
		err := rows.Scan(
			&order.ID,
			&order.UserID,
			&amounts,
			&order.Status,
			&order.CreatedAt,
			&order.CreatedBy,
			&order.UpdatedAt,
			&order.UpdatedBy,
		)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(amounts, &order.Amounts)
		if err != nil {
			return nil, err
		}
		orders = append(orders, &order)
	}
	return orders, nil
}

const updateOrderExternalIDByID = `
UPDATE orders 
SET external_id = ? 
WHERE id = ?
`

func (q *queries) UpdateOrderExternalIDByID(ctx context.Context, id uuid.UUID, externalID string) error {
	_, err := q.x.ExecContext(ctx, updateOrderExternalIDByID, externalID, id)
	return err
}

const findOrderByExternalID = `
SELECT 
	id,
	user_id,
	amounts,
	status,
	created_at,
	created_by,
	updated_at,
	updated_by,
	external_id
FROM orders 
WHERE external_id = ?
`

func (q *queries) FindOrderByExternalID(ctx context.Context, externalID string) (*domain.Order, error) {
	row := q.x.QueryRowContext(ctx, findOrderByExternalID, externalID)
	var order domain.Order
	var amounts json.RawMessage
	err := row.Scan(
		&order.ID,
		&order.UserID,
		&amounts,
		&order.Status,
		&order.CreatedAt,
		&order.CreatedBy,
		&order.UpdatedAt,
		&order.UpdatedBy,
		&order.ExternalID,
	)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(amounts, &order.Amounts)
	if err != nil {
		return nil, err
	}
	return &order, nil
}
