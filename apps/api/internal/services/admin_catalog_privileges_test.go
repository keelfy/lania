package services

import (
	"context"
	stdsql "database/sql"
	"encoding/json"
	"net/http"
	"testing"

	mysql "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/commands"
	"github.com/lania-smp/backend/internal/domain"
	"github.com/lania-smp/backend/internal/storage"
	sql "github.com/lania-smp/backend/internal/storage/main"
	"github.com/lania-smp/backend/internal/utils"
)

// privilegeCatalogQueries keeps privileges, tariffs and products in memory.
type privilegeCatalogQueries struct {
	sql.Queries
	privileges map[uuid.UUID]*domain.Privilege
	// prices holds the amount per currency of every tariff.
	prices    map[domain.ProductPriceName]map[domain.Currency]float64
	products  []*domain.Product
	owners    int
	insertErr error
}

func newPrivilegeCatalogQueries() *privilegeCatalogQueries {
	return &privilegeCatalogQueries{
		privileges: map[uuid.UUID]*domain.Privilege{},
		prices:     map[domain.ProductPriceName]map[domain.Currency]float64{},
	}
}

func (q *privilegeCatalogQueries) withPrices(privilege *domain.Privilege) *domain.Privilege {
	copied := *privilege
	copied.Prices = nil
	for _, currency := range domain.AllowedCurrencies {
		if amount, ok := q.prices[domain.PrivilegePriceName(privilege.ID)][currency]; ok {
			copied.Prices = append(copied.Prices, &domain.ProductPrice{Name: domain.PrivilegePriceName(privilege.ID), Currency: currency, Amount: amount})
		}
	}
	return &copied
}

func (q *privilegeCatalogQueries) FindPrivilegeByID(_ context.Context, id uuid.UUID) (*domain.Privilege, error) {
	privilege, ok := q.privileges[id]
	if !ok {
		return nil, stdsql.ErrNoRows
	}
	return q.withPrices(privilege), nil
}

func (q *privilegeCatalogQueries) FindPrivileges(context.Context) ([]*domain.Privilege, error) {
	privileges := make([]*domain.Privilege, 0, len(q.privileges))
	for _, privilege := range q.privileges {
		privileges = append(privileges, q.withPrices(privilege))
	}
	return privileges, nil
}

func (q *privilegeCatalogQueries) InsertPrivilege(_ context.Context, id uuid.UUID, name string, names domain.CosmeticNames, permission string) error {
	if q.insertErr != nil {
		return q.insertErr
	}
	q.privileges[id] = &domain.Privilege{ID: id, Name: name, Names: names, Permission: permission}
	return nil
}

func (q *privilegeCatalogQueries) UpdatePrivilege(_ context.Context, id uuid.UUID, name string, names domain.CosmeticNames, permission string) error {
	q.privileges[id] = &domain.Privilege{ID: id, Name: name, Names: names, Permission: permission}
	return nil
}

func (q *privilegeCatalogQueries) CountPrivilegeOwners(context.Context, uuid.UUID) (int, error) {
	return q.owners, nil
}

func (q *privilegeCatalogQueries) DeletePrivilege(_ context.Context, id uuid.UUID) error {
	delete(q.privileges, id)
	delete(q.prices, domain.PrivilegePriceName(id))
	return nil
}

func (q *privilegeCatalogQueries) UpsertProductPrices(_ context.Context, name domain.ProductPriceName, prices []*domain.ProductPrice) error {
	if q.prices[name] == nil {
		q.prices[name] = map[domain.Currency]float64{}
	}
	for _, price := range prices {
		q.prices[name][price.Currency] = price.Amount
	}
	return nil
}

func (q *privilegeCatalogQueries) FindPricesByNames(_ context.Context, names []domain.ProductPriceName) ([]*domain.ProductPrice, error) {
	prices := make([]*domain.ProductPrice, 0)
	for _, name := range names {
		for currency, amount := range q.prices[name] {
			prices = append(prices, &domain.ProductPrice{Name: name, Currency: currency, Amount: amount})
		}
	}
	return prices, nil
}

func (q *privilegeCatalogQueries) FindAdminProducts(context.Context) ([]*domain.Product, error) {
	return q.products, nil
}

func (q *privilegeCatalogQueries) InsertProduct(_ context.Context, arg sql.SaveProductParams) error {
	q.products = append(q.products, &domain.Product{ID: arg.ID, Category: arg.Category, PriceName: arg.PriceName, Metadata: arg.Metadata, IsActive: arg.IsActive})
	return nil
}

func (q *privilegeCatalogQueries) UpsertProductLocalization(context.Context, uuid.UUID, string, string, string) error {
	return nil
}

func (q *privilegeCatalogQueries) UpsertEDProduct(context.Context, uuid.UUID, int64) error {
	return nil
}

func (q *privilegeCatalogQueries) DeleteEDProduct(context.Context, uuid.UUID) error { return nil }

type privilegeCatalogStorage struct {
	storage.MainStorage
	queries *privilegeCatalogQueries
	txCalls int
}

func (s *privilegeCatalogStorage) Queries() sql.Queries { return s.queries }

func (s *privilegeCatalogStorage) BeginTx(_ context.Context, fn func(sql.Queries) error) error {
	s.txCalls++
	return fn(s.queries)
}

func newPrivilegeCatalog() (AdminCatalogService, *privilegeCatalogQueries, *privilegeCatalogStorage) {
	queries := newPrivilegeCatalogQueries()
	store := &privilegeCatalogStorage{queries: queries}
	return NewAdminCatalogService(store), queries, store
}

func privilegeCommand(permission string, amount float64) *commands.SavePrivilegeCommand {
	prices := make([]commands.PrivilegePrice, len(domain.AllowedCurrencies))
	for i, currency := range domain.AllowedCurrencies {
		prices[i] = commands.PrivilegePrice{Currency: currency, Amount: amount}
	}
	return &commands.SavePrivilegeCommand{Name: permission, Permission: permission, Prices: prices}
}

func TestAdminCatalogService_CreatePrivilege(t *testing.T) {
	ctx := context.Background()

	t.Run("stores the privilege and its own tariff in one transaction", func(t *testing.T) {
		service, queries, store := newPrivilegeCatalog()

		privilege, err := service.CreatePrivilege(ctx, privilegeCommand("homes.commands.*", 99))
		if err != nil {
			t.Fatal(err)
		}

		if store.txCalls != 1 {
			t.Errorf("got %d transactions, want 1", store.txCalls)
		}
		tariff := domain.PrivilegePriceName(privilege.ID)
		if len(queries.prices[tariff]) != len(domain.AllowedCurrencies) || queries.prices[tariff][domain.CurrencyRUB] != 99 {
			t.Errorf("tariff %s = %v, want 99 in every currency", tariff, queries.prices[tariff])
		}
		if len(privilege.Prices) != len(domain.AllowedCurrencies) {
			t.Errorf("got %d prices back, want one per currency", len(privilege.Prices))
		}
	})

	t.Run("every privilege keeps its own price", func(t *testing.T) {
		service, queries, _ := newPrivilegeCatalog()

		cheap, err := service.CreatePrivilege(ctx, privilegeCommand("homes.commands.*", 99))
		if err != nil {
			t.Fatal(err)
		}
		dear, err := service.CreatePrivilege(ctx, privilegeCommand("essentials.fly", 199))
		if err != nil {
			t.Fatal(err)
		}

		if queries.prices[domain.PrivilegePriceName(cheap.ID)][domain.CurrencyRUB] != 99 || queries.prices[domain.PrivilegePriceName(dear.ID)][domain.CurrencyRUB] != 199 {
			t.Errorf("prices = %v, want 99 and 199", queries.prices)
		}
	})

	t.Run("a name or permission already used is a conflict", func(t *testing.T) {
		service, queries, _ := newPrivilegeCatalog()
		queries.insertErr = &mysql.MySQLError{Number: 1062}

		_, err := service.CreatePrivilege(ctx, privilegeCommand("homes.commands.*", 99))
		if utils.MapCustomErrorToHttpStatus(err) != http.StatusConflict {
			t.Fatalf("got %v, want conflict", err)
		}
	})
}

func TestAdminCatalogService_UpdatePrivilege(t *testing.T) {
	ctx := context.Background()

	t.Run("changes the price in place", func(t *testing.T) {
		service, queries, _ := newPrivilegeCatalog()
		created, err := service.CreatePrivilege(ctx, privilegeCommand("homes.commands.*", 99))
		if err != nil {
			t.Fatal(err)
		}

		cmd := privilegeCommand("homes.commands.*", 149)
		cmd.ID = created.ID
		if _, err := service.UpdatePrivilege(ctx, cmd); err != nil {
			t.Fatal(err)
		}
		if got := queries.prices[domain.PrivilegePriceName(created.ID)][domain.CurrencyRUB]; got != 149 {
			t.Errorf("price = %v, want 149", got)
		}
	})

	t.Run("the permission is fixed while players own the privilege", func(t *testing.T) {
		service, queries, _ := newPrivilegeCatalog()
		created, err := service.CreatePrivilege(ctx, privilegeCommand("homes.commands.*", 99))
		if err != nil {
			t.Fatal(err)
		}
		queries.owners = 1

		cmd := privilegeCommand("homes.commands.set", 99)
		cmd.ID = created.ID
		if _, err := service.UpdatePrivilege(ctx, cmd); utils.MapCustomErrorToHttpStatus(err) != http.StatusConflict {
			t.Fatalf("got %v, want conflict", err)
		}
		if queries.privileges[created.ID].Permission != "homes.commands.*" {
			t.Errorf("permission = %q, want it unchanged", queries.privileges[created.ID].Permission)
		}

		// The name and the price still change.
		cmd = privilegeCommand("homes.commands.*", 120)
		cmd.ID = created.ID
		if _, err := service.UpdatePrivilege(ctx, cmd); err != nil {
			t.Fatalf("got %v, want the same permission accepted", err)
		}
	})

	t.Run("the permission changes when nobody owns the privilege", func(t *testing.T) {
		service, queries, _ := newPrivilegeCatalog()
		created, err := service.CreatePrivilege(ctx, privilegeCommand("homes.commands.*", 99))
		if err != nil {
			t.Fatal(err)
		}

		cmd := privilegeCommand("homes.commands.set", 99)
		cmd.ID = created.ID
		if _, err := service.UpdatePrivilege(ctx, cmd); err != nil {
			t.Fatal(err)
		}
		if queries.privileges[created.ID].Permission != "homes.commands.set" {
			t.Errorf("permission = %q, want the new one", queries.privileges[created.ID].Permission)
		}
	})

	t.Run("an unknown privilege is not found", func(t *testing.T) {
		service, _, _ := newPrivilegeCatalog()

		cmd := privilegeCommand("homes.commands.*", 99)
		cmd.ID = uuid.New()
		if _, err := service.UpdatePrivilege(ctx, cmd); utils.MapCustomErrorToHttpStatus(err) != http.StatusNotFound {
			t.Fatalf("got %v, want not found", err)
		}
	})
}

func TestAdminCatalogService_DeletePrivilege(t *testing.T) {
	ctx := context.Background()

	t.Run("removes the privilege with its tariff", func(t *testing.T) {
		service, queries, _ := newPrivilegeCatalog()
		created, err := service.CreatePrivilege(ctx, privilegeCommand("homes.commands.*", 99))
		if err != nil {
			t.Fatal(err)
		}

		if err := service.DeletePrivilege(ctx, created.ID); err != nil {
			t.Fatal(err)
		}
		if len(queries.privileges) != 0 || len(queries.prices) != 0 {
			t.Errorf("privileges %v, prices %v; want both gone", queries.privileges, queries.prices)
		}
	})

	t.Run("refuses while a product sells it or players own it", func(t *testing.T) {
		for name, arrange := range map[string]func(*privilegeCatalogQueries, uuid.UUID){
			"product": func(q *privilegeCatalogQueries, id uuid.UUID) {
				metadata, _ := json.Marshal(domain.PrivilegeProductMetadata{PrivilegeID: id})
				q.products = []*domain.Product{{ID: uuid.New(), Category: domain.ProductCategoryPrivilege, Metadata: metadata}}
			},
			"owners": func(q *privilegeCatalogQueries, _ uuid.UUID) { q.owners = 2 },
		} {
			t.Run(name, func(t *testing.T) {
				service, queries, _ := newPrivilegeCatalog()
				created, err := service.CreatePrivilege(ctx, privilegeCommand("homes.commands.*", 99))
				if err != nil {
					t.Fatal(err)
				}
				arrange(queries, created.ID)

				if err := service.DeletePrivilege(ctx, created.ID); utils.MapCustomErrorToHttpStatus(err) != http.StatusConflict {
					t.Fatalf("got %v, want conflict", err)
				}
				if len(queries.privileges) != 1 || len(queries.prices) != 1 {
					t.Errorf("privileges %d, tariffs %d; want both kept", len(queries.privileges), len(queries.prices))
				}
			})
		}
	})

	t.Run("an unknown privilege is not found", func(t *testing.T) {
		service, _, _ := newPrivilegeCatalog()

		if err := service.DeletePrivilege(ctx, uuid.New()); utils.MapCustomErrorToHttpStatus(err) != http.StatusNotFound {
			t.Fatalf("got %v, want not found", err)
		}
	})
}

func TestAdminCatalogService_PrivilegeProduct(t *testing.T) {
	ctx := context.Background()
	easyDonateID := int64(7)

	newCommand := func(privilegeID uuid.UUID, active bool) *commands.SaveProductCommand {
		return &commands.SaveProductCommand{
			Category:            domain.ProductCategoryPrivilege,
			PrivilegeID:         &privilegeID,
			PriceName:           domain.PrivilegePriceName(privilegeID),
			IsActive:            active,
			EasyDonateProductID: &easyDonateID,
		}
	}

	t.Run("a product links the privilege and sells at its tariff", func(t *testing.T) {
		service, _, _ := newPrivilegeCatalog()
		privilege, err := service.CreatePrivilege(ctx, privilegeCommand("homes.commands.*", 99))
		if err != nil {
			t.Fatal(err)
		}

		product, err := service.CreateProduct(ctx, newCommand(privilege.ID, true))
		if err != nil {
			t.Fatal(err)
		}

		var metadata domain.PrivilegeProductMetadata
		if err := json.Unmarshal(product.Metadata, &metadata); err != nil || metadata.PrivilegeID != privilege.ID {
			t.Errorf("metadata = %s, want the privilege %s", product.Metadata, privilege.ID)
		}
		if product.PriceName != domain.PrivilegePriceName(privilege.ID) {
			t.Errorf("price name = %q, want the tariff of the privilege", product.PriceName)
		}
	})

	t.Run("one privilege gets one product", func(t *testing.T) {
		service, _, _ := newPrivilegeCatalog()
		privilege, err := service.CreatePrivilege(ctx, privilegeCommand("homes.commands.*", 99))
		if err != nil {
			t.Fatal(err)
		}
		if _, err := service.CreateProduct(ctx, newCommand(privilege.ID, false)); err != nil {
			t.Fatal(err)
		}

		if _, err := service.CreateProduct(ctx, newCommand(privilege.ID, false)); utils.MapCustomErrorToHttpStatus(err) != http.StatusConflict {
			t.Fatalf("got %v, want conflict", err)
		}
	})

	t.Run("a product of an unknown privilege is a bad request", func(t *testing.T) {
		service, queries, _ := newPrivilegeCatalog()
		id := uuid.New()
		queries.prices[domain.PrivilegePriceName(id)] = map[domain.Currency]float64{domain.CurrencyRUB: 1}

		if _, err := service.CreateProduct(ctx, newCommand(id, false)); utils.MapCustomErrorToHttpStatus(err) != http.StatusBadRequest {
			t.Fatalf("got %v, want bad request", err)
		}
	})

	t.Run("an active product needs the price in every currency", func(t *testing.T) {
		service, queries, _ := newPrivilegeCatalog()
		privilege, err := service.CreatePrivilege(ctx, privilegeCommand("homes.commands.*", 99))
		if err != nil {
			t.Fatal(err)
		}
		delete(queries.prices[domain.PrivilegePriceName(privilege.ID)], domain.CurrencyPLN)

		if _, err := service.CreateProduct(ctx, newCommand(privilege.ID, true)); utils.MapCustomErrorToHttpStatus(err) != http.StatusConflict {
			t.Fatalf("got %v, want conflict", err)
		}
		if _, err := service.CreateProduct(ctx, newCommand(privilege.ID, false)); err != nil {
			t.Fatalf("a draft needs no complete tariff, got %v", err)
		}
	})

	t.Run("the product keeps its privilege", func(t *testing.T) {
		service, _, _ := newPrivilegeCatalog()
		first, err := service.CreatePrivilege(ctx, privilegeCommand("homes.commands.*", 99))
		if err != nil {
			t.Fatal(err)
		}
		second, err := service.CreatePrivilege(ctx, privilegeCommand("essentials.fly", 199))
		if err != nil {
			t.Fatal(err)
		}
		product, err := service.CreateProduct(ctx, newCommand(first.ID, false))
		if err != nil {
			t.Fatal(err)
		}

		cmd := newCommand(second.ID, false)
		cmd.ID = product.ID
		if _, err := service.UpdateProduct(ctx, cmd); utils.MapCustomErrorToHttpStatus(err) != http.StatusConflict {
			t.Fatalf("got %v, want conflict", err)
		}
	})
}
