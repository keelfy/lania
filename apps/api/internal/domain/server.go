package domain

import (
	"time"

	"github.com/google/uuid"
)

type Season struct {
	ID   uuid.UUID
	Name string
	// PreviewImage is the screenshot shown on the seasons page. A season without one is not shown there.
	PreviewImage  *string
	StartDate     time.Time
	EndDate       *time.Time
	PublicAddress *string
	// ShellAddress is host:port of the shell service that reaches the season server.
	ShellAddress *string
	// HasServer tells that the API can reach the season server, so it knows who is online there.
	HasServer bool
	// PlanURL is the Plan web interface of the season server, shown to admins only.
	PlanURL          *string
	IsActive         bool
	IsPrimary        bool
	Preregistration  bool
	FreeRegistration bool
	// relations
	Profiles          []*Profile
	ProfileAccesses   []*ProfileAccess
	ProfilePlaytimes  []*ProfilePlaytime
	ProfileViolations []*ProfileViolation
}
