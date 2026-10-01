package domain

import (
	"time"

	"github.com/google/uuid"
)

type PunishmentKind string

const (
	PunishmentKindBan  PunishmentKind = "ban"
	PunishmentKindMute PunishmentKind = "mute"
)

type PunishmentStatus string

const (
	// PunishmentStatusActive is a punishment in force now.
	PunishmentStatusActive PunishmentStatus = "active"
	// PunishmentStatusExpired is a punishment that ran out.
	PunishmentStatusExpired PunishmentStatus = "expired"
	// PunishmentStatusRemoved is a punishment lifted before it ran out.
	PunishmentStatusRemoved PunishmentStatus = "removed"
)

// Punishment is a ban or a mute a player got on the server of one season.
type Punishment struct {
	ID            int64
	Kind          PunishmentKind
	MinecraftUUID uuid.UUID
	Reason        string
	// IssuedBy is the nickname of the moderator, nil when the server console gave it.
	IssuedBy *string
	IssuedAt time.Time
	// ExpiresAt is nil for a permanent punishment.
	ExpiresAt *time.Time
	Status    PunishmentStatus
	// RemovedBy is the nickname of the moderator who lifted it, nil when the console did or nobody.
	RemovedBy     *string
	RemovedReason string
	RemovedAt     *time.Time
}
