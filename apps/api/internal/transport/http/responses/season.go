package responses

import "github.com/google/uuid"

type Season struct {
	ID uuid.UUID `json:"id"`
	// Name is the same in every language.
	Name string `json:"name"`
	// PreviewImage is missing for a season that is not shown on the seasons page.
	PreviewImage     *string `json:"previewImage,omitempty"`
	StartDate        int64   `json:"startDate"`
	EndDate          *int64  `json:"endDate,omitempty"`
	PublicAddress    *string `json:"publicAddress,omitempty"`
	IsActive         bool    `json:"isActive"`
	IsPrimary        bool    `json:"isPrimary"`
	Preregistration  bool    `json:"preregistration"`
	FreeRegistration bool    `json:"freeRegistration"`
}

type AdminSeason struct {
	Season
	ShellAddress *string `json:"shellAddress,omitempty"`
}
