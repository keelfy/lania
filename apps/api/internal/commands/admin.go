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

type SetProfileRoleCommand struct {
	ProfileID uuid.UUID
	Role      domain.Role
}

func (c *SetProfileRoleCommand) Validate() error {
	return validation.ValidateStruct(c,
		validation.Field(&c.ProfileID, notNilUUID),
		validation.Field(&c.Role, validation.Required, validation.By(func(value any) error {
			if !value.(domain.Role).Valid() {
				return errors.New("must be owner, admin, mod or player")
			}
			return nil
		})),
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

type GrantCosmeticCommand struct {
	ProfileID  uuid.UUID
	Type       domain.GrantType
	ItemID     uuid.UUID
	PrefixType domain.ProfilePrefixType
	SeasonID   *uuid.UUID
}

func (c *GrantCosmeticCommand) Validate() error {
	return validation.ValidateStruct(c,
		validation.Field(&c.ProfileID, notNilUUID),
		validation.Field(&c.Type, validation.Required, validation.By(func(value any) error {
			if grantType := value.(domain.GrantType); grantType != domain.GrantTypeNameColor && grantType != domain.GrantTypeNamePrefix {
				return errors.New("must be name-color or name-prefix")
			}
			return nil
		})),
		validation.Field(&c.ItemID, notNilUUID),
		validation.Field(&c.PrefixType, validation.By(func(value any) error {
			prefixType := value.(domain.ProfilePrefixType)
			if c.Type != domain.GrantTypeNamePrefix {
				return nil
			}
			if prefixType != domain.ProfilePrefixTypeGlyth && prefixType != domain.ProfilePrefixTypeSpecial {
				return errors.New("must be glyth or special for a name prefix")
			}
			return nil
		})),
		validation.Field(&c.SeasonID, validation.By(func(value any) error {
			if seasonID, ok := value.(*uuid.UUID); ok && seasonID != nil && *seasonID == uuid.Nil {
				return errors.New("cannot be the zero UUID")
			}
			return nil
		})),
	)
}
