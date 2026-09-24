package requests

import "github.com/google/uuid"

type TransferProfileOwner struct {
	Email string `json:"email"`
}

type SetProfileRole struct {
	// Role is owner, admin, mod or player.
	Role string `json:"role"`
}

type MergeProfiles struct {
	TargetProfileID uuid.UUID `json:"targetProfileId"`
}

type GrantProduct struct {
	ProductID uuid.UUID `json:"productId"`
	SeasonID  uuid.UUID `json:"seasonId"`
}

type GrantCosmetic struct {
	// Type is name-color or name-prefix.
	Type   string    `json:"type"`
	ItemID uuid.UUID `json:"itemId"`
	// PrefixType is glyth or special, used for a name prefix only.
	PrefixType string `json:"prefixType"`
	// SeasonID is empty to grant the item for good.
	SeasonID *uuid.UUID `json:"seasonId"`
}

type SaveSeason struct {
	Name             string  `json:"name"`
	PreviewImage     *string `json:"previewImage"`
	StartDate        string  `json:"startDate"`
	EndDate          *string `json:"endDate"`
	PublicAddress    *string `json:"publicAddress"`
	ShellAddress     *string `json:"shellAddress"`
	PlanURL          *string `json:"planUrl"`
	IsActive         bool    `json:"isActive"`
	IsPrimary        bool    `json:"isPrimary"`
	Preregistration  bool    `json:"preregistration"`
	FreeRegistration bool    `json:"freeRegistration"`
	GameVersion      *string `json:"gameVersion"`
	WorldURL         *string `json:"worldUrl"`
}

type SaveSeasonWorld struct {
	Slug         string  `json:"slug"`
	Name         string  `json:"name"`
	PreviewImage *string `json:"previewImage"`
	MapURL       *string `json:"mapUrl"`
	// ClaimLimit falls back to domain.DefaultClaimLimit when the request leaves it out.
	ClaimLimit      *int     `json:"claimLimit"`
	ClaimDimensions []string `json:"claimDimensions"`
	Position        int      `json:"position"`
}

type SaveSeasonScreenshot struct {
	Image            string      `json:"image"`
	Title            *string     `json:"title"`
	Position         int         `json:"position"`
	AuthorProfileIDs []uuid.UUID `json:"authorProfileIds"`
}

type SaveNameColor struct {
	Name   string   `json:"name"`
	Colors []string `json:"colors"`
}

type SaveNamePrefix struct {
	Name    string `json:"name"`
	Prefix  string `json:"prefix"`
	Image   string `json:"image"`
	NoSpace bool   `json:"noSpace"`
}

type SaveProductLocalization struct {
	Locale      string `json:"locale"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type SaveProduct struct {
	Category            string                    `json:"category"`
	CosmeticID          *uuid.UUID                `json:"cosmeticId"`
	PriceName           string                    `json:"priceName"`
	IsActive            bool                      `json:"isActive"`
	EasyDonateProductID *int64                    `json:"easyDonateProductId"`
	Localizations       []SaveProductLocalization `json:"localizations"`
}
