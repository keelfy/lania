package domain

import (
	"time"

	"github.com/google/uuid"
)

// MetadataRoleKey is the key of the site role in metadata_public of an Ory identity.
const MetadataRoleKey = "role"

// User is an Ory identity. Users are not stored in the API database.
type User struct {
	ID        uuid.UUID
	Email     string
	Username  string
	AvatarURL string
	Role      Role
	CreatedAt *time.Time
	// relations
	Profiles []*Profile
}

// IsAdmin tells whether the role may use the admin panel.
func (r Role) IsAdmin() bool {
	return r == RoleOwner || r == RoleAdmin
}

// RoleFromMetadata reads the site role from metadata_public of an Ory identity.
// A missing or unknown role is a RolePlayer.
func RoleFromMetadata(metadata map[string]any) Role {
	value, ok := metadata[MetadataRoleKey].(string)
	if !ok {
		return RolePlayer
	}

	switch role := Role(value); role {
	case RoleOwner, RoleAdmin, RoleModerator:
		return role
	}
	return RolePlayer
}
