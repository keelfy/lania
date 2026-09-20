package sql

import (
	"strings"
	"testing"
)

// Revoked grants stay in the tables, so every query that reads active grants has to skip them.
func TestActiveGrantReadsSkipRevoked(t *testing.T) {
	queries := map[string]string{
		"checkIfProfileHasAccessBySeasonIDAndMinecraftUUID":       checkIfProfileHasAccessBySeasonIDAndMinecraftUUID,
		"getProfileAccessesBySeasonIDAndOwnerUserID":              getProfileAccessesBySeasonIDAndOwnerUserID,
		"findProfileAccessesByMinecraftUUIDs":                     findProfileAccessesByMinecraftUUIDs,
		"findProfileNameColorOptionsByProfileID":                  findProfileNameColorOptionsByProfileID,
		"findProfileNamePrefixOptionsByProfileIDAndType":          findProfileNamePrefixOptionsByProfileIDAndType,
		"findProfileNameColorOptionByIDAndProfileID":              findProfileNameColorOptionByIDAndProfileID,
		"findProfileNamePrefixOptionByIDAndProfileIDAndType":      findProfileNamePrefixOptionByIDAndProfileIDAndType,
		"findProfileNameColorOptionsByProfileOwnerUserID":         findProfileNameColorOptionsByProfileOwnerUserID,
		"findProfileNamePrefixOptionsByProfileOwnerUserIDAndType": findProfileNamePrefixOptionsByProfileOwnerUserIDAndType,
	}

	for name, query := range queries {
		if !strings.Contains(query, "revoked_at IS NULL") {
			t.Errorf("%s reads grants without skipping revoked ones", name)
		}
	}
}

func TestGrantInsertsReinstateRevoked(t *testing.T) {
	for name, query := range map[string]string{
		"insertProfileAccess":           insertProfileAccess,
		"insertProfileNameColorOption":  insertProfileNameColorOption,
		"insertProfileNamePrefixOption": insertProfileNamePrefixOption,
	} {
		if !strings.Contains(query, "ON DUPLICATE KEY UPDATE revoked_at = NULL, revoked_by = NULL") {
			t.Errorf("%s does not reinstate a revoked grant", name)
		}
	}
}

func TestFindProfileGrantsUnionColumns(t *testing.T) {
	// Each branch of the union binds one placeholder: mc_uuid, profile_id, profile_id.
	if got := strings.Count(findProfileGrants, "?"); got != 3 {
		t.Errorf("got %d placeholders, want 3", got)
	}
}
