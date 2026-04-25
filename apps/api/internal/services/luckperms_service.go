package services

import (
	"context"
	stdsql "database/sql"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/domain"
	"github.com/lania-smp/backend/internal/logger"
	"github.com/lania-smp/backend/internal/storage"
	sql "github.com/lania-smp/backend/internal/storage/main"
	"github.com/lania-smp/backend/internal/utils"
)

type LuckpermsService interface {
	SetUserPrefixByMinecraftUUID(ctx context.Context, queries sql.Queries, minecraftUUID uuid.UUID, prefix string) error
}

type luckpermsService struct {
	storage storage.MainStorage
}

func NewLuckpermsService(storage storage.MainStorage) LuckpermsService {
	return &luckpermsService{storage: storage}
}

const prefixPrefix = "prefix."

func (s *luckpermsService) SetUserPrefixByMinecraftUUID(ctx context.Context, queries sql.Queries, minecraftUUID uuid.UUID, prefix string) error {
	finalPrefix := prefixPrefix + "100." + prefix
	logger.Debugf(ctx, "setting user prefix by minecraft uuid: %s, prefix: %s", minecraftUUID.String(), finalPrefix)

	permissions, err := queries.FindLuckpermsPermissionLikeByMinecraftUUID(ctx, minecraftUUID, prefixPrefix)
	if err != nil && err != stdsql.ErrNoRows {
		return utils.NewInternalServerError("failed to find luckperms permission like by minecraft uuid", err)
	}

	if len(permissions) > 0 {
		ids := make([]int64, len(permissions))
		for i, permission := range permissions {
			ids[i] = permission.ID
		}

		err = queries.DeleteLuckpermsPermissionByIDs(ctx, ids)
		if err != nil {
			return utils.NewInternalServerError("failed to delete luckperms permission by ids", err)
		}
	}

	err = queries.InsertLuckpermsPermission(ctx, sql.InsertLuckpermsPermissionParams{
		UUID:       minecraftUUID,
		Permission: finalPrefix,
		Value:      domain.LuckpermsPermissionValueEnabled,
		Server:     domain.LuckpermsPermissionServerGlobal,
		World:      domain.LuckpermsPermissionWorldGlobal,
		Expiry:     0,
	})
	if err != nil {
		return utils.NewInternalServerError("failed to insert luckperms permission", err)
	}
	return nil
}
