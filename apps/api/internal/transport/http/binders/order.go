package binders

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/commands"
	"github.com/lania-smp/backend/internal/domain"
	"github.com/lania-smp/backend/internal/transport/http/requests"
	"github.com/lania-smp/backend/internal/utils"
)

// BindCreateOrder falls back to defaultSeasonID for products that name no season.
func BindCreateOrder(r *http.Request, defaultSeasonID uuid.UUID) (*commands.CreateOrderCommand, error) {
	userID, err := utils.GetUserIDFromCtx(r.Context())
	if err != nil {
		return nil, err
	}

	req := &requests.CreateOrder{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		return nil, err
	}

	products := make([]*commands.OrderItemCommand, len(req.Products))
	for i, product := range req.Products {
		seasonID := defaultSeasonID
		if product.SeasonID != nil {
			seasonID = *product.SeasonID
		}
		products[i] = &commands.OrderItemCommand{
			ProductID: product.ID,
			ProfileID: product.ProfileID,
			SeasonID:  seasonID,
			Quantity:  1,
		}
	}

	return &commands.CreateOrderCommand{
		PaymentMethod: domain.PaymentMethod(req.PaymentMethod),
		UserID:        userID,
		Items:         products,
	}, nil
}
