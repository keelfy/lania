package handlers

import (
	"net/http"

	"github.com/lania-smp/backend/internal/config"
	"github.com/lania-smp/backend/internal/presenter"
	"github.com/lania-smp/backend/internal/services"
	"github.com/lania-smp/backend/internal/transport/http/binders"
	"github.com/lania-smp/backend/internal/transport/http/responses"
	"github.com/lania-smp/backend/internal/utils"
)

type PurchaseHandler interface {
	GetPurchasedProducts(w http.ResponseWriter, r *http.Request)
}

type purchaseHandler struct {
	profileService          services.ProfileService
	accessService           services.AccessService
	productService          services.ProductService
	profileCosmeticsService services.ProfileCosmeticsService
	purchaseService         services.PurchaseService
}

func NewPurchaseHandler(
	profileService services.ProfileService,
	accessService services.AccessService,
	productService services.ProductService,
	profileCosmeticsService services.ProfileCosmeticsService,
	purchaseService services.PurchaseService,
) PurchaseHandler {
	return &purchaseHandler{
		profileService:          profileService,
		accessService:           accessService,
		productService:          productService,
		profileCosmeticsService: profileCosmeticsService,
		purchaseService:         purchaseService,
	}
}

func (h *purchaseHandler) GetPurchasedProducts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	productIDs, err := binders.BindOptionalQueryParamAsUUIDs(r, "productIds")
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	authUserID, err := utils.GetUserIDFromCtx(ctx)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	seasonID := config.GetActiveSeasonID()

	products, err := h.productService.GetProductsByIDs(ctx, productIDs)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	resList, err := h.purchaseService.GetPurchasesByProducts(ctx, authUserID, seasonID, products)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	res := make([]*responses.PurchasedProduct, len(resList))
	for i, purchasedProduct := range resList {
		res[i] = presenter.PresentPurchasedProduct(purchasedProduct)
	}

	utils.WriteHttpJsonResponse(ctx, w, res)
}
