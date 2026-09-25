package handlers

import (
	"net/http"
	"slices"

	"github.com/lania-smp/backend/internal/domain"
	"github.com/lania-smp/backend/internal/presenter"
	"github.com/lania-smp/backend/internal/services"
	"github.com/lania-smp/backend/internal/transport/http/binders"
	"github.com/lania-smp/backend/internal/utils"
)

type AdminOrderHandler interface {
	GetOrders(w http.ResponseWriter, r *http.Request)
}

type adminOrderHandler struct {
	orderService services.OrderService
}

func NewAdminOrderHandler(orderService services.OrderService) AdminOrderHandler {
	return &adminOrderHandler{orderService: orderService}
}

// GetOrders lists the orders of every user, newest first, with the number of orders in each status.
func (h *adminOrderHandler) GetOrders(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	pagination, err := binders.BindPagination(r)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	filter := domain.OrderFilter{
		Status: domain.OrderStatus(r.URL.Query().Get("status")),
		Search: binders.BindSearch(r),
	}
	if filter.Status != "" && !slices.Contains(domain.OrderStatuses, filter.Status) {
		utils.HttpError(ctx, w, utils.NewBadRequestError("unknown order status", nil))
		return
	}

	orders, count, err := h.orderService.GetAdminOrders(ctx, filter, pagination)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	statusCounts, err := h.orderService.CountOrdersByStatus(ctx)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	utils.WriteHttpJsonResponse(ctx, w, presenter.PresentAdminOrders(pagination, count, orders, statusCounts))
}
