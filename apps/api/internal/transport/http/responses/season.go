package responses

import "github.com/google/uuid"

type Season struct {
	ID uuid.UUID `json:"id"`
	// Name is the same in every language.
	Name string `json:"name"`
	// PreviewImage is missing for a season that is not shown on the seasons page.
	PreviewImage  *string `json:"previewImage,omitempty"`
	StartDate     int64   `json:"startDate"`
	EndDate       *int64  `json:"endDate,omitempty"`
	PublicAddress *string `json:"publicAddress,omitempty"`
	IsActive      bool    `json:"isActive"`
	IsPrimary     bool    `json:"isPrimary"`
	// OnlineAvailable tells whether the API knows who is online in the season: it is active and has a server.
	OnlineAvailable  bool    `json:"onlineAvailable"`
	Preregistration  bool    `json:"preregistration"`
	FreeRegistration bool    `json:"freeRegistration"`
	GameVersion      *string `json:"gameVersion,omitempty"`
	WorldURL         *string `json:"worldUrl,omitempty"`
	// MapURL is the squaremap of the season server, missing when chunk claims are off.
	MapURL     *string `json:"mapUrl,omitempty"`
	ClaimLimit int     `json:"claimLimit"`
}

type AdminSeason struct {
	Season
	ShellAddress *string `json:"shellAddress,omitempty"`
	PlanURL      *string `json:"planUrl,omitempty"`
}
