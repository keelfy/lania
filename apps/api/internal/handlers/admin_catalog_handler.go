package handlers

import (
	"net/http"

	"github.com/lania-smp/backend/internal/commands"
	"github.com/lania-smp/backend/internal/domain"
	"github.com/lania-smp/backend/internal/presenter"
	"github.com/lania-smp/backend/internal/services"
	"github.com/lania-smp/backend/internal/transport/http/binders"
	"github.com/lania-smp/backend/internal/utils"
)

type AdminCatalogHandler interface {
	GetCosmetics(http.ResponseWriter, *http.Request)
	CreateNameColor(http.ResponseWriter, *http.Request)
	UpdateNameColor(http.ResponseWriter, *http.Request)
	CreateNamePrefix(http.ResponseWriter, *http.Request)
	UpdateNamePrefix(http.ResponseWriter, *http.Request)
	GetProducts(http.ResponseWriter, *http.Request)
	CreateProduct(http.ResponseWriter, *http.Request)
	UpdateProduct(http.ResponseWriter, *http.Request)
	DeleteProduct(http.ResponseWriter, *http.Request)
	CreateEasyDonateProduct(http.ResponseWriter, *http.Request)
}

type adminCatalogHandler struct {
	service           services.AdminCatalogService
	easyDonateService services.EasyDonateService
}

func NewAdminCatalogHandler(service services.AdminCatalogService, easyDonateService services.EasyDonateService) AdminCatalogHandler {
	return &adminCatalogHandler{service: service, easyDonateService: easyDonateService}
}

func (h *adminCatalogHandler) GetCosmetics(w http.ResponseWriter, r *http.Request) {
	catalog, err := h.service.GetCosmetics(r.Context())
	if err != nil {
		utils.HttpError(r.Context(), w, err)
		return
	}
	utils.WriteHttpJsonResponse(r.Context(), w, presenter.PresentAdminCosmeticsCatalog(catalog))
}

func (h *adminCatalogHandler) saveNameColor(w http.ResponseWriter, r *http.Request, update bool) {
	var cmd *commands.SaveNameColorCommand
	var err error
	if update {
		cmd, err = binders.BindUpdateNameColor(r)
	} else {
		cmd, err = binders.BindSaveNameColor(r)
	}
	if err != nil {
		utils.HttpError(r.Context(), w, err)
		return
	}
	if err := cmd.Validate(); err != nil {
		utils.HttpError(r.Context(), w, utils.NewBadRequestError("", err))
		return
	}
	var catalog *domain.CosmeticsCatalog
	if update {
		catalog, err = h.service.UpdateNameColor(r.Context(), cmd)
	} else {
		catalog, err = h.service.CreateNameColor(r.Context(), cmd)
	}
	if err != nil {
		utils.HttpError(r.Context(), w, err)
		return
	}
	utils.WriteHttpJsonResponse(r.Context(), w, presenter.PresentAdminCosmeticsCatalog(catalog))
}

func (h *adminCatalogHandler) CreateNameColor(w http.ResponseWriter, r *http.Request) {
	h.saveNameColor(w, r, false)
}
func (h *adminCatalogHandler) UpdateNameColor(w http.ResponseWriter, r *http.Request) {
	h.saveNameColor(w, r, true)
}

func (h *adminCatalogHandler) saveNamePrefix(w http.ResponseWriter, r *http.Request, update bool) {
	var cmd *commands.SaveNamePrefixCommand
	var err error
	if update {
		cmd, err = binders.BindUpdateNamePrefix(r)
	} else {
		cmd, err = binders.BindSaveNamePrefix(r)
	}
	if err != nil {
		utils.HttpError(r.Context(), w, err)
		return
	}
	if err := cmd.Validate(); err != nil {
		utils.HttpError(r.Context(), w, utils.NewBadRequestError("", err))
		return
	}
	var catalog *domain.CosmeticsCatalog
	if update {
		catalog, err = h.service.UpdateNamePrefix(r.Context(), cmd)
	} else {
		catalog, err = h.service.CreateNamePrefix(r.Context(), cmd)
	}
	if err != nil {
		utils.HttpError(r.Context(), w, err)
		return
	}
	utils.WriteHttpJsonResponse(r.Context(), w, presenter.PresentAdminCosmeticsCatalog(catalog))
}

func (h *adminCatalogHandler) CreateNamePrefix(w http.ResponseWriter, r *http.Request) {
	h.saveNamePrefix(w, r, false)
}
func (h *adminCatalogHandler) UpdateNamePrefix(w http.ResponseWriter, r *http.Request) {
	h.saveNamePrefix(w, r, true)
}

func (h *adminCatalogHandler) GetProducts(w http.ResponseWriter, r *http.Request) {
	products, err := h.service.GetProducts(r.Context())
	if err != nil {
		utils.HttpError(r.Context(), w, err)
		return
	}
	utils.WriteHttpJsonResponse(r.Context(), w, presenter.PresentAdminProducts(products))
}

func (h *adminCatalogHandler) saveProduct(w http.ResponseWriter, r *http.Request, update bool) {
	var cmd *commands.SaveProductCommand
	var err error
	if update {
		cmd, err = binders.BindUpdateProduct(r)
	} else {
		cmd, err = binders.BindSaveProduct(r)
	}
	if err != nil {
		utils.HttpError(r.Context(), w, err)
		return
	}
	if err := cmd.Validate(); err != nil {
		utils.HttpError(r.Context(), w, utils.NewBadRequestError("", err))
		return
	}
	var product *domain.Product
	if update {
		product, err = h.service.UpdateProduct(r.Context(), cmd)
	} else {
		product, err = h.service.CreateProduct(r.Context(), cmd)
	}
	if err != nil {
		utils.HttpError(r.Context(), w, err)
		return
	}
	utils.WriteHttpJsonResponse(r.Context(), w, presenter.PresentAdminProducts([]*domain.Product{product})[0])
}

func (h *adminCatalogHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	h.saveProduct(w, r, false)
}
func (h *adminCatalogHandler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	h.saveProduct(w, r, true)
}

func (h *adminCatalogHandler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	productID, err := binders.BindPathVariableAsUUID(r, binders.ProductIDVariable)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}
	if err := h.service.DeleteProduct(ctx, productID); err != nil {
		utils.HttpError(ctx, w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *adminCatalogHandler) CreateEasyDonateProduct(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	r.Body = http.MaxBytesReader(w, r.Body, commands.MaxEDProductImageBytes)

	cmd, err := binders.BindCreateEDProduct(r)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}
	if err := cmd.Validate(); err != nil {
		utils.HttpError(ctx, w, utils.NewBadRequestError("", err))
		return
	}
	easyDonateProductID, err := h.easyDonateService.CreateShopProduct(ctx, cmd)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}
	utils.WriteHttpJsonResponse(ctx, w, presenter.PresentEasyDonateProduct(easyDonateProductID))
}
