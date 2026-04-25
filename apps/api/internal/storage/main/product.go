package sql

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/domain"
)

const findProductsByCategory = `
SELECT 
	id,
	category,
	price_name,
	metadata,
	sold_count,
	created_at,
	updated_at,
	updated_by,
	product_localizations.name,
	product_localizations.description,
	prices.amount
FROM products
LEFT JOIN product_localizations ON products.id = product_localizations.product_id  AND product_localizations.locale = ?
LEFT JOIN product_prices prices ON products.price_name = prices.name AND prices.currency = ?
WHERE category = ?
ORDER BY category DESC, created_at DESC
`

func (q *queries) FindProductsByCategory(ctx context.Context, category domain.ProductCategory, currency domain.Currency, locale string) ([]*domain.Product, error) {
	rows, err := q.x.QueryContext(ctx, findProductsByCategory, locale, currency, category)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	products := make([]*domain.Product, 0)
	for rows.Next() {
		var product domain.Product
		var localization domain.ProductLocalization
		var price domain.ProductPrice
		err := rows.Scan(
			&product.ID,
			&product.Category,
			&product.PriceName,
			&product.Metadata,
			&product.SoldCount,
			&product.CreatedAt,
			&product.UpdatedAt,
			&product.UpdatedBy,
			&localization.Name,
			&localization.Description,
			&price.Amount,
		)
		if err != nil {
			return nil, err
		}

		// enrich localization
		localization.Locale = locale
		localization.ProductID = product.ID
		product.Localizations = append(product.Localizations, &localization)

		// enrich price
		price.Currency = currency
		price.Name = product.PriceName
		product.Prices = append(product.Prices, &price)

		products = append(products, &product)
	}
	return products, nil
}

const findProducts = `
SELECT 
	id,
	price_name,
	category,
	metadata,
	sold_count,
	created_at,
	updated_at,
	updated_by,
	product_localizations.name,
	product_localizations.description,
	prices.amount
FROM products
LEFT JOIN product_localizations ON 
	products.id = product_localizations.product_id 
	AND product_localizations.locale = ?
LEFT JOIN product_prices prices ON products.price_name = prices.name AND prices.currency = ?
ORDER BY category DESC, created_at DESC
`

func (q *queries) FindProducts(ctx context.Context, locale string, currency domain.Currency) ([]*domain.Product, error) {
	rows, err := q.x.QueryContext(ctx, findProducts, locale, currency)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	products := make([]*domain.Product, 0)
	for rows.Next() {
		var product domain.Product
		var localization domain.ProductLocalization
		var price domain.ProductPrice
		err := rows.Scan(
			&product.ID,
			&product.PriceName,
			&product.Category,
			&product.Metadata,
			&product.SoldCount,
			&product.CreatedAt,
			&product.UpdatedAt,
			&product.UpdatedBy,
			&localization.Name,
			&localization.Description,
			&price.Amount,
		)
		if err != nil {
			return nil, err
		}

		// enrich localization
		localization.Locale = locale
		localization.ProductID = product.ID
		product.Localizations = append(product.Localizations, &localization)

		// enrich price
		price.Currency = currency
		price.Name = product.PriceName
		product.Prices = append(product.Prices, &price)

		products = append(products, &product)
	}
	return products, nil
}

const findProductByIDs = `
SELECT 
	id,
	price_name,
	category,
	metadata,
	sold_count,
	created_at,
	updated_at,
	updated_by,
	product_localizations.name,
	product_localizations.description,
	prices.amount
FROM products
LEFT JOIN product_localizations ON 
	products.id = product_localizations.product_id 
	AND product_localizations.locale = ?
LEFT JOIN product_prices prices ON products.price_name = prices.name AND prices.currency = ?
WHERE id IN ('%s')
`

func (q *queries) FindProductByIDs(ctx context.Context, ids uuid.UUIDs, locale string, currency domain.Currency) ([]*domain.Product, error) {
	idsStr := make([]string, len(ids))
	for i, id := range ids {
		idsStr[i] = id.String()
	}
	query := fmt.Sprintf(findProductByIDs, strings.Join(idsStr, "','"))
	rows, err := q.x.QueryContext(ctx, query, locale, currency)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	products := make([]*domain.Product, 0)
	for rows.Next() {
		var product domain.Product
		var localization domain.ProductLocalization
		var price domain.ProductPrice
		err := rows.Scan(
			&product.ID,
			&product.PriceName,
			&product.Category,
			&product.Metadata,
			&product.SoldCount,
			&product.CreatedAt,
			&product.UpdatedAt,
			&product.UpdatedBy,
			&localization.Name,
			&localization.Description,
			&price.Amount,
		)
		if err != nil {
			return nil, err
		}

		// enrich localization
		localization.Locale = locale
		localization.ProductID = product.ID
		product.Localizations = append(product.Localizations, &localization)

		// enrich price
		price.Currency = currency
		price.Name = product.PriceName
		product.Prices = append(product.Prices, &price)

		products = append(products, &product)
	}
	return products, nil
}
