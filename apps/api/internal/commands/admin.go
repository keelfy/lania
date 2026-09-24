package commands

import (
	"errors"
	"regexp"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/domain"
)

var hexColorPattern = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

// imageLocationPattern accepts an s3://bucket/key location, the form seasons use and the one
// glyth previews moved to, as well as a plain http(s) URL for icons still hosted elsewhere.
var imageLocationPattern = regexp.MustCompile(`^(s3|https?)://\S+$`)

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

type MergeProfilesCommand struct {
	SourceProfileID uuid.UUID
	TargetProfileID uuid.UUID
}

func (c *MergeProfilesCommand) Validate() error {
	return validation.ValidateStruct(c,
		validation.Field(&c.SourceProfileID, notNilUUID),
		validation.Field(&c.TargetProfileID, notNilUUID, validation.By(func(value any) error {
			if value.(uuid.UUID) == c.SourceProfileID {
				return errors.New("cannot be the source profile")
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

type SaveNameColorCommand struct {
	ID     uuid.UUID
	Name   string
	Colors []string
}

func (c *SaveNameColorCommand) Validate() error {
	return validation.ValidateStruct(c,
		validation.Field(&c.Name, validation.Required, validation.Length(1, 255)),
		validation.Field(&c.Colors, validation.By(func(value any) error {
			for _, color := range value.([]string) {
				if !hexColorPattern.MatchString(color) {
					return errors.New("must contain only #RRGGBB colors")
				}
			}
			return nil
		})),
	)
}

type SaveNamePrefixCommand struct {
	ID      uuid.UUID
	Name    string
	Prefix  string
	Image   string
	NoSpace bool
}

func (c *SaveNamePrefixCommand) Validate() error {
	return validation.ValidateStruct(c,
		validation.Field(&c.Name, validation.Required, validation.Length(1, 255)),
		validation.Field(&c.Prefix, validation.Required),
		validation.Field(&c.Image, validation.Required, validation.RuneLength(1, 512), validation.Match(imageLocationPattern)),
	)
}

type SaveProductLocalization struct {
	Locale      string
	Name        string
	Description string
}

type SaveProductCommand struct {
	ID                  uuid.UUID
	Category            domain.ProductCategory
	CosmeticID          *uuid.UUID
	PriceName           domain.ProductPriceName
	IsActive            bool
	EasyDonateProductID *int64
	Localizations       []SaveProductLocalization
}

func (c *SaveProductCommand) Validate() error {
	validCategory := c.Category == domain.ProductCategoryUpgrade || c.Category == domain.ProductCategoryNameColor || c.Category == domain.ProductCategoryNamePrefix
	if !validCategory {
		return errors.New("category must be upgrade, name-color or name-prefix")
	}
	expectedPrice := map[domain.ProductCategory]domain.ProductPriceName{
		domain.ProductCategoryUpgrade:    domain.ProductPriceNameSeasonAccess,
		domain.ProductCategoryNameColor:  domain.ProductPriceNameNameColor,
		domain.ProductCategoryNamePrefix: domain.ProductPriceNameNamePrefix,
	}[c.Category]
	if c.PriceName != expectedPrice {
		return errors.New("priceName does not match category")
	}
	if c.Category != domain.ProductCategoryUpgrade && (c.CosmeticID == nil || *c.CosmeticID == uuid.Nil) {
		return errors.New("cosmeticId is required")
	}
	if c.IsActive && (c.EasyDonateProductID == nil || *c.EasyDonateProductID <= 0) {
		return errors.New("easyDonateProductId is required for an active product")
	}
	seen := map[string]bool{}
	for _, localization := range c.Localizations {
		locale := strings.ToLower(localization.Locale)
		if locale != "ru" && locale != "en" {
			return errors.New("localization locale must be ru or en")
		}
		if seen[locale] {
			return errors.New("localization locale must be unique")
		}
		seen[locale] = true
		if strings.TrimSpace(localization.Name) == "" || strings.TrimSpace(localization.Description) == "" {
			return errors.New("localization name and description are required")
		}
	}
	if !seen["ru"] || !seen["en"] {
		return errors.New("ru and en localizations are required")
	}
	return nil
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

// MaxEDProductImageBytes bounds the image forwarded to EasyDonate. Same order of magnitude as a
// glyth preview: the control panel renders it as a small shop thumbnail.
const MaxEDProductImageBytes int64 = 1024 * 1024

type CreateEDProductCommand struct {
	// UserAuth is the value of the admin's user_auth cookie from the control panel. It logs in
	// on its own, and the CSRF token is read from the page it opens. Never log it.
	UserAuth    string
	Name        string
	Description string
	// PriceName selects the tariff whose RUB amount becomes the EasyDonate price, so the shop
	// cannot drift away from lania's own prices.
	PriceName domain.ProductPriceName
	// Image is optional. Empty means the position is created without a picture.
	Image            []byte
	ImageFilename    string
	ImageContentType string
}

func (c *CreateEDProductCommand) Validate() error {
	return validation.ValidateStruct(c,
		validation.Field(&c.UserAuth, validation.Required),
		validation.Field(&c.Name, validation.Required),
		validation.Field(&c.Description, validation.Required),
		validation.Field(&c.PriceName, validation.Required, validation.By(func(value any) error {
			priceName := value.(domain.ProductPriceName)
			if priceName != domain.ProductPriceNameSeasonAccess && priceName != domain.ProductPriceNameNameColor && priceName != domain.ProductPriceNameNamePrefix {
				return errors.New("must be season_access, name_color or name_prefix")
			}
			return nil
		})),
		validation.Field(&c.Image, validation.By(func(value any) error {
			if image, ok := value.([]byte); ok && int64(len(image)) > MaxEDProductImageBytes {
				return errors.New("must not exceed MaxEDProductImageBytes")
			}
			return nil
		})),
	)
}
