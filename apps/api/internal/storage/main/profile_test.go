package sql

import (
	"testing"

	"github.com/google/uuid"
)

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

func TestProfileWhereClause(t *testing.T) {
	where, args := profileWhereClause("", nil)
	if where != "" || len(args) != 0 {
		t.Errorf("empty search = %q %v, want no clause", where, args)
	}

	where, args = profileWhereClause(`50%_a\b`, nil)
	if where == "" || len(args) != 1 {
		t.Fatalf("clause = %q %v", where, args)
	}
	if want := `50\%\_a\\b%`; args[0] != want {
		t.Errorf("arg = %q, want %q", args[0], want)
	}
}

func TestProfileWhereClauseOnly(t *testing.T) {
	first, second := uuid.New(), uuid.New()

	where, args := profileWhereClause("", &uuid.UUIDs{first, second})
	if want := "WHERE p.mc_uuid IN (?, ?)"; where != want {
		t.Errorf("where = %q, want %q", where, want)
	}
	if len(args) != 2 || args[0] != first.String() || args[1] != second.String() {
		t.Errorf("args = %v", args)
	}

	where, args = profileWhereClause("a", &uuid.UUIDs{first})
	if want := "WHERE p.mc_username LIKE ? AND p.mc_uuid IN (?)"; where != want {
		t.Errorf("where = %q, want %q", where, want)
	}
	if len(args) != 2 || args[0] != "a%" || args[1] != first.String() {
		t.Errorf("args = %v", args)
	}

	where, args = profileWhereClause("", &uuid.UUIDs{})
	if where != "WHERE 1 = 0" || len(args) != 0 {
		t.Errorf("empty only = %q %v, want a clause matching nobody", where, args)
	}
}
