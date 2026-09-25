package domain

import (
	"time"

	"github.com/google/uuid"
)

type AccessStatus string

const (
	AccessStatusActive   AccessStatus = "active"
	AccessStatusInactive AccessStatus = "inactive"
	AccessStatusExpired  AccessStatus = "expired"
)

// SeasonAccessStatus is the access of a profile to one season.
type SeasonAccessStatus struct {
	SeasonID uuid.UUID
	Status   AccessStatus
}

// SeasonAccessStatuses lists a status for every active season and for every ended season the profile has access to.
// An active season is active or inactive for the profile, an ended one is expired. Seasons keep the given order.
func SeasonAccessStatuses(seasons []*Season, accesses []*ProfileAccess) []*SeasonAccessStatus {
	accessedSeasonIDs := make(map[uuid.UUID]bool, len(accesses))
	for _, access := range accesses {
		accessedSeasonIDs[access.SeasonID] = true
	}

	statuses := make([]*SeasonAccessStatus, 0, len(seasons))
	for _, season := range seasons {
		hasAccess := accessedSeasonIDs[season.ID]
		switch {
		case season.IsActive && hasAccess:
			statuses = append(statuses, &SeasonAccessStatus{SeasonID: season.ID, Status: AccessStatusActive})
		case season.IsActive:
			statuses = append(statuses, &SeasonAccessStatus{SeasonID: season.ID, Status: AccessStatusInactive})
		case hasAccess:
			statuses = append(statuses, &SeasonAccessStatus{SeasonID: season.ID, Status: AccessStatusExpired})
		}
	}
	return statuses
}

type AccessSource string

const (
	AccessSourceFree         AccessSource = "free"
	AccessSourceRegistration AccessSource = "registration"
	AccessSourceFreekassa    AccessSource = "freekassa"
	AccessSourceAdmin        AccessSource = "admin"
)

type Role string

const (
	RoleOwner     Role = "owner"
	RoleAdmin     Role = "admin"
	RoleModerator Role = "mod"
	RolePlayer    Role = "player"
)

const (
	RolePriorityOwner     = 1
	RolePriorityAdmin     = 2
	RolePriorityModerator = 3
	RolePriorityPlayer    = 4
)

// Valid tells whether the role is one of the known roles.
func (r Role) Valid() bool {
	switch r {
	case RoleOwner, RoleAdmin, RoleModerator, RolePlayer:
		return true
	}
	return false
}

// ProfileRole is the role of one player.
type ProfileRole struct {
	MinecraftUUID uuid.UUID
	Role          Role
}

// StaffRoles are the roles that make a profile part of the server staff.
var StaffRoles = []Role{RoleOwner, RoleAdmin, RoleModerator}

// ProfileFilter narrows down the public profile list.
type ProfileFilter struct {
	// Search matches usernames by prefix.
	Search string
	// OnlineOnly keeps only players that are online right now.
	OnlineOnly bool
	// StaffOnly keeps only players with one of the StaffRoles.
	StaffOnly bool
}

// ProfilesStats sums up the community. Online is nil when the Minecraft server cannot be reached.
type ProfilesStats struct {
	Total       int64
	Online      *int64
	NewLastWeek int64
}

type Profile struct {
	ID                uuid.UUID
	MinecraftUUID     uuid.UUID
	MinecraftUsername string
	OwnerUserID       *uuid.UUID
	FirstSeenAt       *time.Time
	LastSeenAt        *time.Time
	Role              Role
	IsSlimModel       bool
	CreatedAt         time.Time
	UpdatedAt         time.Time
	UpdatedBy         *uuid.UUID
	// LegacyMinecraftUUID is the offline UUID the profile had before it was rekeyed to its Mojang UUID.
	// Plan still has the playtime played under it.
	LegacyMinecraftUUID *uuid.UUID
	// PremiumConflict marks a nickname that was free when checked and is a licensed account of someone else now.
	PremiumConflict bool
	// relations
	Accesses   []*ProfileAccess
	Playtimes  []*ProfilePlaytime
	Violations []*ProfileViolation
}

type ProfileAccess struct {
	MinecraftUUID uuid.UUID
	SeasonID      uuid.UUID
	Source        AccessSource
	CreatedAt     time.Time
	UpdatedAt     time.Time
	UpdatedBy     *uuid.UUID
	// relations
	Profile *Profile
	Season  *Season
}

type ProfilePlaytime struct {
	MinecraftUUID uuid.UUID
	SeasonID      uuid.UUID
	Playtime      int64
	UpdatedAt     time.Time
	// relations
	Profile *Profile
	Season  *Season
}

// ProfileSeasonStats is what one profile did in one season. New metrics are added as fields.
type ProfileSeasonStats struct {
	SeasonID   uuid.UUID
	SeasonName string
	StartDate  time.Time
	EndDate    *time.Time
	IsActive   bool
	IsPrimary  bool
	// Playtime is in milliseconds.
	Playtime int64
}

// MojangLookupTarget is a profile whose Mojang UUID has to be looked up by username.
type MojangLookupTarget struct {
	MinecraftUUID     uuid.UUID
	MinecraftUsername string
	// CheckedBefore is set when an earlier lookup found no Mojang account with the username.
	CheckedBefore bool
}

// PremiumRekeyTarget is a profile whose nickname is a Mojang account while the profile still has another UUID.
type PremiumRekeyTarget struct {
	ProfileID         uuid.UUID
	MinecraftUUID     uuid.UUID
	MinecraftUsername string
	MojangUUID        uuid.UUID
}

// PremiumRekeyReport sums up one run of the premium rekey.
type PremiumRekeyReport struct {
	// Rekeyed are the profiles moved to their Mojang UUID.
	Rekeyed []string
	// Skipped are the profiles left alone: their UUID is neither the offline one nor the Mojang one.
	Skipped []string
	// Failed are the profiles that could not be moved; the run can be repeated.
	Failed []string
	// Unchecked is how many profiles were never looked up on Mojang yet; the background sync gets to them.
	Unchecked int
}

type ProfileViolation struct {
	MinecraftUUID uuid.UUID
	SeasonID      uuid.UUID
	Violation     string
	UpdatedAt     time.Time
	// relations
	Profile *Profile
	Season  *Season
}

type ProfilePrefixType string

const (
	ProfilePrefixTypeGlyth   ProfilePrefixType = "glyth"
	ProfilePrefixTypeSpecial ProfilePrefixType = "special"
)

// ProfileCosmetics is what one profile shows in one season.
type ProfileCosmetics struct {
	// NameColor is the default name color when the player picked none.
	NameColor *NameColor
	Glyth     *NamePrefix
	Special   *NamePrefix
}

type ProfileNameColorOption struct {
	ID          uuid.UUID
	ProfileID   uuid.UUID
	NameColorID uuid.UUID
	OrderItemID *uuid.UUID
	ForSeasonID *uuid.UUID
	// relations
	Profile   *Profile
	NameColor *NameColor
	OrderItem *OrderItem
	Season    *Season
}

type ProfileNamePrefixOption struct {
	ID           uuid.UUID
	ProfileID    uuid.UUID
	NamePrefixID uuid.UUID
	Type         ProfilePrefixType
	OrderItemID  *uuid.UUID
	ForSeasonID  *uuid.UUID
	CreatedAt    time.Time
	// relations
	Profile    *Profile
	NamePrefix *NamePrefix
	OrderItem  *OrderItem
	Season     *Season
}

func GetRolePriority(role Role) int {
	switch role {
	case RoleOwner:
		return RolePriorityOwner
	case RoleAdmin:
		return RolePriorityAdmin
	case RoleModerator:
		return RolePriorityModerator
	case RolePlayer:
		return RolePriorityPlayer
	}
	return RolePriorityPlayer
}
