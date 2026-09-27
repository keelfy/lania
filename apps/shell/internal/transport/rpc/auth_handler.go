package rpc

import (
	"context"
	"errors"

	shellv1 "github.com/lania-smp/shell/internal/gen/lania/shell/v1"
	"github.com/lania-smp/shell/internal/services"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthHandler struct {
	shellv1.UnimplementedAuthServiceServer
	authService services.AuthService
}

func NewAuthHandler(authService services.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) SetPassword(ctx context.Context, req *shellv1.SetPasswordRequest) (*shellv1.SetPasswordResponse, error) {
	mcUUID, err := parseUUID(req.GetMinecraftUuid())
	if err != nil {
		return nil, err
	}

	err = h.authService.SetPassword(ctx, mcUUID, req.GetMinecraftUsername(), req.GetPasswordBcrypt())
	switch {
	case errors.Is(err, services.ErrInvalidPasswordHash), errors.Is(err, services.ErrInvalidUsername):
		return nil, status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, services.ErrPlayerNotRegistered):
		return nil, status.Error(codes.NotFound, err.Error())
	case errors.Is(err, services.ErrPlayerLicensed):
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	case err != nil:
		return nil, internalError(err)
	}
	return &shellv1.SetPasswordResponse{}, nil
}
