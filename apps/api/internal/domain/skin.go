package domain

import "github.com/google/uuid"

// PlayerSkin is the skin a player chose in game with SkinsRestorer, on the server of one season.
type PlayerSkin struct {
	// TextureURL is the https link to the texture, empty when the server has not fetched the skin of the licensed
	// account MojangUUID yet.
	TextureURL string
	// Slim is the model with thin arms (Alex).
	Slim bool
	// MojangUUID is the licensed account the skin was copied from by nickname.
	MojangUUID *uuid.UUID
}

// SkinVariant is the model of a skin given by a file.
type SkinVariant string

const (
	SkinVariantClassic SkinVariant = "classic"
	SkinVariantSlim    SkinVariant = "slim"
)
