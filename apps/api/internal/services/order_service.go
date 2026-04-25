package services

import (
	"context"
	stdsql "database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/commands"
	"github.com/lania-smp/backend/internal/domain"
	"github.com/lania-smp/backend/internal/storage"
	sql "github.com/lania-smp/backend/internal/storage/main"
	"github.com/lania-smp/backend/internal/utils"
)

type OrderService interface {
	CreateOrder(ctx context.Context, queries sql.Queries, amounts []*domain.OrderAmounts, cmd *commands.CreateOrderCommand) (uuid.UUID, error)
	CreateOrderItem(ctx context.Context, queries sql.Queries, orderID uuid.UUID, product *domain.Product, profileID uuid.UUID, seasonID uuid.UUID, quantity int, amounts []*domain.OrderAmounts) error
	GetOrderByID(ctx context.Context, id uuid.UUID) (*domain.Order, error)
	GetOrderItemsByOrderID(ctx context.Context, id uuid.UUID) ([]*domain.OrderItem, error)
	UpdateOrderStatusByID(ctx context.Context, queries sql.Queries, id uuid.UUID, status domain.OrderStatus, updatedBy *uuid.UUID) error
	CompleteOrder(ctx context.Context, order *domain.Order) error
	GetOrdersByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Order, error)
	GetOrderByExternalID(ctx context.Context, externalID string) (*domain.Order, error)
	UpdateOrderExternalIDByID(ctx context.Context, queries sql.Queries, id uuid.UUID, externalID string) error
}

type orderService struct {
	storage                 storage.MainStorage
	freekassaService        FreekassaService
	accessService           AccessService
	profileCosmeticsService ProfileCosmeticsService
	productService          ProductService
	basketService           BasketService
	profileService          ProfileService
}

func NewOrderService(
	storage storage.MainStorage,
	freekassaService FreekassaService,
	accessService AccessService,
	profileCosmeticsService ProfileCosmeticsService,
	productService ProductService,
	basketService BasketService,
	profileService ProfileService,
) OrderService {
	return &orderService{
		storage:                 storage,
		freekassaService:        freekassaService,
		accessService:           accessService,
		profileCosmeticsService: profileCosmeticsService,
		productService:          productService,
		basketService:           basketService,
		profileService:          profileService,
	}
}

func (s *orderService) CreateOrder(ctx context.Context, queries sql.Queries, amounts []*domain.OrderAmounts, cmd *commands.CreateOrderCommand) (uuid.UUID, error) {
	authUserID := utils.GetUserIDFromContextOrNil(ctx)
	orderID, err := queries.InsertOrder(ctx, sql.InsertOrderParams{
		ID:        uuid.New(),
		UserID:    cmd.UserID,
		Amounts:   amounts,
		Status:    domain.OrderStatusCreated,
		CreatedBy: authUserID,
		UpdatedBy: authUserID,
	})
	if err != nil {
		return uuid.Nil, utils.NewInternalServerError("failed to insert order", err)
	}
	return orderID, nil
}

func (s *orderService) CreateOrderItem(ctx context.Context, queries sql.Queries,
	orderID uuid.UUID,
	product *domain.Product,
	profileID, seasonID uuid.UUID,
	quantity int,
	amounts []*domain.OrderAmounts,
) error {
	err := queries.InsertOrderItem(ctx, sql.InsertOrderItemParams{
		OrderID:   orderID,
		ProductID: product.ID,
		ProfileID: profileID,
		SeasonID:  seasonID,
		Amounts:   amounts,
		Quantity:  quantity,
	})
	if err != nil {
		return utils.NewInternalServerError("failed to insert order item", err)
	}
	return nil
}

func (s *orderService) GetOrderByID(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
	order, err := s.storage.Queries().FindOrderByID(ctx, id)
	if err != nil {
		return nil, utils.NewInternalServerError("failed to get order by id", err)
	}
	return order, nil
}

func (s *orderService) GetOrderItemsByOrderID(ctx context.Context, id uuid.UUID) ([]*domain.OrderItem, error) {
	items, err := s.storage.Queries().FindItemsByOrderID(ctx, id)
	if err != nil {
		return nil, utils.NewInternalServerError("failed to get order items by order id", err)
	}
	return items, nil
}

func (s *orderService) UpdateOrderStatusByID(ctx context.Context, queries sql.Queries, id uuid.UUID, status domain.OrderStatus, updatedBy *uuid.UUID) error {
	err := queries.UpdateOrderStatusByID(ctx, id, status, updatedBy)
	if err != nil {
		return utils.NewInternalServerError("failed to update order status by id", err)
	}
	return nil
}

func (s *orderService) GetOrdersByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Order, error) {
	orders, err := s.storage.Queries().FindOrdersByUserID(ctx, userID)
	if err == stdsql.ErrNoRows {
		return []*domain.Order{}, nil
	} else if err != nil {
		return nil, utils.NewInternalServerError("failed to get orders by user id", err)
	}
	return orders, nil
}

func (s *orderService) CompleteOrder(ctx context.Context, order *domain.Order) error {
	if order.Status == domain.OrderStatusCompleted || order.Status == domain.OrderStatusFailed {
		return utils.NewBadRequestError("order already processed", nil)
	}

	err := s.UpdateOrderStatusByID(ctx, s.storage.Queries(), order.ID, domain.OrderStatusProcessing, nil)
	if err != nil {
		return err
	}

	err = s.handleOrderCompletion(ctx, order)
	if err != nil {
		_ = s.UpdateOrderStatusByID(ctx, s.storage.Queries(), order.ID, domain.OrderStatusFailed, nil)
		utils.LogCustomError(ctx, err)
	}

	return nil
}

func (s *orderService) handleOrderCompletion(ctx context.Context, order *domain.Order) error {
	items, err := s.GetOrderItemsByOrderID(ctx, order.ID)
	if err != nil {
		return err
	}

	productIDs := make([]uuid.UUID, len(items))
	for i, item := range items {
		productIDs[i] = item.ProductID
	}

	products, err := s.productService.GetProductsByIDs(ctx, productIDs)
	if err != nil {
		return err
	}

	err = s.storage.BeginTx(ctx, func(queries sql.Queries) error {
		for _, item := range items {
			var product *domain.Product
			for _, p := range products {
				if p.ID == item.ProductID {
					product = p
					break
				}
			}

			err := s.handleOrderItemCompletion(ctx, queries, item, product)
			if err != nil {
				return err
			}
		}

		err := s.UpdateOrderStatusByID(ctx, queries, order.ID, domain.OrderStatusCompleted, nil)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *orderService) handleOrderItemCompletion(ctx context.Context, queries sql.Queries, item *domain.OrderItem, product *domain.Product) error {
	switch product.Category {
	case domain.ProductCategoryUpgrade:
		var metadata domain.UpgradeProductMetadata
		err := json.Unmarshal(product.Metadata, &metadata)
		if err != nil {
			return err
		}

		profile, err := s.profileService.GetProfileByID(ctx, item.ProfileID)
		if err != nil {
			return err
		}

		if metadata.Action == domain.ProductUpgradeActionSeasonAccess {
			// TODO: check if profile has access for the season
			return s.accessService.ObtainAccessForProfile(ctx, item.SeasonID, profile, domain.AccessSourceFreekassa, &item.ID)
		} else {
			return fmt.Errorf("unknown upgrade action: %s", metadata.Action)
		}
	case domain.ProductCategoryNameColor:
		// TODO: avoid hitting unique constraint error
		var metadata domain.NameColorProductMetadata
		err := json.Unmarshal(product.Metadata, &metadata)
		if err != nil {
			return err
		}
		return s.profileCosmeticsService.AddProfileNameColorOption(ctx, queries, item.ProfileID, metadata.NameColorID, &item.SeasonID, &item.ID)
	case domain.ProductCategoryNamePrefix:
		var metadata domain.NamePrefixProductMetadata
		err := json.Unmarshal(product.Metadata, &metadata)
		if err != nil {
			return err
		}
		return s.profileCosmeticsService.AddProfileNameGlythOption(ctx, queries, item.ProfileID, metadata.NamePrefixID, &item.SeasonID, &item.ID)
	default:
		return fmt.Errorf("unknown product category: %s", product.Category)
	}
}

func (s *orderService) GetOrderByExternalID(ctx context.Context, externalID string) (*domain.Order, error) {
	order, err := s.storage.Queries().FindOrderByExternalID(ctx, externalID)
	if err == stdsql.ErrNoRows {
		return nil, utils.NewNotFoundError("order not found", err)
	} else if err != nil {
		return nil, utils.NewInternalServerError("failed to get order by external id", err)
	}
	return order, nil
}

func (s *orderService) UpdateOrderExternalIDByID(ctx context.Context, queries sql.Queries, id uuid.UUID, externalID string) error {
	err := queries.UpdateOrderExternalIDByID(ctx, id, externalID)
	if err == stdsql.ErrNoRows {
		return utils.NewNotFoundError("order not found", err)
	} else if err != nil {
		return utils.NewInternalServerError("failed to update order external id by id", err)
	}
	return nil
}
