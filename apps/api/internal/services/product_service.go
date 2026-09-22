package services

import (
	"context"
	stdsql "database/sql"
	"encoding/json"

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
	GetProductsByIDsIncludingInactive(ctx context.Context, ids uuid.UUIDs) ([]*domain.Product, error)
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
	return products, s.hydrateCosmeticMetadata(ctx, products)
}

func (s *productService) GetProductsByIDsIncludingInactive(ctx context.Context, ids uuid.UUIDs) ([]*domain.Product, error) {
	locale := utils.GetLocaleFromCtx(ctx)
	currency := utils.GetCurrencyFromCtx(ctx)
	products, err := s.storage.Queries().FindProductByIDsIncludingInactive(ctx, ids, locale, currency)
	if err != nil {
		return nil, utils.NewInternalServerError("failed to get products by ids", err)
	}
	return products, s.hydrateCosmeticMetadata(ctx, products)
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
	return products, s.hydrateCosmeticMetadata(ctx, products)
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
	return products, s.hydrateCosmeticMetadata(ctx, products)
}

func (s *productService) hydrateCosmeticMetadata(ctx context.Context, products []*domain.Product) error {
	needColors, needPrefixes := false, false
	for _, product := range products {
		needColors = needColors || product.Category == domain.ProductCategoryNameColor
		needPrefixes = needPrefixes || product.Category == domain.ProductCategoryNamePrefix
	}
	colors := map[uuid.UUID]*domain.NameColor{}
	if needColors {
		items, err := s.storage.Queries().FindNameColors(ctx)
		if err != nil {
			return utils.NewInternalServerError("failed to load name colors", err)
		}
		for _, item := range items {
			colors[item.ID] = item
		}
	}
	prefixes := map[uuid.UUID]*domain.NamePrefix{}
	if needPrefixes {
		items, err := s.storage.Queries().FindNamePrefixes(ctx)
		if err != nil {
			return utils.NewInternalServerError("failed to load name prefixes", err)
		}
		for _, item := range items {
			prefixes[item.ID] = item
		}
	}
	for _, product := range products {
		switch product.Category {
		case domain.ProductCategoryNameColor:
			var metadata domain.NameColorProductMetadata
			if err := json.Unmarshal(product.Metadata, &metadata); err != nil {
				return utils.NewInternalServerError("failed to decode name color product", err)
			}
			item := colors[metadata.NameColorID]
			if item == nil {
				return utils.NewInternalServerError("name color product references missing cosmetic", nil)
			}
			metadata.Colors = item.Metadata.Colors
			payload, err := json.Marshal(metadata)
			if err != nil {
				return utils.NewInternalServerError("failed to encode name color product", err)
			}
			product.Metadata = payload
		case domain.ProductCategoryNamePrefix:
			var metadata domain.NamePrefixProductMetadata
			if err := json.Unmarshal(product.Metadata, &metadata); err != nil {
				return utils.NewInternalServerError("failed to decode name prefix product", err)
			}
			item := prefixes[metadata.NamePrefixID]
			if item == nil {
				return utils.NewInternalServerError("name prefix product references missing cosmetic", nil)
			}
			metadata.Prefix = item.Metadata.Image
			payload, err := json.Marshal(metadata)
			if err != nil {
				return utils.NewInternalServerError("failed to encode name prefix product", err)
			}
			product.Metadata = payload
		}
	}
	return nil
}

func (s *productService) GetPricesByNames(ctx context.Context, names []domain.ProductPriceName) ([]*domain.ProductPrice, error) {
	prices, err := s.storage.Queries().FindPricesByNames(ctx, names)
	if err != nil {
		return nil, utils.NewInternalServerError("failed to get prices by names", err)
	}
	return prices, nil
}
