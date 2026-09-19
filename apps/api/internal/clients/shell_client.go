package clients

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/config"
	"github.com/lania-smp/backend/internal/domain"
	shellv1 "github.com/lania-smp/backend/internal/gen/lania/shell/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

// ShellAPI is the only way the API reaches the Minecraft server and its plugins.
type ShellAPI interface {
	GetOnlineStatus(ctx context.Context, mcUUIDs uuid.UUIDs) (map[uuid.UUID]bool, error)
	// ListChangedPlaytimes returns playtime of players whose last session ended at or after sinceMs.
	ListChangedPlaytimes(ctx context.Context, sinceMs int64) (map[uuid.UUID]*domain.Playtime, error)
	GetPlayerGroups(ctx context.Context, mcUUIDs uuid.UUIDs) (map[uuid.UUID][]string, error)
	SetPlayerPrefix(ctx context.Context, mcUUID uuid.UUID, prefix string) error
	AddToWhitelist(ctx context.Context, mcUUID uuid.UUID, username string) error
	RemoveFromWhitelist(ctx context.Context, mcUUID uuid.UUID) error
}

type shellAPI struct {
	player     shellv1.PlayerServiceClient
	permission shellv1.PermissionServiceClient
	whitelist  shellv1.WhitelistServiceClient
}

// shellCallTimeout keeps pages responsive when the Minecraft host is slow.
const shellCallTimeout = 3 * time.Second

func NewShellAPI(ctx context.Context) (ShellAPI, func(), error) {
	// Shell runs on the same host inside a private docker network, so plaintext is fine.
	conn, err := grpc.NewClient(
		config.GetShellAddress(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(shellUnaryInterceptor),
	)
	if err != nil {
		return nil, nil, err
	}

	cleanup := func() {
		_ = conn.Close()
	}

	return &shellAPI{
		player:     shellv1.NewPlayerServiceClient(conn),
		permission: shellv1.NewPermissionServiceClient(conn),
		whitelist:  shellv1.NewWhitelistServiceClient(conn),
	}, cleanup, nil
}

func shellUnaryInterceptor(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	ctx, cancel := context.WithTimeout(ctx, shellCallTimeout)
	defer cancel()

	if token := config.GetShellToken(); token != "" {
		ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)
	}
	return invoker(ctx, method, req, reply, cc, opts...)
}

func (api *shellAPI) GetOnlineStatus(ctx context.Context, mcUUIDs uuid.UUIDs) (map[uuid.UUID]bool, error) {
	res, err := api.player.GetOnlineStatus(ctx, &shellv1.GetOnlineStatusRequest{MinecraftUuids: mcUUIDs.Strings()})
	if err != nil {
		return nil, err
	}

	online := make(map[uuid.UUID]bool, len(res.GetOnline()))
	for key, isOnline := range res.GetOnline() {
		mcUUID, err := uuid.Parse(key)
		if err != nil {
			return nil, err
		}
		online[mcUUID] = isOnline
	}
	return online, nil
}

func (api *shellAPI) ListChangedPlaytimes(ctx context.Context, sinceMs int64) (map[uuid.UUID]*domain.Playtime, error) {
	res, err := api.player.ListChangedPlaytimes(ctx, &shellv1.ListChangedPlaytimesRequest{SinceMs: sinceMs})
	if err != nil {
		return nil, err
	}

	playtimes := make(map[uuid.UUID]*domain.Playtime, len(res.GetPlaytimes()))
	for key, playtime := range res.GetPlaytimes() {
		mcUUID, err := uuid.Parse(key)
		if err != nil {
			return nil, err
		}
		playtimes[mcUUID] = &domain.Playtime{
			TotalPlaytime:     playtime.GetTotalMs(),
			FirstSessionStart: playtime.FirstSeenMs,
			LastSessionEnd:    playtime.LastSeenMs,
		}
	}
	return playtimes, nil
}

func (api *shellAPI) GetPlayerGroups(ctx context.Context, mcUUIDs uuid.UUIDs) (map[uuid.UUID][]string, error) {
	res, err := api.permission.GetPlayerGroups(ctx, &shellv1.GetPlayerGroupsRequest{MinecraftUuids: mcUUIDs.Strings()})
	if err != nil {
		return nil, err
	}

	groups := make(map[uuid.UUID][]string, len(res.GetGroups()))
	for key, playerGroups := range res.GetGroups() {
		mcUUID, err := uuid.Parse(key)
		if err != nil {
			return nil, err
		}
		groups[mcUUID] = playerGroups.GetNames()
	}
	return groups, nil
}

func (api *shellAPI) SetPlayerPrefix(ctx context.Context, mcUUID uuid.UUID, prefix string) error {
	_, err := api.permission.SetPlayerPrefix(ctx, &shellv1.SetPlayerPrefixRequest{
		MinecraftUuid: mcUUID.String(),
		Prefix:        prefix,
	})
	return err
}

func (api *shellAPI) AddToWhitelist(ctx context.Context, mcUUID uuid.UUID, username string) error {
	_, err := api.whitelist.AddPlayer(ctx, &shellv1.AddPlayerRequest{
		MinecraftUuid:     mcUUID.String(),
		MinecraftUsername: username,
	})
	return err
}

func (api *shellAPI) RemoveFromWhitelist(ctx context.Context, mcUUID uuid.UUID) error {
	_, err := api.whitelist.RemovePlayer(ctx, &shellv1.RemovePlayerRequest{MinecraftUuid: mcUUID.String()})
	return err
}
