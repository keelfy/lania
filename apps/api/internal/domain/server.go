package domain

import (
	"time"

	"github.com/google/uuid"
)

type Season struct {
	ID   uuid.UUID
	Name string
	// PreviewImage is the screenshot shown on the seasons page. A season without one is not shown there.
	PreviewImage     *string
	StartDate        time.Time
	EndDate          *time.Time
	PublicAddress    *string
	SystemAddress    *string
	RCONPort         *uint16
	IsActive         bool
	IsPrimary        bool
	Preregistration  bool
	FreeRegistration bool
	RCONPassword     *string
	// relations
	Profiles          []*Profile
	ProfileAccesses   []*ProfileAccess
	ProfilePlaytimes  []*ProfilePlaytime
	ProfileViolations []*ProfileViolation
}
