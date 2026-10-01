package domain

import "github.com/google/uuid"

type PunishmentKind int

const (
	PunishmentKindBan PunishmentKind = iota + 1
	PunishmentKindMute
)

type PunishmentStatus int

const (
	// PunishmentStatusActive is a punishment in force now.
	PunishmentStatusActive PunishmentStatus = iota + 1
	// PunishmentStatusExpired is a punishment that ran out.
	PunishmentStatusExpired
	// PunishmentStatusRemoved is a punishment lifted before it ran out.
	PunishmentStatusRemoved
)

// Punishment is a ban or a mute LiteBans gave a player.
type Punishment struct {
	ID     int64
	Kind   PunishmentKind
	MCUUID uuid.UUID
	Reason string
	// IssuedBy is the nickname of the moderator, nil when the console gave it.
	IssuedBy   *string
	IssuedAtMs int64
	// ExpiresAtMs is nil for a permanent punishment.
	ExpiresAtMs *int64
	Status      PunishmentStatus
	// RemovedBy is the nickname of the moderator who lifted it, nil when the console did or nobody.
	RemovedBy     *string
	RemovedReason string
	RemovedAtMs   *int64
}
