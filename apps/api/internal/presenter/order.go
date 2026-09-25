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

func presentOrderAmountsList(amounts []*domain.OrderAmounts) []*responses.OrderAmounts {
	res := make([]*responses.OrderAmounts, len(amounts))
	for i, amount := range amounts {
		res[i] = PresentOrderAmounts(amount)
	}
	return res
}

func PresentAdminOrder(order *domain.Order) *responses.AdminOrder {
	items := make([]*responses.AdminOrderItem, len(order.Items))
	for i, item := range order.Items {
		res := &responses.AdminOrderItem{
			ID:        item.ID,
			ProductID: item.ProductID,
			ProfileID: item.ProfileID,
			SeasonID:  item.SeasonID,
			Amounts:   presentOrderAmountsList(item.Amounts),
			Quantity:  item.Quantity,
		}
		if item.Product != nil && len(item.Product.Localizations) > 0 {
			res.ProductName = item.Product.Localizations[0].Name
		}
		if item.Profile != nil {
			res.Username = item.Profile.MinecraftUsername
		}
		if item.Season != nil {
			res.SeasonName = item.Season.Name
		}
		items[i] = res
	}
	return &responses.AdminOrder{
		ID:         order.ID,
		UserID:     order.UserID,
		Status:     string(order.Status),
		Amounts:    presentOrderAmountsList(order.Amounts),
		ExternalID: order.ExternalID,
		CreatedAt:  order.CreatedAt.UnixMilli(),
		UpdatedAt:  order.UpdatedAt.UnixMilli(),
		Items:      items,
	}
}

func PresentAdminOrders(pagination *domain.Pagination, count int64, orders []*domain.Order, statusCounts map[domain.OrderStatus]int64) *responses.AdminOrders {
	content := make([]*responses.AdminOrder, len(orders))
	for i, order := range orders {
		content[i] = PresentAdminOrder(order)
	}
	counts := make(map[string]int64, len(domain.OrderStatuses))
	for _, status := range domain.OrderStatuses {
		counts[string(status)] = statusCounts[status]
	}
	return &responses.AdminOrders{
		Paginated:    *PresentPaginatedResponse(pagination, count, content),
		StatusCounts: counts,
	}
}
