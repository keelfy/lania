package commands

import (
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"github.com/google/uuid"
)

type AddBasketItemCommand struct {
	UserID    uuid.UUID
	ProductID uuid.UUID
	ProfileID uuid.UUID
	Quantity  int
}

func (c *AddBasketItemCommand) Validate() error {
	return validation.ValidateStruct(c,
		validation.Field(&c.UserID, validation.Required, is.UUID),
		validation.Field(&c.ProductID, validation.Required, is.UUID),
		validation.Field(&c.ProfileID, validation.Required, is.UUID),
		validation.Field(&c.Quantity, validation.Required, validation.Max(1)),
	)
}
