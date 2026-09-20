package requests

import "github.com/google/uuid"

type CreateOrder struct {
	PaymentMethod string          `json:"paymentMethod"`
	Products      []*OrderProduct `json:"products"`
}

type OrderProduct struct {
	ID        uuid.UUID `json:"id"`
	ProfileID uuid.UUID `json:"profileId"`
	// SeasonID defaults to the primary season.
	SeasonID *uuid.UUID `json:"seasonId"`
}
