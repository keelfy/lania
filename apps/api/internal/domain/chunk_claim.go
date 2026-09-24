package domain

import (
	"time"

	"github.com/google/uuid"
)

// DefaultClaimLimit is how many chunks one profile may claim in a world that sets no limit of its own.
const DefaultClaimLimit = 100

// MaxClaimBatch is the most chunks one request may claim or release.
const MaxClaimBatch = 256

// ChunkBlocks is the side of a chunk in blocks.
const ChunkBlocks = 16

// ChunkPos is a chunk index, the block coordinate divided by ChunkBlocks and rounded down, as F3 shows it.
type ChunkPos struct {
	X int
	Z int
}

// ChunkClaim is a chunk a player reserved on the site map. Nothing is protected in game, the claim only
// records who reserved the chunk and since when.
type ChunkClaim struct {
	ID      uuid.UUID
	WorldID uuid.UUID
	// Dimension is the squaremap world name of the world server, e.g. minecraft_overworld.
	Dimension string
	Chunk     ChunkPos
	ProfileID uuid.UUID
	ClaimedAt time.Time
	// Profile is populated by the query that loads the claim, with the id and the username only.
	Profile *Profile
}
