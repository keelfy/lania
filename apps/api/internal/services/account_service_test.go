package services

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/clients"
	"github.com/lania-smp/backend/internal/domain"
	"github.com/lania-smp/backend/internal/storage"
	sql "github.com/lania-smp/backend/internal/storage/main"
)

type accountQueries struct {
	sql.Queries
	revokedCosmetics []uuid.UUID
	keptColors       []uuid.UUID
	roles            map[uuid.UUID]domain.Role
	released         []uuid.UUID
	basketCleared    bool
	notesDeleted     bool
}

func (q *accountQueries) RevokeProfileCosmetics(_ context.Context, profileID, keepNameColorID uuid.UUID, _ *uuid.UUID) error {
	q.revokedCosmetics = append(q.revokedCosmetics, profileID)
	q.keptColors = append(q.keptColors, keepNameColorID)
	return nil
}

func (q *accountQueries) SetProfileRole(_ context.Context, profileID uuid.UUID, role domain.Role, _ *uuid.UUID) error {
	q.roles[profileID] = role
	return nil
}

func (q *accountQueries) SetProfileOwner(_ context.Context, profileID uuid.UUID, ownerUserID, _ *uuid.UUID) error {
	if ownerUserID == nil {
		q.released = append(q.released, profileID)
	}
	return nil
}

func (q *accountQueries) ClearBasketItemsByUserID(context.Context, uuid.UUID) error {
	q.basketCleared = true
	return nil
}

func (q *accountQueries) DeleteNotificationsByUserID(context.Context, uuid.UUID) error {
	q.notesDeleted = true
	return nil
}

type accountStorage struct {
	storage.MainStorage
	queries *accountQueries
}

func (s *accountStorage) Queries() sql.Queries { return s.queries }

func (s *accountStorage) BeginTx(_ context.Context, fn func(sql.Queries) error) error {
	return fn(s.queries)
}

type accountProfiles struct {
	ProfileService
	profiles []*domain.Profile
}

func (p *accountProfiles) GetProfilesByOwnerUserID(context.Context, uuid.UUID) ([]*domain.Profile, error) {
	return p.profiles, nil
}

type accountCosmetics struct {
	ProfileCosmeticsService
	prunedProfiles []uuid.UUID
}

func (c *accountCosmetics) PruneProfileSelections(_ context.Context, _ sql.Queries, profileID uuid.UUID) error {
	c.prunedProfiles = append(c.prunedProfiles, profileID)
	return nil
}

type accountResync struct {
	ProfileResyncService
	done chan uuid.UUID
}

func (r *accountResync) ResyncProfile(_ context.Context, profileID uuid.UUID) (*domain.ProfileResync, error) {
	r.done <- profileID
	return &domain.ProfileResync{}, nil
}

type accountOry struct {
	clients.OryAPI
	deleted   []string
	deleteErr error
}

func (o *accountOry) DeleteIdentity(_ context.Context, identityID string) error {
	o.deleted = append(o.deleted, identityID)
	return o.deleteErr
}

type accountFixture struct {
	svc       *accountService
	queries   *accountQueries
	cosmetics *accountCosmetics
	resync    *accountResync
	ory       *accountOry
	now       time.Time
}

func newAccountFixture(t *testing.T, profiles ...*domain.Profile) *accountFixture {
	t.Setenv("DEFAULT_NAME_COLOR_ID", defaultColorID.String())
	f := &accountFixture{
		queries:   &accountQueries{roles: map[uuid.UUID]domain.Role{}},
		cosmetics: &accountCosmetics{},
		resync:    &accountResync{done: make(chan uuid.UUID, len(profiles))},
		ory:       &accountOry{},
		now:       time.Date(2026, 9, 21, 12, 0, 0, 0, time.UTC),
	}
	f.svc = NewAccountService(
		&accountStorage{queries: f.queries},
		f.ory,
		&accountProfiles{profiles: profiles},
		f.cosmetics,
		f.resync,
	).(*accountService)
	f.svc.now = func() time.Time { return f.now }
	return f
}

var defaultColorID = uuid.MustParse("2628bf9d-5b7c-438b-900a-67753261a823")

func TestDeleteAccountReleasesProfilesAndDeletesIdentity(t *testing.T) {
	staff := &domain.Profile{ID: uuid.New(), Role: domain.RoleModerator}
	player := &domain.Profile{ID: uuid.New(), Role: domain.RolePlayer}
	f := newAccountFixture(t, staff, player)
	userID := uuid.New()

	err := f.svc.DeleteAccount(context.Background(), userID, domain.RoleAdmin, f.now.Add(-time.Minute))
	if err != nil {
		t.Fatal(err)
	}

	if len(f.queries.released) != 2 || len(f.queries.revokedCosmetics) != 2 {
		t.Errorf("released %d and revoked cosmetics of %d profiles, want 2 and 2", len(f.queries.released), len(f.queries.revokedCosmetics))
	}
	for _, kept := range f.queries.keptColors {
		if kept != defaultColorID {
			t.Errorf("kept color %s, want the default color", kept)
		}
	}
	if len(f.cosmetics.prunedProfiles) != 2 {
		t.Errorf("reset the selection of %d profiles, want both", len(f.cosmetics.prunedProfiles))
	}
	if role, ok := f.queries.roles[staff.ID]; !ok || role != domain.RolePlayer {
		t.Errorf("staff profile role is %q, want player", role)
	}
	if _, ok := f.queries.roles[player.ID]; ok {
		t.Error("the role of a player was written again, it pushes the role to every server for nothing")
	}
	if !f.queries.basketCleared || !f.queries.notesDeleted {
		t.Error("basket or notifications are left")
	}
	if len(f.ory.deleted) != 1 || f.ory.deleted[0] != userID.String() {
		t.Errorf("deleted identities %v, want only the user", f.ory.deleted)
	}

	for range 2 {
		select {
		case <-f.resync.done:
		case <-time.After(time.Second):
			t.Fatal("released profiles were not resynced")
		}
	}
}

func TestDeleteAccountRefusesOwnerAndOldSession(t *testing.T) {
	for name, tt := range map[string]struct {
		role     domain.Role
		signedIn time.Duration
		status   int
	}{
		"owner":             {domain.RoleOwner, time.Minute, http.StatusConflict},
		"old session":       {domain.RolePlayer, AccountDeletionMaxAuthAge + time.Second, http.StatusForbidden},
		"no sign in moment": {domain.RolePlayer, -1, http.StatusForbidden},
	} {
		t.Run(name, func(t *testing.T) {
			f := newAccountFixture(t, &domain.Profile{ID: uuid.New(), Role: domain.RolePlayer})
			authenticatedAt := f.now.Add(-tt.signedIn)
			if tt.signedIn < 0 {
				authenticatedAt = time.Time{}
			}

			err := f.svc.DeleteAccount(context.Background(), uuid.New(), tt.role, authenticatedAt)
			if statusOf(err) != tt.status {
				t.Fatalf("got %v, want status %d", err, tt.status)
			}
			if len(f.queries.released) != 0 || len(f.ory.deleted) != 0 {
				t.Error("something was deleted after a refusal")
			}
		})
	}
}

func TestDeleteAccountTreatsMissingIdentityAsDeleted(t *testing.T) {
	f := newAccountFixture(t)
	f.ory.deleteErr = clients.ErrIdentityNotFound

	if err := f.svc.DeleteAccount(context.Background(), uuid.New(), domain.RolePlayer, f.now); err != nil {
		t.Fatalf("got %v, want a repeated request to succeed", err)
	}
}
