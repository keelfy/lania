package domain

import "testing"

func TestHighestRole(t *testing.T) {
	tests := []struct {
		name   string
		groups []string
		want   Role
	}{
		{"no groups", nil, RolePlayer},
		{"default group only", []string{"default"}, RolePlayer},
		{"single role", []string{"admin"}, RoleAdmin},
		{"owner beats admin", []string{"admin", "owner"}, RoleOwner},
		{"order does not matter", []string{"owner", "admin", "mod"}, RoleOwner},
		{"mod beats player", []string{"default", "mod"}, RoleModerator},
		{"unknown groups are ignored", []string{"vip", "mod", "builder"}, RoleModerator},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HighestRole(tt.groups); got != tt.want {
				t.Errorf("HighestRole(%v) = %q, want %q", tt.groups, got, tt.want)
			}
		})
	}
}
