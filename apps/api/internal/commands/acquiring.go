package commands

import (
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"github.com/google/uuid"
)

type FreekassaResultCommand struct {
	MerchantID  int64
	Amount      int64
	Currency    string
	OrderID     uuid.UUID
	Signature   string
	StatusCheck bool
}

func (c *FreekassaResultCommand) Validate() error {
	return validation.ValidateStruct(c,
		validation.Field(&c.Amount, validation.Required),
		validation.Field(&c.OrderID, validation.Required, is.UUID),
		validation.Field(&c.Currency, validation.Required),
		validation.Field(&c.Signature, validation.Required),
	)
}

type EasyDonateResultCommand struct {
	PaymentID int64
	Cost      int64
	Customer  string
	Signature string
}

func (c *EasyDonateResultCommand) Validate() error {
	return validation.ValidateStruct(c,
		validation.Field(&c.PaymentID, validation.Required),
		validation.Field(&c.Cost, validation.Required),
		validation.Field(&c.Customer, validation.Required),
		validation.Field(&c.Signature, validation.Required),
	)
}
