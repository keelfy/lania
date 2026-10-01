package storage

import (
	"context"
	stdsql "database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/lania-smp/shell/internal/config"
)

// SkinsRestorer skin types, as it keeps them in skin_type.
const (
	SkinTypePlayer = "PLAYER"
	SkinTypeURL    = "URL"
	SkinTypeCustom = "CUSTOM"
	SkinTypeLegacy = "LEGACY"
)

// SkinRecord is the skin a player chose with SkinsRestorer.
type SkinRecord struct {
	// Type is one of the SkinType constants.
	Type string
	// Identifier is the Mojang UUID for PLAYER, the link for URL and the name for CUSTOM and LEGACY.
	Identifier string
	// Variant is CLASSIC or SLIM, nil when SkinsRestorer picked none.
	Variant *string
	// Value is the base64 textures property, nil when SkinsRestorer has not stored the skin data yet.
	Value *string
}

// SkinsRestorerStorage reads the skins players chose with SkinsRestorer. SkinsRestorer is the only writer: shell
// changes skins with its commands.
type SkinsRestorerStorage interface {
	// FindPlayerSkins returns the skin of every player that chose one.
	FindPlayerSkins(ctx context.Context, mcUUIDs uuid.UUIDs) (map[uuid.UUID]*SkinRecord, error)
}

type skinsRestorerStorage struct {
	db *stdsql.DB
}

func NewSkinsRestorerStorage(db *stdsql.DB) SkinsRestorerStorage {
	return &skinsRestorerStorage{db: db}
}

// A link skin is stored per variant. A player who chose no variant wears the one SkinsRestorer detected for the
// link, kept in url_index.
const findPlayerSkins = `
SELECT p.uuid, p.skin_type, p.skin_identifier, p.skin_variant, COALESCE(ps.value, us.value, cs.value, ls.value)
FROM %[1]s p
LEFT JOIN %[2]s ps ON p.skin_type = 'PLAYER' AND ps.uuid = p.skin_identifier
LEFT JOIN %[3]s ui ON p.skin_type = 'URL' AND ui.url = p.skin_identifier
LEFT JOIN %[4]s us ON p.skin_type = 'URL' AND us.url = p.skin_identifier AND us.skin_variant = COALESCE(p.skin_variant, ui.skin_variant)
LEFT JOIN %[5]s cs ON p.skin_type = 'CUSTOM' AND cs.name = p.skin_identifier
LEFT JOIN %[6]s ls ON p.skin_type = 'LEGACY' AND ls.name = p.skin_identifier
WHERE p.uuid IN (%[7]s) AND p.skin_identifier IS NOT NULL AND p.skin_type IS NOT NULL
`

func (s *skinsRestorerStorage) FindPlayerSkins(ctx context.Context, mcUUIDs uuid.UUIDs) (map[uuid.UUID]*SkinRecord, error) {
	skins := make(map[uuid.UUID]*SkinRecord)
	if len(mcUUIDs) == 0 {
		return skins, nil
	}

	placeholders, args := uuidArgs(mcUUIDs)
	query := fmt.Sprintf(findPlayerSkins,
		skinsRestorerTable("players"),
		skinsRestorerTable("player_skins"),
		skinsRestorerTable("url_index"),
		skinsRestorerTable("url_skins"),
		skinsRestorerTable("custom_skins"),
		skinsRestorerTable("legacy_skins"),
		placeholders,
	)
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var mcUUID uuid.UUID
		var variant, value stdsql.NullString
		skin := &SkinRecord{}
		if err := rows.Scan(&mcUUID, &skin.Type, &skin.Identifier, &variant, &value); err != nil {
			return nil, err
		}
		if variant.Valid {
			skin.Variant = &variant.String
		}
		if value.Valid && value.String != "" {
			skin.Value = &value.String
		}
		skins[mcUUID] = skin
	}
	return skins, rows.Err()
}

func skinsRestorerTable(name string) string {
	return fmt.Sprintf("`%s%s`", config.GetSkinsRestorerTablePrefix(), name)
}
