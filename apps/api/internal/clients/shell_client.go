package clients

import (
	"context"
	"fmt"
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
	// SetPassword replaces the in-game password of an unlicensed player, turns off its two-factor login and kicks
	// the player from the network. It fails with codes.NotFound when the player never registered in game, and with codes.FailedPrecondition
	// when the player logs in with a licensed account.
	SetPassword(ctx context.Context, mcUUID uuid.UUID, username, passwordBcrypt string) error
	// GetPlayerSkins returns the SkinsRestorer skin of every player that chose one.
	GetPlayerSkins(ctx context.Context, mcUUIDs uuid.UUIDs) (map[uuid.UUID]*domain.PlayerSkin, error)
	// SetPlayerSkin gives the player the skin of a link, or of the licensed account mojangUUID by its nickname, and
	// waits until the server applies it. An empty variant lets the server guess the model of a link. It fails with
	// codes.DeadlineExceeded when the server does not apply the skin in time, and with codes.InvalidArgument when
	// the skin is neither a link nor a nickname.
	SetPlayerSkin(ctx context.Context, mcUUID uuid.UUID, skin string, variant domain.SkinVariant, mojangUUID *uuid.UUID) (*domain.PlayerSkin, error)
	// ClearPlayerSkin takes the SkinsRestorer skin off the player. It fails with codes.DeadlineExceeded when the
	// server does not do it in time.
	ClearPlayerSkin(ctx context.Context, mcUUID uuid.UUID) error
	// GetPlayerPunishments returns the LiteBans bans and mutes of the players, newest first.
	GetPlayerPunishments(ctx context.Context, mcUUIDs uuid.UUIDs) ([]*domain.Punishment, error)
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
	auth       shellv1.AuthServiceClient
	skin       shellv1.SkinServiceClient
	punishment shellv1.PunishmentServiceClient
}

// shellCallTimeout keeps pages responsive when the Minecraft host is slow.
const shellCallTimeout = 3 * time.Second

// shellApplyTimeout is for calls that wait for the server to apply a change in background. Shell gives up a bit
// earlier and answers with codes.DeadlineExceeded.
const shellApplyTimeout = 20 * time.Second

var shellCallTimeouts = map[string]time.Duration{
	shellv1.SkinService_SetPlayerSkin_FullMethodName:   shellApplyTimeout,
	shellv1.SkinService_ClearPlayerSkin_FullMethodName: shellApplyTimeout,
}

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
		auth:       shellv1.NewAuthServiceClient(conn),
		skin:       shellv1.NewSkinServiceClient(conn),
		punishment: shellv1.NewPunishmentServiceClient(conn),
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
	timeout, ok := shellCallTimeouts[method]
	if !ok {
		timeout = shellCallTimeout
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
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
			Deaths:            playtime.GetDeaths(),
			MobKills:          playtime.GetMobKills(),
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

func (api *shellAPI) SetPassword(ctx context.Context, mcUUID uuid.UUID, username, passwordBcrypt string) error {
	_, err := api.auth.SetPassword(ctx, &shellv1.SetPasswordRequest{
		MinecraftUuid:     mcUUID.String(),
		MinecraftUsername: username,
		PasswordBcrypt:    passwordBcrypt,
	})
	return err
}

func (api *shellAPI) GetPlayerSkins(ctx context.Context, mcUUIDs uuid.UUIDs) (map[uuid.UUID]*domain.PlayerSkin, error) {
	res, err := api.skin.GetPlayerSkins(ctx, &shellv1.GetPlayerSkinsRequest{MinecraftUuids: mcUUIDs.Strings()})
	if err != nil {
		return nil, err
	}

	skins := make(map[uuid.UUID]*domain.PlayerSkin, len(res.GetSkins()))
	for key, skin := range res.GetSkins() {
		mcUUID, err := uuid.Parse(key)
		if err != nil {
			return nil, err
		}
		if skins[mcUUID], err = parsePlayerSkin(skin); err != nil {
			return nil, err
		}
	}
	return skins, nil
}

var shellSkinVariants = map[domain.SkinVariant]shellv1.SkinVariant{
	domain.SkinVariantClassic: shellv1.SkinVariant_SKIN_VARIANT_CLASSIC,
	domain.SkinVariantSlim:    shellv1.SkinVariant_SKIN_VARIANT_SLIM,
}

func (api *shellAPI) SetPlayerSkin(ctx context.Context, mcUUID uuid.UUID, skin string, variant domain.SkinVariant, mojangUUID *uuid.UUID) (*domain.PlayerSkin, error) {
	req := &shellv1.SetPlayerSkinRequest{
		MinecraftUuid: mcUUID.String(),
		Skin:          skin,
		Variant:       shellSkinVariants[variant],
	}
	if mojangUUID != nil {
		value := mojangUUID.String()
		req.MojangUuid = &value
	}
	res, err := api.skin.SetPlayerSkin(ctx, req)
	if err != nil {
		return nil, err
	}
	return parsePlayerSkin(res.GetSkin())
}

func (api *shellAPI) ClearPlayerSkin(ctx context.Context, mcUUID uuid.UUID) error {
	_, err := api.skin.ClearPlayerSkin(ctx, &shellv1.ClearPlayerSkinRequest{MinecraftUuid: mcUUID.String()})
	return err
}

func parsePlayerSkin(skin *shellv1.PlayerSkin) (*domain.PlayerSkin, error) {
	res := &domain.PlayerSkin{TextureURL: skin.GetTextureUrl(), Slim: skin.GetSlim()}
	if skin.MojangUuid != nil {
		mojangUUID, err := uuid.Parse(skin.GetMojangUuid())
		if err != nil {
			return nil, err
		}
		res.MojangUUID = &mojangUUID
	}
	return res, nil
}

var shellPunishmentKinds = map[shellv1.PunishmentKind]domain.PunishmentKind{
	shellv1.PunishmentKind_PUNISHMENT_KIND_BAN:  domain.PunishmentKindBan,
	shellv1.PunishmentKind_PUNISHMENT_KIND_MUTE: domain.PunishmentKindMute,
}

var shellPunishmentStatuses = map[shellv1.PunishmentStatus]domain.PunishmentStatus{
	shellv1.PunishmentStatus_PUNISHMENT_STATUS_ACTIVE:  domain.PunishmentStatusActive,
	shellv1.PunishmentStatus_PUNISHMENT_STATUS_EXPIRED: domain.PunishmentStatusExpired,
	shellv1.PunishmentStatus_PUNISHMENT_STATUS_REMOVED: domain.PunishmentStatusRemoved,
}

func (api *shellAPI) GetPlayerPunishments(ctx context.Context, mcUUIDs uuid.UUIDs) ([]*domain.Punishment, error) {
	res, err := api.punishment.GetPlayerPunishments(ctx, &shellv1.GetPlayerPunishmentsRequest{MinecraftUuids: mcUUIDs.Strings()})
	if err != nil {
		return nil, err
	}

	punishments := make([]*domain.Punishment, 0, len(res.GetPunishments()))
	for _, punishment := range res.GetPunishments() {
		kind, ok := shellPunishmentKinds[punishment.GetKind()]
		if !ok {
			return nil, fmt.Errorf("unknown punishment kind %s", punishment.GetKind())
		}
		status, ok := shellPunishmentStatuses[punishment.GetStatus()]
		if !ok {
			return nil, fmt.Errorf("unknown punishment status %s", punishment.GetStatus())
		}
		mcUUID, err := uuid.Parse(punishment.GetMinecraftUuid())
		if err != nil {
			return nil, err
		}
		punishments = append(punishments, &domain.Punishment{
			ID:            punishment.GetId(),
			Kind:          kind,
			MinecraftUUID: mcUUID,
			Reason:        punishment.GetReason(),
			IssuedBy:      punishment.IssuedBy,
			IssuedAt:      time.UnixMilli(punishment.GetIssuedAtMs()),
			ExpiresAt:     millisToTime(punishment.ExpiresAtMs),
			Status:        status,
			RemovedBy:     punishment.RemovedBy,
			RemovedReason: punishment.GetRemovedReason(),
			RemovedAt:     millisToTime(punishment.RemovedAtMs),
		})
	}
	return punishments, nil
}

func millisToTime(ms *int64) *time.Time {
	if ms == nil {
		return nil
	}
	t := time.UnixMilli(*ms)
	return &t
}
