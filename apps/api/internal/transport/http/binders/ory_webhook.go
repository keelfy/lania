package binders

import (
	"encoding/json"
	"net/http"

	"github.com/lania-smp/backend/internal/commands"
	"github.com/lania-smp/backend/internal/transport/http/requests"
)

func BindAfterOryRegistrationWebhook(r *http.Request) (*commands.CreateUserCommand, error) {
	req := &requests.AfterOryRegistrationWebhook{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		return nil, err
	}

	return &commands.CreateUserCommand{
		IdentityID: req.IdentityID,
		Email:      req.Traits.Email,
		Username:   req.Traits.Username,
		AvatarURL:  req.Traits.AvatarURL,
	}, nil
}
