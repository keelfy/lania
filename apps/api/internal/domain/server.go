package domain

import (
	"time"

	"github.com/google/uuid"
)

type Season struct {
	ID           uuid.UUID
	SeasonNumber int
	Name         string
	// PreviewImage is the screenshot shown on the seasons page. A season without one is not shown there.
	PreviewImage *string
	StartDate    time.Time
	EndDate      *time.Time
	ServerIP     *string
	ServerPort   *uint16
	IsActive     bool
	RCONPassword *string
	// relations
	Profiles          []*Profile
	ProfileAccesses   []*ProfileAccess
	ProfilePlaytimes  []*ProfilePlaytime
	ProfileViolations []*ProfileViolation
}
