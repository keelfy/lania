package domain

import (
	"reflect"
	"testing"

	"github.com/google/uuid"
)

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

func TestSeasonAccessStatuses(t *testing.T) {
	active, other, ended, oldEnded := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	seasons := []*Season{
		{ID: active, IsActive: true},
		{ID: other, IsActive: true},
		{ID: ended, IsActive: false},
		{ID: oldEnded, IsActive: false},
	}

	tests := []struct {
		name     string
		accesses []*ProfileAccess
		want     []*SeasonAccessStatus
	}{
		{
			"no access",
			nil,
			[]*SeasonAccessStatus{{active, AccessStatusInactive}, {other, AccessStatusInactive}},
		},
		{
			"access to one of the active seasons",
			[]*ProfileAccess{{SeasonID: other}},
			[]*SeasonAccessStatus{{active, AccessStatusInactive}, {other, AccessStatusActive}},
		},
		{
			"ended season with access is expired",
			[]*ProfileAccess{{SeasonID: active}, {SeasonID: ended}},
			[]*SeasonAccessStatus{{active, AccessStatusActive}, {other, AccessStatusInactive}, {ended, AccessStatusExpired}},
		},
		{
			"access to an unknown season is ignored",
			[]*ProfileAccess{{SeasonID: uuid.New()}},
			[]*SeasonAccessStatus{{active, AccessStatusInactive}, {other, AccessStatusInactive}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SeasonAccessStatuses(seasons, tt.accesses); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SeasonAccessStatuses() = %v, want %v", got, tt.want)
			}
		})
	}
}
