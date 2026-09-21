package domain

import "github.com/google/uuid"

// ResyncPart is a kind of data the season servers keep about a profile.
type ResyncPart string

const (
	ResyncPartRole      ResyncPart = "role"
	ResyncPartCosmetics ResyncPart = "cosmetics"
	ResyncPartAccess    ResyncPart = "access"
)

// PartResync is the outcome of writing one part of a profile to one season server.
type PartResync struct {
	Part ResyncPart
	// Err is nil when the part was written.
	Err error
}

// SeasonResync tells what happened on the server of one season.
type SeasonResync struct {
	SeasonID   uuid.UUID
	SeasonName string
	Parts      []PartResync
}

// OK tells whether every part was written to the season server.
func (r *SeasonResync) OK() bool {
	for _, part := range r.Parts {
		if part.Err != nil {
			return false
		}
	}
	return true
}

// ProfileResync is the outcome of a resync of one profile, for every active season that has a server.
type ProfileResync struct {
	Seasons []*SeasonResync
}

// OK tells whether every season server got every part.
func (r *ProfileResync) OK() bool {
	for _, season := range r.Seasons {
		if !season.OK() {
			return false
		}
	}
	return true
}
