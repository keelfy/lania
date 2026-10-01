package handlers

import (
	"net/http"
	"strconv"

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
	GetProductCatalog(w http.ResponseWriter, r *http.Request)
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

const (
	defaultCatalogPageSize = 24
	maxCatalogPageSize     = 60
)

// GetProductCatalog returns one page of the shop catalog. The cursor is the nextCursor of the previous page.
func (h *productHandler) GetProductCatalog(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	category := domain.ProductCategory(binders.BindOptionalQueryParamAsString(r, "category", ""))

	limit, err := strconv.Atoi(binders.BindOptionalQueryParamAsString(r, "limit", strconv.Itoa(defaultCatalogPageSize)))
	if err != nil || limit < 1 {
		limit = defaultCatalogPageSize
	}
	limit = min(limit, maxCatalogPageSize)

	var cursor *domain.ProductCursor
	if token := binders.BindOptionalQueryParamAsString(r, "cursor", ""); token != "" {
		if cursor, err = domain.DecodeProductCursor(token); err != nil {
			utils.HttpError(ctx, w, utils.NewBadRequestError("invalid cursor", err))
			return
		}
	}

	page, err := h.productService.GetProductsPage(ctx, category, cursor, limit)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	// The counts ignore the category filter, they label the category tabs.
	counts, err := h.productService.GetProductCounts(ctx)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	utils.WriteHttpJsonResponse(ctx, w, presenter.PresentProductCatalog(page, counts))
}
