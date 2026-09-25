package commands

import (
	validation "github.com/go-ozzo/ozzo-validation"
)

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
