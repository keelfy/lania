package services

import (
	"context"
	"slices"
	"testing"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/domain"
	sql "github.com/lania-smp/backend/internal/storage/main"
	"github.com/lania-smp/backend/internal/utils"
)

type fakeRekeyQueries struct {
	sql.Queries
	targets []*domain.PremiumRekeyTarget
	// rekeyed maps a profile to the UUID it was moved to.
	rekeyed map[uuid.UUID]uuid.UUID
}

func (q *fakeRekeyQueries) FindPremiumRekeyTargets(context.Context) ([]*domain.PremiumRekeyTarget, error) {
	return q.targets, nil
}

func (q *fakeRekeyQueries) CountUncheckedMojangProfiles(context.Context) (int, error) { return 3, nil }

func (q *fakeRekeyQueries) RekeyProfile(_ context.Context, profileID, _, newMcUUID uuid.UUID) (bool, error) {
	q.rekeyed[profileID] = newMcUUID
	return true, nil
}

func TestRekeyPremiumProfiles(t *testing.T) {
	ctx := context.Background()
	offline, _ := utils.GetOfflinePlayerUUID("keelfy")
	premium := &domain.PremiumRekeyTarget{ProfileID: uuid.New(), MinecraftUUID: offline, MinecraftUsername: "keelfy", MojangUUID: uuid.New()}
	// Its UUID was set by hand, so it is neither the offline one nor the Mojang one.
	manual := &domain.PremiumRekeyTarget{ProfileID: uuid.New(), MinecraftUUID: uuid.New(), MinecraftUsername: "other", MojangUUID: uuid.New()}

	queries := &fakeRekeyQueries{targets: []*domain.PremiumRekeyTarget{premium, manual}, rekeyed: map[uuid.UUID]uuid.UUID{}}
	resync := &stubMergeProfileResyncService{}
	service := NewPremiumUUIDService(&stubMainStorage{queries: queries}, resync)

	report, err := service.RekeyPremiumProfiles(ctx)
	if err != nil {
		t.Fatalf("RekeyPremiumProfiles() error = %v", err)
	}

	if !slices.Equal(report.Rekeyed, []string{"keelfy"}) || !slices.Equal(report.Skipped, []string{"other"}) || report.Unchecked != 3 {
		t.Errorf("report = %+v, want keelfy rekeyed, other skipped, 3 unchecked", report)
	}
	if queries.rekeyed[premium.ProfileID] != premium.MojangUUID {
		t.Errorf("profile moved to %v, want the Mojang UUID", queries.rekeyed[premium.ProfileID])
	}
	if _, moved := queries.rekeyed[manual.ProfileID]; moved {
		t.Error("a profile with a hand-set UUID was moved")
	}
	if len(resync.moved) != 1 || resync.moved[premium.ProfileID] != offline {
		t.Errorf("moved on servers = %v, want only keelfy away from the offline UUID", resync.moved)
	}
}
