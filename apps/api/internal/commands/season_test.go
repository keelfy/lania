package commands

import (
	"testing"
	"time"
)

func TestSaveSeasonCommand_Validate(t *testing.T) {
	t.Parallel()

	validIP := "203.0.113.10"
	validShell := "shell.internal:9090"
	shellWithoutPort := "shell.internal"
	invalidAddress := "bad host!"
	zeroPortShell := "shell.internal:0"
	start := time.Date(2026, 10, 9, 0, 0, 0, 0, time.UTC)
	beforeStart := start.AddDate(0, 0, -1)

	tests := []struct {
		name    string
		command SaveSeasonCommand
		wantErr bool
	}{
		{
			name: "valid technical settings",
			command: SaveSeasonCommand{
				Name: "Lania V", StartDate: start,
				PublicAddress: &validIP,
				ShellAddress:  &validShell,
			},
		},
		{
			name: "invalid address",
			command: SaveSeasonCommand{
				Name: "Lania V", StartDate: start,
				PublicAddress: &invalidAddress,
			},
			wantErr: true,
		},
		{
			name: "shell address without port",
			command: SaveSeasonCommand{
				Name: "Lania V", StartDate: start,
				ShellAddress: &shellWithoutPort,
			},
			wantErr: true,
		},
		{
			name: "zero shell port",
			command: SaveSeasonCommand{
				Name: "Lania V", StartDate: start,
				ShellAddress: &zeroPortShell,
			},
			wantErr: true,
		},
		{
			name: "end before start",
			command: SaveSeasonCommand{
				Name: "Lania V", StartDate: start, EndDate: &beforeStart,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.command.Validate()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
