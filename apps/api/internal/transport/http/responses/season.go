package responses

import "github.com/google/uuid"

type Season struct {
	ID uuid.UUID `json:"id"`
	// Name is the same in every language.
	Name string `json:"name"`
	// PreviewImage is missing for a season that is not shown on the seasons page.
	PreviewImage *string `json:"previewImage,omitempty"`
	StartDate    int64   `json:"startDate"`
	EndDate      *int64  `json:"endDate,omitempty"`
	// IsActive marks the season that runs on the Minecraft server.
	IsActive bool `json:"isActive"`
}

type AdminSeason struct {
	Season
	SeasonNumber    int     `json:"seasonNumber"`
	ServerIP        *string `json:"serverIp,omitempty"`
	ServerPort      *uint16 `json:"serverPort,omitempty"`
	RCONPasswordSet bool    `json:"rconPasswordSet"`
}
