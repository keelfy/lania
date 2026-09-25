package domain

import (
	"time"

	"github.com/google/uuid"
)

type OrderStatus string

const (
	OrderStatusCreated    OrderStatus = "created"
	OrderStatusProcessing OrderStatus = "processing"
	OrderStatusCompleted  OrderStatus = "completed"
	OrderStatusFailed     OrderStatus = "failed"
)

type PaymentMethod string

const (
	PaymentMethodEasyDonate PaymentMethod = "easy-donate"
)

type OrderAmounts struct {
	Currency Currency `json:"c"`
	Amount   float64  `json:"a"`
}

type Order struct {
	ID            uuid.UUID
	UserID        uuid.UUID
	Amounts       []*OrderAmounts
	Status        OrderStatus
	PaymentMethod PaymentMethod
	ExternalID    string
	CreatedAt     time.Time
	CreatedBy     *uuid.UUID
	UpdatedAt     time.Time
	UpdatedBy     *uuid.UUID
	// relations
	Items []*OrderItem
}

type OrderItem struct {
	ID        uuid.UUID
	OrderID   uuid.UUID
	ProductID uuid.UUID
	ProfileID uuid.UUID
	SeasonID  uuid.UUID
	Amounts   []*OrderAmounts
	Quantity  int
	// relations
	Order   *Order
	Product *Product
	Profile *Profile
	Season  *Season
}

// OrderStatuses lists every order status in the order an order goes through them.
var OrderStatuses = []OrderStatus{OrderStatusCreated, OrderStatusProcessing, OrderStatusCompleted, OrderStatusFailed}

// OrderFilter narrows the admin order list. An empty status keeps every status.
// Search matches the order ID, the EasyDonate payment ID or a recipient username, all by prefix.
type OrderFilter struct {
	Status OrderStatus
	Search string
}
