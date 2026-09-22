package services

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/domain"
	"github.com/lania-smp/backend/internal/storage"
	sql "github.com/lania-smp/backend/internal/storage/main"
	"github.com/lania-smp/backend/internal/utils"
	ory "github.com/ory/client-go"
)

// fakeMergeQueries answers only the calls a profile merge makes. Everything else panics, which is fine as
// long as a test never exercises it.
type fakeMergeQueries struct {
	sql.Queries
	profiles           map[uuid.UUID]*domain.Profile
	liveSeasons        map[uuid.UUID][]string
	mergeDataCalls     int
	deleteProfileCalls int
	insertMergeCalls   int
	setRoleCalls       int
	lastRole           domain.Role
	notifications      []sql.InsertNotificationParams
}

func (q *fakeMergeQueries) FindProfileByID(_ context.Context, id uuid.UUID) (*domain.Profile, error) {
	if profile, ok := q.profiles[id]; ok {
		copied := *profile
		return &copied, nil
	}
	return nil, utils.NewNotFoundError("profile not found", nil)
}

func (q *fakeMergeQueries) LockProfilesForMerge(_ context.Context, _, _ uuid.UUID) error {
	return nil
}

func (q *fakeMergeQueries) FindLiveSyncedSeasonNamesWithPlaytime(_ context.Context, mcUUID uuid.UUID) ([]string, error) {
	return q.liveSeasons[mcUUID], nil
}

func (q *fakeMergeQueries) MergeProfileData(_ context.Context, _, _, _, _ uuid.UUID) (*domain.ProfileMergeCounts, error) {
	q.mergeDataCalls++
	return &domain.ProfileMergeCounts{PlaytimeMoved: 1}, nil
}

func (q *fakeMergeQueries) UpdateProfileAfterMerge(_ context.Context, _ uuid.UUID, _ *uuid.UUID, _, _ *time.Time, _ uuid.UUID) error {
	return nil
}

func (q *fakeMergeQueries) SetProfileRole(_ context.Context, _ uuid.UUID, role domain.Role, _ *uuid.UUID) error {
	q.setRoleCalls++
	q.lastRole = role
	return nil
}

func (q *fakeMergeQueries) DeleteProfile(_ context.Context, _ uuid.UUID) error {
	q.deleteProfileCalls++
	return nil
}

func (q *fakeMergeQueries) InsertProfileMerge(_ context.Context, _ sql.InsertProfileMergeParams) error {
	q.insertMergeCalls++
	return nil
}

func (q *fakeMergeQueries) InsertNotification(_ context.Context, arg sql.InsertNotificationParams) error {
	q.notifications = append(q.notifications, arg)
	return nil
}

// fakeMergeStorage runs BeginTx against the fake queries directly, with no real database.
type fakeMergeStorage struct {
	storage.MainStorage
	queries *fakeMergeQueries
}

// BeginTx simulates rollback: a real transaction would undo every write the closure made when it returns
// an error, so the fake restores the write counters it tracks in that case too.
func (s *fakeMergeStorage) BeginTx(_ context.Context, fn func(sql.Queries) error) error {
	before := *s.queries
	err := fn(s.queries)
	if err != nil {
		*s.queries = before
	}
	return err
}

func (s *fakeMergeStorage) Queries() sql.Queries {
	return s.queries
}

// stubMergeSeasonService reports no seasons, so resyncAfterMerge skips every shell call.
type stubMergeSeasonService struct {
	SeasonService
}

func (s *stubMergeSeasonService) GetSeasons(_ context.Context) ([]*domain.Season, error) {
	return nil, nil
}

type stubMergeProfileResyncService struct {
	ProfileResyncService
}

func (s *stubMergeProfileResyncService) ResyncProfile(_ context.Context, profileID uuid.UUID) (*domain.ProfileResync, error) {
	return &domain.ProfileResync{}, nil
}

func adminContext(adminID uuid.UUID) context.Context {
	session := &ory.Session{Identity: &ory.Identity{Id: adminID.String()}}
	return context.WithValue(context.Background(), "req.session", session)
}

func newTestMergeService(queries *fakeMergeQueries) AdminProfileMergeService {
	return NewAdminProfileMergeService(
		&fakeMergeStorage{queries: queries},
		&stubMergeSeasonService{},
		nil,
		&stubMergeProfileResyncService{},
		NewNotificationService(nil),
	)
}

func TestAdminProfileMergeServiceMergeProfiles(t *testing.T) {
	admin := uuid.New()
	owner := uuid.New()

	t.Run("moves data and deletes the source", func(t *testing.T) {
		source := &domain.Profile{ID: uuid.New(), MinecraftUUID: uuid.New(), MinecraftUsername: "player2", Role: domain.RolePlayer, OwnerUserID: &owner}
		target := &domain.Profile{ID: uuid.New(), MinecraftUUID: uuid.New(), MinecraftUsername: "player1", Role: domain.RolePlayer}

		queries := &fakeMergeQueries{profiles: map[uuid.UUID]*domain.Profile{source.ID: source, target.ID: target}}
		svc := newTestMergeService(queries)

		summary, _, err := svc.MergeProfiles(adminContext(admin), source.ID, target.ID)
		if err != nil {
			t.Fatalf("got error %v, want a successful merge", err)
		}
		if !summary.CanMerge() || len(summary.Blockers) != 0 {
			t.Errorf("got blockers %+v, want none", summary.Blockers)
		}
		if queries.mergeDataCalls != 1 || queries.deleteProfileCalls != 1 || queries.insertMergeCalls != 1 {
			t.Errorf("merge data / delete / audit calls = %d/%d/%d, want 1/1/1", queries.mergeDataCalls, queries.deleteProfileCalls, queries.insertMergeCalls)
		}
		if summary.OwnerUserID == nil || *summary.OwnerUserID != owner {
			t.Errorf("owner = %v, want %v carried over from the source", summary.OwnerUserID, owner)
		}
		if len(queries.notifications) != 1 {
			t.Errorf("notified %d times, want 1", len(queries.notifications))
		}
	})

	t.Run("picks the higher role of the two profiles", func(t *testing.T) {
		source := &domain.Profile{ID: uuid.New(), MinecraftUUID: uuid.New(), MinecraftUsername: "player2", Role: domain.RoleModerator}
		target := &domain.Profile{ID: uuid.New(), MinecraftUUID: uuid.New(), MinecraftUsername: "player1", Role: domain.RolePlayer}

		queries := &fakeMergeQueries{profiles: map[uuid.UUID]*domain.Profile{source.ID: source, target.ID: target}}
		svc := newTestMergeService(queries)

		summary, _, err := svc.MergeProfiles(adminContext(admin), source.ID, target.ID)
		if err != nil {
			t.Fatal(err)
		}
		if summary.RoleAfter != domain.RoleModerator {
			t.Errorf("role after merge = %s, want %s", summary.RoleAfter, domain.RoleModerator)
		}
		if queries.lastRole != domain.RoleModerator {
			t.Errorf("role written to storage = %s, want %s", queries.lastRole, domain.RoleModerator)
		}
	})

	t.Run("refuses to merge a profile into itself", func(t *testing.T) {
		profile := &domain.Profile{ID: uuid.New(), MinecraftUUID: uuid.New(), MinecraftUsername: "player1"}
		queries := &fakeMergeQueries{profiles: map[uuid.UUID]*domain.Profile{profile.ID: profile}}
		svc := newTestMergeService(queries)

		_, _, err := svc.MergeProfiles(adminContext(admin), profile.ID, profile.ID)
		if utils.MapCustomErrorToHttpStatus(err) != http.StatusConflict {
			t.Fatalf("got error %v, want a conflict", err)
		}
		if queries.mergeDataCalls != 0 || queries.deleteProfileCalls != 0 {
			t.Errorf("merge data / delete calls = %d/%d, want no writes", queries.mergeDataCalls, queries.deleteProfileCalls)
		}
	})

	t.Run("refuses profiles with different owners", func(t *testing.T) {
		ownerA, ownerB := uuid.New(), uuid.New()
		source := &domain.Profile{ID: uuid.New(), MinecraftUUID: uuid.New(), MinecraftUsername: "player2", OwnerUserID: &ownerA}
		target := &domain.Profile{ID: uuid.New(), MinecraftUUID: uuid.New(), MinecraftUsername: "player1", OwnerUserID: &ownerB}

		queries := &fakeMergeQueries{profiles: map[uuid.UUID]*domain.Profile{source.ID: source, target.ID: target}}
		svc := newTestMergeService(queries)

		_, _, err := svc.MergeProfiles(adminContext(admin), source.ID, target.ID)
		if utils.MapCustomErrorToHttpStatus(err) != http.StatusConflict {
			t.Fatalf("got error %v, want a conflict", err)
		}
		if queries.mergeDataCalls != 0 || queries.deleteProfileCalls != 0 {
			t.Errorf("merge data / delete calls = %d/%d, want no writes", queries.mergeDataCalls, queries.deleteProfileCalls)
		}
	})

	t.Run("refuses a source with playtime in a live season", func(t *testing.T) {
		source := &domain.Profile{ID: uuid.New(), MinecraftUUID: uuid.New(), MinecraftUsername: "player2"}
		target := &domain.Profile{ID: uuid.New(), MinecraftUUID: uuid.New(), MinecraftUsername: "player1"}

		queries := &fakeMergeQueries{
			profiles:    map[uuid.UUID]*domain.Profile{source.ID: source, target.ID: target},
			liveSeasons: map[uuid.UUID][]string{source.MinecraftUUID: {"Season 5"}},
		}
		svc := newTestMergeService(queries)

		_, _, err := svc.MergeProfiles(adminContext(admin), source.ID, target.ID)
		if utils.MapCustomErrorToHttpStatus(err) != http.StatusConflict {
			t.Fatalf("got error %v, want a conflict", err)
		}
		if queries.mergeDataCalls != 0 {
			t.Errorf("merge data calls = %d, want 0", queries.mergeDataCalls)
		}
	})
}

func TestAdminProfileMergeServicePreviewMergeProfiles(t *testing.T) {
	admin := uuid.New()
	source := &domain.Profile{ID: uuid.New(), MinecraftUUID: uuid.New(), MinecraftUsername: "player2"}
	target := &domain.Profile{ID: uuid.New(), MinecraftUUID: uuid.New(), MinecraftUsername: "player1"}

	queries := &fakeMergeQueries{profiles: map[uuid.UUID]*domain.Profile{source.ID: source, target.ID: target}}
	svc := newTestMergeService(queries)

	summary, err := svc.PreviewMergeProfiles(adminContext(admin), source.ID, target.ID)
	if err != nil {
		t.Fatalf("got error %v, want a preview", err)
	}
	if !summary.CanMerge() {
		t.Errorf("got blockers %+v, want none", summary.Blockers)
	}
	// A preview must never persist anything, even though the merge itself ran to compute the summary.
	if queries.deleteProfileCalls != 0 || queries.insertMergeCalls != 0 {
		t.Errorf("delete / audit calls = %d/%d, want 0/0 for a preview", queries.deleteProfileCalls, queries.insertMergeCalls)
	}
	if len(queries.notifications) != 0 {
		t.Errorf("notified %d times, want 0 for a preview", len(queries.notifications))
	}
}
