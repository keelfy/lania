package rpc

import (
	"context"

	shellv1 "github.com/lania-smp/shell/internal/gen/lania/shell/v1"
	"github.com/lania-smp/shell/internal/services"
)

type PlayerHandler struct {
	shellv1.UnimplementedPlayerServiceServer
	playerService services.PlayerService
}

func NewPlayerHandler(playerService services.PlayerService) *PlayerHandler {
	return &PlayerHandler{playerService: playerService}
}

func (h *PlayerHandler) GetOnlineStatus(ctx context.Context, req *shellv1.GetOnlineStatusRequest) (*shellv1.GetOnlineStatusResponse, error) {
	mcUUIDs, err := parseUUIDs(req.GetMinecraftUuids())
	if err != nil {
		return nil, err
	}

	online, err := h.playerService.GetOnlineStatus(ctx, mcUUIDs)
	if err != nil {
		return nil, internalError(err)
	}

	res := &shellv1.GetOnlineStatusResponse{Online: make(map[string]bool, len(online))}
	for mcUUID, isOnline := range online {
		res.Online[mcUUID.String()] = isOnline
	}
	return res, nil
}

func (h *PlayerHandler) GetPlaytime(ctx context.Context, req *shellv1.GetPlaytimeRequest) (*shellv1.GetPlaytimeResponse, error) {
	mcUUIDs, err := parseUUIDs(req.GetMinecraftUuids())
	if err != nil {
		return nil, err
	}

	playtimes, err := h.playerService.GetPlaytime(ctx, mcUUIDs)
	if err != nil {
		return nil, internalError(err)
	}

	res := &shellv1.GetPlaytimeResponse{Playtimes: make(map[string]*shellv1.Playtime, len(playtimes))}
	for mcUUID, playtime := range playtimes {
		res.Playtimes[mcUUID.String()] = &shellv1.Playtime{
			TotalMs:     playtime.TotalMs,
			FirstSeenMs: playtime.FirstSeenMs,
			LastSeenMs:  playtime.LastSeenMs,
		}
	}
	return res, nil
}
