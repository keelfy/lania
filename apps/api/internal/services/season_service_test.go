package services

import (
	"context"
	stdsql "database/sql"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/commands"
	"github.com/lania-smp/backend/internal/domain"
	"github.com/lania-smp/backend/internal/storage"
	sql "github.com/lania-smp/backend/internal/storage/main"
	"github.com/lania-smp/backend/internal/utils"
	"github.com/redis/go-redis/v9"
)

type fakeSeasonQueries struct {
	sql.Queries
	seasons map[uuid.UUID]*domain.Season
}

func (q *fakeSeasonQueries) FindSeasonByID(_ context.Context, id uuid.UUID) (*domain.Season, error) {
	return q.seasons[id], nil
}

func (q *fakeSeasonQueries) FindPrimarySeasonID(context.Context) (uuid.UUID, error) {
	for id, season := range q.seasons {
		if season.IsPrimary {
			return id, nil
		}
	}
	return uuid.Nil, stdsql.ErrNoRows
}

func (q *fakeSeasonQueries) FindSeasons(context.Context) ([]*domain.Season, error) {
	seasons := make([]*domain.Season, 0, len(q.seasons))
	for _, season := range q.seasons {
		seasons = append(seasons, season)
	}
	return seasons, nil
}

func (q *fakeSeasonQueries) SetSeasonShellAddress(_ context.Context, id uuid.UUID, address string) error {
	q.seasons[id].ShellAddress = &address
	return nil
}

func (q *fakeSeasonQueries) InsertSeason(_ context.Context, arg sql.InsertSeasonParams) error {
	q.seasons[arg.ID] = &domain.Season{ID: arg.ID, Name: arg.Name, ShellAddress: arg.ShellAddress, StartDate: arg.StartDate}
	return nil
}

func (q *fakeSeasonQueries) UpdateSeason(_ context.Context, arg sql.UpdateSeasonParams) error {
	season := q.seasons[arg.ID]
	season.Name = arg.Name
	season.StartDate = arg.StartDate
	season.IsActive = arg.IsActive
	season.Preregistration = arg.Preregistration
	season.FreeRegistration = arg.FreeRegistration
	season.ShellAddress = arg.ShellAddress
	return nil
}

func (q *fakeSeasonQueries) ClearPrimarySeasons(context.Context) error {
	for _, season := range q.seasons {
		season.IsPrimary = false
	}
	return nil
}

func (q *fakeSeasonQueries) SetSeasonPrimary(_ context.Context, id uuid.UUID) (bool, error) {
	season, found := q.seasons[id]
	if found {
		season.IsPrimary = true
	}
	return found, nil
}

type fakeCache struct {
	storage.CacheStorage
	values map[string]string
}

func newFakeCache() *fakeCache { return &fakeCache{values: map[string]string{}} }

func (c *fakeCache) GetKey(_ context.Context, key string) (string, error) {
	if value, ok := c.values[key]; ok {
		return value, nil
	}
	return "", redis.Nil
}

func (c *fakeCache) SetKey(_ context.Context, key string, value interface{}, _ time.Duration) error {
	c.values[key] = value.(string)
	return nil
}

func (c *fakeCache) DeleteKey(_ context.Context, key string) error {
	delete(c.values, key)
	return nil
}

type fakeSeasonStorage struct {
	storage.MainStorage
	queries *fakeSeasonQueries
}

func (s *fakeSeasonStorage) Queries() sql.Queries { return s.queries }

func (s *fakeSeasonStorage) BeginTx(_ context.Context, fn func(sql.Queries) error) error {
	return fn(s.queries)
}

func TestSeasonService_UpdateSeason_ActivityAndPrimaryAreIndependent(t *testing.T) {
	primaryID := uuid.New()
	secondaryID := uuid.New()
	start := time.Date(2026, 10, 9, 0, 0, 0, 0, time.UTC)
	queries := &fakeSeasonQueries{seasons: map[uuid.UUID]*domain.Season{
		primaryID: {
			ID: primaryID, Name: "Primary", StartDate: start,
			IsActive: true, IsPrimary: true,
		},
		secondaryID: {
			ID: secondaryID, Name: "Secondary", StartDate: start,
			IsActive: true,
		},
	}}
	service := NewSeasonService(&fakeSeasonStorage{queries: queries}, newFakeCache())

	updated, err := service.UpdateSeason(t.Context(), &commands.SaveSeasonCommand{
		ID: primaryID, Name: "Primary", StartDate: start,
		IsActive: false, IsPrimary: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated.IsActive || !updated.IsPrimary {
		t.Fatalf("updated season = %+v, want inactive primary", updated)
	}
	if !queries.seasons[secondaryID].IsActive {
		t.Fatal("updating primary deactivated another active season")
	}
}

func TestSeasonService_UpdateSeason_SelectsInactivePrimary(t *testing.T) {
	oldPrimaryID := uuid.New()
	newPrimaryID := uuid.New()
	start := time.Date(2026, 10, 9, 0, 0, 0, 0, time.UTC)
	queries := &fakeSeasonQueries{seasons: map[uuid.UUID]*domain.Season{
		oldPrimaryID: {ID: oldPrimaryID, Name: "Old", StartDate: start, IsPrimary: true},
		newPrimaryID: {ID: newPrimaryID, Name: "New", StartDate: start},
	}}
	service := NewSeasonService(&fakeSeasonStorage{queries: queries}, newFakeCache())

	updated, err := service.UpdateSeason(t.Context(), &commands.SaveSeasonCommand{
		ID: newPrimaryID, Name: "New", StartDate: start,
		IsPrimary: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated.IsActive || !updated.IsPrimary {
		t.Fatalf("updated season = %+v, want inactive primary", updated)
	}
	if queries.seasons[oldPrimaryID].IsPrimary {
		t.Fatal("old primary season remained primary")
	}
}

func TestSeasonService_GetPrimarySeasonID(t *testing.T) {
	primaryID := uuid.New()
	queries := &fakeSeasonQueries{seasons: map[uuid.UUID]*domain.Season{
		primaryID:  {ID: primaryID, IsPrimary: true},
		uuid.New(): {},
	}}
	cache := newFakeCache()
	service := NewSeasonService(&fakeSeasonStorage{queries: queries}, cache)

	got, err := service.GetPrimarySeasonID(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if got != primaryID {
		t.Fatalf("primary season id = %v, want %v", got, primaryID)
	}
	if cache.values[primarySeasonCacheKey] != primaryID.String() {
		t.Fatalf("cache = %v, want primary season id stored", cache.values)
	}

	queries.seasons[primaryID].IsPrimary = false
	if got, _ := service.GetPrimarySeasonID(t.Context()); got != primaryID {
		t.Fatal("cached primary season id was not served")
	}

	delete(cache.values, primarySeasonCacheKey)
	if _, err := service.GetPrimarySeasonID(t.Context()); err == nil {
		t.Fatal("expected error when no season is primary")
	}
}

func TestSeasonService_UpdateSeason_DropsPrimaryCache(t *testing.T) {
	oldPrimaryID := uuid.New()
	newPrimaryID := uuid.New()
	start := time.Date(2026, 10, 9, 0, 0, 0, 0, time.UTC)
	queries := &fakeSeasonQueries{seasons: map[uuid.UUID]*domain.Season{
		oldPrimaryID: {ID: oldPrimaryID, StartDate: start, IsPrimary: true},
		newPrimaryID: {ID: newPrimaryID, StartDate: start},
	}}
	cache := newFakeCache()
	service := NewSeasonService(&fakeSeasonStorage{queries: queries}, cache)

	if _, err := service.GetPrimarySeasonID(t.Context()); err != nil {
		t.Fatal(err)
	}
	if _, err := service.UpdateSeason(t.Context(), &commands.SaveSeasonCommand{
		ID: newPrimaryID, StartDate: start, IsPrimary: true,
	}); err != nil {
		t.Fatal(err)
	}

	got, err := service.GetPrimarySeasonID(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if got != newPrimaryID {
		t.Fatalf("primary season id = %v, want %v after switch", got, newPrimaryID)
	}
}

func TestSeasonService_ShellAddressBelongsToOneSeason(t *testing.T) {
	address := "shell-a:9000"
	otherAddress := "shell-b:9000"
	firstID, secondID := uuid.New(), uuid.New()
	start := time.Date(2026, 10, 9, 0, 0, 0, 0, time.UTC)
	newService := func() SeasonService {
		return NewSeasonService(&fakeSeasonStorage{queries: &fakeSeasonQueries{seasons: map[uuid.UUID]*domain.Season{
			firstID:  {ID: firstID, Name: "First", StartDate: start, ShellAddress: &address},
			secondID: {ID: secondID, Name: "Second", StartDate: start, ShellAddress: &otherAddress},
		}}}, newFakeCache())
	}

	t.Run("create rejects an address another season uses", func(t *testing.T) {
		_, err := newService().CreateSeason(t.Context(), &commands.SaveSeasonCommand{Name: "Third", StartDate: start, ShellAddress: &address})
		if utils.MapCustomErrorToHttpStatus(err) != http.StatusConflict || !strings.Contains(err.Error(), "First") {
			t.Fatalf("error = %v, want a conflict naming the season that holds the address", err)
		}
	})

	t.Run("update rejects an address another season uses", func(t *testing.T) {
		_, err := newService().UpdateSeason(t.Context(), &commands.SaveSeasonCommand{ID: secondID, Name: "Second", StartDate: start, ShellAddress: &address})
		if utils.MapCustomErrorToHttpStatus(err) != http.StatusConflict {
			t.Fatalf("error = %v, want a conflict", err)
		}
	})

	t.Run("update keeps the season's own address", func(t *testing.T) {
		if _, err := newService().UpdateSeason(t.Context(), &commands.SaveSeasonCommand{ID: firstID, Name: "Renamed", StartDate: start, ShellAddress: &address}); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("seasons without a shell do not collide", func(t *testing.T) {
		service := newService()
		for range 2 {
			if _, err := service.CreateSeason(t.Context(), &commands.SaveSeasonCommand{Name: "No shell", StartDate: start}); err != nil {
				t.Fatal(err)
			}
		}
	})
}

func TestSeasonService_SeedShellAddressSkipsAddressInUse(t *testing.T) {
	t.Setenv("SHELL_ADDRESS", "shell-a:9000")
	address := "shell-a:9000"
	primaryID, otherID := uuid.New(), uuid.New()
	start := time.Date(2026, 10, 9, 0, 0, 0, 0, time.UTC)
	queries := &fakeSeasonQueries{seasons: map[uuid.UUID]*domain.Season{
		primaryID: {ID: primaryID, Name: "Primary", StartDate: start, IsPrimary: true},
		otherID:   {ID: otherID, Name: "Other", StartDate: start, ShellAddress: &address},
	}}

	if err := NewSeasonService(&fakeSeasonStorage{queries: queries}, newFakeCache()).InitializePrimarySeason(t.Context()); err != nil {
		t.Fatal(err)
	}
	if queries.seasons[primaryID].ShellAddress != nil {
		t.Fatalf("primary season got address %s that another season serves", *queries.seasons[primaryID].ShellAddress)
	}
}
