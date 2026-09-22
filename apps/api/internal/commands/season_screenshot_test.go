package commands

import (
	"strings"
	"testing"
)

func TestSaveSeasonScreenshotCommand_Validate(t *testing.T) {
	t.Parallel()

	title := "Spawn build"
	tooLongTitle := strings.Repeat("a", 256)

	tests := []struct {
		name    string
		command SaveSeasonScreenshotCommand
		wantErr bool
	}{
		{
			name:    "valid with title",
			command: SaveSeasonScreenshotCommand{Image: "s3://bucket/key.jpg", Title: &title},
		},
		{
			name:    "valid without title",
			command: SaveSeasonScreenshotCommand{Image: "s3://bucket/key.jpg"},
		},
		{
			name:    "empty image",
			command: SaveSeasonScreenshotCommand{Image: ""},
			wantErr: true,
		},
		{
			name:    "title too long",
			command: SaveSeasonScreenshotCommand{Image: "s3://bucket/key.jpg", Title: &tooLongTitle},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.command.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
