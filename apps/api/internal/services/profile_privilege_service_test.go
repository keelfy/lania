package services

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/domain"
	"github.com/lania-smp/backend/internal/storage"
	sql "github.com/lania-smp/backend/internal/storage/main"
)

type privilegeAddCall struct {
	profile               *domain.Profile
	privilegeID, seasonID uuid.UUID
	orderItemID           *uuid.UUID
}

// recordingPrivileges stands in for ProfilePrivilegeService in the tests of the services that use it.
type recordingPrivileges struct {
	ProfilePrivilegeService
	added []privilegeAddCall
	// owned is what the profile has, whatever the season.
	owned []*domain.ProfilePrivilege
	// synced and cleared hold the season of every call.
	synced  []uuid.UUID
	cleared []uuid.UUID
	// clearedUUIDs holds the player of every clear.
	clearedUUIDs []uuid.UUID
	addErr       error
	syncErr      error
}

func (p *recordingPrivileges) AddProfilePrivilege(_ context.Context, _ sql.Queries, profile *domain.Profile, privilegeID, seasonID uuid.UUID, orderItemID *uuid.UUID) error {
	if p.addErr != nil {
		return p.addErr
	}
	p.added = append(p.added, privilegeAddCall{profile, privilegeID, seasonID, orderItemID})
	return nil
}

func (p *recordingPrivileges) GetProfilePrivileges(context.Context, uuid.UUID, uuid.UUID) ([]*domain.ProfilePrivilege, error) {
	return p.owned, nil
}

func (p *recordingPrivileges) GetProfilePrivilegesByOwnerUserID(context.Context, uuid.UUID, uuid.UUID) ([]*domain.ProfilePrivilege, error) {
	return p.owned, nil
}

func (p *recordingPrivileges) SyncProfileInSeason(_ context.Context, _ *domain.Profile, seasonID uuid.UUID) error {
	p.synced = append(p.synced, seasonID)
	return p.syncErr
}

func (p *recordingPrivileges) ClearPlayerInSeason(_ context.Context, mcUUID, seasonID uuid.UUID) error {
	p.cleared = append(p.cleared, seasonID)
	p.clearedUUIDs = append(p.clearedUUIDs, mcUUID)
	return nil
}

type privilegeQueries struct {
	sql.Queries
	catalog   []*domain.Privilege
	owned     []*domain.ProfilePrivilege
	inserted  []sql.InsertProfilePrivilegeParams
	insertErr error
}

func (q *privilegeQueries) FindPrivilegeByID(_ context.Context, id uuid.UUID) (*domain.Privilege, error) {
	for _, privilege := range q.catalog {
		if privilege.ID == id {
			return privilege, nil
		}
	}
	return nil, errors.New("privilege not found")
}

func (q *privilegeQueries) FindPrivileges(context.Context) ([]*domain.Privilege, error) {
	return q.catalog, nil
}

func (q *privilegeQueries) FindProfilePrivileges(context.Context, uuid.UUID, uuid.UUID) ([]*domain.ProfilePrivilege, error) {
	return q.owned, nil
}

func (q *privilegeQueries) FindProfilePrivilegesByOwnerUserID(context.Context, uuid.UUID, uuid.UUID) ([]*domain.ProfilePrivilege, error) {
	return q.owned, nil
}

func (q *privilegeQueries) InsertProfilePrivilege(_ context.Context, arg sql.InsertProfilePrivilegeParams) error {
	if q.insertErr != nil {
		return q.insertErr
	}
	q.inserted = append(q.inserted, arg)
	return nil
}

type privilegeStorage struct {
	storage.MainStorage
	queries *privilegeQueries
}

func (s *privilegeStorage) Queries() sql.Queries { return s.queries }

type privilegeCall struct {
	seasonID, mcUUID uuid.UUID
	add, remove      []string
}

type privilegeMinecraft struct {
	MinecraftService
	calls []privilegeCall
	err   error
}

func (m *privilegeMinecraft) SetPrivilegesInSeason(_ context.Context, seasonID, mcUUID uuid.UUID, add, remove []string) error {
	m.calls = append(m.calls, privilegeCall{seasonID, mcUUID, add, remove})
	return m.err
}

func TestProfilePrivilegeService(t *testing.T) {
	ctx := context.Background()
	profile := &domain.Profile{ID: uuid.New(), MinecraftUUID: uuid.New()}
	seasonID := uuid.New()
	homes := &domain.Privilege{ID: uuid.New(), Name: "Homes", Permission: "homes.commands.*"}
	fly := &domain.Privilege{ID: uuid.New(), Name: "Fly", Permission: "essentials.fly"}

	newService := func() (ProfilePrivilegeService, *privilegeQueries, *privilegeMinecraft) {
		queries := &privilegeQueries{catalog: []*domain.Privilege{homes, fly}}
		minecraft := &privilegeMinecraft{}
		return NewProfilePrivilegeService(&privilegeStorage{queries: queries}, minecraft), queries, minecraft
	}

	t.Run("adding stores the grant and writes the node to the season server", func(t *testing.T) {
		svc, queries, minecraft := newService()
		orderItemID := uuid.New()

		if err := svc.AddProfilePrivilege(ctx, queries, profile, homes.ID, seasonID, &orderItemID); err != nil {
			t.Fatal(err)
		}

		want := sql.InsertProfilePrivilegeParams{ProfileID: profile.ID, PrivilegeID: homes.ID, SeasonID: seasonID, OrderItemID: &orderItemID}
		if len(queries.inserted) != 1 || queries.inserted[0] != want {
			t.Errorf("inserted = %+v, want %+v", queries.inserted, want)
		}
		if len(minecraft.calls) != 1 {
			t.Fatalf("got %d server writes, want 1", len(minecraft.calls))
		}
		call := minecraft.calls[0]
		if call.seasonID != seasonID || call.mcUUID != profile.MinecraftUUID || !slices.Equal(call.add, []string{"homes.commands.*"}) || len(call.remove) != 0 {
			t.Errorf("server write = %+v, want only homes.commands.* added in the season", call)
		}
	})

	t.Run("an unreachable server does not fail the grant", func(t *testing.T) {
		svc, queries, minecraft := newService()
		minecraft.err = errors.New("shell down")

		if err := svc.AddProfilePrivilege(ctx, queries, profile, homes.ID, seasonID, nil); err != nil {
			t.Fatalf("got %v, want the grant kept", err)
		}
		if len(queries.inserted) != 1 {
			t.Errorf("inserted = %d, want 1", len(queries.inserted))
		}
	})

	t.Run("a grant that cannot be stored writes nothing to the server", func(t *testing.T) {
		svc, queries, minecraft := newService()
		queries.insertErr = errors.New("db down")

		if err := svc.AddProfilePrivilege(ctx, queries, profile, homes.ID, seasonID, nil); err == nil {
			t.Fatal("got no error")
		}
		if len(minecraft.calls) != 0 {
			t.Errorf("got %d server writes, want none", len(minecraft.calls))
		}
	})

	t.Run("a privilege that is not in the catalog fails", func(t *testing.T) {
		svc, queries, _ := newService()

		if err := svc.AddProfilePrivilege(ctx, queries, profile, uuid.New(), seasonID, nil); err == nil {
			t.Fatal("got no error")
		}
		if len(queries.inserted) != 0 {
			t.Errorf("inserted = %d, want none", len(queries.inserted))
		}
	})

	t.Run("sync adds owned nodes and removes the other catalog nodes", func(t *testing.T) {
		svc, queries, minecraft := newService()
		queries.owned = []*domain.ProfilePrivilege{{ProfileID: profile.ID, PrivilegeID: homes.ID, SeasonID: seasonID, Permission: homes.Permission}}

		if err := svc.SyncProfileInSeason(ctx, profile, seasonID); err != nil {
			t.Fatal(err)
		}
		if len(minecraft.calls) != 1 {
			t.Fatalf("got %d server writes, want 1", len(minecraft.calls))
		}
		call := minecraft.calls[0]
		if !slices.Equal(call.add, []string{"homes.commands.*"}) || !slices.Equal(call.remove, []string{"essentials.fly"}) {
			t.Errorf("add %v, remove %v; want homes added and fly removed", call.add, call.remove)
		}
	})

	t.Run("sync removes every node once the profile owns nothing", func(t *testing.T) {
		svc, _, minecraft := newService()

		if err := svc.SyncProfileInSeason(ctx, profile, seasonID); err != nil {
			t.Fatal(err)
		}
		call := minecraft.calls[0]
		if len(call.add) != 0 || !slices.Equal(call.remove, []string{"homes.commands.*", "essentials.fly"}) {
			t.Errorf("add %v, remove %v; want every node removed", call.add, call.remove)
		}
	})

	t.Run("sync with an empty catalog skips the server", func(t *testing.T) {
		svc, queries, minecraft := newService()
		queries.catalog = nil

		if err := svc.SyncProfileInSeason(ctx, profile, seasonID); err != nil {
			t.Fatal(err)
		}
		if len(minecraft.calls) != 0 {
			t.Errorf("got %d server writes, want none", len(minecraft.calls))
		}
	})

	t.Run("sync reports an unreachable server", func(t *testing.T) {
		svc, _, minecraft := newService()
		minecraft.err = errors.New("shell down")

		if err := svc.SyncProfileInSeason(ctx, profile, seasonID); err == nil {
			t.Fatal("got no error")
		}
	})

	t.Run("clearing a player removes every catalog node from the uuid", func(t *testing.T) {
		svc, _, minecraft := newService()
		old := uuid.New()

		if err := svc.ClearPlayerInSeason(ctx, old, seasonID); err != nil {
			t.Fatal(err)
		}
		call := minecraft.calls[0]
		if call.mcUUID != old || len(call.add) != 0 || !slices.Equal(call.remove, []string{"homes.commands.*", "essentials.fly"}) {
			t.Errorf("server write = %+v, want every node removed from the old uuid", call)
		}
	})
}
