package sql

import (
	"context"
	stdsql "database/sql"
	"encoding/json"
	"fmt"
	"strings"

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

// orderWhereClause filters by status and matches the search by prefix against the order ID,
// the EasyDonate payment ID and the usernames of the order recipients.
func orderWhereClause(filter domain.OrderFilter) (string, []any) {
	conditions := make([]string, 0, 2)
	args := make([]any, 0)
	if filter.Status != "" {
		conditions = append(conditions, "o.status = ?")
		args = append(args, string(filter.Status))
	}
	if filter.Search != "" {
		pattern := profileSearchEscaper.Replace(filter.Search) + "%"
		conditions = append(conditions, `(
	CAST(o.id AS CHAR) LIKE ?
	OR o.external_id LIKE ?
	OR EXISTS (
		SELECT 1 FROM order_items oi
		JOIN profiles p ON p.id = oi.profile_id
		WHERE oi.order_id = o.id AND p.mc_username LIKE ?
	)
)`)
		args = append(args, pattern, pattern, pattern)
	}
	if len(conditions) == 0 {
		return "", nil
	}
	return "WHERE " + strings.Join(conditions, " AND "), args
}

const findAdminOrders = `
SELECT
	o.id,
	o.user_id,
	o.amounts,
	o.status,
	o.external_id,
	o.created_at,
	o.created_by,
	o.updated_at,
	o.updated_by
FROM orders o
%s
ORDER BY o.created_at DESC, o.id
LIMIT ? OFFSET ?
`

// FindAdminOrders returns a page of orders of every user, newest first.
func (q *queries) FindAdminOrders(ctx context.Context, filter domain.OrderFilter, size, from int) ([]*domain.Order, error) {
	where, args := orderWhereClause(filter)
	rows, err := q.x.QueryContext(ctx, fmt.Sprintf(findAdminOrders, where), append(args, size, from)...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders := make([]*domain.Order, 0)
	for rows.Next() {
		var order domain.Order
		var amounts json.RawMessage
		var externalID stdsql.NullString
		err := rows.Scan(
			&order.ID,
			&order.UserID,
			&amounts,
			&order.Status,
			&externalID,
			&order.CreatedAt,
			&order.CreatedBy,
			&order.UpdatedAt,
			&order.UpdatedBy,
		)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(amounts, &order.Amounts); err != nil {
			return nil, err
		}
		order.ExternalID = externalID.String
		orders = append(orders, &order)
	}
	return orders, rows.Err()
}

const countAdminOrders = `
SELECT COUNT(o.id) FROM orders o
%s
`

func (q *queries) CountAdminOrders(ctx context.Context, filter domain.OrderFilter) (int64, error) {
	where, args := orderWhereClause(filter)
	var count int64
	err := q.x.QueryRowContext(ctx, fmt.Sprintf(countAdminOrders, where), args...).Scan(&count)
	return count, err
}

const countOrdersByStatus = `
SELECT status, COUNT(id) FROM orders GROUP BY status
`

// CountOrdersByStatus counts the orders of every user in each status.
func (q *queries) CountOrdersByStatus(ctx context.Context) (map[domain.OrderStatus]int64, error) {
	rows, err := q.x.QueryContext(ctx, countOrdersByStatus)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	counts := make(map[domain.OrderStatus]int64)
	for rows.Next() {
		var status domain.OrderStatus
		var count int64
		if err := rows.Scan(&status, &count); err != nil {
			return nil, err
		}
		counts[status] = count
	}
	return counts, rows.Err()
}

// The product name falls back to any localization when the product has none in the requested locale.
const findAdminOrderItems = `
SELECT
	oi.id,
	oi.order_id,
	oi.product_id,
	oi.profile_id,
	oi.season_id,
	oi.amounts,
	oi.quantity,
	p.mc_username,
	s.name,
	(
		SELECT pl.name FROM product_localizations pl
		WHERE pl.product_id = oi.product_id
		ORDER BY pl.locale = ? DESC, pl.locale
		LIMIT 1
	)
FROM order_items oi
LEFT JOIN profiles p ON p.id = oi.profile_id
LEFT JOIN seasons s ON s.id = oi.season_id
WHERE oi.order_id IN (%s)
ORDER BY oi.order_id, oi.id
`

// FindAdminOrderItems returns the items of the orders with the recipient username, the season name
// and the product name in the locale.
func (q *queries) FindAdminOrderItems(ctx context.Context, orderIDs uuid.UUIDs, locale string) ([]*domain.OrderItem, error) {
	if len(orderIDs) == 0 {
		return []*domain.OrderItem{}, nil
	}
	placeholders := make([]string, len(orderIDs))
	args := make([]any, 0, len(orderIDs)+1)
	args = append(args, locale)
	for i, id := range orderIDs {
		placeholders[i] = "?"
		args = append(args, id)
	}

	rows, err := q.x.QueryContext(ctx, fmt.Sprintf(findAdminOrderItems, strings.Join(placeholders, ", ")), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]*domain.OrderItem, 0)
	for rows.Next() {
		var item domain.OrderItem
		var amounts json.RawMessage
		var username, seasonName, productName stdsql.NullString
		err := rows.Scan(
			&item.ID,
			&item.OrderID,
			&item.ProductID,
			&item.ProfileID,
			&item.SeasonID,
			&amounts,
			&item.Quantity,
			&username,
			&seasonName,
			&productName,
		)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(amounts, &item.Amounts); err != nil {
			return nil, err
		}
		item.Profile = &domain.Profile{ID: item.ProfileID, MinecraftUsername: username.String}
		item.Season = &domain.Season{ID: item.SeasonID, Name: seasonName.String}
		item.Product = &domain.Product{ID: item.ProductID}
		if productName.Valid {
			item.Product.Localizations = []*domain.ProductLocalization{{ProductID: item.ProductID, Locale: locale, Name: productName.String}}
		}
		items = append(items, &item)
	}
	return items, rows.Err()
}
