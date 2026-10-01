package rpc

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/lania-smp/shell/internal/domain"
	shellv1 "github.com/lania-smp/shell/internal/gen/lania/shell/v1"
	"github.com/lania-smp/shell/internal/services"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type SkinHandler struct {
	shellv1.UnimplementedSkinServiceServer
	skinService services.SkinService
}

func NewSkinHandler(skinService services.SkinService) *SkinHandler {
	return &SkinHandler{skinService: skinService}
}

func (h *SkinHandler) GetPlayerSkins(ctx context.Context, req *shellv1.GetPlayerSkinsRequest) (*shellv1.GetPlayerSkinsResponse, error) {
	mcUUIDs, err := parseUUIDs(req.GetMinecraftUuids())
	if err != nil {
		return nil, err
	}

	skins, err := h.skinService.GetPlayerSkins(ctx, mcUUIDs)
	if err != nil {
		return nil, internalError(err)
	}

	res := &shellv1.GetPlayerSkinsResponse{Skins: make(map[string]*shellv1.PlayerSkin, len(skins))}
	for mcUUID, skin := range skins {
		res.Skins[mcUUID.String()] = toPlayerSkin(skin)
	}
	return res, nil
}

var skinVariants = map[shellv1.SkinVariant]string{
	shellv1.SkinVariant_SKIN_VARIANT_UNSPECIFIED: services.SkinVariantAny,
	shellv1.SkinVariant_SKIN_VARIANT_CLASSIC:     services.SkinVariantClassic,
	shellv1.SkinVariant_SKIN_VARIANT_SLIM:        services.SkinVariantSlim,
}

func (h *SkinHandler) SetPlayerSkin(ctx context.Context, req *shellv1.SetPlayerSkinRequest) (*shellv1.SetPlayerSkinResponse, error) {
	mcUUID, err := parseUUID(req.GetMinecraftUuid())
	if err != nil {
		return nil, err
	}
	var mojangUUID *uuid.UUID
	if req.MojangUuid != nil {
		parsed, err := parseUUID(req.GetMojangUuid())
		if err != nil {
			return nil, err
		}
		mojangUUID = &parsed
	}
	variant, ok := skinVariants[req.GetVariant()]
	if !ok {
		return nil, status.Errorf(codes.InvalidArgument, "unknown skin variant %d", req.GetVariant())
	}

	skin, err := h.skinService.SetPlayerSkin(ctx, mcUUID, req.GetSkin(), variant, mojangUUID)
	if err != nil {
		return nil, skinError(err)
	}
	return &shellv1.SetPlayerSkinResponse{Skin: toPlayerSkin(skin)}, nil
}

func (h *SkinHandler) ClearPlayerSkin(ctx context.Context, req *shellv1.ClearPlayerSkinRequest) (*shellv1.ClearPlayerSkinResponse, error) {
	mcUUID, err := parseUUID(req.GetMinecraftUuid())
	if err != nil {
		return nil, err
	}

	if err := h.skinService.ClearPlayerSkin(ctx, mcUUID); err != nil {
		return nil, skinError(err)
	}
	return &shellv1.ClearPlayerSkinResponse{}, nil
}

func skinError(err error) error {
	switch {
	case errors.Is(err, services.ErrInvalidSkin):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, services.ErrSkinNotApplied):
		return status.Error(codes.DeadlineExceeded, err.Error())
	default:
		return internalError(err)
	}
}

func toPlayerSkin(skin *domain.PlayerSkin) *shellv1.PlayerSkin {
	res := &shellv1.PlayerSkin{TextureUrl: skin.TextureURL, Slim: skin.Slim}
	if skin.MojangUUID != nil {
		mojangUUID := skin.MojangUUID.String()
		res.MojangUuid = &mojangUUID
	}
	return res
}
