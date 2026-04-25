package presenter

import (
	"github.com/lania-smp/backend/internal/domain"
	"github.com/lania-smp/backend/internal/transport/http/responses"
)

func PresentBasketItem(item *domain.BasketItem) *responses.BasketItem {
	return &responses.BasketItem{
		ID:        item.ID,
		ProductID: item.ProductID,
		ProfileID: item.ProfileID,
		Quantity:  item.Quantity,
	}
}
