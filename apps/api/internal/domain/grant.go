package domain

import (
	"time"

	"github.com/google/uuid"
)

type GrantType string

const (
	GrantTypeAccess     GrantType = "access"
	GrantTypeNameColor  GrantType = "name-color"
	GrantTypeNamePrefix GrantType = "name-prefix"
)

func (t GrantType) IsValid() bool {
	switch t {
	case GrantTypeAccess, GrantTypeNameColor, GrantTypeNamePrefix:
		return true
	}
	return false
}

// Grant is something a profile was given: access to a season, a name color or a name prefix.
// A revoked grant stays as history.
type Grant struct {
	ID        uuid.UUID
	Type      GrantType
	ProfileID uuid.UUID
	// SeasonID is empty for a grant that is not tied to a season.
	SeasonID *uuid.UUID
	// ItemID is the ID of the name color or name prefix, empty for access.
	ItemID uuid.UUID
	// PrefixType is set for a name prefix only.
	PrefixType ProfilePrefixType
	// Name is the name of the name color or name prefix, empty for access.
	Name string
	// Source tells how access was obtained, empty for a name color and a name prefix.
	Source AccessSource
	// OrderItemID is set when a name color or a name prefix came from an order.
	OrderItemID *uuid.UUID
	// GrantedBy is the user who caused the grant, empty for grants made by a payment callback.
	GrantedBy *uuid.UUID
	CreatedAt time.Time
	RevokedAt *time.Time
	RevokedBy *uuid.UUID
}

func (g *Grant) IsRevoked() bool {
	return g.RevokedAt != nil
}
