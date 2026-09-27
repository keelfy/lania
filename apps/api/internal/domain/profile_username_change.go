package domain

import (
	"time"

	"github.com/google/uuid"
)

// ProfileUsernameChangeSource tells who changed the nickname of a profile.
type ProfileUsernameChangeSource string

const (
	// ProfileUsernameChangeSourceOwner is a change the owner made on the site.
	ProfileUsernameChangeSourceOwner ProfileUsernameChangeSource = "owner"
	// ProfileUsernameChangeSourceMojang is the new name of a licensed account, picked up from Mojang.
	ProfileUsernameChangeSourceMojang ProfileUsernameChangeSource = "mojang"
)

// ProfileUsernameChange is one nickname change of a profile. The UUIDs are equal for a licensed account,
// which keeps its Mojang UUID on a rename.
type ProfileUsernameChange struct {
	ID               uuid.UUID
	ProfileID        uuid.UUID
	OldUsername      string
	NewUsername      string
	OldMinecraftUUID uuid.UUID
	NewMinecraftUUID uuid.UUID
	Source           ProfileUsernameChangeSource
	// ChangedBy is the user who made the change, nil for a change picked up from Mojang.
	ChangedBy *uuid.UUID
	CreatedAt time.Time
}
