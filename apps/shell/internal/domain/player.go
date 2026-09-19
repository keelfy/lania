package domain

type Playtime struct {
	// TotalMs is active playtime in milliseconds, AFK time excluded.
	TotalMs int64
	// FirstSeenMs is Unix epoch milliseconds of the first session start.
	FirstSeenMs *int64
	// LastSeenMs is Unix epoch milliseconds of the last session end.
	LastSeenMs *int64
}
