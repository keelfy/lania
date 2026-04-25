package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/config"
	"github.com/lania-smp/backend/internal/domain"
	"github.com/lania-smp/backend/internal/storage"
	"github.com/lania-smp/backend/internal/utils/hashHelper"
)

type FreekassaService interface {
	ConstructPaymentURL(ctx context.Context, email string, orderID uuid.UUID, amount float64, currency domain.Currency) string
}

type freekassaService struct {
	storage storage.MainStorage
}

func NewFreekassaService(storage storage.MainStorage) FreekassaService {
	return &freekassaService{storage: storage}
}

var basePaymentURL = config.GetFreekassaBasePaymentURL()
var merchantID = config.GetFreekassaMerchantID()
var merchantPassword = config.GetFreekassaMerchantPassword1()

func (s *freekassaService) ConstructPaymentURL(ctx context.Context, email string, orderID uuid.UUID, amount float64, currency domain.Currency) string {
	signatureSource := fmt.Sprintf("%d:%0.2f:%s:%s:%s", merchantID, amount, merchantPassword, currency, orderID.String())
	signatureMD5 := hashHelper.Hash(signatureSource)
	// https://pay.fk.money?m=%d&oa=%d&o=%d&currency=%s&em=%s&s=%s&pay=PAY
	paymentURL := fmt.Sprintf(basePaymentURL+"?m=%d&oa=%d&o=%s&currency=%s&em=%s&s=%s&pay=PAY&i=&phone=", merchantID, amount, orderID.String(), currency, email, signatureMD5)
	return paymentURL
}
