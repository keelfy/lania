package binders

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/commands"
	"github.com/lania-smp/backend/internal/logger"
	"github.com/lania-smp/backend/internal/transport/http/requests"
	"github.com/lania-smp/backend/internal/utils"
)

func BindFreekassaResult(r *http.Request) (*commands.FreekassaResultCommand, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, utils.NewBadRequestError("body is required", err)
	}

	values, err := url.ParseQuery(string(body))
	if err != nil {
		return nil, utils.NewBadRequestError("body is not a valid url encoded query", err)
	}

	logger.Infof(r.Context(), "freekassa result: %v", values)

	if values.Get("status_check") == "1" {
		return &commands.FreekassaResultCommand{
			StatusCheck: true,
		}, nil
	}

	orderID, err := uuid.Parse(values.Get("MERCHANT_ORDER_ID"))
	if err != nil {
		return nil, utils.NewBadRequestError("order id is required", err)
	}

	if !values.Has("SIGN") {
		return nil, utils.NewBadRequestError("signature value is required", nil)
	}

	if !values.Has("AMOUNT") {
		return nil, utils.NewBadRequestError("out sum is required", nil)
	}

	merchantID, err := strconv.ParseInt(values.Get("MERCHANT_ID"), 10, 64)
	if err != nil {
		return nil, utils.NewBadRequestError("merchant id is required", err)
	}

	amount, err := strconv.ParseInt(values.Get("AMOUNT"), 10, 64)
	if err != nil {
		return nil, utils.NewBadRequestError("amount is required", err)
	}

	return &commands.FreekassaResultCommand{
		StatusCheck: false,
		MerchantID:  merchantID,
		Amount:      amount,
		Currency:    values.Get("CUR_ID"),
		OrderID:     orderID,
		Signature:   values.Get("SIGN"),
	}, nil
}

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
