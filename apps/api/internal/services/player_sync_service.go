package services

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/domain"
	"github.com/lania-smp/backend/internal/logger"
	"github.com/lania-smp/backend/internal/storage"
	sql "github.com/lania-smp/backend/internal/storage/main"
)

const (
	playerSyncInterval = time.Minute
	// playerSyncOverlap re-reads sessions near the cursor, because PLAN may save a session after a later one.
	playerSyncOverlap = 10 * time.Minute
	// playerSyncFullInterval re-reads every player, so profiles created after their first session are picked up too.
	playerSyncFullInterval = 24 * time.Hour
)

// PlayerSyncService copies player data of the Minecraft server into profiles.
type PlayerSyncService interface {
	// RunPlayerSync pulls playtime and seen dates from shell until ctx is done.
	RunPlayerSync(ctx context.Context)
}

type playerSyncService struct {
	storage          storage.MainStorage
	seasonService    SeasonService
	minecraftService MinecraftService
}

func NewPlayerSyncService(storage storage.MainStorage, seasonService SeasonService, minecraftService MinecraftService) PlayerSyncService {
	return &playerSyncService{
		storage:          storage,
		seasonService:    seasonService,
		minecraftService: minecraftService,
	}
}

func (s *playerSyncService) RunPlayerSync(ctx context.Context) {
	// cursorsMs holds the latest session end already synced per season. Missing means full sync.
	cursorsMs := make(map[uuid.UUID]int64)
	var lastFullSync time.Time
	wait := time.Duration(0)
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(wait):
		}

		wait = playerSyncInterval
		if time.Since(lastFullSync) >= playerSyncFullInterval {
			clear(cursorsMs)
			lastFullSync = time.Now()
		}

		seasons, err := s.seasonService.GetSeasons(ctx)
		if err != nil {
			if !errors.Is(err, context.Canceled) {
				logger.Errorf(ctx, "[PLAYER SYNC] Failed to list seasons: %v", err)
			}
			continue
		}

		for _, season := range seasons {
			// A season without shell has no server to read from.
			if !season.IsActive || season.ShellAddress == nil {
				continue
			}

			sinceMs := max(cursorsMs[season.ID]-playerSyncOverlap.Milliseconds(), 0)
			latestMs, err := s.syncPlayers(ctx, season.ID, sinceMs)
			switch {
			case err == nil:
				cursorsMs[season.ID] = max(cursorsMs[season.ID], latestMs)
			case errors.Is(err, context.Canceled):
			default:
				logger.Errorf(ctx, "[PLAYER SYNC] Failed to sync players of season %s: %v", season.ID, err)
			}
		}
	}
}

// syncPlayers stores players of the season changed since sinceMs and returns the latest session end among them.
func (s *playerSyncService) syncPlayers(ctx context.Context, seasonID uuid.UUID, sinceMs int64) (int64, error) {
	playtimes, err := s.minecraftService.ListChangedPlaytimes(ctx, seasonID, sinceMs)
	if err != nil {
		return 0, err
	}
	if err := s.addLegacyPlaytimes(ctx, seasonID, playtimes); err != nil {
		return 0, err
	}

	var latestMs int64
	err = s.storage.BeginTx(ctx, func(queries sql.Queries) error {
		for mcUUID, playtime := range playtimes {
			if err := syncPlayer(ctx, queries, seasonID, mcUUID, playtime); err != nil {
				return err
			}
			if playtime.LastSessionEnd != nil {
				latestMs = max(latestMs, *playtime.LastSessionEnd)
			}
		}
		return nil
	})
	if err != nil {
		return 0, err
	}

	if len(playtimes) > 0 {
		logger.Debugf(ctx, "[PLAYER SYNC] Synced %d players of season %s changed since %d", len(playtimes), seasonID, sinceMs)
	}
	return latestMs, nil
}

// addLegacyPlaytimes sums playtime a rekeyed profile played under its offline UUID into its current UUID.
// Either UUID may be missing from the changed ones, so both are read again for every such profile.
func (s *playerSyncService) addLegacyPlaytimes(ctx context.Context, seasonID uuid.UUID, playtimes map[uuid.UUID]*domain.Playtime) error {
	changed := make(uuid.UUIDs, 0, len(playtimes))
	for mcUUID := range playtimes {
		changed = append(changed, mcUUID)
	}
	legacy, err := s.storage.Queries().FindLegacyMinecraftUUIDs(ctx, changed)
	if err != nil || len(legacy) == 0 {
		return err
	}

	both := make(uuid.UUIDs, 0, 2*len(legacy))
	for legacyUUID, mcUUID := range legacy {
		both = append(both, legacyUUID, mcUUID)
	}
	fresh, err := s.minecraftService.GetPlaytimesInSeason(ctx, seasonID, both)
	if err != nil {
		return err
	}
	for legacyUUID, mcUUID := range legacy {
		delete(playtimes, legacyUUID)
		playtimes[mcUUID] = sumPlaytimes(fresh[mcUUID], fresh[legacyUUID])
	}
	return nil
}

// sumPlaytimes adds up playtime of one player under two UUIDs: the first session of both, the last session of both.
func sumPlaytimes(a, b *domain.Playtime) *domain.Playtime {
	if a == nil {
		a = &domain.Playtime{}
	}
	if b == nil {
		b = &domain.Playtime{}
	}
	sum := &domain.Playtime{TotalPlaytime: a.TotalPlaytime + b.TotalPlaytime}
	sum.FirstSessionStart = pickMillis(a.FirstSessionStart, b.FirstSessionStart, func(x, y int64) int64 { return min(x, y) })
	sum.LastSessionEnd = pickMillis(a.LastSessionEnd, b.LastSessionEnd, func(x, y int64) int64 { return max(x, y) })
	return sum
}

// pickMillis chooses between two optional dates, a missing one never wins.
func pickMillis(a, b *int64, choose func(x, y int64) int64) *int64 {
	switch {
	case a == nil:
		return b
	case b == nil:
		return a
	}
	picked := choose(*a, *b)
	return &picked
}

func syncPlayer(ctx context.Context, queries sql.Queries, seasonID, mcUUID uuid.UUID, playtime *domain.Playtime) error {
	lastSeenAt := millisToTime(playtime.LastSessionEnd)
	if err := queries.UpsertProfilePlaytime(ctx, mcUUID, seasonID, playtime.TotalPlaytime, lastSeenAt); err != nil {
		return err
	}
	// The profile keeps the latest date over every season for the admin panel, the public pages read the season one.
	return queries.UpdateProfileSeenAt(ctx, mcUUID, millisToTime(playtime.FirstSessionStart), lastSeenAt)
}

func millisToTime(millis *int64) *time.Time {
	if millis == nil {
		return nil
	}
	t := time.UnixMilli(*millis)
	return &t
}
