package requests

import "github.com/google/uuid"

type TransferProfileOwner struct {
	Email string `json:"email"`
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
	SeasonNumber int     `json:"seasonNumber"`
	Name         string  `json:"name"`
	PreviewImage *string `json:"previewImage"`
	StartDate    string  `json:"startDate"`
	EndDate      *string `json:"endDate"`
	ServerIP     *string `json:"serverIp"`
	ServerPort   *uint16 `json:"serverPort"`
	IsActive     bool    `json:"isActive"`
	// Nil keeps the current password during update. Empty clears it.
	RCONPassword *string `json:"rconPassword"`
}
