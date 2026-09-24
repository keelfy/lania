package commands

import (
	"errors"
	"regexp"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/domain"
)

// dimensionPattern is the shape of a squaremap world name, e.g. minecraft_overworld.
var dimensionPattern = regexp.MustCompile(`^[a-z0-9_]{1,64}$`)

// maxChunkIndex is the last chunk inside the 30 million block world border.
const maxChunkIndex = 30_000_000 / domain.ChunkBlocks

// ClaimChunksCommand claims or releases chunks of one dimension of a world. ProfileID is the profile acting on
// its own claims; it is uuid.Nil for an admin releasing whoever holds the chunks.
type ClaimChunksCommand struct {
	WorldID   uuid.UUID
	UserID    uuid.UUID
	ProfileID uuid.UUID
	Dimension string
	Chunks    []domain.ChunkPos
}

func (c *ClaimChunksCommand) Validate() error {
	return validation.ValidateStruct(c,
		validation.Field(&c.Dimension, validation.Required, validation.Match(dimensionPattern)),
		validation.Field(&c.Chunks, validation.Required, validation.Length(1, domain.MaxClaimBatch),
			validation.By(func(any) error {
				seen := make(map[domain.ChunkPos]bool, len(c.Chunks))
				for _, chunk := range c.Chunks {
					if chunk.X < -maxChunkIndex || chunk.X > maxChunkIndex || chunk.Z < -maxChunkIndex || chunk.Z > maxChunkIndex {
						return errors.New("chunk is outside the world border")
					}
					if seen[chunk] {
						return errors.New("chunk is listed twice")
					}
					seen[chunk] = true
				}
				return nil
			}),
		),
	)
}
