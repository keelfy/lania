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
