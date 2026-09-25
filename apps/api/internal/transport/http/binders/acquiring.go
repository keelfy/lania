package binders

import (
	"encoding/json"
	"net/http"

	"github.com/lania-smp/backend/internal/commands"
	"github.com/lania-smp/backend/internal/transport/http/requests"
	"github.com/lania-smp/backend/internal/utils"
)

func BindEasyDonateResult(r *http.Request) (*commands.EasyDonateResultCommand, error) {
	var request requests.EasyDonateCallbackRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return nil, utils.NewBadRequestError("body is not a valid json", err)
	}

	return &commands.EasyDonateResultCommand{
		PaymentID: request.PaymentID,
		Cost:      request.Cost,
		Customer:  request.Customer,
		Signature: request.Signature,
	}, nil
}
