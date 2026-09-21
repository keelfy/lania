package rpc

import (
	"context"
	"errors"

	shellv1 "github.com/lania-smp/shell/internal/gen/lania/shell/v1"
	"github.com/lania-smp/shell/internal/services"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type WhitelistHandler struct {
	shellv1.UnimplementedWhitelistServiceServer
	whitelistService services.WhitelistService
}

func NewWhitelistHandler(whitelistService services.WhitelistService) *WhitelistHandler {
	return &WhitelistHandler{whitelistService: whitelistService}
}

func (h *WhitelistHandler) AddPlayer(ctx context.Context, req *shellv1.AddPlayerRequest) (*shellv1.AddPlayerResponse, error) {
	mcUUID, err := parseUUID(req.GetMinecraftUuid())
	if err != nil {
		return nil, err
	}

	if err := h.whitelistService.AddPlayer(ctx, mcUUID, req.GetMinecraftUsername()); err != nil {
		return nil, whitelistError(err)
	}
	return &shellv1.AddPlayerResponse{}, nil
}

func (h *WhitelistHandler) RemovePlayer(ctx context.Context, req *shellv1.RemovePlayerRequest) (*shellv1.RemovePlayerResponse, error) {
	mcUUID, err := parseUUID(req.GetMinecraftUuid())
	if err != nil {
		return nil, err
	}

	if err := h.whitelistService.RemovePlayer(ctx, mcUUID, req.GetMinecraftUsername()); err != nil {
		return nil, whitelistError(err)
	}
	return &shellv1.RemovePlayerResponse{}, nil
}

func whitelistError(err error) error {
	if errors.Is(err, services.ErrInvalidUsername) {
		return status.Error(codes.InvalidArgument, err.Error())
	}
	return internalError(err)
}
