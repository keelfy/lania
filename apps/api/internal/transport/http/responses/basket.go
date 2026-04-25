package responses

import "github.com/google/uuid"

type BasketItem struct {
	ID        uuid.UUID `json:"id"`
	ProductID uuid.UUID `json:"productId"`
	ProfileID uuid.UUID `json:"profileId"`
	Quantity  int       `json:"quantity"`
}
