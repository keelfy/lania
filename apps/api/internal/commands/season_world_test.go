package commands

import "testing"

func TestSaveSeasonWorldCommand_Validate(t *testing.T) {
	mapURL := "https://survival-map.lania.network"
	badURL := "survival-map.lania.network"
	tests := []struct {
		name    string
		cmd     SaveSeasonWorldCommand
		wantErr bool
	}{
		{name: "valid", cmd: SaveSeasonWorldCommand{Slug: "survival", Name: "Выживание", MapURL: &mapURL, ClaimLimit: 100, ClaimDimensions: []string{"minecraft_overworld"}}},
		{name: "view-only without a map", cmd: SaveSeasonWorldCommand{Slug: "creative-2", Name: "Креатив"}},
		{name: "slug with a slash", cmd: SaveSeasonWorldCommand{Slug: "survival/1", Name: "x"}, wantErr: true},
		{name: "uppercase slug", cmd: SaveSeasonWorldCommand{Slug: "Survival", Name: "x"}, wantErr: true},
		{name: "no name", cmd: SaveSeasonWorldCommand{Slug: "survival"}, wantErr: true},
		{name: "map without a scheme", cmd: SaveSeasonWorldCommand{Slug: "survival", Name: "x", MapURL: &badURL}, wantErr: true},
		{name: "negative limit", cmd: SaveSeasonWorldCommand{Slug: "survival", Name: "x", ClaimLimit: -1}, wantErr: true},
		{name: "dimension with a path", cmd: SaveSeasonWorldCommand{Slug: "survival", Name: "x", ClaimDimensions: []string{"../x"}}, wantErr: true},
		{name: "dimension twice", cmd: SaveSeasonWorldCommand{Slug: "survival", Name: "x", ClaimDimensions: []string{"minecraft_overworld", "minecraft_overworld"}}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.cmd.Validate(); (err != nil) != tt.wantErr {
				t.Fatalf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
