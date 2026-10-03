package sql

import (
	"context"
	"fmt"
	"strings"

	"github.com/lania-smp/backend/internal/domain"
)

const findPriceByNameAndCurrency = `
SELECT name, currency, amount
FROM product_prices
WHERE name = ? AND currency = ?
`

func (q *queries) FindPriceByNameAndCurrency(ctx context.Context, name domain.ProductPriceName, currency domain.Currency) (*domain.ProductPrice, error) {
	row := q.x.QueryRowContext(ctx, findPriceByNameAndCurrency, name, currency)
	var price domain.ProductPrice
	err := row.Scan(&price.Name, &price.Currency, &price.Amount)
	if err != nil {
		return nil, err
	}
	return &price, nil
}

const findPricesByNames = `
SELECT name, currency, amount
FROM product_prices
WHERE name IN ('%s')
`

func (q *queries) FindPricesByNames(ctx context.Context, names []domain.ProductPriceName) ([]*domain.ProductPrice, error) {
	namesStr := make([]string, len(names))
	for i, name := range names {
		namesStr[i] = string(name)
	}
	query := fmt.Sprintf(findPricesByNames, strings.Join(namesStr, "','"))
	rows, err := q.x.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	prices := make([]*domain.ProductPrice, 0)
	for rows.Next() {
		var price domain.ProductPrice
		err := rows.Scan(&price.Name, &price.Currency, &price.Amount)
		if err != nil {
			return nil, err
		}
		prices = append(prices, &price)
	}
	return prices, nil
}

const upsertProductPrice = `
INSERT INTO product_prices (name, currency, amount)
VALUES (?, ?, ?)
ON DUPLICATE KEY UPDATE amount = VALUES(amount)
`

// UpsertProductPrices writes the prices of the tariff, one row per currency. Other currencies of the tariff stay.
func (q *queries) UpsertProductPrices(ctx context.Context, name domain.ProductPriceName, prices []*domain.ProductPrice) error {
	for _, price := range prices {
		if _, err := q.x.ExecContext(ctx, upsertProductPrice, name, price.Currency, price.Amount); err != nil {
			return err
		}
	}
	return nil
}

// DeleteProductPrices removes the tariff with every currency.
func (q *queries) DeleteProductPrices(ctx context.Context, name domain.ProductPriceName) error {
	_, err := q.x.ExecContext(ctx, "DELETE FROM product_prices WHERE name = ?", name)
	return err
}
