package presenter

import (
	"github.com/lania-smp/backend/internal/domain"
	"github.com/lania-smp/backend/internal/transport/http/responses"
)

func PresentPurchasedProduct(purchasedProduct *domain.PurchasedProduct) *responses.PurchasedProduct {
	return &responses.PurchasedProduct{
		ProductID: purchasedProduct.ProductID,
		ProfileID: purchasedProduct.ProfileID,
		SeasonID:  purchasedProduct.SeasonID,
	}
}
