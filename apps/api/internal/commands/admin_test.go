package commands

import (
	"testing"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/domain"
)

func TestSetProfileRoleCommand_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		command SetProfileRoleCommand
		wantErr bool
	}{
		{"staff role", SetProfileRoleCommand{ProfileID: uuid.New(), Role: domain.RoleModerator}, false},
		{"player role", SetProfileRoleCommand{ProfileID: uuid.New(), Role: domain.RolePlayer}, false},
		{"unknown role", SetProfileRoleCommand{ProfileID: uuid.New(), Role: "vip"}, true},
		{"blank role", SetProfileRoleCommand{ProfileID: uuid.New()}, true},
		{"nil profile", SetProfileRoleCommand{Role: domain.RoleAdmin}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := tt.command.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMergeProfilesCommand_Validate(t *testing.T) {
	t.Parallel()

	source, target := uuid.New(), uuid.New()
	tests := []struct {
		name    string
		command MergeProfilesCommand
		wantErr bool
	}{
		{"distinct profiles", MergeProfilesCommand{SourceProfileID: source, TargetProfileID: target}, false},
		{"same profile", MergeProfilesCommand{SourceProfileID: source, TargetProfileID: source}, true},
		{"nil source", MergeProfilesCommand{TargetProfileID: target}, true},
		{"nil target", MergeProfilesCommand{SourceProfileID: source}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := tt.command.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
