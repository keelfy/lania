package responses

import (
	"github.com/google/uuid"
)

type SeasonAccess struct {
	SeasonID uuid.UUID `json:"seasonId"`
	Status   string    `json:"status"`
}

type Profile struct {
	ID            uuid.UUID         `json:"id"`
	MinecraftUUID uuid.UUID         `json:"mcUuid"`
	Username      string            `json:"username"`
	Cosmetics     *ProfileCosmetics `json:"cosmetics"`
	AccessStatus  string            `json:"accessStatus"`
	Accesses      []*SeasonAccess   `json:"accesses"`
	MojangUUID    *uuid.UUID        `json:"mojangUuid,omitempty"`
	// Skin is the skin the player chose in game in the season, absent when the player wears the skin of the
	// licensed account or the default one.
	Skin *PlayerSkin `json:"skin,omitempty"`
	// Verified: the owner proved in game that the licensed account is theirs.
	Verified bool `json:"verified"`
	// UsernameChangeAvailableAt is when the owner can move to an unlicensed nickname again (a change of one, or
	// leaving the license), absent when the owner can now.
	UsernameChangeAvailableAt *int64 `json:"usernameChangeAvailableAt,omitempty"`
	// UsernameChangeCooldownDays is how long the owner of an unlicensed profile waits after a change, absent
	// when there is no cooldown.
	UsernameChangeCooldownDays int `json:"usernameChangeCooldownDays,omitempty"`
}

type PublicProfile struct {
	ID            uuid.UUID         `json:"id"`
	MinecraftUUID uuid.UUID         `json:"mcUuid"`
	Username      string            `json:"username"`
	Cosmetics     *ProfileCosmetics `json:"cosmetics"`
	Role          string            `json:"role"`
	IsOnline      bool              `json:"isOnline"`
	Playtime      int64             `json:"playtime"`
	LastSeenAt    *int64            `json:"lastSeenAt,omitempty"`
	MojangUUID    *uuid.UUID        `json:"mojangUuid,omitempty"`
	// Skin is the skin the player chose in game in the season, absent when the player wears the skin of the
	// licensed account or the default one.
	Skin *PlayerSkin `json:"skin,omitempty"`
	// Verified: the owner proved in game that the licensed account is theirs.
	Verified bool `json:"verified"`
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
	Accesses      []*SeasonAccess   `json:"accesses"`
	Playtime      int64             `json:"playtime"`
	IsOnline      bool              `json:"isOnline"`
	MojangUUID    *uuid.UUID        `json:"mojangUuid,omitempty"`
	// Skin is the skin the player chose in game in the season, absent when the player wears the skin of the
	// licensed account or the default one.
	Skin *PlayerSkin `json:"skin,omitempty"`
	// Verified: the owner proved in game that the licensed account is theirs.
	Verified bool `json:"verified"`
}

type PlayerSkin struct {
	// TextureURL is the https link to the skin texture. It is absent when the server has not fetched the skin of
	// the licensed account MojangUUID yet.
	TextureURL string `json:"textureUrl,omitempty"`
	Slim       bool   `json:"slim"`
	// MojangUUID is the licensed account the skin was copied from by nickname.
	MojangUUID *uuid.UUID `json:"mojangUuid,omitempty"`
}

// SkinChangeStatus tells whether the server applied a skin change.
type SkinChangeStatus string

const (
	SkinChangeStatusApplied SkinChangeStatus = "applied"
	// SkinChangeStatusPending means the server did not apply the change in time. It may still apply it.
	SkinChangeStatusPending SkinChangeStatus = "pending"
)

type SkinChange struct {
	Status SkinChangeStatus `json:"status"`
	// Skin is the applied skin, absent for a pending change and for a cleared skin.
	Skin *PlayerSkin `json:"skin,omitempty"`
}

type ProfileSeasonStats struct {
	SeasonID   uuid.UUID `json:"seasonId"`
	SeasonName string    `json:"seasonName"`
	StartDate  int64     `json:"startDate"`
	EndDate    *int64    `json:"endDate,omitempty"`
	IsActive   bool      `json:"isActive"`
	IsPrimary  bool      `json:"isPrimary"`
	// Playtime is in milliseconds.
	Playtime int64 `json:"playtime"`
	Deaths   int64 `json:"deaths"`
	MobKills int64 `json:"mobKills"`
}

type ProfileStats struct {
	// TotalPlaytime is summed over all seasons, in milliseconds.
	TotalPlaytime int64                 `json:"totalPlaytime"`
	TotalDeaths   int64                 `json:"totalDeaths"`
	TotalMobKills int64                 `json:"totalMobKills"`
	Seasons       []*ProfileSeasonStats `json:"seasons"`
}

// ProfileViolations are the bans and mutes of a profile in one season. Available is false when the server of the
// season is over or cannot be reached, so nothing is known.
type ProfileViolations struct {
	Available  bool                `json:"available"`
	Violations []*ProfileViolation `json:"violations"`
}

// ProfileViolation times are in epoch milliseconds.
type ProfileViolation struct {
	// ID is unique within a kind.
	ID     int64  `json:"id"`
	Kind   string `json:"kind"`
	Reason string `json:"reason"`
	// IssuedBy is null when the server console gave it.
	IssuedBy  *string `json:"issuedBy"`
	IssuedAt  int64   `json:"issuedAt"`
	ExpiresAt *int64  `json:"expiresAt"`
	Status    string  `json:"status"`
	// RemovedBy is null when the console lifted it or nobody did.
	RemovedBy     *string `json:"removedBy"`
	RemovedReason string  `json:"removedReason"`
	RemovedAt     *int64  `json:"removedAt"`
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
	// Premium is true when a Mojang account has the nickname: the server lets only its owner in.
	Premium bool `json:"premium"`
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

type ProfilesStats struct {
	Total       int64  `json:"total"`
	Online      *int64 `json:"online,omitempty"`
	NewLastWeek int64  `json:"newLastWeek"`
}

// ProfileResync tells per season what a resync wrote to the server of the season.
type ProfileResync struct {
	// OK is true when every part was written on every server.
	OK      bool            `json:"ok"`
	Seasons []*SeasonResync `json:"seasons"`
}

type SeasonResync struct {
	SeasonID   uuid.UUID     `json:"seasonId"`
	SeasonName string        `json:"seasonName"`
	OK         bool          `json:"ok"`
	Parts      []*PartResync `json:"parts"`
}

type PartResync struct {
	// Part is role, cosmetics or access.
	Part string `json:"part"`
	OK   bool   `json:"ok"`
	// Error is missing when the part was written.
	Error *string `json:"error,omitempty"`
}

type ProfileVerification struct {
	ExpiresAt int64 `json:"expiresAt"`
}

// VerificationLogin has the code the proxy shows on the kick screen; empty lets the player in.
type VerificationLogin struct {
	Code string `json:"code,omitempty"`
}
