package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/lania-smp/backend/internal/config"
	"github.com/lania-smp/backend/internal/services"
	"github.com/lania-smp/backend/internal/storage"
	"github.com/lania-smp/backend/internal/transport/http/binders"
	"github.com/lania-smp/backend/internal/utils"
	"github.com/lania-smp/backend/internal/utils/hashHelper"
)

type AcquiringHandler interface {
	HandleFreekassaResult(w http.ResponseWriter, r *http.Request)
	HandleEasyDonateResult(w http.ResponseWriter, r *http.Request)
}

type acquiringHandler struct {
	storage      storage.MainStorage
	orderService services.OrderService
}

func NewAcquiringHandler(
	storage storage.MainStorage,
	orderService services.OrderService,
) AcquiringHandler {
	return &acquiringHandler{
		storage:      storage,
		orderService: orderService,
	}
}

func (h *acquiringHandler) HandleFreekassaResult(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	cmd, err := binders.BindFreekassaResult(r)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	if cmd.StatusCheck {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("YES"))
		return
	}

	if err := cmd.Validate(); err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	merchantID := config.GetFreekassaMerchantID()
	password2 := config.GetFreekassaMerchantPassword2()
	signatureSource := fmt.Sprintf("%d:%d:%s:%s", merchantID, cmd.Amount, password2, cmd.OrderID.String())
	signature := hashHelper.Hash(signatureSource)

	if !strings.EqualFold(signature, cmd.Signature) {
		utils.HttpError(ctx, w, utils.NewBadRequestError("invalid signature", nil))
		return
	}

	order, err := h.orderService.GetOrderByID(ctx, cmd.OrderID)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	err = h.orderService.CompleteOrder(ctx, order)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("YES"))
}

func (h *acquiringHandler) HandleEasyDonateResult(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	cmd, err := binders.BindEasyDonateResult(r)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	if err := cmd.Validate(); err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	if !config.IsEasyDonateSignatureVerificationSkipped() {
		signatureSource := fmt.Sprintf("%d@%d@%s", cmd.PaymentID, cmd.Cost, cmd.Customer)
		signature := hashHelper.HashHMAC(signatureSource, config.GetEasyDonateKey())
		if !strings.EqualFold(signature, cmd.Signature) {
			utils.HttpError(ctx, w, utils.NewBadRequestError("invalid signature", nil))
			return
		}
	}

	order, err := h.orderService.GetOrderByExternalID(ctx, strconv.FormatInt(cmd.PaymentID, 10))
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	err = h.orderService.CompleteOrder(ctx, order)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}
