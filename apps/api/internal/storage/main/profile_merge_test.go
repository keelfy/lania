package sql

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// mergeHandledTables are the tables MergeProfileData moves, sums or dedupes. profile_merges is the audit
// ledger itself: it references the target profile on purpose and must never be rewritten by a merge.
var mergeHandledTables = map[string]bool{
	"profile_playtimes":           true,
	"profile_accesses":            true,
	"profile_violations":          true,
	"profile_mojang_uuids":        true,
	"profile_name_color_options":  true,
	"profile_name_prefix_options": true,
	"profile_season_cosmetics":    true,
	"profile_prefixes":            true,
	"order_items":                 true,
	"basket_items":                true,
	"profile_merges":              true,
}

var createTablePattern = regexp.MustCompile(`(?is)CREATE TABLE IF NOT EXISTS\s+(\w+)\s*\((.*?)\n\);`)

// TestMergeProfileDataCoversEveryProfileTable fails when a migration adds a table referencing profiles(id) or
// profiles(mc_uuid) that MergeProfileData does not know about, so a new profile table cannot be forgotten by a merge.
func TestMergeProfileDataCoversEveryProfileTable(t *testing.T) {
	migrationDir := filepath.Join("..", "..", "..", "db", "migration")
	entries, err := os.ReadDir(migrationDir)
	if err != nil {
		t.Fatalf("failed to read migration directory: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".up.sql") {
			continue
		}

		content, err := os.ReadFile(filepath.Join(migrationDir, entry.Name()))
		if err != nil {
			t.Fatalf("failed to read %s: %v", entry.Name(), err)
		}

		for _, match := range createTablePattern.FindAllStringSubmatch(string(content), -1) {
			table, body := match[1], match[2]
			if !strings.Contains(body, "REFERENCES profiles(") {
				continue
			}
			if !mergeHandledTables[table] {
				t.Errorf("table %s references profiles but is not handled by a profile merge; add it to MergeProfileData and mergeHandledTables", table)
			}
		}
	}
}
