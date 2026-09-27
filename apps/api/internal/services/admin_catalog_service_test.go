package services

import (
	"context"
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

type fakeCatalogQueries struct {
	sql.Queries
	products  []*domain.Product
	deleted   []uuid.UUID
	deleteErr error
	owners    int
	prefixes  []*domain.NamePrefix
}

func (q *fakeCatalogQueries) FindNameColorByID(_ context.Context, id uuid.UUID) (*domain.NameColor, error) {
	return &domain.NameColor{ID: id}, nil
}

func (q *fakeCatalogQueries) CountNameColorOwners(context.Context, uuid.UUID) (int, error) {
	return q.owners, nil
}

func (q *fakeCatalogQueries) DeleteNameColor(_ context.Context, id, _ uuid.UUID) error {
	if q.deleteErr != nil {
		return q.deleteErr
	}
	q.deleted = append(q.deleted, id)
	return nil
}

func (q *fakeCatalogQueries) FindNameColors(context.Context) ([]*domain.NameColor, error) {
	return nil, nil
}

func (q *fakeCatalogQueries) FindNamePrefixes(context.Context) ([]*domain.NamePrefix, error) {
	return q.prefixes, nil
}

func (q *fakeCatalogQueries) FindAdminProducts(context.Context) ([]*domain.Product, error) {
	return q.products, nil
}

func (q *fakeCatalogQueries) DeleteProduct(_ context.Context, id uuid.UUID) error {
	if q.deleteErr != nil {
		return q.deleteErr
	}
	q.deleted = append(q.deleted, id)
	return nil
}

type fakeCatalogStorage struct {
	storage.MainStorage
	queries *fakeCatalogQueries
}

func (s *fakeCatalogStorage) Queries() sql.Queries { return s.queries }

func (s *fakeCatalogStorage) BeginTx(_ context.Context, fn func(sql.Queries) error) error {
	return fn(s.queries)
}

func nameColorProduct(t *testing.T, colorID uuid.UUID, active bool) *domain.Product {
	t.Helper()
	metadata, err := json.Marshal(domain.NameColorProductMetadata{NameColorID: colorID})
	if err != nil {
		t.Fatal(err)
	}
	return &domain.Product{ID: uuid.New(), Category: domain.ProductCategoryNameColor, Metadata: metadata, IsActive: active}
}

func TestAdminCatalogService_CreateProduct_RejectsDuplicateCosmetic(t *testing.T) {
	colorID := uuid.New()
	queries := &fakeCatalogQueries{products: []*domain.Product{nameColorProduct(t, colorID, false)}}
	service := NewAdminCatalogService(&fakeCatalogStorage{queries: queries})

	_, err := service.CreateProduct(t.Context(), &commands.SaveProductCommand{Category: domain.ProductCategoryNameColor, CosmeticID: &colorID})
	if utils.MapCustomErrorToHttpStatus(err) != http.StatusConflict {
		t.Fatalf("CreateProduct() error = %v, want conflict", err)
	}
}

func TestAdminCatalogService_DeleteProduct(t *testing.T) {
	t.Run("draft is deleted", func(t *testing.T) {
		draft := nameColorProduct(t, uuid.New(), false)
		queries := &fakeCatalogQueries{products: []*domain.Product{draft}}
		service := NewAdminCatalogService(&fakeCatalogStorage{queries: queries})

		if err := service.DeleteProduct(t.Context(), draft.ID); err != nil {
			t.Fatalf("DeleteProduct() error = %v", err)
		}
		if len(queries.deleted) != 1 || queries.deleted[0] != draft.ID {
			t.Fatalf("deleted = %v, want [%s]", queries.deleted, draft.ID)
		}
	})

	tests := []struct {
		name      string
		active    bool
		deleteErr error
		missing   bool
		status    int
	}{
		{name: "published product", active: true, status: http.StatusConflict},
		{name: "ordered draft", deleteErr: &mysql.MySQLError{Number: 1451}, status: http.StatusConflict},
		{name: "unknown product", missing: true, status: http.StatusNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			product := nameColorProduct(t, uuid.New(), tt.active)
			queries := &fakeCatalogQueries{products: []*domain.Product{product}, deleteErr: tt.deleteErr}
			service := NewAdminCatalogService(&fakeCatalogStorage{queries: queries})

			id := product.ID
			if tt.missing {
				id = uuid.New()
			}
			err := service.DeleteProduct(t.Context(), id)
			if utils.MapCustomErrorToHttpStatus(err) != tt.status {
				t.Fatalf("DeleteProduct() error = %v, want status %d", err, tt.status)
			}
			if len(queries.deleted) != 0 {
				t.Fatalf("deleted = %v, want none", queries.deleted)
			}
		})
	}
}

func TestAdminCatalogService_DeleteNameColor(t *testing.T) {
	defaultColorID := uuid.New()
	t.Setenv("DEFAULT_NAME_COLOR_ID", defaultColorID.String())

	t.Run("unused color is deleted", func(t *testing.T) {
		colorID := uuid.New()
		queries := &fakeCatalogQueries{}
		service := NewAdminCatalogService(&fakeCatalogStorage{queries: queries})

		if _, err := service.DeleteNameColor(t.Context(), colorID); err != nil {
			t.Fatalf("DeleteNameColor() error = %v", err)
		}
		if len(queries.deleted) != 1 || queries.deleted[0] != colorID {
			t.Fatalf("deleted = %v, want [%s]", queries.deleted, colorID)
		}
	})

	colorID := uuid.New()
	tests := []struct {
		name      string
		id        uuid.UUID
		owners    int
		products  []*domain.Product
		deleteErr error
	}{
		{name: "owned color", id: colorID, owners: 1},
		{name: "color sold by a product", id: colorID, products: []*domain.Product{nameColorProduct(t, colorID, false)}},
		{name: "granted meanwhile", id: colorID, deleteErr: &mysql.MySQLError{Number: 1451}},
		{name: "default color", id: defaultColorID},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			queries := &fakeCatalogQueries{owners: tt.owners, products: tt.products, deleteErr: tt.deleteErr}
			service := NewAdminCatalogService(&fakeCatalogStorage{queries: queries})

			_, err := service.DeleteNameColor(t.Context(), tt.id)
			if utils.MapCustomErrorToHttpStatus(err) != http.StatusConflict {
				t.Fatalf("DeleteNameColor() error = %v, want conflict", err)
			}
			if len(queries.deleted) != 0 {
				t.Fatalf("deleted = %v, want none", queries.deleted)
			}
		})
	}
}
