package domain

import (
	"time"

	"github.com/google/uuid"
)

// ProfileMergeBlockerKind is why a profile merge cannot run.
type ProfileMergeBlockerKind string

const (
	// ProfileMergeBlockerSameProfile means the source and the target are the same profile.
	ProfileMergeBlockerSameProfile ProfileMergeBlockerKind = "same-profile"
	// ProfileMergeBlockerDifferentOwners means both profiles have an owner, and the owners differ.
	ProfileMergeBlockerDifferentOwners ProfileMergeBlockerKind = "different-owners"
	// ProfileMergeBlockerLiveSeasonPlaytime means the source has playtime in a season a shell still syncs,
	// so the sync would overwrite the merged sum with the source's own value again.
	ProfileMergeBlockerLiveSeasonPlaytime ProfileMergeBlockerKind = "live-season-playtime"
)

// ProfileMergeBlocker stops a merge from running. SeasonNames is set for ProfileMergeBlockerLiveSeasonPlaytime.
type ProfileMergeBlocker struct {
	Kind        ProfileMergeBlockerKind
	SeasonNames []string
}

// ProfileMergeCounts is how many rows of each table a merge moved, summed or dropped as a duplicate.
type ProfileMergeCounts struct {
	PlaytimeMoved            int64 `json:"playtimeMoved"`
	PlaytimeSummed           int64 `json:"playtimeSummed"`
	AccessesMoved            int64 `json:"accessesMoved"`
	AccessesDropped          int64 `json:"accessesDropped"`
	ViolationsMoved          int64 `json:"violationsMoved"`
	NameColorOptionsMoved    int64 `json:"nameColorOptionsMoved"`
	NameColorOptionsDropped  int64 `json:"nameColorOptionsDropped"`
	NamePrefixOptionsMoved   int64 `json:"namePrefixOptionsMoved"`
	NamePrefixOptionsDropped int64 `json:"namePrefixOptionsDropped"`
	SeasonCosmeticsMoved     int64 `json:"seasonCosmeticsMoved"`
	SeasonCosmeticsDropped   int64 `json:"seasonCosmeticsDropped"`
	PrefixesMoved            int64 `json:"prefixesMoved"`
	PrefixesDropped          int64 `json:"prefixesDropped"`
	OrderItemsMoved          int64 `json:"orderItemsMoved"`
	BasketItemsMoved         int64 `json:"basketItemsMoved"`
	BasketItemsDropped       int64 `json:"basketItemsDropped"`
	NotificationsRepointed   int64 `json:"notificationsRepointed"`
	ScreenshotAuthorsMoved   int64 `json:"screenshotAuthorsMoved"`
	ScreenshotAuthorsDropped int64 `json:"screenshotAuthorsDropped"`
	ChunkClaimsMoved         int64 `json:"chunkClaimsMoved"`
}

// ProfileMergeSummary is what a merge did, or would do in a preview.
type ProfileMergeSummary struct {
	SourceProfileID uuid.UUID              `json:"sourceProfileId"`
	SourceUsername  string                 `json:"sourceUsername"`
	TargetProfileID uuid.UUID              `json:"targetProfileId"`
	TargetUsername  string                 `json:"targetUsername"`
	RoleBefore      Role                   `json:"roleBefore"`
	RoleAfter       Role                   `json:"roleAfter"`
	OwnerUserID     *uuid.UUID             `json:"ownerUserId,omitempty"`
	Counts          ProfileMergeCounts     `json:"counts"`
	Blockers        []*ProfileMergeBlocker `json:"blockers,omitempty"`
}

// CanMerge tells whether nothing blocks the merge.
func (s *ProfileMergeSummary) CanMerge() bool {
	return len(s.Blockers) == 0
}

// ProfileMerge is an audit row: a profile that was merged into another one and then deleted.
// There is no relation to the source profile, it no longer exists once the merge is done.
type ProfileMerge struct {
	ID                  uuid.UUID
	SourceProfileID     uuid.UUID
	SourceMinecraftUUID uuid.UUID
	SourceUsername      string
	TargetProfileID     uuid.UUID
	MergedBy            uuid.UUID
	Summary             ProfileMergeCounts
	CreatedAt           time.Time
}
