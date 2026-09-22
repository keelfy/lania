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
}
