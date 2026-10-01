package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/domain"
	"github.com/lania-smp/backend/internal/presenter"
	"github.com/lania-smp/backend/internal/services"
	"github.com/lania-smp/backend/internal/transport/http/binders"
	"github.com/lania-smp/backend/internal/transport/http/requests"
	"github.com/lania-smp/backend/internal/transport/http/responses"
	"github.com/lania-smp/backend/internal/utils"
)

// ProfileSkinHandler lets the owner change the in-game skin of a profile in the season the seasonId query parameter
// names, the primary one by default. A change answers 200 when the server applied it, and 202 with the status
// pending when the server did not apply it in time.
type ProfileSkinHandler interface {
	// UploadSkin takes a multipart form with a PNG file and the variant classic or slim.
	UploadSkin(w http.ResponseWriter, r *http.Request)
	// SetSkinNickname copies the skin of a licensed account by its nickname.
	SetSkinNickname(w http.ResponseWriter, r *http.Request)
	ClearSkin(w http.ResponseWriter, r *http.Request)
	// GetSkinFile serves a skin file uploaded on the site, for the server to download.
	GetSkinFile(w http.ResponseWriter, r *http.Request)
}

type profileSkinHandler struct {
	profileSkinService services.ProfileSkinService
	seasonService      services.SeasonService
}

func NewProfileSkinHandler(profileSkinService services.ProfileSkinService, seasonService services.SeasonService) ProfileSkinHandler {
	return &profileSkinHandler{profileSkinService: profileSkinService, seasonService: seasonService}
}

// bindSkinChange returns the user, the profile and the season of a skin change.
func (h *profileSkinHandler) bindSkinChange(r *http.Request) (userID, profileID, seasonID uuid.UUID, err error) {
	if userID, err = utils.GetUserIDFromCtx(r.Context()); err != nil {
		return
	}
	if profileID, err = binders.BindPathVariableAsUUID(r, binders.ProfileIDVariable); err != nil {
		return
	}
	seasonID, err = seasonIDFromRequest(r, h.seasonService)
	return
}

func (h *profileSkinHandler) UploadSkin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, profileID, seasonID, err := h.bindSkinChange(r)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	// The form carries the variant next to the file.
	r.Body = http.MaxBytesReader(w, r.Body, services.MaxSkinFileBytes+4*1024)
	if err := r.ParseMultipartForm(services.MaxSkinFileBytes); err != nil {
		utils.HttpError(ctx, w, utils.NewBadRequestError(services.ErrSkinFileInvalidMessage, err))
		return
	}
	defer r.MultipartForm.RemoveAll()
	file, _, err := r.FormFile("file")
	if err != nil {
		utils.HttpError(ctx, w, utils.NewBadRequestError("file is required", err))
		return
	}
	defer file.Close()
	content, err := io.ReadAll(file)
	if err != nil {
		utils.HttpError(ctx, w, utils.NewBadRequestError("failed to read file", err))
		return
	}

	variant := domain.SkinVariant(strings.ToLower(r.FormValue("variant")))
	skin, err := h.profileSkinService.SetOwnedSkinFile(ctx, userID, profileID, seasonID, content, variant)
	writeSkinChange(w, r, skin, err)
}

func (h *profileSkinHandler) SetSkinNickname(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, profileID, seasonID, err := h.bindSkinChange(r)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}
	req := &requests.SetSkinNickname{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		utils.HttpError(ctx, w, utils.NewBadRequestError("request body is invalid", err))
		return
	}

	skin, err := h.profileSkinService.SetOwnedSkinNickname(ctx, userID, profileID, seasonID, req.Nickname)
	writeSkinChange(w, r, skin, err)
}

func (h *profileSkinHandler) ClearSkin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, profileID, seasonID, err := h.bindSkinChange(r)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	err = h.profileSkinService.ClearOwnedSkin(ctx, userID, profileID, seasonID)
	writeSkinChange(w, r, nil, err)
}

func writeSkinChange(w http.ResponseWriter, r *http.Request, skin *domain.PlayerSkin, err error) {
	ctx := r.Context()
	if errors.Is(err, services.ErrSkinPending) {
		w.Header().Set(utils.HeaderContentType, utils.ApplicationJsonType)
		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(&responses.SkinChange{Status: responses.SkinChangeStatusPending})
		return
	} else if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}
	utils.WriteHttpJsonResponse(ctx, w, &responses.SkinChange{
		Status: responses.SkinChangeStatusApplied,
		Skin:   presenter.PresentPlayerSkin(skin),
	})
}

func (h *profileSkinHandler) GetSkinFile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	fileID, err := binders.BindPathVariable(r, "fileId")
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}
	content, err := h.profileSkinService.GetSkinFile(ctx, fileID)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	// A file is named by its content, so it never changes.
	w.Header().Set(utils.HeaderContentType, "image/png")
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	_, _ = w.Write(content)
}
