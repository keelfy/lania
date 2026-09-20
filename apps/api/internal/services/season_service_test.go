package services

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/commands"
	"github.com/lania-smp/backend/internal/domain"
	"github.com/lania-smp/backend/internal/storage"
	sql "github.com/lania-smp/backend/internal/storage/main"
)

type fakeSeasonQueries struct {
	sql.Queries
	seasons map[uuid.UUID]*domain.Season
}

func (q *fakeSeasonQueries) FindSeasonByID(_ context.Context, id uuid.UUID) (*domain.Season, error) {
	return q.seasons[id], nil
}

func (q *fakeSeasonQueries) UpdateSeason(_ context.Context, arg sql.UpdateSeasonParams) error {
	season := q.seasons[arg.ID]
	season.Name = arg.Name
	season.StartDate = arg.StartDate
	season.IsActive = arg.IsActive
	season.Preregistration = arg.Preregistration
	season.FreeRegistration = arg.FreeRegistration
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
	service := NewSeasonService(&fakeSeasonStorage{queries: queries})

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
	service := NewSeasonService(&fakeSeasonStorage{queries: queries})

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
