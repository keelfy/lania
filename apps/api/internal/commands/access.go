package commands

import (
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/domain"
	is2 "github.com/lania-smp/backend/internal/utils/is"
)

type ObtainAccessByUsernamesCommand struct {
	SeasonID    uuid.UUID
	Source      domain.AccessSource
	OwnerUserID uuid.UUID
	Usernames   []string
}

func (c *ObtainAccessByUsernamesCommand) Validate() error {
	return validation.ValidateStruct(c,
		validation.Field(&c.SeasonID, validation.Required, is.UUID),
		validation.Field(&c.Source, validation.Required, is2.AccessSource),
		validation.Field(&c.OwnerUserID, validation.Required, is.UUID),
		validation.Field(&c.Usernames, validation.Required, validation.Each(is2.MinecraftUsername)),
	)
}

type CheckUsernamesCommand struct {
	SeasonID  *uuid.UUID
	Usernames []string
}

func (c *CheckUsernamesCommand) Validate() error {
	return validation.ValidateStruct(c,
		validation.Field(&c.SeasonID, validation.NilOrNotEmpty, is.UUID),
		validation.Field(&c.Usernames, validation.Required, validation.Each(is2.MinecraftUsername)),
	)
}
