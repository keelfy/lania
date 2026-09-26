package domain

import "github.com/google/uuid"

type NameColorMetadata struct {
	Colors []string `json:"colors"`
}

type NamePrefixMetadata struct {
	Prefix  string `json:"prefix"`
	Image   string `json:"image"`
	NoSpace bool   `json:"noSpace"`
}

// CosmeticNameLocales are the locales a cosmetic name is translated into. English is the main name itself.
var CosmeticNameLocales = []string{"ru"}

// CosmeticNames maps a locale to the translation of the main name.
type CosmeticNames map[string]string

// Localized returns the name in the locale, or fallback when the locale has none.
func (n CosmeticNames) Localized(locale, fallback string) string {
	if name := n[locale]; name != "" {
		return name
	}
	return fallback
}

type NameColor struct {
	ID uuid.UUID
	// Name is the unique main name, shown in English and wherever Names has no translation.
	Name     string
	Names    CosmeticNames
	Metadata NameColorMetadata
}

// LocalizedName is the name players see in the locale.
func (c *NameColor) LocalizedName(locale string) string {
	return c.Names.Localized(locale, c.Name)
}

type NamePrefix struct {
	ID uuid.UUID
	// Name is the unique main name, shown in English and wherever Names has no translation.
	Name     string
	Names    CosmeticNames
	Metadata NamePrefixMetadata
}

// LocalizedName is the name players see in the locale.
func (p *NamePrefix) LocalizedName(locale string) string {
	return p.Names.Localized(locale, p.Name)
}

// CosmeticsCatalog lists every name color and name prefix that exists, whether it is for sale or not.
type CosmeticsCatalog struct {
	NameColors   []*NameColor
	NamePrefixes []*NamePrefix
}
