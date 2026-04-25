package handlers

import (
	"net/http"

	"github.com/lania-smp/backend/internal/domain"
	"github.com/lania-smp/backend/internal/presenter"
	"github.com/lania-smp/backend/internal/services"
	"github.com/lania-smp/backend/internal/transport/http/binders"
	"github.com/lania-smp/backend/internal/transport/http/responses"
	"github.com/lania-smp/backend/internal/utils"
)

type ProductHandler interface {
	GetProductByID(w http.ResponseWriter, r *http.Request)
	GetProducts(w http.ResponseWriter, r *http.Request)
}

type productHandler struct {
	productService services.ProductService
}

func NewProductHandler(
	productService services.ProductService,
) ProductHandler {
	return &productHandler{
		productService: productService,
	}
}

func (h *productHandler) GetProductByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id, err := binders.BindPathVariableAsUUID(r, "productId")
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	product, err := h.productService.GetProductByID(ctx, id)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	utils.WriteHttpJsonResponse(ctx, w, presenter.PresentProduct(product, product.Prices[0]))
}

func (h *productHandler) GetProducts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	category := binders.BindOptionalQueryParamAsString(r, "category", "")

	productIDs, err := binders.BindOptionalQueryParamAsUUIDs(r, "ids")
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	var products []*domain.Product

	if category != "" {
		products, err = h.productService.GetProductsByCategory(ctx, domain.ProductCategory(category))
	} else if len(productIDs) > 0 {
		products, err = h.productService.GetProductsByIDs(ctx, productIDs)
	} else {
		products, err = h.productService.GetProducts(ctx)
	}

	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	res := make([]*responses.Product, len(products))
	for i, product := range products {
		res[i] = presenter.PresentProduct(product, product.Prices[0])
	}

	utils.WriteHttpJsonResponse(ctx, w, res)
}
