package plandomain

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID   int64
	UUID uuid.UUID
}

type Session struct {
	ID           int64
	UserID       int64
	SessionStart time.Time
	SessionEnd   time.Time
	AfkTime      time.Duration
}

type Playtime struct {
	TotalPlaytime     int64
	FirstSessionStart *int64
	LastSessionEnd    *int64
}
