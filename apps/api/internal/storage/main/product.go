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
WHERE category = ? AND products.is_active = true
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
WHERE products.is_active = true
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
WHERE (? = false OR products.is_active = true) AND id IN ('%s')
`

func (q *queries) FindProductByIDs(ctx context.Context, ids uuid.UUIDs, locale string, currency domain.Currency) ([]*domain.Product, error) {
	return q.findProductByIDs(ctx, ids, locale, currency, true)
}

func (q *queries) FindProductByIDsIncludingInactive(ctx context.Context, ids uuid.UUIDs, locale string, currency domain.Currency) ([]*domain.Product, error) {
	return q.findProductByIDs(ctx, ids, locale, currency, false)
}

func (q *queries) findProductByIDs(ctx context.Context, ids uuid.UUIDs, locale string, currency domain.Currency, activeOnly bool) ([]*domain.Product, error) {
	idsStr := make([]string, len(ids))
	for i, id := range ids {
		idsStr[i] = id.String()
	}
	query := fmt.Sprintf(findProductByIDs, strings.Join(idsStr, "','"))
	rows, err := q.x.QueryContext(ctx, query, locale, currency, activeOnly)
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

const findProductsPage = `
SELECT
	products.id,
	products.price_name,
	products.category,
	products.metadata,
	products.sold_count,
	products.created_at,
	products.updated_at,
	products.updated_by,
	product_localizations.name,
	product_localizations.description,
	prices.amount
FROM products
LEFT JOIN product_localizations ON
	products.id = product_localizations.product_id
	AND product_localizations.locale = ?
LEFT JOIN product_prices prices ON products.price_name = prices.name AND prices.currency = ?
WHERE products.is_active = true%s
ORDER BY products.category DESC, products.created_at DESC, products.id DESC
LIMIT ?
`

// FindProductsPage returns up to limit active products after the cursor, in catalog order.
// An empty category matches all of them, a nil cursor starts from the first product.
func (q *queries) FindProductsPage(ctx context.Context, category domain.ProductCategory, cursor *domain.ProductCursor, limit int, locale string, currency domain.Currency) ([]*domain.Product, error) {
	args := []any{locale, currency}
	var conditions strings.Builder
	if category != "" {
		conditions.WriteString(" AND products.category = ?")
		args = append(args, category)
	}
	if cursor != nil {
		conditions.WriteString(" AND (products.category, products.created_at, products.id) < (?, ?, ?)")
		args = append(args, cursor.Category, cursor.CreatedAt, cursor.ID)
	}
	args = append(args, limit)

	rows, err := q.x.QueryContext(ctx, fmt.Sprintf(findProductsPage, conditions.String()), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	products := make([]*domain.Product, 0, limit)
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

		localization.Locale = locale
		localization.ProductID = product.ID
		product.Localizations = append(product.Localizations, &localization)

		price.Currency = currency
		price.Name = product.PriceName
		product.Prices = append(product.Prices, &price)

		products = append(products, &product)
	}
	return products, rows.Err()
}

const countProductsByCategory = `
SELECT category, COUNT(*) FROM products WHERE is_active = true GROUP BY category
`

func (q *queries) CountProductsByCategory(ctx context.Context) (map[domain.ProductCategory]int64, error) {
	rows, err := q.x.QueryContext(ctx, countProductsByCategory)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	counts := make(map[domain.ProductCategory]int64)
	for rows.Next() {
		var category domain.ProductCategory
		var count int64
		if err := rows.Scan(&category, &count); err != nil {
			return nil, err
		}
		counts[category] = count
	}
	return counts, rows.Err()
}
