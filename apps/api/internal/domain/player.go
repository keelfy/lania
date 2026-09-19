package domain

// Playtime is live playtime on the current Minecraft server, in milliseconds.
type Playtime struct {
	TotalPlaytime     int64
	FirstSessionStart *int64
	LastSessionEnd    *int64
}
