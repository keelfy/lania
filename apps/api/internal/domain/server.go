package domain

import (
	"time"

	"github.com/google/uuid"
)

type Season struct {
	ID           uuid.UUID
	SeasonNumber int
	StartDate    time.Time
	EndDate      *time.Time
	// relations
	Profiles          []*Profile
	ProfileAccesses   []*ProfileAccess
	ProfilePlaytimes  []*ProfilePlaytime
	ProfileViolations []*ProfileViolation
}
