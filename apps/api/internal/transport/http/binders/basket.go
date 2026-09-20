package binders

import (
	"encoding/json"
	"net/http"

	"github.com/lania-smp/backend/internal/commands"
	"github.com/lania-smp/backend/internal/config"
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

	seasonID := config.GetPrimarySeasonID()
	if req.SeasonID != nil {
		seasonID = *req.SeasonID
	}

	return &commands.AddBasketItemCommand{
		UserID:    authUserID,
		ProductID: req.ProductID,
		ProfileID: req.ProfileID,
		SeasonID:  seasonID,
		Quantity:  1,
	}, nil
}
