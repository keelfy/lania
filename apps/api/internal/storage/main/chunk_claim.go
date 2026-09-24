package sql

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/domain"
)

// chunkPositionsFilter returns "(chunk_x, chunk_z) IN ((?, ?), …)" and its arguments.
func chunkPositionsFilter(chunks []domain.ChunkPos) (string, []any) {
	placeholders := make([]string, len(chunks))
	args := make([]any, 0, len(chunks)*2)
	for i, chunk := range chunks {
		placeholders[i] = "(?, ?)"
		args = append(args, chunk.X, chunk.Z)
	}
	return "(chunk_x, chunk_z) IN (" + strings.Join(placeholders, ", ") + ")", args
}

const findActiveChunkClaims = `
SELECT c.id, c.season_id, c.world, c.chunk_x, c.chunk_z, c.profile_id, c.claimed_at, p.mc_username
FROM chunk_claims c
JOIN profiles p ON p.id = c.profile_id
WHERE c.season_id = ? AND c.world = ? AND c.active = 1
`

// FindActiveChunkClaims returns every claim that holds in the world of the season, each with its profile.
func (q *queries) FindActiveChunkClaims(ctx context.Context, seasonID uuid.UUID, world string) ([]*domain.ChunkClaim, error) {
	rows, err := q.x.QueryContext(ctx, findActiveChunkClaims, seasonID, world)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	claims := make([]*domain.ChunkClaim, 0)
	for rows.Next() {
		claim := &domain.ChunkClaim{Profile: &domain.Profile{}}
		err := rows.Scan(
			&claim.ID, &claim.SeasonID, &claim.World, &claim.Chunk.X, &claim.Chunk.Z,
			&claim.ProfileID, &claim.ClaimedAt, &claim.Profile.MinecraftUsername,
		)
		if err != nil {
			return nil, err
		}
		claim.Profile.ID = claim.ProfileID
		claims = append(claims, claim)
	}
	return claims, rows.Err()
}

const countActiveChunkClaimsByProfile = `
SELECT COUNT(*) FROM chunk_claims WHERE profile_id = ? AND season_id = ? AND active = 1
`

func (q *queries) CountActiveChunkClaimsByProfile(ctx context.Context, profileID, seasonID uuid.UUID) (int, error) {
	var count int
	err := q.x.QueryRowContext(ctx, countActiveChunkClaimsByProfile, profileID, seasonID).Scan(&count)
	return count, err
}

const countActiveChunkClaimsAt = `
SELECT COUNT(*) FROM chunk_claims WHERE season_id = ? AND world = ? AND active = 1 AND `

// CountActiveChunkClaimsAt returns how many of the chunks are already claimed by anyone.
func (q *queries) CountActiveChunkClaimsAt(ctx context.Context, seasonID uuid.UUID, world string, chunks []domain.ChunkPos) (int, error) {
	filter, filterArgs := chunkPositionsFilter(chunks)
	args := append([]any{seasonID, world}, filterArgs...)
	var count int
	err := q.x.QueryRowContext(ctx, countActiveChunkClaimsAt+filter, args...).Scan(&count)
	return count, err
}

// InsertChunkClaims claims every chunk for the profile in one statement. A chunk already claimed fails the
// whole statement on idx_chunk_claims_active.
func (q *queries) InsertChunkClaims(ctx context.Context, seasonID uuid.UUID, world string, profileID uuid.UUID, chunks []domain.ChunkPos) error {
	placeholders := make([]string, len(chunks))
	args := make([]any, 0, len(chunks)*5)
	for i, chunk := range chunks {
		placeholders[i] = "(?, ?, ?, ?, ?)"
		args = append(args, seasonID, world, chunk.X, chunk.Z, profileID)
	}
	query := "INSERT INTO chunk_claims (season_id, world, chunk_x, chunk_z, profile_id) VALUES " + strings.Join(placeholders, ", ")
	_, err := q.x.ExecContext(ctx, query, args...)
	return err
}

const releaseChunkClaims = `
UPDATE chunk_claims SET released_at = now(), released_by = ?
WHERE season_id = ? AND world = ? AND active = 1 AND `

// ReleaseChunkClaims ends the active claims on the chunks and returns how many it ended. A non-nil profileID
// limits it to the claims of that profile.
func (q *queries) ReleaseChunkClaims(ctx context.Context, seasonID uuid.UUID, world string, profileID *uuid.UUID, chunks []domain.ChunkPos, releasedBy uuid.UUID) (int64, error) {
	filter, filterArgs := chunkPositionsFilter(chunks)
	query := releaseChunkClaims + filter
	args := append([]any{releasedBy, seasonID, world}, filterArgs...)
	if profileID != nil {
		query += " AND profile_id = ?"
		args = append(args, *profileID)
	}
	result, err := q.x.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

const lockProfile = `
SELECT id FROM profiles WHERE id = ? FOR UPDATE
`

// LockProfile holds the profile row until the transaction ends, so two claims of one profile count its
// claims one after another and cannot both slip under the limit.
func (q *queries) LockProfile(ctx context.Context, profileID uuid.UUID) error {
	var id uuid.UUID
	return q.x.QueryRowContext(ctx, lockProfile, profileID).Scan(&id)
}
