package commands

import (
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
)

type CreateUserCommand struct {
	IdentityID string
	Email      string
	Username   string
	AvatarURL  string
}

func (c *CreateUserCommand) Validate() error {
	return validation.ValidateStruct(c,
		validation.Field(&c.IdentityID, validation.Required, is.UUID),
		validation.Field(&c.Email, validation.Required, is.Email),
	)
}
