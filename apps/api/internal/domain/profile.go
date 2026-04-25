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

type AccessSource string

const (
	AccessSourceFree      AccessSource = "free"
	AccessSourceFreekassa AccessSource = "freekassa"
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

type Profile struct {
	ID                uuid.UUID
	MinecraftUUID     uuid.UUID
	MinecraftUsername string
	OwnerUserID       uuid.UUID
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
	OrderItemID   *uuid.UUID
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
