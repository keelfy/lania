package commands

import (
	"testing"
	"time"
)

func TestSaveSeasonCommand_Validate(t *testing.T) {
	t.Parallel()

	validIP := "203.0.113.10"
	invalidIP := "minecraft.example.com"
	port := uint16(25565)
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
				SeasonNumber: 5, Name: "Lania V", StartDate: start,
				ServerIP: &validIP, ServerPort: &port,
			},
		},
		{
			name: "port without IP",
			command: SaveSeasonCommand{
				SeasonNumber: 5, Name: "Lania V", StartDate: start, ServerPort: &port,
			},
			wantErr: true,
		},
		{
			name: "invalid IP",
			command: SaveSeasonCommand{
				SeasonNumber: 5, Name: "Lania V", StartDate: start,
				ServerIP: &invalidIP, ServerPort: &port,
			},
			wantErr: true,
		},
		{
			name: "end before start",
			command: SaveSeasonCommand{
				SeasonNumber: 5, Name: "Lania V", StartDate: start, EndDate: &beforeStart,
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
