package presenter

import (
	"github.com/lania-smp/backend/internal/domain"
	"github.com/lania-smp/backend/internal/transport/http/responses"
)

func PresentOrderAmounts(amounts *domain.OrderAmounts) *responses.OrderAmounts {
	return &responses.OrderAmounts{
		Currency: string(amounts.Currency),
		Amount:   amounts.Amount,
	}
}

func PresentOrder(order *domain.Order, items []*responses.OrderItem) *responses.Order {
	amounts := make([]*responses.OrderAmounts, len(order.Amounts))
	for i, amount := range order.Amounts {
		amounts[i] = PresentOrderAmounts(amount)
	}
	return &responses.Order{
		ID:        order.ID,
		Status:    string(order.Status),
		Amounts:   amounts,
		Items:     items,
		CreatedAt: order.CreatedAt,
	}
}

func PresentOrderItem(item *domain.OrderItem) *responses.OrderItem {
	amounts := make([]*responses.OrderAmounts, len(item.Amounts))
	for i, amount := range item.Amounts {
		amounts[i] = PresentOrderAmounts(amount)
	}
	return &responses.OrderItem{
		ID:        item.ID,
		ProductID: item.ProductID,
		ProfileID: item.ProfileID,
		SeasonID:  item.SeasonID,
		Amounts:   amounts,
		Quantity:  item.Quantity,
	}
}
