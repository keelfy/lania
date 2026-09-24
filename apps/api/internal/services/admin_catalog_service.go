package services

import (
	"context"
	stdsql "database/sql"
	"encoding/json"
	"errors"
	"time"

	mysql "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/commands"
	"github.com/lania-smp/backend/internal/domain"
	"github.com/lania-smp/backend/internal/storage"
	sql "github.com/lania-smp/backend/internal/storage/main"
	"github.com/lania-smp/backend/internal/utils"
)

type AdminCatalogService interface {
	GetCosmetics(ctx context.Context) (*domain.CosmeticsCatalog, error)
	CreateNameColor(ctx context.Context, cmd *commands.SaveNameColorCommand) (*domain.CosmeticsCatalog, error)
	UpdateNameColor(ctx context.Context, cmd *commands.SaveNameColorCommand) (*domain.CosmeticsCatalog, error)
	CreateNamePrefix(ctx context.Context, cmd *commands.SaveNamePrefixCommand) (*domain.CosmeticsCatalog, error)
	UpdateNamePrefix(ctx context.Context, cmd *commands.SaveNamePrefixCommand) (*domain.CosmeticsCatalog, error)
	GetProducts(ctx context.Context) ([]*domain.Product, error)
	CreateProduct(ctx context.Context, cmd *commands.SaveProductCommand) (*domain.Product, error)
	UpdateProduct(ctx context.Context, cmd *commands.SaveProductCommand) (*domain.Product, error)
	DeleteProduct(ctx context.Context, id uuid.UUID) error
}

type adminCatalogService struct{ storage storage.MainStorage }

func NewAdminCatalogService(storage storage.MainStorage) AdminCatalogService {
	return &adminCatalogService{storage: storage}
}

func (s *adminCatalogService) GetCosmetics(ctx context.Context) (*domain.CosmeticsCatalog, error) {
	colors, err := s.storage.Queries().FindNameColors(ctx)
	if err != nil {
		return nil, utils.NewInternalServerError("failed to get name colors", err)
	}
	prefixes, err := s.storage.Queries().FindNamePrefixes(ctx)
	if err != nil {
		return nil, utils.NewInternalServerError("failed to get name prefixes", err)
	}
	return &domain.CosmeticsCatalog{NameColors: colors, NamePrefixes: prefixes}, nil
}

func catalogWriteError(message string, err error) error {
	var customErr *utils.CustomError
	if errors.As(err, &customErr) {
		return err
	}
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
		return utils.NewConflictError("name or EasyDonate ID is already used", err)
	}
	return utils.NewInternalServerError(message, err)
}

func (s *adminCatalogService) CreateNameColor(ctx context.Context, cmd *commands.SaveNameColorCommand) (*domain.CosmeticsCatalog, error) {
	if err := s.storage.Queries().InsertNameColor(ctx, uuid.New(), cmd.Name, cmd.Colors); err != nil {
		return nil, catalogWriteError("failed to create name color", err)
	}
	return s.GetCosmetics(ctx)
}

func (s *adminCatalogService) UpdateNameColor(ctx context.Context, cmd *commands.SaveNameColorCommand) (*domain.CosmeticsCatalog, error) {
	if _, err := s.storage.Queries().FindNameColorByID(ctx, cmd.ID); errors.Is(err, stdsql.ErrNoRows) {
		return nil, utils.NewNotFoundError("name color not found", err)
	} else if err != nil {
		return nil, utils.NewInternalServerError("failed to find name color", err)
	}
	if err := s.storage.Queries().UpdateNameColor(ctx, cmd.ID, cmd.Name, cmd.Colors); err != nil {
		return nil, catalogWriteError("failed to update name color", err)
	}
	return s.GetCosmetics(ctx)
}

func prefixMetadata(cmd *commands.SaveNamePrefixCommand) domain.NamePrefixMetadata {
	return domain.NamePrefixMetadata{Prefix: cmd.Prefix, Image: cmd.Image, NoSpace: cmd.NoSpace}
}

func (s *adminCatalogService) CreateNamePrefix(ctx context.Context, cmd *commands.SaveNamePrefixCommand) (*domain.CosmeticsCatalog, error) {
	if err := s.storage.Queries().InsertNamePrefix(ctx, uuid.New(), cmd.Name, prefixMetadata(cmd)); err != nil {
		return nil, catalogWriteError("failed to create name prefix", err)
	}
	return s.GetCosmetics(ctx)
}

func (s *adminCatalogService) UpdateNamePrefix(ctx context.Context, cmd *commands.SaveNamePrefixCommand) (*domain.CosmeticsCatalog, error) {
	if _, err := s.storage.Queries().FindNamePrefixByID(ctx, cmd.ID); errors.Is(err, stdsql.ErrNoRows) {
		return nil, utils.NewNotFoundError("name prefix not found", err)
	} else if err != nil {
		return nil, utils.NewInternalServerError("failed to find name prefix", err)
	}
	if err := s.storage.Queries().UpdateNamePrefix(ctx, cmd.ID, cmd.Name, prefixMetadata(cmd)); err != nil {
		return nil, catalogWriteError("failed to update name prefix", err)
	}
	return s.GetCosmetics(ctx)
}

func (s *adminCatalogService) GetProducts(ctx context.Context) ([]*domain.Product, error) {
	products, err := s.storage.Queries().FindAdminProducts(ctx)
	if err != nil {
		return nil, utils.NewInternalServerError("failed to get admin products", err)
	}
	return products, nil
}

func productMetadata(ctx context.Context, queries sql.Queries, cmd *commands.SaveProductCommand) (json.RawMessage, error) {
	var value any
	switch cmd.Category {
	case domain.ProductCategoryUpgrade:
		value = domain.UpgradeProductMetadata{Action: domain.ProductUpgradeActionSeasonAccess}
	case domain.ProductCategoryNameColor:
		if _, err := queries.FindNameColorByID(ctx, *cmd.CosmeticID); errors.Is(err, stdsql.ErrNoRows) {
			return nil, utils.NewBadRequestError("name color not found", err)
		} else if err != nil {
			return nil, err
		}
		value = domain.NameColorProductMetadata{NameColorID: *cmd.CosmeticID}
	case domain.ProductCategoryNamePrefix:
		if _, err := queries.FindNamePrefixByID(ctx, *cmd.CosmeticID); errors.Is(err, stdsql.ErrNoRows) {
			return nil, utils.NewBadRequestError("name prefix not found", err)
		} else if err != nil {
			return nil, err
		}
		value = domain.NamePrefixProductMetadata{NamePrefixID: *cmd.CosmeticID}
	}
	payload, err := json.Marshal(value)
	return payload, err
}

func linkedCosmeticID(product *domain.Product) *uuid.UUID {
	var id uuid.UUID
	switch product.Category {
	case domain.ProductCategoryNameColor:
		var metadata domain.NameColorProductMetadata
		if json.Unmarshal(product.Metadata, &metadata) == nil {
			id = metadata.NameColorID
		}
	case domain.ProductCategoryNamePrefix:
		var metadata domain.NamePrefixProductMetadata
		if json.Unmarshal(product.Metadata, &metadata) == nil {
			id = metadata.NamePrefixID
		}
	}
	if id == uuid.Nil {
		return nil
	}
	return &id
}

func sameUUID(left, right *uuid.UUID) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}

func (s *adminCatalogService) validateTariff(ctx context.Context, cmd *commands.SaveProductCommand) error {
	prices, err := s.storage.Queries().FindPricesByNames(ctx, []domain.ProductPriceName{cmd.PriceName})
	if err != nil {
		return utils.NewInternalServerError("failed to get product tariff", err)
	}
	if len(prices) == 0 {
		return utils.NewBadRequestError("product tariff not found", nil)
	}
	if cmd.IsActive {
		currencies := map[domain.Currency]bool{}
		for _, price := range prices {
			currencies[price.Currency] = true
		}
		for _, currency := range domain.AllowedCurrencies {
			if !currencies[currency] {
				return utils.NewConflictError("product tariff has no "+string(currency)+" price", nil)
			}
		}
	}
	return nil
}

func (s *adminCatalogService) saveProduct(ctx context.Context, cmd *commands.SaveProductCommand, create bool) error {
	if err := s.validateTariff(ctx, cmd); err != nil {
		return err
	}
	err := s.storage.BeginTx(ctx, func(queries sql.Queries) error {
		metadata, err := productMetadata(ctx, queries, cmd)
		if err != nil {
			return err
		}
		params := sql.SaveProductParams{ID: cmd.ID, Category: cmd.Category, PriceName: cmd.PriceName, Metadata: metadata, IsActive: cmd.IsActive, UpdatedBy: utils.GetUserIDFromContextOrNil(ctx), UpdatedAt: time.Now()}
		if create {
			err = queries.InsertProduct(ctx, params)
		} else {
			err = queries.UpdateProduct(ctx, params)
		}
		if err != nil {
			return err
		}
		for _, localization := range cmd.Localizations {
			if err := queries.UpsertProductLocalization(ctx, cmd.ID, localization.Locale, localization.Name, localization.Description); err != nil {
				return err
			}
		}
		if cmd.EasyDonateProductID == nil {
			return queries.DeleteEDProduct(ctx, cmd.ID)
		}
		return queries.UpsertEDProduct(ctx, cmd.ID, *cmd.EasyDonateProductID)
	})
	if err != nil {
		return catalogWriteError("failed to save product", err)
	}
	return nil
}

func (s *adminCatalogService) productByID(ctx context.Context, id uuid.UUID) (*domain.Product, error) {
	products, err := s.GetProducts(ctx)
	if err != nil {
		return nil, err
	}
	for _, product := range products {
		if product.ID == id {
			return product, nil
		}
	}
	return nil, utils.NewNotFoundError("product not found", nil)
}

func (s *adminCatalogService) CreateProduct(ctx context.Context, cmd *commands.SaveProductCommand) (*domain.Product, error) {
	if cmd.CosmeticID != nil {
		products, err := s.GetProducts(ctx)
		if err != nil {
			return nil, err
		}
		for _, product := range products {
			if product.Category == cmd.Category && sameUUID(linkedCosmeticID(product), cmd.CosmeticID) {
				return nil, utils.NewConflictError("product for this cosmetic already exists", nil)
			}
		}
	}
	cmd.ID = uuid.New()
	if err := s.saveProduct(ctx, cmd, true); err != nil {
		return nil, err
	}
	return s.productByID(ctx, cmd.ID)
}

func (s *adminCatalogService) UpdateProduct(ctx context.Context, cmd *commands.SaveProductCommand) (*domain.Product, error) {
	current, err := s.productByID(ctx, cmd.ID)
	if err != nil {
		return nil, err
	}
	if current.Category != cmd.Category || !sameUUID(linkedCosmeticID(current), cmd.CosmeticID) {
		return nil, utils.NewConflictError("product category and cosmetic cannot be changed", nil)
	}
	if err := s.saveProduct(ctx, cmd, false); err != nil {
		return nil, err
	}
	return s.productByID(ctx, cmd.ID)
}

// DeleteProduct removes a draft. Published products and products that were ever ordered stay.
func (s *adminCatalogService) DeleteProduct(ctx context.Context, id uuid.UUID) error {
	product, err := s.productByID(ctx, id)
	if err != nil {
		return err
	}
	if product.IsActive {
		return utils.NewConflictError("only draft products can be deleted", nil)
	}
	err = s.storage.BeginTx(ctx, func(queries sql.Queries) error {
		return queries.DeleteProduct(ctx, id)
	})
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1451 {
		return utils.NewConflictError("product has orders and cannot be deleted", err)
	}
	if err != nil {
		return utils.NewInternalServerError("failed to delete product", err)
	}
	return nil
}
