package services

import (
	"context"
	stdsql "database/sql"
	"errors"

	mysql "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/commands"
	"github.com/lania-smp/backend/internal/domain"
	sql "github.com/lania-smp/backend/internal/storage/main"
	"github.com/lania-smp/backend/internal/utils"
)

func (s *adminCatalogService) GetPrivileges(ctx context.Context) ([]*domain.Privilege, error) {
	privileges, err := s.storage.Queries().FindPrivileges(ctx)
	if err != nil {
		return nil, utils.NewInternalServerError("failed to get privileges", err)
	}
	return privileges, nil
}

func privilegeWriteError(message string, err error) error {
	var customErr *utils.CustomError
	if errors.As(err, &customErr) {
		return err
	}
	var mysqlErr *mysql.MySQLError
	switch {
	case errors.As(err, &mysqlErr) && mysqlErr.Number == 1062:
		return utils.NewConflictError("name or permission is already used", err)
	case errors.As(err, &mysqlErr) && mysqlErr.Number == 1451:
		return utils.NewConflictError("privilege is owned by players", err)
	}
	return utils.NewInternalServerError(message, err)
}

// savePrivilegePrices writes the price of the privilege under its own tariff, so every privilege sells at its own price.
func savePrivilegePrices(ctx context.Context, queries sql.Queries, id uuid.UUID, prices []commands.PrivilegePrice) error {
	rows := make([]*domain.ProductPrice, len(prices))
	for i, price := range prices {
		rows[i] = &domain.ProductPrice{Name: domain.PrivilegePriceName(id), Currency: price.Currency, Amount: price.Amount}
	}
	return queries.UpsertProductPrices(ctx, domain.PrivilegePriceName(id), rows)
}

func (s *adminCatalogService) CreatePrivilege(ctx context.Context, cmd *commands.SavePrivilegeCommand) (*domain.Privilege, error) {
	id := uuid.New()
	err := s.storage.BeginTx(ctx, func(queries sql.Queries) error {
		if err := queries.InsertPrivilege(ctx, id, cmd.Name, cmd.Names, cmd.Permission); err != nil {
			return err
		}
		return savePrivilegePrices(ctx, queries, id, cmd.Prices)
	})
	if err != nil {
		return nil, privilegeWriteError("failed to create privilege", err)
	}
	return s.privilegeByID(ctx, id)
}

func (s *adminCatalogService) UpdatePrivilege(ctx context.Context, cmd *commands.SavePrivilegeCommand) (*domain.Privilege, error) {
	err := s.storage.BeginTx(ctx, func(queries sql.Queries) error {
		current, err := queries.FindPrivilegeByID(ctx, cmd.ID)
		if errors.Is(err, stdsql.ErrNoRows) {
			return utils.NewNotFoundError("privilege not found", err)
		} else if err != nil {
			return err
		}
		// Owners hold the old node in the game, so it would stay there while the new one is never written.
		if current.Permission != cmd.Permission {
			owners, err := queries.CountPrivilegeOwners(ctx, cmd.ID)
			if err != nil {
				return err
			}
			if owners > 0 {
				return utils.NewConflictError("permission cannot be changed while players own the privilege", nil)
			}
		}
		if err := queries.UpdatePrivilege(ctx, cmd.ID, cmd.Name, cmd.Names, cmd.Permission); err != nil {
			return err
		}
		return savePrivilegePrices(ctx, queries, cmd.ID, cmd.Prices)
	})
	if err != nil {
		return nil, privilegeWriteError("failed to update privilege", err)
	}
	return s.privilegeByID(ctx, cmd.ID)
}

func (s *adminCatalogService) privilegeByID(ctx context.Context, id uuid.UUID) (*domain.Privilege, error) {
	privilege, err := s.storage.Queries().FindPrivilegeByID(ctx, id)
	if errors.Is(err, stdsql.ErrNoRows) {
		return nil, utils.NewNotFoundError("privilege not found", err)
	} else if err != nil {
		return nil, utils.NewInternalServerError("failed to find privilege", err)
	}
	return privilege, nil
}

// DeletePrivilege removes a privilege that no product sells and no player has, with its tariff.
func (s *adminCatalogService) DeletePrivilege(ctx context.Context, id uuid.UUID) error {
	err := s.storage.BeginTx(ctx, func(queries sql.Queries) error {
		if _, err := queries.FindPrivilegeByID(ctx, id); errors.Is(err, stdsql.ErrNoRows) {
			return utils.NewNotFoundError("privilege not found", err)
		} else if err != nil {
			return err
		}
		products, err := queries.FindAdminProducts(ctx)
		if err != nil {
			return err
		}
		for _, product := range products {
			if product.Category == domain.ProductCategoryPrivilege && sameUUID(linkedCosmeticID(product), &id) {
				return utils.NewConflictError("privilege is linked to a product", nil)
			}
		}
		owners, err := queries.CountPrivilegeOwners(ctx, id)
		if err != nil {
			return err
		}
		if owners > 0 {
			return utils.NewConflictError("privilege is owned by players", nil)
		}
		return queries.DeletePrivilege(ctx, id)
	})
	if err != nil {
		return privilegeWriteError("failed to delete privilege", err)
	}
	return nil
}
