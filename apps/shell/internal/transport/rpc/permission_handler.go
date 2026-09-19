package rpc

import (
	"context"

	shellv1 "github.com/lania-smp/shell/internal/gen/lania/shell/v1"
	"github.com/lania-smp/shell/internal/services"
)

type PermissionHandler struct {
	shellv1.UnimplementedPermissionServiceServer
	permissionService services.PermissionService
}

func NewPermissionHandler(permissionService services.PermissionService) *PermissionHandler {
	return &PermissionHandler{permissionService: permissionService}
}

func (h *PermissionHandler) GetPlayerGroups(ctx context.Context, req *shellv1.GetPlayerGroupsRequest) (*shellv1.GetPlayerGroupsResponse, error) {
	mcUUIDs, err := parseUUIDs(req.GetMinecraftUuids())
	if err != nil {
		return nil, err
	}

	groups, err := h.permissionService.GetPlayerGroups(ctx, mcUUIDs)
	if err != nil {
		return nil, internalError(err)
	}

	res := &shellv1.GetPlayerGroupsResponse{Groups: make(map[string]*shellv1.PlayerGroups, len(groups))}
	for mcUUID, names := range groups {
		res.Groups[mcUUID.String()] = &shellv1.PlayerGroups{Names: names}
	}
	return res, nil
}

func (h *PermissionHandler) ListPlayersByGroups(ctx context.Context, req *shellv1.ListPlayersByGroupsRequest) (*shellv1.ListPlayersByGroupsResponse, error) {
	mcUUIDs, err := h.permissionService.ListPlayersByGroups(ctx, req.GetGroups())
	if err != nil {
		return nil, internalError(err)
	}

	return &shellv1.ListPlayersByGroupsResponse{MinecraftUuids: mcUUIDs.Strings()}, nil
}

func (h *PermissionHandler) SetPlayerPrefix(ctx context.Context, req *shellv1.SetPlayerPrefixRequest) (*shellv1.SetPlayerPrefixResponse, error) {
	mcUUID, err := parseUUID(req.GetMinecraftUuid())
	if err != nil {
		return nil, err
	}

	if err := h.permissionService.SetPlayerPrefix(ctx, mcUUID, req.GetPrefix()); err != nil {
		return nil, internalError(err)
	}
	return &shellv1.SetPlayerPrefixResponse{}, nil
}
