package binders

import (
	"encoding/json"
	"net/http"

	"github.com/lania-smp/backend/internal/transport/http/requests"
	"github.com/lania-smp/backend/internal/utils"
)

func BindConfirmProfileVerification(r *http.Request) (*requests.ConfirmProfileVerification, error) {
	req := &requests.ConfirmProfileVerification{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		return nil, utils.NewBadRequestError("request body is invalid", err)
	}
	return req, nil
}

func BindVerificationLogin(r *http.Request) (*requests.VerificationLogin, error) {
	req := &requests.VerificationLogin{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		return nil, utils.NewBadRequestError("request body is invalid", err)
	}
	if req.Username == "" {
		return nil, utils.NewBadRequestError("username is required", nil)
	}
	return req, nil
}
