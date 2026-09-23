package commands

import (
	"errors"
	"fmt"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/google/uuid"
)

type UploadImageKind string

const (
	UploadImageKindGlythPreview     UploadImageKind = "glyth-preview"
	UploadImageKindSeasonScreenshot UploadImageKind = "season-screenshot"
	UploadImageKindSeasonPreview    UploadImageKind = "season-preview"
)

// MaxUploadImageBytes bounds the request body accepted for each kind, checked before the image is
// decoded. The API container runs with a 512m memory limit, which is why a screenshot is capped
// well below what a phone camera can produce.
var MaxUploadImageBytes = map[UploadImageKind]int64{
	UploadImageKindGlythPreview:     512 * 1024,
	UploadImageKindSeasonScreenshot: 8 * 1024 * 1024,
	UploadImageKindSeasonPreview:    8 * 1024 * 1024,
}

func (k UploadImageKind) Valid() bool {
	_, ok := MaxUploadImageBytes[k]
	return ok
}

type UploadImageCommand struct {
	Kind    UploadImageKind
	Content []byte
	// Token is the in-game glyth token (":glyth_popcat:"), used to name a glyth preview's key.
	Token string
	// Name is the cosmetic's display name, used to name a glyth preview's key when Token does not
	// match the :glyth_<snake_case>: pattern.
	Name string
	// SeasonID is required for a season screenshot, which is keyed per season.
	SeasonID uuid.UUID
}

func (c *UploadImageCommand) Validate() error {
	return validation.ValidateStruct(c,
		validation.Field(&c.Kind, validation.Required, validation.By(func(value any) error {
			if !value.(UploadImageKind).Valid() {
				return errors.New("must be glyth-preview, season-screenshot or season-preview")
			}
			return nil
		})),
		validation.Field(&c.Content, validation.Required, validation.By(func(value any) error {
			if max, ok := MaxUploadImageBytes[c.Kind]; ok && int64(len(value.([]byte))) > max {
				return fmt.Errorf("must not exceed %d bytes", max)
			}
			return nil
		})),
		validation.Field(&c.SeasonID, validation.By(func(value any) error {
			if c.Kind != UploadImageKindSeasonScreenshot {
				return nil
			}
			if value.(uuid.UUID) == uuid.Nil {
				return errors.New("is required for a season screenshot")
			}
			return nil
		})),
	)
}
