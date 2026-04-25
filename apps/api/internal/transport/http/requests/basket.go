package requests

import "github.com/google/uuid"

type AddBasketItem struct {
	ProductID uuid.UUID `json:"productId"`
	ProfileID uuid.UUID `json:"profileId"`
}
