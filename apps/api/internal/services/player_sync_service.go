package services

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/clients"
	"github.com/lania-smp/backend/internal/config"
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
	storage  storage.MainStorage
	shellAPI clients.ShellAPI
}

func NewPlayerSyncService(storage storage.MainStorage, shellAPI clients.ShellAPI) PlayerSyncService {
	return &playerSyncService{
		storage:  storage,
		shellAPI: shellAPI,
	}
}

func (s *playerSyncService) RunPlayerSync(ctx context.Context) {
	// cursorMs is the latest session end already synced. Zero means full sync.
	var cursorMs int64
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
			cursorMs = 0
			lastFullSync = time.Now()
		}

		sinceMs := max(cursorMs-playerSyncOverlap.Milliseconds(), 0)
		latestMs, err := s.syncPlayers(ctx, sinceMs)
		switch {
		case err == nil:
			cursorMs = max(cursorMs, latestMs)
		case errors.Is(err, context.Canceled):
		default:
			logger.Errorf(ctx, "[PLAYER SYNC] Failed to sync players: %v", err)
		}
	}
}

// syncPlayers stores players changed since sinceMs and returns the latest session end among them.
func (s *playerSyncService) syncPlayers(ctx context.Context, sinceMs int64) (int64, error) {
	playtimes, err := s.shellAPI.ListChangedPlaytimes(ctx, sinceMs)
	if err != nil {
		return 0, err
	}

	seasonID := config.GetPrimarySeasonID()
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
		logger.Debugf(ctx, "[PLAYER SYNC] Synced %d players changed since %d", len(playtimes), sinceMs)
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
