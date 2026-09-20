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
	NameColorID       uuid.UUID
	CreatedAt         time.Time
	UpdatedAt         time.Time
	UpdatedBy         *uuid.UUID
	// relations
	Accesses   []*ProfileAccess
	Playtimes  []*ProfilePlaytime
	Violations []*ProfileViolation
	NameColor  *NameColor
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

// MojangLookupTarget is a profile whose Mojang UUID has to be looked up by username.
type MojangLookupTarget struct {
	MinecraftUUID     uuid.UUID
	MinecraftUsername string
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

type ProfilePrefix struct {
	ProfileID    uuid.UUID
	NamePrefixID uuid.UUID
	Type         ProfilePrefixType
	CreatedAt    time.Time
	CreatedBy    *uuid.UUID
	// relations
	Profile    *Profile
	NamePrefix *NamePrefix
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

// HighestRole picks the role with the smallest priority number among groups.
// Groups that are not roles are ignored, so a player without role groups is a RolePlayer.
func HighestRole(groups []string) Role {
	role := RolePlayer
	rolePriority := RolePriorityPlayer

	for _, group := range groups {
		if priority := GetRolePriority(Role(group)); priority < rolePriority {
			role = Role(group)
			rolePriority = priority
		}
	}
	return role
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
