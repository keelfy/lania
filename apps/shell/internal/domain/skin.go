package domain

import "github.com/google/uuid"

// PlayerSkin is the skin a player wears in game.
type PlayerSkin struct {
	// TextureURL is the https link to the texture, empty when it is not known yet.
	TextureURL string
	// Slim is the model with thin arms (Alex).
	Slim bool
	// MojangUUID is the licensed account the skin was copied from by nickname.
	MojangUUID *uuid.UUID
}
