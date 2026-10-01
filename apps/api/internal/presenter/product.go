package presenter

import (
	"github.com/lania-smp/backend/internal/domain"
	"github.com/lania-smp/backend/internal/transport/http/responses"
)

func PresentProduct(product *domain.Product, price *domain.ProductPrice) *responses.Product {
	name := ""
	description := ""

	if len(product.Localizations) > 0 {
		name = product.Localizations[0].Name
		description = product.Localizations[0].Description
	}

	return &responses.Product{
		ID:          product.ID,
		Price:       price.Amount,
		Category:    string(product.Category),
		Name:        name,
		Description: description,
		Metadata:    product.Metadata,
		SoldCount:   product.SoldCount,
		CreatedAt:   product.CreatedAt,
	}
}

func PresentProductCatalog(page *domain.ProductPage, counts map[domain.ProductCategory]int64) *responses.ProductCatalog {
	res := &responses.ProductCatalog{
		Content: make([]*responses.Product, len(page.Products)),
		Counts:  make(map[string]int64, len(counts)),
	}
	for i, product := range page.Products {
		res.Content[i] = PresentProduct(product, product.Prices[0])
	}
	for category, count := range counts {
		res.Counts[string(category)] = count
	}
	if page.NextCursor != nil {
		res.NextCursor = page.NextCursor.Encode()
	}
	return res
}
