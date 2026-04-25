package handlers

import (
	"net/http"

	"github.com/lania-smp/backend/internal/presenter"
	"github.com/lania-smp/backend/internal/services"
	"github.com/lania-smp/backend/internal/storage"
	"github.com/lania-smp/backend/internal/transport/http/binders"
	"github.com/lania-smp/backend/internal/transport/http/responses"
	"github.com/lania-smp/backend/internal/utils"
)

type BasketHandler interface {
	GetBasketItems(w http.ResponseWriter, r *http.Request)
	DeleteBasketItem(w http.ResponseWriter, r *http.Request)
	AddBasketItem(w http.ResponseWriter, r *http.Request)
}

type basketHandler struct {
	basketService services.BasketService
	storage       storage.MainStorage
}

func NewBasketHandler(
	basketService services.BasketService,
	storage storage.MainStorage,
) BasketHandler {
	return &basketHandler{
		basketService: basketService,
		storage:       storage,
	}
}

func (h *basketHandler) GetBasketItems(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	authUserID, err := utils.GetUserIDFromCtx(ctx)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	items, err := h.basketService.GetBasketItemsByUserID(ctx, authUserID)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	res := make([]*responses.BasketItem, len(items))
	for i, item := range items {
		res[i] = presenter.PresentBasketItem(item)
	}
	utils.WriteHttpJsonResponse(ctx, w, res)
}

func (h *basketHandler) DeleteBasketItem(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	authUserID, err := utils.GetUserIDFromCtx(ctx)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	itemIDs, err := binders.BindOptionalQueryParamAsUUIDs(r, "itemIds")
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	if len(itemIDs) == 0 {
		err = h.basketService.ClearBasketItemsByUserID(ctx, h.storage.Queries(), authUserID)
		if err != nil {
			utils.HttpError(ctx, w, err)
			return
		}
	} else {
		err = h.basketService.DeleteBasketItemByIDs(ctx, h.storage.Queries(), itemIDs)
		if err != nil {
			utils.HttpError(ctx, w, err)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}

func (h *basketHandler) AddBasketItem(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	cmd, err := binders.BindAddBasketItem(r)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	if err := cmd.Validate(); err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	err = h.basketService.AddBasketItem(ctx, h.storage.Queries(), cmd)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}
