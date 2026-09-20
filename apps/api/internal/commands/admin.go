package commands

import (
	"errors"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/domain"
)

// notNilUUID rejects the zero UUID. validation.Required does not, because it reads a UUID as a non-empty string.
var notNilUUID = validation.By(func(value any) error {
	if id, ok := value.(uuid.UUID); ok && id == uuid.Nil {
		return errors.New("cannot be blank")
	}
	return nil
})

type TransferProfileOwnerCommand struct {
	ProfileID uuid.UUID
	Email     string
}

func (c *TransferProfileOwnerCommand) Validate() error {
	return validation.ValidateStruct(c,
		validation.Field(&c.ProfileID, notNilUUID),
		validation.Field(&c.Email, validation.Required, is.Email),
	)
}

type GrantProductCommand struct {
	ProfileID uuid.UUID
	ProductID uuid.UUID
	SeasonID  uuid.UUID
}

func (c *GrantProductCommand) Validate() error {
	return validation.ValidateStruct(c,
		validation.Field(&c.ProfileID, notNilUUID),
		validation.Field(&c.ProductID, notNilUUID),
		validation.Field(&c.SeasonID, notNilUUID),
	)
}

type RevokeGrantCommand struct {
	ProfileID uuid.UUID
	GrantType domain.GrantType
	GrantID   uuid.UUID
}

func (c *RevokeGrantCommand) Validate() error {
	return validation.ValidateStruct(c,
		validation.Field(&c.ProfileID, notNilUUID),
		validation.Field(&c.GrantType, validation.Required, validation.By(func(value any) error {
			if !value.(domain.GrantType).IsValid() {
				return errors.New("must be access, name-color or name-prefix")
			}
			return nil
		})),
		validation.Field(&c.GrantID, notNilUUID),
	)
}
