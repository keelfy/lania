package sql

import "testing"

func TestProfileOrderBy(t *testing.T) {
	tests := []struct {
		name     string
		col      string
		dir      string
		wantJoin bool
		want     string
	}{
		{"default", "", "", false, "p.created_at ASC, p.id"},
		{"unknown column", "id; DROP TABLE profiles", "desc", false, "p.created_at DESC, p.id"},
		{"unknown direction", "created_at", "sideways", false, "p.created_at ASC, p.id"},
		{"username", "username", "desc", false, "p.mc_username DESC, p.id"},
		{"first seen", "first_seen_at", "asc", false, "p.first_seen_at IS NULL, p.first_seen_at ASC, p.id"},
		{"last seen", "last_seen_at", "DESC", false, "p.last_seen_at IS NULL, p.last_seen_at DESC, p.id"},
		{"playtime", "playtime", "desc", true, "COALESCE(pt.total_playtime, 0) DESC, p.id"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			join, got := profileOrderBy(tt.col, tt.dir)
			if got != tt.want {
				t.Errorf("orderBy = %q, want %q", got, tt.want)
			}
			if (join != "") != tt.wantJoin {
				t.Errorf("join = %q, wantJoin %v", join, tt.wantJoin)
			}
		})
	}
}

func TestProfileSearchClause(t *testing.T) {
	where, args := profileSearchClause("")
	if where != "" || len(args) != 0 {
		t.Errorf("empty search = %q %v, want no clause", where, args)
	}

	where, args = profileSearchClause(`50%_a\b`)
	if where == "" || len(args) != 1 {
		t.Fatalf("clause = %q %v", where, args)
	}
	if want := `50\%\_a\\b%`; args[0] != want {
		t.Errorf("arg = %q, want %q", args[0], want)
	}
}
