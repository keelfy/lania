package services

import (
	"context"
	stdsql "database/sql"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/domain"
	"github.com/lania-smp/backend/internal/storage"
	sql "github.com/lania-smp/backend/internal/storage/main"
	"github.com/lania-smp/backend/internal/utils"
)

// renameQueries keeps profiles, former UUIDs and nickname changes in memory.
type renameQueries struct {
	sql.Queries
	profiles    map[uuid.UUID]*domain.Profile
	mojangUUIDs map[uuid.UUID]*uuid.UUID
	// formerUUIDs maps a UUID a profile left to that profile.
	formerUUIDs map[uuid.UUID]uuid.UUID
	changes     []sql.InsertProfileUsernameChangeParams
	lastOwnerAt *time.Time
	conflicts   uuid.UUIDs
}

func (q *renameQueries) FindProfileByID(_ context.Context, profileID uuid.UUID) (*domain.Profile, error) {
	profile, ok := q.profiles[profileID]
	if !ok {
		return nil, stdsql.ErrNoRows
	}
	copied := *profile
	return &copied, nil
}

func (q *renameQueries) FindProfileByUsername(_ context.Context, username string) (*domain.Profile, error) {
	for _, profile := range q.profiles {
		if strings.EqualFold(profile.MinecraftUsername, username) {
			copied := *profile
			return &copied, nil
		}
	}
	return nil, stdsql.ErrNoRows
}

func (q *renameQueries) FindProfileMojangUUIDsByMinecraftUUIDs(_ context.Context, mcUUIDs uuid.UUIDs) (map[uuid.UUID]uuid.UUID, error) {
	res := make(map[uuid.UUID]uuid.UUID)
	for _, mcUUID := range mcUUIDs {
		if mojangUUID := q.mojangUUIDs[mcUUID]; mojangUUID != nil {
			res[mcUUID] = *mojangUUID
		}
	}
	return res, nil
}

func (q *renameQueries) LockProfile(context.Context, uuid.UUID) error { return nil }

func (q *renameQueries) FindLastProfileUsernameChangeAt(context.Context, uuid.UUID, domain.ProfileUsernameChangeSource) (*time.Time, error) {
	return q.lastOwnerAt, nil
}

func (q *renameQueries) FindProfileIDsHoldingMinecraftUUID(_ context.Context, mcUUID uuid.UUID) (uuid.UUIDs, error) {
	holders := uuid.UUIDs{}
	for _, profile := range q.profiles {
		if profile.MinecraftUUID == mcUUID {
			holders = append(holders, profile.ID)
		}
	}
	if profileID, ok := q.formerUUIDs[mcUUID]; ok {
		holders = append(holders, profileID)
	}
	return holders, nil
}

func (q *renameQueries) RenameProfile(_ context.Context, profileID, oldMcUUID, newMcUUID uuid.UUID, newUsername string, _ *uuid.UUID) (bool, error) {
	profile := q.profiles[profileID]
	if profile.MinecraftUUID != oldMcUUID {
		return false, nil
	}
	// The mc_uuid foreign keys cascade, the Mojang lookup row included.
	if lookup, ok := q.mojangUUIDs[oldMcUUID]; ok && oldMcUUID != newMcUUID {
		delete(q.mojangUUIDs, oldMcUUID)
		q.mojangUUIDs[newMcUUID] = lookup
	}
	profile.MinecraftUUID, profile.MinecraftUsername = newMcUUID, newUsername
	return true, nil
}

func (q *renameQueries) UpsertProfileMojangUUID(_ context.Context, mcUUID uuid.UUID, mojangUUID *uuid.UUID) error {
	q.mojangUUIDs[mcUUID] = mojangUUID
	return nil
}

func (q *renameQueries) InsertProfileFormerUUID(_ context.Context, mcUUID, profileID uuid.UUID) error {
	q.formerUUIDs[mcUUID] = profileID
	return nil
}

func (q *renameQueries) DeleteProfileFormerUUID(_ context.Context, mcUUID, profileID uuid.UUID) error {
	if q.formerUUIDs[mcUUID] == profileID {
		delete(q.formerUUIDs, mcUUID)
	}
	return nil
}

func (q *renameQueries) InsertProfileUsernameChange(_ context.Context, arg sql.InsertProfileUsernameChangeParams) error {
	q.changes = append(q.changes, arg)
	return nil
}

func (q *renameQueries) SetProfilePremiumConflict(_ context.Context, mcUUID uuid.UUID) error {
	q.conflicts = append(q.conflicts, mcUUID)
	return nil
}

func (q *renameQueries) FindPremiumNameCheckTargets(context.Context, time.Time, int) ([]*domain.Profile, error) {
	targets := make([]*domain.Profile, 0)
	for _, profile := range q.profiles {
		if mojangUUID := q.mojangUUIDs[profile.MinecraftUUID]; mojangUUID != nil && *mojangUUID == profile.MinecraftUUID {
			targets = append(targets, &domain.Profile{ID: profile.ID, MinecraftUUID: profile.MinecraftUUID, MinecraftUsername: profile.MinecraftUsername})
		}
	}
	return targets, nil
}

// renameStorage runs a transaction straight on the fake queries; no test here needs a rollback.
type renameStorage struct {
	storage.MainStorage
	queries *renameQueries
}

func (s *renameStorage) Queries() sql.Queries { return s.queries }

func (s *renameStorage) BeginTx(_ context.Context, fn func(sql.Queries) error) error {
	return fn(s.queries)
}

// renameMojang knows licensed accounts by lowercase name, and the current name of every account by UUID.
type renameMojang struct {
	MojangService
	accounts    map[string]uuid.UUID
	names       map[uuid.UUID]string
	unavailable bool
}

func (m *renameMojang) ResolveGameUUID(_ context.Context, username string) (uuid.UUID, *uuid.UUID, error) {
	if m.unavailable {
		return uuid.Nil, nil, utils.NewServiceUnavailableError("mojang is down", nil)
	}
	if mojangUUID, ok := m.accounts[strings.ToLower(username)]; ok {
		return mojangUUID, &mojangUUID, nil
	}
	offlineUUID, err := utils.GetOfflinePlayerUUID(username)
	return offlineUUID, nil, err
}

func (m *renameMojang) LookupUsernameByMojangUUID(_ context.Context, mojangUUID uuid.UUID) (string, error) {
	return m.names[mojangUUID], nil
}

func offlineUUID(t *testing.T, username string) uuid.UUID {
	t.Helper()
	mcUUID, err := utils.GetOfflinePlayerUUID(username)
	if err != nil {
		t.Fatal(err)
	}
	return mcUUID
}

func TestChangeOwnedUsername(t *testing.T) {
	ctx := context.Background()
	owner := uuid.New()
	now := time.Date(2026, 9, 27, 12, 0, 0, 0, time.UTC)
	licensedUUID, otherAccount := uuid.New(), uuid.New()

	type setup struct {
		queries  *renameQueries
		offline  *domain.Profile
		licensed *domain.Profile
		other    *domain.Profile
	}
	newSetup := func(t *testing.T) *setup {
		offline := &domain.Profile{ID: uuid.New(), MinecraftUUID: offlineUUID(t, "freenick"), MinecraftUsername: "freenick", OwnerUserID: &owner}
		licensed := &domain.Profile{ID: uuid.New(), MinecraftUUID: licensedUUID, MinecraftUsername: "OldPremium", OwnerUserID: &owner}
		other := &domain.Profile{ID: uuid.New(), MinecraftUUID: offlineUUID(t, "taken"), MinecraftUsername: "taken"}
		return &setup{
			offline: offline, licensed: licensed, other: other,
			queries: &renameQueries{
				profiles:    map[uuid.UUID]*domain.Profile{offline.ID: offline, licensed.ID: licensed, other.ID: other},
				mojangUUIDs: map[uuid.UUID]*uuid.UUID{offline.MinecraftUUID: nil, licensedUUID: &licensedUUID, other.MinecraftUUID: nil},
				formerUUIDs: map[uuid.UUID]uuid.UUID{},
			},
		}
	}
	mojang := &renameMojang{
		accounts: map[string]uuid.UUID{"newpremium": licensedUUID, "notch": otherAccount},
		names:    map[uuid.UUID]string{licensedUUID: "NewPremium"},
	}

	tests := []struct {
		name string
		// arrange changes the setup before the call; it returns the profile renamed and the nickname asked for.
		arrange func(t *testing.T, s *setup) (*domain.Profile, string)
		mojang  *renameMojang
		status  int
		// check runs after a successful rename.
		check func(t *testing.T, s *setup, renamed *domain.Profile)
	}{
		{
			name:    "offline profile moves to the offline uuid of the new nickname",
			arrange: func(_ *testing.T, s *setup) (*domain.Profile, string) { return s.offline, "brandnew" },
			check: func(t *testing.T, s *setup, renamed *domain.Profile) {
				newUUID := offlineUUID(t, "brandnew")
				if renamed.MinecraftUUID != newUUID || renamed.MinecraftUsername != "brandnew" {
					t.Errorf("renamed to %s (%s), want brandnew at its offline uuid", renamed.MinecraftUsername, renamed.MinecraftUUID)
				}
				if s.queries.formerUUIDs[offlineUUID(t, "freenick")] != s.offline.ID {
					t.Errorf("former uuids = %v, want the old uuid kept for the profile", s.queries.formerUUIDs)
				}
				if lookup, ok := s.queries.mojangUUIDs[newUUID]; !ok || lookup != nil {
					t.Errorf("mojang lookup of the new uuid = %v, want stored as not found", lookup)
				}
				if len(s.queries.changes) != 1 || s.queries.changes[0].Source != domain.ProfileUsernameChangeSourceOwner || *s.queries.changes[0].ChangedBy != owner {
					t.Errorf("changes = %+v, want one owner change", s.queries.changes)
				}
			},
		},
		{
			name: "going back to an own old nickname takes its uuid back",
			arrange: func(t *testing.T, s *setup) (*domain.Profile, string) {
				s.queries.formerUUIDs[offlineUUID(t, "mine")] = s.offline.ID
				return s.offline, "mine"
			},
			check: func(t *testing.T, s *setup, renamed *domain.Profile) {
				if _, kept := s.queries.formerUUIDs[offlineUUID(t, "mine")]; kept {
					t.Error("the uuid taken back is still a former uuid")
				}
				if renamed.MinecraftUUID != offlineUUID(t, "mine") {
					t.Errorf("uuid = %s, want the old one of the nickname", renamed.MinecraftUUID)
				}
			},
		},
		{
			name:    "a changed letter case is a new offline uuid too",
			arrange: func(_ *testing.T, s *setup) (*domain.Profile, string) { return s.offline, "FreeNick" },
			check: func(t *testing.T, _ *setup, renamed *domain.Profile) {
				if renamed.MinecraftUUID != offlineUUID(t, "FreeNick") {
					t.Errorf("uuid = %s, want the offline uuid of FreeNick", renamed.MinecraftUUID)
				}
			},
		},
		{
			name:    "licensed profile takes the current mojang name in its case",
			arrange: func(_ *testing.T, s *setup) (*domain.Profile, string) { return s.licensed, "newpremium" },
			check: func(t *testing.T, s *setup, renamed *domain.Profile) {
				if renamed.MinecraftUUID != licensedUUID || renamed.MinecraftUsername != "NewPremium" {
					t.Errorf("renamed to %s (%s), want NewPremium at the same uuid", renamed.MinecraftUsername, renamed.MinecraftUUID)
				}
				if len(s.queries.formerUUIDs) != 0 {
					t.Errorf("former uuids = %v, want none for a licensed rename", s.queries.formerUUIDs)
				}
			},
		},
		{
			name:    "licensed profile cannot take a name its account does not have",
			arrange: func(_ *testing.T, s *setup) (*domain.Profile, string) { return s.licensed, "Notch" },
			status:  http.StatusConflict,
		},
		{
			name: "licensed name held by another profile marks that profile",
			arrange: func(_ *testing.T, s *setup) (*domain.Profile, string) {
				s.other.MinecraftUsername = "newpremium"
				return s.licensed, "NewPremium"
			},
			status: http.StatusConflict,
			check: func(t *testing.T, s *setup, _ *domain.Profile) {
				if len(s.queries.conflicts) != 1 || s.queries.conflicts[0] != s.other.MinecraftUUID {
					t.Errorf("conflicts = %v, want the holder marked", s.queries.conflicts)
				}
			},
		},
		{
			name:    "offline profile cannot take a licensed nickname",
			arrange: func(_ *testing.T, s *setup) (*domain.Profile, string) { return s.offline, "Notch" },
			status:  http.StatusConflict,
		},
		{
			name:    "nickname of another profile is taken",
			arrange: func(_ *testing.T, s *setup) (*domain.Profile, string) { return s.offline, "TAKEN" },
			status:  http.StatusConflict,
		},
		{
			name: "nickname another profile left is reserved",
			arrange: func(t *testing.T, s *setup) (*domain.Profile, string) {
				s.queries.formerUUIDs[offlineUUID(t, "leftover")] = s.other.ID
				return s.offline, "leftover"
			},
			status: http.StatusConflict,
		},
		{
			name: "cooldown after the last change",
			arrange: func(_ *testing.T, s *setup) (*domain.Profile, string) {
				last := now.Add(-24 * time.Hour)
				s.queries.lastOwnerAt = &last
				return s.offline, "brandnew"
			},
			status: http.StatusTooManyRequests,
		},
		{
			name: "cooldown over",
			arrange: func(_ *testing.T, s *setup) (*domain.Profile, string) {
				last := now.Add(-31 * 24 * time.Hour)
				s.queries.lastOwnerAt = &last
				return s.offline, "brandnew"
			},
		},
		{
			name: "licensed profile has no cooldown",
			arrange: func(_ *testing.T, s *setup) (*domain.Profile, string) {
				last := now.Add(-time.Hour)
				s.queries.lastOwnerAt = &last
				return s.licensed, "NewPremium"
			},
		},
		{
			name:    "mojang unavailable",
			arrange: func(_ *testing.T, s *setup) (*domain.Profile, string) { return s.offline, "brandnew" },
			mojang:  &renameMojang{unavailable: true},
			status:  http.StatusServiceUnavailable,
		},
		{
			name:    "not the owner",
			arrange: func(_ *testing.T, s *setup) (*domain.Profile, string) { return s.other, "brandnew" },
			status:  http.StatusForbidden,
		},
		{
			name:    "invalid nickname",
			arrange: func(_ *testing.T, s *setup) (*domain.Profile, string) { return s.offline, "no spaces" },
			status:  http.StatusBadRequest,
		},
		{
			name:    "same nickname",
			arrange: func(_ *testing.T, s *setup) (*domain.Profile, string) { return s.offline, "freenick" },
			status:  http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newSetup(t)
			profile, username := tt.arrange(t, s)
			before := *profile
			resync := &stubMergeProfileResyncService{}
			service := &profileRenameService{
				storage:              &renameStorage{queries: s.queries},
				mojangService:        mojang,
				profileResyncService: resync,
				cooldown:             30 * 24 * time.Hour,
				now:                  func() time.Time { return now },
			}
			if tt.mojang != nil {
				service.mojangService = tt.mojang
			}

			renamed, err := service.ChangeOwnedUsername(ctx, owner, profile.ID, username)

			if tt.status != 0 {
				if utils.MapCustomErrorToHttpStatus(err) != tt.status {
					t.Fatalf("error = %v, want status %d", err, tt.status)
				}
				if s.queries.profiles[profile.ID].MinecraftUsername != before.MinecraftUsername || len(resync.moved) != 0 {
					t.Error("a refused rename changed the profile or the servers")
				}
			} else {
				if err != nil {
					t.Fatalf("error = %v, want a rename", err)
				}
				if resync.moved[profile.ID] != before.MinecraftUUID {
					t.Errorf("moved on servers = %v, want the profile moved away from %s", resync.moved, before.MinecraftUUID)
				}
			}
			if tt.check != nil {
				tt.check(t, s, renamed)
			}
		})
	}
}

func TestSyncPremiumNames(t *testing.T) {
	licensedUUID := uuid.New()
	profile := &domain.Profile{ID: uuid.New(), MinecraftUUID: licensedUUID, MinecraftUsername: "OldPremium"}
	queries := &renameQueries{
		profiles:    map[uuid.UUID]*domain.Profile{profile.ID: profile},
		mojangUUIDs: map[uuid.UUID]*uuid.UUID{licensedUUID: &licensedUUID},
		formerUUIDs: map[uuid.UUID]uuid.UUID{},
	}
	resync := &stubMergeProfileResyncService{}
	service := &profileRenameService{
		storage:              &renameStorage{queries: queries},
		mojangService:        &renameMojang{names: map[uuid.UUID]string{licensedUUID: "NewPremium"}},
		profileResyncService: resync,
		now:                  time.Now,
	}

	if err := service.syncPremiumNames(context.Background()); err != nil {
		t.Fatal(err)
	}

	if profile.MinecraftUsername != "NewPremium" || profile.MinecraftUUID != licensedUUID {
		t.Errorf("profile = %s (%s), want NewPremium at the same uuid", profile.MinecraftUsername, profile.MinecraftUUID)
	}
	if len(queries.changes) != 1 || queries.changes[0].Source != domain.ProfileUsernameChangeSourceMojang || queries.changes[0].ChangedBy != nil {
		t.Errorf("changes = %+v, want one change from mojang", queries.changes)
	}
	if _, moved := resync.moved[profile.ID]; !moved {
		t.Error("the new name was not registered on the servers")
	}
}
