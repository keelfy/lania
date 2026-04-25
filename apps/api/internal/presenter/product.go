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
