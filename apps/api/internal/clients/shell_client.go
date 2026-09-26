package clients

import (
	"context"
	"sync"
	"time"

	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/config"
	"github.com/lania-smp/backend/internal/domain"
	shellv1 "github.com/lania-smp/backend/internal/gen/lania/shell/v1"
	"github.com/lania-smp/backend/internal/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// ShellAPI is the only way the API reaches a Minecraft server and its plugins. One ShellAPI serves one season.
type ShellAPI interface {
	GetOnlineStatus(ctx context.Context, mcUUIDs uuid.UUIDs) (map[uuid.UUID]bool, error)
	// ListOnlinePlayers returns every player that is online right now.
	ListOnlinePlayers(ctx context.Context) (uuid.UUIDs, error)
	// ListChangedPlaytimes returns playtime of players whose last session ended at or after sinceMs.
	ListChangedPlaytimes(ctx context.Context, sinceMs int64) (map[uuid.UUID]*domain.Playtime, error)
	// GetPlaytimes returns playtime of every player asked for, zero for one that never played. A Plan server name
	// counts only that server of the network; nil counts every server.
	GetPlaytimes(ctx context.Context, mcUUIDs uuid.UUIDs, serverName *string) (map[uuid.UUID]*domain.Playtime, error)
	SetPlayerPrefix(ctx context.Context, mcUUID uuid.UUID, prefix string) error
	// SetPlayerRoles makes every player belong to exactly the role group it maps to among roleGroups.
	// An empty group leaves the player in no role group.
	SetPlayerRoles(ctx context.Context, roleGroups []string, roles map[uuid.UUID]string) error
	// RegisterPlayer makes LuckPerms resolve the username to the player before the first join.
	RegisterPlayer(ctx context.Context, mcUUID uuid.UUID, username string) error
	AddToWhitelist(ctx context.Context, mcUUID uuid.UUID, username string) error
	// RemoveFromWhitelist forbids the player to join. It does nothing for a player that is not whitelisted.
	RemoveFromWhitelist(ctx context.Context, mcUUID uuid.UUID, username string) error
}

// ShellPool gives the ShellAPI of a season by the address of its shell service.
type ShellPool interface {
	Get(address string) (ShellAPI, error)
}

type shellPool struct {
	mu    sync.Mutex
	conns map[string]*grpc.ClientConn
	apis  map[string]ShellAPI
}

type shellAPI struct {
	player     shellv1.PlayerServiceClient
	permission shellv1.PermissionServiceClient
	whitelist  shellv1.WhitelistServiceClient
}

// shellCallTimeout keeps pages responsive when the Minecraft host is slow.
const shellCallTimeout = 3 * time.Second

func NewShellPool(ctx context.Context) (ShellPool, func(), error) {
	pool := &shellPool{
		conns: make(map[string]*grpc.ClientConn),
		apis:  make(map[string]ShellAPI),
	}
	cleanup := func() {
		pool.mu.Lock()
		defer pool.mu.Unlock()
		for _, conn := range pool.conns {
			_ = conn.Close()
		}
	}
	return pool, cleanup, nil
}

// Get connects on first use of an address. gRPC dials lazily, so an unreachable shell fails at the call.
func (p *shellPool) Get(address string) (ShellAPI, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if api, ok := p.apis[address]; ok {
		return api, nil
	}

	// Shell runs on the same host as its Minecraft server inside a private docker network, so plaintext is fine.
	conn, err := grpc.NewClient(
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(shellLoggingInterceptor, shellUnaryInterceptor),
	)
	if err != nil {
		return nil, err
	}

	api := &shellAPI{
		player:     shellv1.NewPlayerServiceClient(conn),
		permission: shellv1.NewPermissionServiceClient(conn),
		whitelist:  shellv1.NewWhitelistServiceClient(conn),
	}
	p.conns[address] = conn
	p.apis[address] = api
	return api, nil
}

// shellLoggingInterceptor logs every call to a shell. Errors are warnings: callers decide whether a failure matters.
func shellLoggingInterceptor(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	start := time.Now()
	err := invoker(ctx, method, req, reply, cc, opts...)
	elapsed := time.Since(start)

	if code := status.Code(err); code == codes.OK {
		logger.Infof(ctx, "[SHELL] %s %s: %s in %s", cc.Target(), method, code, elapsed)
	} else {
		logger.Warnf(ctx, "[SHELL] %s %s: %s in %s: %v", cc.Target(), method, code, elapsed, err)
	}
	return err
}

func shellUnaryInterceptor(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	ctx, cancel := context.WithTimeout(ctx, shellCallTimeout)
	defer cancel()

	if token := config.GetShellToken(); token != "" {
		ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)
	}
	// The shell tags its logs with the id, so a page request can be followed across both services.
	if requestID := chiMiddleware.GetReqID(ctx); requestID != "" {
		ctx = metadata.AppendToOutgoingContext(ctx, "x-request-id", requestID)
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

func (api *shellAPI) ListOnlinePlayers(ctx context.Context) (uuid.UUIDs, error) {
	res, err := api.player.ListOnlinePlayers(ctx, &shellv1.ListOnlinePlayersRequest{})
	if err != nil {
		return nil, err
	}
	return parseUUIDs(res.GetMinecraftUuids())
}

func parseUUIDs(values []string) (uuid.UUIDs, error) {
	mcUUIDs := make(uuid.UUIDs, len(values))
	for i, value := range values {
		mcUUID, err := uuid.Parse(value)
		if err != nil {
			return nil, err
		}
		mcUUIDs[i] = mcUUID
	}
	return mcUUIDs, nil
}

func (api *shellAPI) ListChangedPlaytimes(ctx context.Context, sinceMs int64) (map[uuid.UUID]*domain.Playtime, error) {
	res, err := api.player.ListChangedPlaytimes(ctx, &shellv1.ListChangedPlaytimesRequest{SinceMs: sinceMs})
	if err != nil {
		return nil, err
	}
	return parsePlaytimes(res.GetPlaytimes())
}

func (api *shellAPI) GetPlaytimes(ctx context.Context, mcUUIDs uuid.UUIDs, serverName *string) (map[uuid.UUID]*domain.Playtime, error) {
	res, err := api.player.GetPlaytime(ctx, &shellv1.GetPlaytimeRequest{MinecraftUuids: mcUUIDs.Strings(), ServerName: serverName})
	if err != nil {
		return nil, err
	}
	return parsePlaytimes(res.GetPlaytimes())
}

func parsePlaytimes(res map[string]*shellv1.Playtime) (map[uuid.UUID]*domain.Playtime, error) {
	playtimes := make(map[uuid.UUID]*domain.Playtime, len(res))
	for key, playtime := range res {
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

func (api *shellAPI) SetPlayerPrefix(ctx context.Context, mcUUID uuid.UUID, prefix string) error {
	_, err := api.permission.SetPlayerPrefix(ctx, &shellv1.SetPlayerPrefixRequest{
		MinecraftUuid: mcUUID.String(),
		Prefix:        prefix,
	})
	return err
}

func (api *shellAPI) SetPlayerRoles(ctx context.Context, roleGroups []string, roles map[uuid.UUID]string) error {
	req := &shellv1.SetPlayerRolesRequest{RoleGroups: roleGroups, Roles: make(map[string]string, len(roles))}
	for mcUUID, group := range roles {
		req.Roles[mcUUID.String()] = group
	}
	_, err := api.permission.SetPlayerRoles(ctx, req)
	return err
}

func (api *shellAPI) RegisterPlayer(ctx context.Context, mcUUID uuid.UUID, username string) error {
	_, err := api.permission.RegisterPlayer(ctx, &shellv1.RegisterPlayerRequest{
		MinecraftUuid: mcUUID.String(),
		Username:      username,
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

func (api *shellAPI) RemoveFromWhitelist(ctx context.Context, mcUUID uuid.UUID, username string) error {
	_, err := api.whitelist.RemovePlayer(ctx, &shellv1.RemovePlayerRequest{
		MinecraftUuid:     mcUUID.String(),
		MinecraftUsername: username,
	})
	return err
}
