package domain

import (
	"reflect"
	"testing"

	"github.com/google/uuid"
)

func TestRoleValid(t *testing.T) {
	for _, role := range []Role{RoleOwner, RoleAdmin, RoleModerator, RolePlayer} {
		if !role.Valid() {
			t.Errorf("Role(%q).Valid() = false, want true", role)
		}
	}
	for _, role := range []Role{"", "vip", "Admin", "default"} {
		if role.Valid() {
			t.Errorf("Role(%q).Valid() = true, want false", role)
		}
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
