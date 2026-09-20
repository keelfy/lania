package domain

import "testing"

func TestRoleFromMetadata(t *testing.T) {
	tests := []struct {
		name     string
		metadata map[string]any
		want     Role
	}{
		{"nil metadata", nil, RolePlayer},
		{"no role key", map[string]any{"other": "owner"}, RolePlayer},
		{"role is not a string", map[string]any{"role": 1}, RolePlayer},
		{"unknown role", map[string]any{"role": "root"}, RolePlayer},
		{"player", map[string]any{"role": "player"}, RolePlayer},
		{"moderator", map[string]any{"role": "mod"}, RoleModerator},
		{"admin", map[string]any{"role": "admin"}, RoleAdmin},
		{"owner", map[string]any{"role": "owner"}, RoleOwner},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RoleFromMetadata(tt.metadata); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRoleIsAdmin(t *testing.T) {
	for role, want := range map[Role]bool{
		RoleOwner: true, RoleAdmin: true, RoleModerator: false, RolePlayer: false, Role(""): false,
	} {
		if got := role.IsAdmin(); got != want {
			t.Errorf("Role(%q).IsAdmin() = %v, want %v", role, got, want)
		}
	}
}
