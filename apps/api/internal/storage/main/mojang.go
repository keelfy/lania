package sql

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/domain"
)

const findProfileMojangUUIDsByMinecraftUUIDs = `
SELECT mc_uuid, mojang_uuid
FROM profile_mojang_uuids
WHERE mojang_uuid IS NOT NULL AND mc_uuid IN (%s)
`

// FindProfileMojangUUIDsByMinecraftUUIDs returns Mojang UUIDs by minecraft UUIDs.
// Profiles without a known Mojang account are absent from the result.
func (q *queries) FindProfileMojangUUIDsByMinecraftUUIDs(ctx context.Context, mcUUIDs uuid.UUIDs) (map[uuid.UUID]uuid.UUID, error) {
	res := make(map[uuid.UUID]uuid.UUID, len(mcUUIDs))
	if len(mcUUIDs) == 0 {
		return res, nil
	}

	args := make([]any, len(mcUUIDs))
	for i, mcUUID := range mcUUIDs {
		args[i] = mcUUID
	}
	query := strings.Replace(findProfileMojangUUIDsByMinecraftUUIDs, "%s", strings.TrimSuffix(strings.Repeat("?,", len(mcUUIDs)), ","), 1)
	rows, err := q.x.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var mcUUID, mojangUUID uuid.UUID
		if err := rows.Scan(&mcUUID, &mojangUUID); err != nil {
			return nil, err
		}
		res[mcUUID] = mojangUUID
	}
	return res, rows.Err()
}

// Profiles that never were looked up go first, then not found ones that were checked before notCheckedSince.
const findMojangLookupTargets = `
SELECT p.mc_uuid, p.mc_username
FROM profiles p
LEFT JOIN profile_mojang_uuids m ON m.mc_uuid = p.mc_uuid
WHERE m.mc_uuid IS NULL OR (m.mojang_uuid IS NULL AND m.checked_at < ?)
ORDER BY m.checked_at IS NOT NULL, m.checked_at
LIMIT ?
`

func (q *queries) FindMojangLookupTargets(ctx context.Context, notCheckedSince time.Time, limit int) ([]*domain.MojangLookupTarget, error) {
	rows, err := q.x.QueryContext(ctx, findMojangLookupTargets, notCheckedSince, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	targets := make([]*domain.MojangLookupTarget, 0)
	for rows.Next() {
		var target domain.MojangLookupTarget
		if err := rows.Scan(&target.MinecraftUUID, &target.MinecraftUsername); err != nil {
			return nil, err
		}
		targets = append(targets, &target)
	}
	return targets, rows.Err()
}

const upsertProfileMojangUUID = `
INSERT INTO profile_mojang_uuids (mc_uuid, mojang_uuid, checked_at)
VALUES (?, ?, now())
ON DUPLICATE KEY UPDATE mojang_uuid = VALUES(mojang_uuid), checked_at = VALUES(checked_at)
`

// UpsertProfileMojangUUID stores the lookup result. A nil mojangUUID means Mojang has no such username.
func (q *queries) UpsertProfileMojangUUID(ctx context.Context, mcUUID uuid.UUID, mojangUUID *uuid.UUID) error {
	_, err := q.x.ExecContext(ctx, upsertProfileMojangUUID, mcUUID, mojangUUID)
	return err
}
