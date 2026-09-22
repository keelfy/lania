package presenter

import (
	"testing"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/domain"
)

func TestPresentProfileMergeSummary(t *testing.T) {
	owner := uuid.New()
	summary := &domain.ProfileMergeSummary{
		SourceProfileID: uuid.New(), SourceUsername: "player2",
		TargetProfileID: uuid.New(), TargetUsername: "player1",
		RoleBefore: domain.RolePlayer, RoleAfter: domain.RoleModerator, OwnerUserID: &owner,
		Counts: domain.ProfileMergeCounts{PlaytimeMoved: 2, PlaytimeSummed: 1},
	}

	t.Run("a merge without blockers can run", func(t *testing.T) {
		got := PresentProfileMergeSummary(summary, nil)

		if !got.CanMerge || len(got.Blockers) != 0 {
			t.Errorf("got %+v, want no blockers", got)
		}
		if got.RoleBefore != "player" || got.RoleAfter != "mod" || got.OwnerUserID == nil || *got.OwnerUserID != owner {
			t.Errorf("got %+v", got)
		}
		if got.Counts.PlaytimeMoved != 2 || got.Counts.PlaytimeSummed != 1 {
			t.Errorf("got counts %+v", got.Counts)
		}
		if got.Resync != nil {
			t.Errorf("got resync %+v, want none without one", got.Resync)
		}
	})

	t.Run("a blocked merge cannot run", func(t *testing.T) {
		blocked := &domain.ProfileMergeSummary{
			SourceProfileID: summary.SourceProfileID, TargetProfileID: summary.TargetProfileID,
			Blockers: []*domain.ProfileMergeBlocker{{Kind: domain.ProfileMergeBlockerLiveSeasonPlaytime, SeasonNames: []string{"Season 5"}}},
		}

		got := PresentProfileMergeSummary(blocked, nil)

		if got.CanMerge || len(got.Blockers) != 1 {
			t.Fatalf("got %+v, want one blocker", got)
		}
		if got.Blockers[0].Kind != string(domain.ProfileMergeBlockerLiveSeasonPlaytime) || len(got.Blockers[0].SeasonNames) != 1 {
			t.Errorf("got blocker %+v", got.Blockers[0])
		}
	})

	t.Run("the resync report is attached once the merge ran", func(t *testing.T) {
		resync := &domain.ProfileResync{Seasons: []*domain.SeasonResync{{SeasonID: uuid.New(), SeasonName: "Season 5"}}}

		got := PresentProfileMergeSummary(summary, resync)

		if got.Resync == nil || !got.Resync.OK || len(got.Resync.Seasons) != 1 {
			t.Errorf("got resync %+v", got.Resync)
		}
	})
}

func TestPresentProfileMerges(t *testing.T) {
	merges := []*domain.ProfileMerge{
		{ID: uuid.New(), SourceProfileID: uuid.New(), SourceMinecraftUUID: uuid.New(), SourceUsername: "player2", MergedBy: uuid.New()},
	}

	got := PresentProfileMerges(merges)

	if len(got) != 1 || got[0].SourceUsername != "player2" {
		t.Errorf("got %+v", got)
	}
}
