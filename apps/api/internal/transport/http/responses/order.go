package responses

import (
	"time"

	"github.com/google/uuid"
)

type CreateOrder struct {
	PaymentURL string `json:"paymentUrl"`
}

type OrderItem struct {
	ID        uuid.UUID       `json:"id"`
	ProductID uuid.UUID       `json:"productId"`
	ProfileID uuid.UUID       `json:"profileId"`
	SeasonID  uuid.UUID       `json:"seasonId"`
	Amounts   []*OrderAmounts `json:"amounts"`
	Quantity  int             `json:"quantity"`
}

type OrderAmounts struct {
	Currency string  `json:"currency"`
	Amount   float64 `json:"amount"`
}

type Order struct {
	ID        uuid.UUID       `json:"id"`
	Status    string          `json:"status"`
	Amounts   []*OrderAmounts `json:"amounts"`
	Items     []*OrderItem    `json:"items"`
	CreatedAt time.Time       `json:"createdAt"`
}

type AdminOrderItem struct {
	ID          uuid.UUID       `json:"id"`
	ProductID   uuid.UUID       `json:"productId"`
	ProductName string          `json:"productName"`
	ProfileID   uuid.UUID       `json:"profileId"`
	Username    string          `json:"username"`
	SeasonID    uuid.UUID       `json:"seasonId"`
	SeasonName  string          `json:"seasonName"`
	Amounts     []*OrderAmounts `json:"amounts"`
	Quantity    int             `json:"quantity"`
}

type AdminOrder struct {
	ID      uuid.UUID       `json:"id"`
	UserID  uuid.UUID       `json:"userId"`
	Status  string          `json:"status"`
	Amounts []*OrderAmounts `json:"amounts"`
	// ExternalID is the EasyDonate payment ID, empty until the payment is created.
	ExternalID string            `json:"externalId,omitempty"`
	CreatedAt  int64             `json:"createdAt"`
	UpdatedAt  int64             `json:"updatedAt"`
	Items      []*AdminOrderItem `json:"items"`
}

// AdminOrders is a page of orders with the number of orders in each status, whatever the filter.
type AdminOrders struct {
	Paginated[*AdminOrder]
	StatusCounts map[string]int64 `json:"statusCounts"`
}
