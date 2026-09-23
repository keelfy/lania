package handlers

import (
	"net/http"

	"github.com/lania-smp/backend/internal/commands"
	"github.com/lania-smp/backend/internal/presenter"
	"github.com/lania-smp/backend/internal/services"
	"github.com/lania-smp/backend/internal/transport/http/binders"
	"github.com/lania-smp/backend/internal/utils"
)

// UploadHandler puts admin-uploaded images (glyth previews, season screenshots, season previews)
// into the project's S3 bucket, so nobody has to do it by hand.
type UploadHandler interface {
	UploadGlythPreview(http.ResponseWriter, *http.Request)
	UploadSeasonScreenshot(http.ResponseWriter, *http.Request)
	UploadSeasonPreview(http.ResponseWriter, *http.Request)
}

type uploadHandler struct{ service services.UploadService }

func NewUploadHandler(service services.UploadService) UploadHandler {
	return &uploadHandler{service: service}
}

func (h *uploadHandler) upload(w http.ResponseWriter, r *http.Request, kind commands.UploadImageKind) {
	ctx := r.Context()
	r.Body = http.MaxBytesReader(w, r.Body, commands.MaxUploadImageBytes[kind])

	cmd, err := binders.BindUploadImage(r, kind)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}
	if err := cmd.Validate(); err != nil {
		utils.HttpError(ctx, w, utils.NewBadRequestError("", err))
		return
	}
	image, err := h.service.UploadImage(ctx, cmd)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}
	utils.WriteHttpJsonResponse(ctx, w, presenter.PresentUploadedImage(image))
}

func (h *uploadHandler) UploadGlythPreview(w http.ResponseWriter, r *http.Request) {
	h.upload(w, r, commands.UploadImageKindGlythPreview)
}

func (h *uploadHandler) UploadSeasonScreenshot(w http.ResponseWriter, r *http.Request) {
	h.upload(w, r, commands.UploadImageKindSeasonScreenshot)
}

func (h *uploadHandler) UploadSeasonPreview(w http.ResponseWriter, r *http.Request) {
	h.upload(w, r, commands.UploadImageKindSeasonPreview)
}
