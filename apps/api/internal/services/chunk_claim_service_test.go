package services

import (
	"context"
	stdsql "database/sql"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/commands"
	"github.com/lania-smp/backend/internal/domain"
	sql "github.com/lania-smp/backend/internal/storage/main"
	"github.com/lania-smp/backend/internal/utils"
)

type fakeClaimQueries struct {
	sql.Queries
	season  *domain.Season
	worlds  map[uuid.UUID]*domain.SeasonWorld
	profile *domain.Profile
	// held is how many claims the profile holds, by world.
	held     map[uuid.UUID]int
	inserted []domain.ChunkPos
}

func (q *fakeClaimQueries) FindSeasonByID(context.Context, uuid.UUID) (*domain.Season, error) {
	return q.season, nil
}

func (q *fakeClaimQueries) FindSeasonWorldByID(_ context.Context, id uuid.UUID) (*domain.SeasonWorld, error) {
	if world, ok := q.worlds[id]; ok {
		return world, nil
	}
	return nil, stdsql.ErrNoRows
}

func (q *fakeClaimQueries) FindProfileByID(context.Context, uuid.UUID) (*domain.Profile, error) {
	return q.profile, nil
}

func (q *fakeClaimQueries) CheckIfProfileHasAccessBySeasonIDAndMinecraftUUID(context.Context, uuid.UUID, uuid.UUID) (bool, error) {
	return true, nil
}

func (q *fakeClaimQueries) LockProfile(context.Context, uuid.UUID) error { return nil }

func (q *fakeClaimQueries) CountActiveChunkClaimsByProfile(_ context.Context, _ uuid.UUID, worldID uuid.UUID) (int, error) {
	return q.held[worldID], nil
}

func (q *fakeClaimQueries) CountActiveChunkClaimsAt(context.Context, uuid.UUID, string, []domain.ChunkPos) (int, error) {
	return 0, nil
}

func (q *fakeClaimQueries) InsertChunkClaims(_ context.Context, _ uuid.UUID, _ string, _ uuid.UUID, chunks []domain.ChunkPos) error {
	q.inserted = append(q.inserted, chunks...)
	return nil
}

type fakeClaimStorage struct {
	fakeSeasonStorage
	queries *fakeClaimQueries
}

func (s *fakeClaimStorage) Queries() sql.Queries { return s.queries }

func (s *fakeClaimStorage) BeginTx(_ context.Context, fn func(sql.Queries) error) error {
	return fn(s.queries)
}

func TestChunkClaimService_ClaimChunks_WorldRules(t *testing.T) {
	userID := uuid.New()
	mapURL := "https://survival-map.lania.network"
	survival := &domain.SeasonWorld{
		ID: uuid.New(), MapURL: &mapURL, ClaimLimit: 3,
		ClaimDimensions: []string{domain.OverworldDimension},
	}
	farms := &domain.SeasonWorld{
		ID: uuid.New(), MapURL: &mapURL, ClaimLimit: 3,
		ClaimDimensions: []string{domain.OverworldDimension},
	}
	noMap := &domain.SeasonWorld{ID: uuid.New(), ClaimLimit: 3, ClaimDimensions: []string{domain.OverworldDimension}}
	twoChunks := []domain.ChunkPos{{X: 0, Z: 0}, {X: 1, Z: 0}}

	tests := []struct {
		name       string
		inactive   bool
		world      *domain.SeasonWorld
		dimension  string
		wantStatus int
	}{
		{name: "claims in the overworld", world: survival, dimension: domain.OverworldDimension},
		{name: "the limit is per world", world: farms, dimension: domain.OverworldDimension},
		{name: "over the limit of the world", world: survival, dimension: domain.OverworldDimension, wantStatus: http.StatusConflict},
		{name: "dimension without claims", world: survival, dimension: "minecraft_the_nether", wantStatus: http.StatusBadRequest},
		{name: "world without a map", world: noMap, dimension: domain.OverworldDimension, wantStatus: http.StatusBadRequest},
		{name: "inactive season", inactive: true, world: survival, dimension: domain.OverworldDimension, wantStatus: http.StatusBadRequest},
		{name: "unknown world", world: &domain.SeasonWorld{ID: uuid.New()}, dimension: domain.OverworldDimension, wantStatus: http.StatusNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			held := map[uuid.UUID]int{survival.ID: 2}
			if tt.name == "claims in the overworld" {
				held[survival.ID] = 1
			}
			queries := &fakeClaimQueries{
				season: &domain.Season{ID: uuid.New(), IsActive: !tt.inactive},
				worlds: map[uuid.UUID]*domain.SeasonWorld{
					survival.ID: survival, farms.ID: farms, noMap.ID: noMap,
				},
				profile: &domain.Profile{ID: uuid.New(), OwnerUserID: &userID},
				held:    held,
			}
			store := &fakeClaimStorage{queries: queries}
			seasons := NewSeasonService(store, newFakeCache())
			service := NewChunkClaimService(store, seasons, NewSeasonWorldService(store, seasons))

			err := service.ClaimChunks(t.Context(), &commands.ClaimChunksCommand{
				WorldID: tt.world.ID, UserID: userID, ProfileID: queries.profile.ID,
				Dimension: tt.dimension, Chunks: twoChunks,
			})
			if tt.wantStatus == 0 {
				if err != nil {
					t.Fatal(err)
				}
				if len(queries.inserted) != len(twoChunks) {
					t.Fatalf("inserted %d chunks, want %d", len(queries.inserted), len(twoChunks))
				}
				return
			}
			if status := utils.MapCustomErrorToHttpStatus(err); status != tt.wantStatus {
				t.Fatalf("status %d (%v), want %d", status, err, tt.wantStatus)
			}
			if len(queries.inserted) != 0 {
				t.Fatal("inserted chunks on a refused claim")
			}
		})
	}
}
