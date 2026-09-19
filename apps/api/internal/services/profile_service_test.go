package services

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/domain"
)

type stubMinecraftService struct {
	MinecraftService
	online    uuid.UUIDs
	staff     uuid.UUIDs
	err       error
	lastGroup []string
}

func (s *stubMinecraftService) ListOnlineMinecraftUUIDs(context.Context) (uuid.UUIDs, error) {
	return s.online, s.err
}

func (s *stubMinecraftService) ListMinecraftUUIDsByGroups(_ context.Context, groups []string) (uuid.UUIDs, error) {
	s.lastGroup = groups
	return s.staff, s.err
}

func TestFilterMinecraftUUIDs(t *testing.T) {
	first, second, third := uuid.New(), uuid.New(), uuid.New()
	minecraft := &stubMinecraftService{online: uuid.UUIDs{first, second}, staff: uuid.UUIDs{second, third}}
	service := &profileService{minecraftService: minecraft}
	ctx := context.Background()

	only, err := service.filterMinecraftUUIDs(ctx, domain.ProfileFilter{Search: "a"})
	if err != nil || only != nil {
		t.Errorf("filter without server state = %v %v, want no restriction", only, err)
	}

	only, err = service.filterMinecraftUUIDs(ctx, domain.ProfileFilter{OnlineOnly: true})
	if err != nil || only == nil || len(*only) != 2 {
		t.Errorf("online = %v %v, want 2 players", only, err)
	}

	only, err = service.filterMinecraftUUIDs(ctx, domain.ProfileFilter{StaffOnly: true})
	if err != nil || only == nil || len(*only) != 2 {
		t.Errorf("staff = %v %v, want 2 players", only, err)
	}
	if want := []string{"owner", "admin", "mod"}; len(minecraft.lastGroup) != 3 || minecraft.lastGroup[0] != want[0] || minecraft.lastGroup[1] != want[1] || minecraft.lastGroup[2] != want[2] {
		t.Errorf("groups = %v, want %v", minecraft.lastGroup, want)
	}

	only, err = service.filterMinecraftUUIDs(ctx, domain.ProfileFilter{OnlineOnly: true, StaffOnly: true})
	if err != nil || only == nil || len(*only) != 1 || (*only)[0] != second {
		t.Errorf("online staff = %v %v, want only the second player", only, err)
	}

	minecraft.staff = uuid.UUIDs{third}
	only, err = service.filterMinecraftUUIDs(ctx, domain.ProfileFilter{OnlineOnly: true, StaffOnly: true})
	if err != nil || only == nil || len(*only) != 0 {
		t.Errorf("no online staff = %v %v, want an empty restriction", only, err)
	}

	minecraft.err = errors.New("shell down")
	if _, err = service.filterMinecraftUUIDs(ctx, domain.ProfileFilter{OnlineOnly: true}); err == nil {
		t.Error("filter must fail when the server is unreachable")
	}
}
