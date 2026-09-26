package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/presenter"
	"github.com/lania-smp/backend/internal/services"
	"github.com/lania-smp/backend/internal/transport/http/binders"
	"github.com/lania-smp/backend/internal/transport/http/requests"
	"github.com/lania-smp/backend/internal/transport/http/responses"
	"github.com/lania-smp/backend/internal/utils"
)

type NotificationHandler interface {
	GetNotifications(w http.ResponseWriter, r *http.Request)
	MarkNotificationsRead(w http.ResponseWriter, r *http.Request)
	// SendAnnouncement is for admins: it sends news to every user.
	SendAnnouncement(w http.ResponseWriter, r *http.Request)
}

type notificationHandler struct {
	notificationService services.NotificationService
	announcementService services.AnnouncementService
}

func NewNotificationHandler(
	notificationService services.NotificationService,
	announcementService services.AnnouncementService,
) NotificationHandler {
	return &notificationHandler{
		notificationService: notificationService,
		announcementService: announcementService,
	}
}

func (h *notificationHandler) GetNotifications(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	authUserID, err := utils.GetUserIDFromCtx(ctx)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	h.writeNotifications(w, r, authUserID)
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

	h.writeNotifications(w, r, authUserID)
}

func (h *notificationHandler) SendAnnouncement(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	cmd, err := binders.BindSendAnnouncement(r)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}
	if err := cmd.Validate(); err != nil {
		utils.HttpError(ctx, w, utils.NewBadRequestError("", err))
		return
	}

	recipients, err := h.announcementService.SendAnnouncement(ctx, cmd.Payload())
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	utils.WriteHttpJsonResponse(ctx, w, &responses.AnnouncementSent{Recipients: recipients})
}

// writeNotifications answers with the list the query asks for, so marking as read needs no second request.
func (h *notificationHandler) writeNotifications(w http.ResponseWriter, r *http.Request, userID uuid.UUID) {
	ctx := r.Context()

	notifications, hasMore, err := h.notificationService.GetNotifications(ctx, userID, binders.BindNotificationFilter(r))
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	unreadCount, err := h.notificationService.CountUnreadNotifications(ctx, userID)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	utils.WriteHttpJsonResponse(ctx, w, presenter.PresentNotificationList(notifications, unreadCount, hasMore))
}
