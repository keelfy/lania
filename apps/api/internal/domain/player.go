package domain

// Playtime is playtime on the current Minecraft server as reported by shell, in milliseconds.
type Playtime struct {
	TotalPlaytime     int64
	FirstSessionStart *int64
	LastSessionEnd    *int64
	// Deaths and MobKills are summed over the sessions counted.
	Deaths   int64
	MobKills int64
}
