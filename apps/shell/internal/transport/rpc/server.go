package rpc

import (
	"context"
	"crypto/subtle"
	"strings"

	"github.com/google/uuid"
	"github.com/lania-smp/shell/internal/config"
	shellv1 "github.com/lania-smp/shell/internal/gen/lania/shell/v1"
	"github.com/lania-smp/shell/internal/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func NewServer(
	playerHandler *PlayerHandler,
	permissionHandler *PermissionHandler,
	whitelistHandler *WhitelistHandler,
) *grpc.Server {
	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			loggingInterceptor,
			authInterceptor,
		),
	)

	shellv1.RegisterPlayerServiceServer(server, playerHandler)
	shellv1.RegisterPermissionServiceServer(server, permissionHandler)
	shellv1.RegisterWhitelistServiceServer(server, whitelistHandler)
	healthpb.RegisterHealthServer(server, health.NewServer())

	return server
}

func authInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	if strings.HasPrefix(info.FullMethod, "/"+healthpb.Health_ServiceDesc.ServiceName+"/") {
		return handler(ctx, req)
	}

	token := config.GetToken()
	if token == "" {
		return handler(ctx, req)
	}

	md, _ := metadata.FromIncomingContext(ctx)
	for _, value := range md.Get("authorization") {
		provided := strings.TrimPrefix(value, "Bearer ")
		if subtle.ConstantTimeCompare([]byte(provided), []byte(token)) == 1 {
			return handler(ctx, req)
		}
	}
	return nil, status.Error(codes.Unauthenticated, "invalid token")
}

func loggingInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	res, err := handler(ctx, req)
	if err != nil {
		logger.Errorf(ctx, "%s: %v", info.FullMethod, err)
	} else {
		logger.Debugf(ctx, "%s: ok", info.FullMethod)
	}
	return res, err
}

func parseUUIDs(values []string) (uuid.UUIDs, error) {
	mcUUIDs := make(uuid.UUIDs, len(values))
	for i, value := range values {
		mcUUID, err := parseUUID(value)
		if err != nil {
			return nil, err
		}
		mcUUIDs[i] = mcUUID
	}
	return mcUUIDs, nil
}

func parseUUID(value string) (uuid.UUID, error) {
	mcUUID, err := uuid.Parse(value)
	if err != nil {
		return uuid.Nil, status.Errorf(codes.InvalidArgument, "invalid minecraft uuid %q", value)
	}
	return mcUUID, nil
}

func internalError(err error) error {
	return status.Error(codes.Internal, err.Error())
}
