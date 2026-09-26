package rpc

import (
	"context"
	"net"
	"slices"
	"testing"

	"github.com/google/uuid"
	"github.com/lania-smp/shell/internal/domain"
	shellv1 "github.com/lania-smp/shell/internal/gen/lania/shell/v1"
	"github.com/lania-smp/shell/internal/services"
	"github.com/lania-smp/shell/internal/storage"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

type fakePlanStorage struct {
	playtimes map[uuid.UUID]*domain.Playtime
	// servers holds playtime on one server of the network, by Plan server name.
	servers map[string]map[uuid.UUID]*domain.Playtime
}

func (s *fakePlanStorage) FindPlaytimes(_ context.Context, _ uuid.UUIDs, serverName *string) (map[uuid.UUID]*domain.Playtime, error) {
	if serverName == nil {
		return s.playtimes, nil
	}
	found := make(map[uuid.UUID]*domain.Playtime)
	for mcUUID, playtime := range s.servers[*serverName] {
		found[mcUUID] = playtime
	}
	return found, nil
}

func (s *fakePlanStorage) FindPlaytimesChangedSince(_ context.Context, sinceMs int64) (map[uuid.UUID]*domain.Playtime, error) {
	changed := make(map[uuid.UUID]*domain.Playtime)
	for mcUUID, playtime := range s.playtimes {
		if playtime.LastSeenMs != nil && *playtime.LastSeenMs >= sinceMs {
			changed[mcUUID] = playtime
		}
	}
	return changed, nil
}

type fakeFlectoneStorage struct {
	online map[uuid.UUID]bool
}

func (s *fakeFlectoneStorage) FindOnline(_ context.Context, _ uuid.UUIDs) (map[uuid.UUID]bool, error) {
	return s.online, nil
}

func (s *fakeFlectoneStorage) FindOnlineUUIDs(_ context.Context) (uuid.UUIDs, error) {
	online := uuid.UUIDs{}
	for mcUUID, isOnline := range s.online {
		if isOnline {
			online = append(online, mcUUID)
		}
	}
	return online, nil
}

type fakeLuckpermsStorage struct {
	nodes           map[uuid.UUID][]string
	lastPermissions []string
}

func (s *fakeLuckpermsStorage) FindPermissionsWithPrefix(_ context.Context, _ uuid.UUIDs, _ string) (map[uuid.UUID][]string, error) {
	return s.nodes, nil
}

func (s *fakeLuckpermsStorage) FindPlayersWithPermissions(_ context.Context, permissions []string) (uuid.UUIDs, error) {
	s.lastPermissions = permissions
	mcUUIDs := uuid.UUIDs{}
	for mcUUID, nodes := range s.nodes {
		for _, node := range nodes {
			if slices.Contains(permissions, node) {
				mcUUIDs = append(mcUUIDs, mcUUID)
				break
			}
		}
	}
	return mcUUIDs, nil
}

func (s *fakeLuckpermsStorage) ReplacePermissions(context.Context, []storage.PermissionReplacement) error {
	return nil
}

func (s *fakeLuckpermsStorage) SavePlayer(context.Context, uuid.UUID, string) error {
	return nil
}

type fakeConsole struct {
	commands []string
}

func (c *fakeConsole) Execute(_ context.Context, command string) (string, error) {
	c.commands = append(c.commands, command)
	return "", nil
}

var (
	knownUUID   = uuid.MustParse("0f0c2a3e-5a4b-4d4c-9f6e-3b1a2c3d4e5f")
	unknownUUID = uuid.MustParse("1a2b3c4d-5e6f-4a1b-8c2d-3e4f5a6b7c8d")
)

type testEnv struct {
	conn      *grpc.ClientConn
	luckperms *fakeLuckpermsStorage
	console   *fakeConsole
}

func newTestEnv(t *testing.T) *testEnv {
	t.Helper()

	first := int64(1000)
	last := int64(5000)
	luckperms := &fakeLuckpermsStorage{nodes: map[uuid.UUID][]string{knownUUID: {"group.admin", "group.default"}}}
	console := &fakeConsole{}

	server := NewServer(
		NewPlayerHandler(services.NewPlayerService(
			&fakePlanStorage{
				playtimes: map[uuid.UUID]*domain.Playtime{knownUUID: {TotalMs: 3000, FirstSeenMs: &first, LastSeenMs: &last}},
				servers:   map[string]map[uuid.UUID]*domain.Playtime{"farms": {knownUUID: {TotalMs: 1000, FirstSeenMs: &first, LastSeenMs: &first}}},
			},
			&fakeFlectoneStorage{online: map[uuid.UUID]bool{knownUUID: true}},
		)),
		NewPermissionHandler(services.NewPermissionService(luckperms, console)),
		NewWhitelistHandler(services.NewWhitelistService(console)),
	)

	listener := bufconn.Listen(1024 * 1024)
	go func() {
		_ = server.Serve(listener)
	}()
	t.Cleanup(server.Stop)

	conn, err := grpc.NewClient(
		"passthrough:///bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return listener.DialContext(ctx)
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	return &testEnv{conn: conn, luckperms: luckperms, console: console}
}

func TestPlayerService(t *testing.T) {
	env := newTestEnv(t)
	client := shellv1.NewPlayerServiceClient(env.conn)
	ctx := context.Background()
	uuids := []string{knownUUID.String(), unknownUUID.String()}

	online, err := client.GetOnlineStatus(ctx, &shellv1.GetOnlineStatusRequest{MinecraftUuids: uuids})
	if err != nil {
		t.Fatal(err)
	}
	if !online.GetOnline()[knownUUID.String()] {
		t.Errorf("known player must be online")
	}
	if isOnline, ok := online.GetOnline()[unknownUUID.String()]; !ok || isOnline {
		t.Errorf("unknown player must be present and offline, got ok=%v online=%v", ok, isOnline)
	}

	playtimes, err := client.GetPlaytime(ctx, &shellv1.GetPlaytimeRequest{MinecraftUuids: uuids})
	if err != nil {
		t.Fatal(err)
	}
	known := playtimes.GetPlaytimes()[knownUUID.String()]
	if known.GetTotalMs() != 3000 || known.GetFirstSeenMs() != 1000 || known.GetLastSeenMs() != 5000 {
		t.Errorf("unexpected known playtime: %v", known)
	}
	unknown, ok := playtimes.GetPlaytimes()[unknownUUID.String()]
	if !ok || unknown.GetTotalMs() != 0 || unknown.FirstSeenMs != nil || unknown.LastSeenMs != nil {
		t.Errorf("unknown player must have empty playtime, got %v", unknown)
	}

	farms := "farms"
	onServer, err := client.GetPlaytime(ctx, &shellv1.GetPlaytimeRequest{MinecraftUuids: uuids, ServerName: &farms})
	if err != nil {
		t.Fatal(err)
	}
	if got := onServer.GetPlaytimes()[knownUUID.String()].GetTotalMs(); got != 1000 {
		t.Errorf("playtime on the farms server must be 1000, got %d", got)
	}
	if unknown, ok := onServer.GetPlaytimes()[unknownUUID.String()]; !ok || unknown.GetTotalMs() != 0 {
		t.Errorf("unknown player must have empty playtime on the server, got %v", unknown)
	}

	onlinePlayers, err := client.ListOnlinePlayers(ctx, &shellv1.ListOnlinePlayersRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if got := onlinePlayers.GetMinecraftUuids(); len(got) != 1 || got[0] != knownUUID.String() {
		t.Errorf("unexpected online players: %v", got)
	}

	changed, err := client.ListChangedPlaytimes(ctx, &shellv1.ListChangedPlaytimesRequest{SinceMs: 5000})
	if err != nil {
		t.Fatal(err)
	}
	if got := changed.GetPlaytimes()[knownUUID.String()]; len(changed.GetPlaytimes()) != 1 || got.GetTotalMs() != 3000 {
		t.Errorf("unexpected changed playtimes: %v", changed.GetPlaytimes())
	}

	changed, err = client.ListChangedPlaytimes(ctx, &shellv1.ListChangedPlaytimesRequest{SinceMs: 5001})
	if err != nil {
		t.Fatal(err)
	}
	if len(changed.GetPlaytimes()) != 0 {
		t.Errorf("players seen before since must be skipped, got %v", changed.GetPlaytimes())
	}
}

func TestPermissionService(t *testing.T) {
	env := newTestEnv(t)
	client := shellv1.NewPermissionServiceClient(env.conn)
	ctx := context.Background()

	groups, err := client.GetPlayerGroups(ctx, &shellv1.GetPlayerGroupsRequest{MinecraftUuids: []string{knownUUID.String(), unknownUUID.String()}})
	if err != nil {
		t.Fatal(err)
	}
	names := groups.GetGroups()[knownUUID.String()].GetNames()
	if len(names) != 2 || names[0] != "admin" || names[1] != "default" {
		t.Errorf("unexpected groups: %v", names)
	}
	if _, ok := groups.GetGroups()[unknownUUID.String()]; !ok {
		t.Errorf("unknown player must be present")
	}

	members, err := client.ListPlayersByGroups(ctx, &shellv1.ListPlayersByGroupsRequest{Groups: []string{"admin", "mod"}})
	if err != nil {
		t.Fatal(err)
	}
	if got := members.GetMinecraftUuids(); len(got) != 1 || got[0] != knownUUID.String() {
		t.Errorf("unexpected group members: %v", got)
	}
	if want := []string{"group.admin", "group.mod"}; !slices.Equal(env.luckperms.lastPermissions, want) {
		t.Errorf("groups must be queried as permission nodes, got %v want %v", env.luckperms.lastPermissions, want)
	}

	_, err = client.SetPlayerPrefix(ctx, &shellv1.SetPlayerPrefixRequest{MinecraftUuid: knownUUID.String(), Prefix: "<red>[A]"})
	if err != nil {
		t.Fatal(err)
	}
	wantCommands := []string{
		"lp user " + knownUUID.String() + " meta clear prefix",
		"lp user " + knownUUID.String() + ` meta addprefix 100 "<red>[A]"`,
	}
	if !slices.Equal(env.console.commands, wantCommands) {
		t.Errorf("prefix must be written with console commands, got %q want %q", env.console.commands, wantCommands)
	}

	_, err = client.SetPlayerPrefix(ctx, &shellv1.SetPlayerPrefixRequest{MinecraftUuid: knownUUID.String(), Prefix: `[A" x`})
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("prefix with a quote must be invalid, got %v", err)
	}
}

func TestWhitelistService(t *testing.T) {
	env := newTestEnv(t)
	client := shellv1.NewWhitelistServiceClient(env.conn)
	ctx := context.Background()

	_, err := client.AddPlayer(ctx, &shellv1.AddPlayerRequest{MinecraftUuid: knownUUID.String(), MinecraftUsername: "Steve"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.RemovePlayer(ctx, &shellv1.RemovePlayerRequest{MinecraftUuid: knownUUID.String(), MinecraftUsername: "Steve"})
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"whitelist add Steve", "whitelist remove Steve"}; !slices.Equal(env.console.commands, want) {
		t.Errorf("expected commands %v, got %v", want, env.console.commands)
	}

	for _, username := range []string{"", "St", "Steve\nop Steve", "Steve; stop"} {
		_, err = client.AddPlayer(ctx, &shellv1.AddPlayerRequest{MinecraftUuid: knownUUID.String(), MinecraftUsername: username})
		if status.Code(err) != codes.InvalidArgument {
			t.Errorf("username %q: expected InvalidArgument, got %v", username, err)
		}
	}
	if len(env.console.commands) != 2 {
		t.Errorf("invalid usernames must not reach the console, got %v", env.console.commands)
	}

	_, err = client.AddPlayer(ctx, &shellv1.AddPlayerRequest{MinecraftUuid: "not-a-uuid", MinecraftUsername: "Steve"})
	if status.Code(err) != codes.InvalidArgument {
		t.Errorf("expected InvalidArgument, got %v", err)
	}
}

func TestAuth(t *testing.T) {
	t.Setenv("SHELL_TOKEN", "secret")
	env := newTestEnv(t)
	client := shellv1.NewPlayerServiceClient(env.conn)
	req := &shellv1.GetOnlineStatusRequest{MinecraftUuids: []string{knownUUID.String()}}

	_, err := client.GetOnlineStatus(context.Background(), req)
	if status.Code(err) != codes.Unauthenticated {
		t.Errorf("expected Unauthenticated without token, got %v", err)
	}

	ctx := metadata.AppendToOutgoingContext(context.Background(), "authorization", "Bearer wrong")
	_, err = client.GetOnlineStatus(ctx, req)
	if status.Code(err) != codes.Unauthenticated {
		t.Errorf("expected Unauthenticated with wrong token, got %v", err)
	}

	ctx = metadata.AppendToOutgoingContext(context.Background(), "authorization", "Bearer secret")
	if _, err = client.GetOnlineStatus(ctx, req); err != nil {
		t.Errorf("expected success with valid token, got %v", err)
	}
}
