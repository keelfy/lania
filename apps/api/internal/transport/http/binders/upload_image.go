package binders

import (
	"io"
	"net/http"

	"github.com/lania-smp/backend/internal/commands"
	"github.com/lania-smp/backend/internal/utils"
)

// BindUploadImage reads a multipart image upload. The caller must wrap r.Body in
// http.MaxBytesReader with the kind's byte limit first, so an oversized request fails fast instead
// of spilling to a temp file. The "token" and "name" form fields are only used for a glyth
// preview, to derive the object key.
func BindUploadImage(r *http.Request, kind commands.UploadImageKind) (*commands.UploadImageCommand, error) {
	if err := r.ParseMultipartForm(commands.MaxUploadImageBytes[kind]); err != nil {
		return nil, utils.NewBadRequestError("upload is invalid or too large", err)
	}
	defer r.MultipartForm.RemoveAll()

	file, _, err := r.FormFile("file")
	if err != nil {
		return nil, utils.NewBadRequestError("file is required", err)
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return nil, utils.NewBadRequestError("failed to read file", err)
	}

	cmd := &commands.UploadImageCommand{
		Kind:    kind,
		Content: content,
		Token:   r.FormValue("token"),
		Name:    r.FormValue("name"),
	}
	if kind == commands.UploadImageKindSeasonScreenshot {
		cmd.SeasonID, err = BindPathVariableAsUUID(r, SeasonIDVariable)
		if err != nil {
			return nil, err
		}
	}
	return cmd, nil
}
