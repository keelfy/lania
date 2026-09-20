package requests

import "github.com/google/uuid"

type TransferProfileOwner struct {
	Email string `json:"email"`
}

type GrantProduct struct {
	ProductID uuid.UUID `json:"productId"`
	SeasonID  uuid.UUID `json:"seasonId"`
}
