package sql

import (
	"context"
	stdsql "database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/domain"
)

func (q *queries) FindAdminProducts(ctx context.Context) ([]*domain.Product, error) {
	rows, err := q.x.QueryContext(ctx, `
SELECT p.id, p.category, p.price_name, p.metadata, p.sold_count, p.created_at,
       p.updated_at, p.updated_by, p.is_active, ep.ed_product_id,
       l.locale, l.name, l.description, pp.currency, pp.amount
FROM products p
LEFT JOIN ed_products ep ON ep.product_id = p.id
LEFT JOIN product_localizations l ON l.product_id = p.id
LEFT JOIN product_prices pp ON pp.name = p.price_name
ORDER BY p.created_at DESC, l.locale, pp.currency`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	byID := make(map[uuid.UUID]*domain.Product)
	products := make([]*domain.Product, 0)
	locales := make(map[uuid.UUID]map[string]bool)
	prices := make(map[uuid.UUID]map[domain.Currency]bool)
	for rows.Next() {
		var row domain.Product
		var edID stdsql.NullInt64
		var locale, name, description, currency stdsql.NullString
		var amount stdsql.NullFloat64
		if err := rows.Scan(&row.ID, &row.Category, &row.PriceName, &row.Metadata, &row.SoldCount, &row.CreatedAt, &row.UpdatedAt, &row.UpdatedBy, &row.IsActive, &edID, &locale, &name, &description, &currency, &amount); err != nil {
			return nil, err
		}
		product := byID[row.ID]
		if product == nil {
			product = &row
			if edID.Valid {
				value := edID.Int64
				product.EasyDonateProductID = &value
			}
			byID[row.ID] = product
			products = append(products, product)
			locales[row.ID] = make(map[string]bool)
			prices[row.ID] = make(map[domain.Currency]bool)
		}
		if locale.Valid && !locales[row.ID][locale.String] {
			product.Localizations = append(product.Localizations, &domain.ProductLocalization{ProductID: row.ID, Locale: locale.String, Name: name.String, Description: description.String})
			locales[row.ID][locale.String] = true
		}
		priceCurrency := domain.Currency(currency.String)
		if currency.Valid && amount.Valid && !prices[row.ID][priceCurrency] {
			product.Prices = append(product.Prices, &domain.ProductPrice{Name: row.PriceName, Currency: priceCurrency, Amount: amount.Float64})
			prices[row.ID][priceCurrency] = true
		}
	}
	return products, rows.Err()
}

type SaveProductParams struct {
	ID        uuid.UUID
	Category  domain.ProductCategory
	PriceName domain.ProductPriceName
	Metadata  json.RawMessage
	IsActive  bool
	UpdatedBy *uuid.UUID
	UpdatedAt time.Time
}

func (q *queries) InsertProduct(ctx context.Context, arg SaveProductParams) error {
	_, err := q.x.ExecContext(ctx, `INSERT INTO products (id, category, price_name, metadata, is_active, updated_by, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)`, arg.ID, arg.Category, arg.PriceName, arg.Metadata, arg.IsActive, arg.UpdatedBy, arg.UpdatedAt)
	return err
}

func (q *queries) UpdateProduct(ctx context.Context, arg SaveProductParams) error {
	_, err := q.x.ExecContext(ctx, `UPDATE products SET price_name = ?, is_active = ?, updated_by = ?, updated_at = ? WHERE id = ?`, arg.PriceName, arg.IsActive, arg.UpdatedBy, arg.UpdatedAt, arg.ID)
	return err
}

func (q *queries) UpsertProductLocalization(ctx context.Context, productID uuid.UUID, locale, name, description string) error {
	_, err := q.x.ExecContext(ctx, `INSERT INTO product_localizations (product_id, locale, name, description) VALUES (?, ?, ?, ?) ON DUPLICATE KEY UPDATE name = VALUES(name), description = VALUES(description)`, productID, locale, name, description)
	return err
}

func (q *queries) UpsertEDProduct(ctx context.Context, productID uuid.UUID, edProductID int64) error {
	_, err := q.x.ExecContext(ctx, `INSERT INTO ed_products (product_id, ed_product_id) VALUES (?, ?) ON DUPLICATE KEY UPDATE ed_product_id = VALUES(ed_product_id)`, productID, edProductID)
	return err
}

func (q *queries) DeleteEDProduct(ctx context.Context, productID uuid.UUID) error {
	_, err := q.x.ExecContext(ctx, "DELETE FROM ed_products WHERE product_id = ?", productID)
	return err
}

// DeleteProduct removes the product with its localizations, EasyDonate link and basket items.
// Order items keep their foreign key, so a product that was ever ordered fails with 1451.
func (q *queries) DeleteProduct(ctx context.Context, productID uuid.UUID) error {
	for _, query := range []string{
		"DELETE FROM basket_items WHERE product_id = ?",
		"DELETE FROM ed_products WHERE product_id = ?",
		"DELETE FROM product_localizations WHERE product_id = ?",
		"DELETE FROM products WHERE id = ?",
	} {
		if _, err := q.x.ExecContext(ctx, query, productID); err != nil {
			return err
		}
	}
	return nil
}
