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
