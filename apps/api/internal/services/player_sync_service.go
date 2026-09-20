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

func syncPlayer(ctx context.Context, queries sql.Queries, seasonID, mcUUID uuid.UUID, playtime *domain.Playtime) error {
	if err := queries.UpsertProfilePlaytime(ctx, mcUUID, seasonID, playtime.TotalPlaytime); err != nil {
		return err
	}
	return queries.UpdateProfileSeenAt(ctx, mcUUID, millisToTime(playtime.FirstSessionStart), millisToTime(playtime.LastSessionEnd))
}

func millisToTime(millis *int64) *time.Time {
	if millis == nil {
		return nil
	}
	t := time.UnixMilli(*millis)
	return &t
}
