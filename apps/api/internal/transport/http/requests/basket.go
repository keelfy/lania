package requests

import "github.com/google/uuid"

type AddBasketItem struct {
	ProductID uuid.UUID `json:"productId"`
	ProfileID uuid.UUID `json:"profileId"`
	// SeasonID defaults to the primary season.
	SeasonID *uuid.UUID `json:"seasonId"`
}
