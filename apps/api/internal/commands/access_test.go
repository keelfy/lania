package commands

import (
	"testing"

	"github.com/lania-smp/backend/internal/domain"
)

func TestExceedsProfileLimit(t *testing.T) {
	tests := []struct {
		role     domain.Role
		count    int
		expected bool
	}{
		{domain.RolePlayer, 1, false},
		{domain.RolePlayer, 2, true},
		{domain.RoleModerator, 2, true},
		{domain.RoleAdmin, 5, false},
		{domain.RoleOwner, 5, false},
	}
	for _, tt := range tests {
		cmd := &ObtainAccessByUsernamesCommand{OwnerRole: tt.role}
		if got := cmd.ExceedsProfileLimit(tt.count, 2); got != tt.expected {
			t.Errorf("role %q with %d profiles: expected %v, got %v", tt.role, tt.count, tt.expected, got)
		}
	}
}
