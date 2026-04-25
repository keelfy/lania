package commands

import (
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/domain"
	is2 "github.com/lania-smp/backend/internal/utils/is"
)

type CreateOrderCommand struct {
	PaymentMethod domain.PaymentMethod
	UserID        uuid.UUID
	Items         []*OrderItemCommand
}

func (c *CreateOrderCommand) Validate() error {
	for _, item := range c.Items {
		if err := item.Validate(); err != nil {
			return err
		}
	}

	return validation.ValidateStruct(c,
		validation.Field(&c.PaymentMethod, validation.Required, is2.PaymentMethod),
		validation.Field(&c.UserID, validation.Required, is.UUID),
		validation.Field(&c.Items, validation.Required),
	)
}

type OrderItemCommand struct {
	ProductID uuid.UUID
	ProfileID uuid.UUID
	SeasonID  uuid.UUID
	Quantity  int
}

func (c *OrderItemCommand) Validate() error {
	return validation.ValidateStruct(c,
		validation.Field(&c.ProductID, validation.Required, is.UUID),
		validation.Field(&c.ProfileID, validation.Required, is.UUID),
		validation.Field(&c.SeasonID, validation.Required, is.UUID),
		validation.Field(&c.Quantity, validation.Required, validation.Min(1)),
	)
}
