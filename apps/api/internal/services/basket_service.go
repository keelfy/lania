package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/commands"
	"github.com/lania-smp/backend/internal/domain"
	"github.com/lania-smp/backend/internal/storage"
	sql "github.com/lania-smp/backend/internal/storage/main"
	"github.com/lania-smp/backend/internal/utils"
)

type BasketService interface {
	GetBasketItemsByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.BasketItem, error)
	ClearBasketItemsByUserID(ctx context.Context, queries sql.Queries, userID uuid.UUID) error
	DeleteBasketItemByIDs(ctx context.Context, queries sql.Queries, ids []uuid.UUID) error
	AddBasketItem(ctx context.Context, queries sql.Queries, cmd *commands.AddBasketItemCommand) error
}

type basketService struct {
	storage storage.MainStorage
}

func NewBasketService(storage storage.MainStorage) BasketService {
	return &basketService{storage: storage}
}

func (s *basketService) GetBasketItemsByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.BasketItem, error) {
	items, err := s.storage.Queries().FindBasketItemsByUserID(ctx, userID)
	if err != nil {
		return nil, utils.NewInternalServerError("failed to get basket items by user id", err)
	}
	return items, nil
}

func (s *basketService) ClearBasketItemsByUserID(ctx context.Context, queries sql.Queries, userID uuid.UUID) error {
	err := queries.ClearBasketItemsByUserID(ctx, userID)
	if err != nil {
		return utils.NewInternalServerError("failed to clear basket items by user id", err)
	}
	return nil
}

func (s *basketService) DeleteBasketItemByIDs(ctx context.Context, queries sql.Queries, ids []uuid.UUID) error {
	err := queries.DeleteBasketItemByIDs(ctx, ids)
	if err != nil {
		return utils.NewInternalServerError("failed to delete basket item by ids", err)
	}
	return nil
}

func (s *basketService) AddBasketItem(ctx context.Context, queries sql.Queries, cmd *commands.AddBasketItemCommand) error {
	authUserID, err := utils.GetUserIDFromCtx(ctx)
	if err != nil {
		return err
	}

	err = queries.InsertBasketItem(ctx, sql.InsertBasketItemParams{
		UserID:    cmd.UserID,
		ProductID: cmd.ProductID,
		ProfileID: cmd.ProfileID,
		Quantity:  cmd.Quantity,
		CreatedBy: authUserID,
	})
	if err != nil {
		return utils.NewInternalServerError("failed to insert basket item", err)
	}
	return nil
}
