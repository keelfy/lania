package responses

import (
	"github.com/google/uuid"
)

type Profile struct {
	ID            uuid.UUID         `json:"id"`
	MinecraftUUID uuid.UUID         `json:"mcUuid"`
	Username      string            `json:"username"`
	Cosmetics     *ProfileCosmetics `json:"cosmetics"`
	AccessStatus  string            `json:"accessStatus"`
	MojangUUID    *uuid.UUID        `json:"mojangUuid,omitempty"`
}

type PublicProfile struct {
	ID            uuid.UUID         `json:"id"`
	MinecraftUUID uuid.UUID         `json:"mcUuid"`
	Username      string            `json:"username"`
	Cosmetics     *ProfileCosmetics `json:"cosmetics"`
	Role          string            `json:"role"`
	IsOnline      bool              `json:"isOnline"`
	MojangUUID    *uuid.UUID        `json:"mojangUuid,omitempty"`
}

type ProfileDetails struct {
	ID            uuid.UUID         `json:"id"`
	MinecraftUUID uuid.UUID         `json:"mcUuid"`
	Username      string            `json:"username"`
	Cosmetics     *ProfileCosmetics `json:"cosmetics"`
	IsSlimModel   bool              `json:"isSlimModel"`
	FirstSeenAt   *int64            `json:"firstSeenAt,omitempty"`
	LastSeenAt    *int64            `json:"lastSeenAt,omitempty"`
	Role          string            `json:"role"`
	AccessStatus  string            `json:"accessStatus"`
	Playtime      int64             `json:"playtime"`
	IsOnline      bool              `json:"isOnline"`
	MojangUUID    *uuid.UUID        `json:"mojangUuid,omitempty"`
}

type NameColor struct {
	ID     uuid.UUID `json:"id"`
	Name   string    `json:"name"`
	Colors []string  `json:"colors"`
}

type NamePrefix struct {
	ID     uuid.UUID `json:"id"`
	Name   string    `json:"name"`
	Prefix string    `json:"prefix"`
	Image  string    `json:"image"`
}

type ProfileNameCosmetics struct {
	Colors        *NameColor  `json:"colors"`
	GlythPrefix   *NamePrefix `json:"glythPrefix,omitempty"`
	SpecialPrefix *NamePrefix `json:"specialPrefix,omitempty"`
}

type ProfileCosmetics struct {
	Name *ProfileNameCosmetics `json:"name"`
}

type UsernameStatus string

const (
	UsernameStatusAvailable  UsernameStatus = "available"
	UsernameStatusTaken      UsernameStatus = "taken"
	UsernameStatusOwnedByYou UsernameStatus = "owned_by_you"
)

type CheckUsername struct {
	Status    UsernameStatus `json:"status"`
	HasAccess bool           `json:"hasAccess"`
}

type ProfileNameColorOption struct {
	ID          uuid.UUID  `json:"id"`
	NameColorID uuid.UUID  `json:"nameColorId"`
	Name        string     `json:"name"`
	ProfileID   uuid.UUID  `json:"profileId"`
	Colors      []string   `json:"colors"`
	ForSeasonID *uuid.UUID `json:"forSeasonId,omitempty"`
}

type ProfileNamePrefixOption struct {
	ID           uuid.UUID  `json:"id"`
	NamePrefixID uuid.UUID  `json:"namePrefixId"`
	Name         string     `json:"name"`
	ProfileID    uuid.UUID  `json:"profileId"`
	Prefix       string     `json:"prefix"`
	Image        string     `json:"image"`
	ForSeasonID  *uuid.UUID `json:"forSeasonId,omitempty"`
}

type ProfileNameCosmeticOptions struct {
	Colors          []*ProfileNameColorOption  `json:"colors"`
	GlythPrefixes   []*ProfileNamePrefixOption `json:"glythPrefixes"`
	SpecialPrefixes []*ProfileNamePrefixOption `json:"specialPrefixes"`
}

type ProfileCosmeticOptions struct {
	Name *ProfileNameCosmeticOptions `json:"name"`
}
