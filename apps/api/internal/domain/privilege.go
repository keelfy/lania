package domain

import (
	"time"

	"github.com/google/uuid"
)

// Privilege is one LuckPerms permission node a player buys for a season, e.g. homes.commands.*.
type Privilege struct {
	ID uuid.UUID
	// Name is the unique main name, shown in English and wherever Names has no translation.
	Name  string
	Names CosmeticNames
	// Permission is the LuckPerms node written to the season server while the player has the privilege.
	Permission string
	// Prices holds one price per currency, under the tariff PrivilegePriceName(ID).
	Prices    []*ProductPrice
	CreatedAt time.Time
}

// LocalizedName is the name players see in the locale.
func (p *Privilege) LocalizedName(locale string) string {
	return p.Names.Localized(locale, p.Name)
}

// ProfilePrivilege is a privilege a profile has for one season.
type ProfilePrivilege struct {
	ID          uuid.UUID
	ProfileID   uuid.UUID
	PrivilegeID uuid.UUID
	SeasonID    uuid.UUID
	Permission  string
}
