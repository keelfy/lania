package services

import (
	"context"

	"github.com/google/uuid"
	plandomain "github.com/lania-smp/backend/internal/domain/plan"
	"github.com/lania-smp/backend/internal/storage"
	"github.com/lania-smp/backend/internal/utils"
)

type PlanService interface {
	CountPlaytimeByMinecaftUUIDs(ctx context.Context, mcUUIDs uuid.UUIDs) (map[uuid.UUID]*plandomain.Playtime, error)
}

type planService struct {
	storage storage.PlanStorage
}

func NewPlanService(storage storage.PlanStorage) PlanService {
	return &planService{
		storage: storage,
	}
}

func (s *planService) CountPlaytimeByMinecaftUUIDs(ctx context.Context, mcUUIDs uuid.UUIDs) (map[uuid.UUID]*plandomain.Playtime, error) {
	userIDMap, err := s.storage.Queries().FindPLANUserByMinecraftUUID(ctx, mcUUIDs)
	if err != nil {
		return nil, utils.NewInternalServerError("failed to find plan user by minecraft uuid", err)
	}

	userIDs := make([]int64, 0)
	for _, userID := range userIDMap {
		userIDs = append(userIDs, userID.ID)
	}

	playtimes, err := s.storage.Queries().FindPlaytimeByUserIDs(ctx, userIDs)
	if err != nil {
		return nil, utils.NewInternalServerError("failed to count playtime by minecraft uuid", err)
	}

	playtimeMap := make(map[uuid.UUID]*plandomain.Playtime)
	for _, mcUUID := range mcUUIDs {
		if _, ok := userIDMap[mcUUID]; !ok {
			playtimeMap[mcUUID] = &plandomain.Playtime{
				TotalPlaytime:     0,
				FirstSessionStart: nil,
				LastSessionEnd:    nil,
			}
			continue
		}
		playtimeMap[mcUUID] = playtimes[userIDMap[mcUUID].ID]
	}

	return playtimeMap, nil
}
