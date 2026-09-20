package handlers

import (
	"net/http"

	"github.com/lania-smp/backend/internal/presenter"
	"github.com/lania-smp/backend/internal/services"
	"github.com/lania-smp/backend/internal/transport/http/binders"
	"github.com/lania-smp/backend/internal/transport/http/responses"
	"github.com/lania-smp/backend/internal/utils"
)

type AdminUserHandler interface {
	GetUsers(w http.ResponseWriter, r *http.Request)
	GetUserDetails(w http.ResponseWriter, r *http.Request)
}

type adminUserHandler struct {
	adminUserService services.AdminUserService
}

func NewAdminUserHandler(
	adminUserService services.AdminUserService,
) AdminUserHandler {
	return &adminUserHandler{
		adminUserService: adminUserService,
	}
}

func (h *adminUserHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	pagination, err := binders.BindPagination(r)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	users, nextPageToken, err := h.adminUserService.ListUsers(ctx, binders.BindEmailSearch(r), binders.BindPageToken(r), pagination.Size)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	content := make([]*responses.AdminUser, len(users))
	for i, user := range users {
		content[i] = presenter.PresentAdminUser(user)
	}
	utils.WriteHttpJsonResponse(ctx, w, &responses.TokenPaginated[*responses.AdminUser]{
		Content:       content,
		NextPageToken: nextPageToken,
	})
}

func (h *adminUserHandler) GetUserDetails(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := binders.BindPathVariableAsUUID(r, binders.UserIDVariable)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	user, err := h.adminUserService.GetUserByID(ctx, userID)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	utils.WriteHttpJsonResponse(ctx, w, presenter.PresentAdminUserDetails(user))
}
