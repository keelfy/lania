package commands

import (
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/google/uuid"
)

type SaveSeasonScreenshotCommand struct {
	ID               uuid.UUID
	SeasonID         uuid.UUID
	Image            string
	Title            *string
	Position         int
	AuthorProfileIDs uuid.UUIDs
}

func (c *SaveSeasonScreenshotCommand) Validate() error {
	return validation.ValidateStruct(c,
		validation.Field(&c.Image, validation.Required, validation.RuneLength(1, 512)),
		validation.Field(&c.Title, validation.By(validateRuneLength(c.Title, 0, 255))),
	)
}
