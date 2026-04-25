package services

import (
	"context"
	stdsql "database/sql"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/domain"
	"github.com/lania-smp/backend/internal/storage"
	"github.com/lania-smp/backend/internal/utils"
)

type ProductService interface {
	GetProductByID(ctx context.Context, id uuid.UUID) (*domain.Product, error)
	GetProductsByCategory(ctx context.Context, category domain.ProductCategory) ([]*domain.Product, error)
	GetProducts(ctx context.Context) ([]*domain.Product, error)
	GetProductsByIDs(ctx context.Context, ids uuid.UUIDs) ([]*domain.Product, error)
	GetPricesByNames(ctx context.Context, names []domain.ProductPriceName) ([]*domain.ProductPrice, error)
}

type productService struct {
	storage storage.MainStorage
}

func NewProductService(storage storage.MainStorage) ProductService {
	return &productService{storage: storage}
}

func (s *productService) GetProductByID(ctx context.Context, id uuid.UUID) (*domain.Product, error) {
	products, err := s.GetProductsByIDs(ctx, uuid.UUIDs{id})
	if err != nil {
		return nil, err
	} else if len(products) == 0 {
		return nil, utils.NewNotFoundError("product not found", nil)
	}
	return products[0], nil
}

func (s *productService) GetProductsByIDs(ctx context.Context, ids uuid.UUIDs) ([]*domain.Product, error) {
	locale := utils.GetLocaleFromCtx(ctx)
	currency := utils.GetCurrencyFromCtx(ctx)
	products, err := s.storage.Queries().FindProductByIDs(ctx, ids, locale, currency)
	if err == stdsql.ErrNoRows {
		return []*domain.Product{}, nil
	} else if err != nil {
		return nil, utils.NewInternalServerError("failed to get products by ids", err)
	}
	return products, nil
}

func (s *productService) GetProductsByCategory(ctx context.Context, category domain.ProductCategory) ([]*domain.Product, error) {
	locale := utils.GetLocaleFromCtx(ctx)
	currency := utils.GetCurrencyFromCtx(ctx)
	products, err := s.storage.Queries().FindProductsByCategory(ctx, category, currency, locale)
	if err == stdsql.ErrNoRows {
		return []*domain.Product{}, nil
	} else if err != nil {
		return nil, utils.NewInternalServerError("failed to get products by category", err)
	}
	return products, nil
}

func (s *productService) GetProducts(ctx context.Context) ([]*domain.Product, error) {
	locale := utils.GetLocaleFromCtx(ctx)
	currency := utils.GetCurrencyFromCtx(ctx)
	products, err := s.storage.Queries().FindProducts(ctx, locale, currency)
	if err == stdsql.ErrNoRows {
		return []*domain.Product{}, nil
	} else if err != nil {
		return nil, utils.NewInternalServerError("failed to get products", err)
	}
	return products, nil
}

func (s *productService) GetPricesByNames(ctx context.Context, names []domain.ProductPriceName) ([]*domain.ProductPrice, error) {
	prices, err := s.storage.Queries().FindPricesByNames(ctx, names)
	if err != nil {
		return nil, utils.NewInternalServerError("failed to get prices by names", err)
	}
	return prices, nil
}
