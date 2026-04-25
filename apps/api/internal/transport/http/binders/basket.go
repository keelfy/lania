package binders

import (
	"encoding/json"
	"net/http"

	"github.com/lania-smp/backend/internal/commands"
	"github.com/lania-smp/backend/internal/transport/http/requests"
	"github.com/lania-smp/backend/internal/utils"
)

func BindAddBasketItem(r *http.Request) (*commands.AddBasketItemCommand, error) {
	authUserID, err := utils.GetUserIDFromCtx(r.Context())
	if err != nil {
		return nil, err
	}

	req := &requests.AddBasketItem{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		return nil, err
	}

	return &commands.AddBasketItemCommand{
		UserID:    authUserID,
		ProductID: req.ProductID,
		ProfileID: req.ProfileID,
		Quantity:  1,
	}, nil
}
