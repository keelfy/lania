package domain

import "github.com/google/uuid"

type PurchasedProduct struct {
	ProductID uuid.UUID `json:"productId"`
	ProfileID uuid.UUID `json:"profileId"`
	SeasonID  uuid.UUID `json:"seasonId"`
}
