package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/presenter"
	"github.com/lania-smp/backend/internal/services"
	"github.com/lania-smp/backend/internal/transport/http/binders"
	"github.com/lania-smp/backend/internal/transport/http/requests"
	"github.com/lania-smp/backend/internal/utils"
)

type NotificationHandler interface {
	GetNotifications(w http.ResponseWriter, r *http.Request)
	MarkNotificationsRead(w http.ResponseWriter, r *http.Request)
}

type notificationHandler struct {
	notificationService services.NotificationService
}

func NewNotificationHandler(notificationService services.NotificationService) NotificationHandler {
	return &notificationHandler{notificationService: notificationService}
}

func (h *notificationHandler) GetNotifications(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	authUserID, err := utils.GetUserIDFromCtx(ctx)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	limit, err := strconv.Atoi(binders.BindOptionalQueryParamAsString(r, binders.LimitQueryParam, strconv.Itoa(binders.DefaultLimit)))
	if err != nil {
		limit = binders.DefaultLimit
	}

	h.writeNotifications(w, r, authUserID, limit)
}

func (h *notificationHandler) MarkNotificationsRead(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	authUserID, err := utils.GetUserIDFromCtx(ctx)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	// An empty body marks every unread notification, so the client can send no body at all.
	req := &requests.MarkNotificationsRead{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil && !errors.Is(err, io.EOF) {
		utils.HttpError(ctx, w, utils.NewBadRequestError("failed to read the request body", err))
		return
	}

	if err := h.notificationService.MarkNotificationsRead(ctx, authUserID, req.IDs); err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	h.writeNotifications(w, r, authUserID, binders.DefaultLimit)
}

// writeNotifications answers with the current list, so the client needs no second request.
func (h *notificationHandler) writeNotifications(w http.ResponseWriter, r *http.Request, userID uuid.UUID, limit int) {
	ctx := r.Context()

	notifications, err := h.notificationService.GetNotifications(ctx, userID, limit)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	unreadCount, err := h.notificationService.CountUnreadNotifications(ctx, userID)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	utils.WriteHttpJsonResponse(ctx, w, presenter.PresentNotificationList(notifications, unreadCount))
}
