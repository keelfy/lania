package commands

import (
	"errors"
	"regexp"
	"slices"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/google/uuid"
)

// worldSlugPattern is the shape of the key in /worlds/<slug>.
var worldSlugPattern = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

// maxDimensions caps a dimension list; a squaremap has three vanilla dimensions and a few custom ones at most.
const maxDimensions = 16

type SaveSeasonWorldCommand struct {
	ID                    uuid.UUID
	SeasonID              uuid.UUID
	Slug                  string
	Name                  string
	PreviewImage          *string
	MapURL                *string
	ClaimLimit            int
	ClaimDimensions       []string
	HiddenDimensions      []string
	PlanServer            *string
	ClaimMinPlaytimeHours int
	Position              int
}

func (c *SaveSeasonWorldCommand) Validate() error {
	return validation.ValidateStruct(c,
		validation.Field(&c.Slug, validation.Required, validation.Length(1, 32), validation.Match(worldSlugPattern)),
		validation.Field(&c.Name, validation.Required, validation.RuneLength(1, 64)),
		validation.Field(&c.PreviewImage, validation.By(validateRuneLength(c.PreviewImage, 1, 512))),
		validation.Field(&c.MapURL, validation.By(validateRuneLength(c.MapURL, 1, 512)), validation.By(validateHTTPURL(c.MapURL))),
		validation.Field(&c.ClaimLimit, validation.Min(0), validation.Max(100000)),
		validation.Field(&c.PlanServer, validation.By(validateRuneLength(c.PlanServer, 1, 100))),
		validation.Field(&c.ClaimMinPlaytimeHours, validation.Min(0), validation.Max(10000)),
		validation.Field(&c.ClaimDimensions, validation.Length(0, maxDimensions), validation.By(validateDimensions)),
		validation.Field(&c.HiddenDimensions, validation.Length(0, maxDimensions), validation.By(validateDimensions),
			validation.By(func(any) error {
				for _, dimension := range c.HiddenDimensions {
					if slices.Contains(c.ClaimDimensions, dimension) {
						return errors.New("a claim dimension cannot be hidden")
					}
				}
				return nil
			}),
		),
	)
}

// validateDimensions checks a list of squaremap world names.
func validateDimensions(value any) error {
	dimensions, _ := value.([]string)
	seen := make(map[string]bool, len(dimensions))
	for _, dimension := range dimensions {
		if !dimensionPattern.MatchString(dimension) {
			return errors.New("must be squaremap world names")
		}
		if seen[dimension] {
			return errors.New("dimension is listed twice")
		}
		seen[dimension] = true
	}
	return nil
}
