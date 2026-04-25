package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/config"
	"github.com/lania-smp/backend/internal/domain"
	"github.com/lania-smp/backend/internal/logger"
	"github.com/lania-smp/backend/internal/storage"
	sql "github.com/lania-smp/backend/internal/storage/main"
	"github.com/lania-smp/backend/internal/utils"
)

type EasyDonateService interface {
	ConstructPaymentURL(ctx context.Context, queries sql.Queries, params EDConstructPaymentURLParams) (string, error)
}

type easyDonateService struct {
	storage        storage.MainStorage
	profileService ProfileService
	orderService   OrderService
}

func NewEasyDonateService(
	storage storage.MainStorage,
	profileService ProfileService,
	orderService OrderService,
) EasyDonateService {
	return &easyDonateService{
		storage:        storage,
		profileService: profileService,
		orderService:   orderService,
	}
}

type edResponse struct {
	Success bool `json:"success"`
}

type easyDonatePaymentCreateRes struct {
	Success  bool `json:"success"`
	Response struct {
		URL     string `json:"url"`
		Payment struct {
			ID int64 `json:"id"`
		}
	}
}

type EDConstructPaymentURLParams struct {
	Email      string
	OrderID    uuid.UUID
	ProductIDs uuid.UUIDs
	ProfileID  uuid.UUID
	Amount     float64
	Currency   domain.Currency
}

func (s *easyDonateService) ConstructPaymentURL(ctx context.Context, queries sql.Queries, params EDConstructPaymentURLParams) (string, error) {
	profile, err := s.profileService.GetProfileByID(ctx, params.ProfileID)
	if err != nil {
		return "", err
	}

	edProductIDs, err := s.storage.Queries().FindEDProductsByProductIDs(ctx, params.ProductIDs)
	if err != nil {
		return "", utils.NewInternalServerError("failed to find ed products by product ids", err)
	}

	products := make(map[string]int)
	for _, edProductID := range edProductIDs {
		products[strconv.FormatInt(edProductID, 10)] = 1
	}
	productsStr, err := json.Marshal(products)
	if err != nil {
		return "", utils.NewInternalServerError("failed to marshal products", err)
	}

	queryParams := url.Values{}
	queryParams.Add("customer", profile.MinecraftUsername)
	queryParams.Add("server_id", strconv.FormatInt(config.GetEasyDonateProxyServerID(), 10))
	queryParams.Add("products", string(productsStr))
	queryParams.Add("email", params.Email)
	queryParams.Add("success_url", fmt.Sprintf(config.GetEasyDonateSuccessURL(), params.OrderID.String()))
	shopKey := config.GetEasyDonateKey()

	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("GET", config.GetEasyDonateCreatePaymentEndpoint(), nil)
	if err != nil {
		return "", utils.NewInternalServerError("failed to create request", err)
	}
	req.URL.RawQuery = queryParams.Encode()
	logger.Debugf(ctx, "easy donate create payment url params: %s", req.URL.String())
	req.Header.Set(config.GetEasyDonateShopKeyHeader(), shopKey)
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", utils.NewInternalServerError("failed to do Easy Donate request", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", utils.NewInternalServerError("failed to read Easy Donate response body", err)
		}
		return "", utils.NewInternalServerError("failed to do Easy Donate request", fmt.Errorf("status code: %d, body: %s", resp.StatusCode, string(errBody)))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Errorf(ctx, "failed to read Easy Donate response body: %s", string(body))
		return "", utils.NewInternalServerError("failed to read EasyDonate response body", err)
	}

	var basicResponse edResponse
	err = json.NewDecoder(bytes.NewReader(body)).Decode(&basicResponse)
	if err != nil {
		logger.Errorf(ctx, "failed to decode Easy Donate response: %s", string(body))
		return "", utils.NewInternalServerError("failed to decode EasyDonate response", err)
	}

	if !basicResponse.Success {
		return "", utils.NewInternalServerError("failed to do Easy Donate request", fmt.Errorf("status code: %d, body: %s", resp.StatusCode, string(body)))
	}

	var response easyDonatePaymentCreateRes
	err = json.NewDecoder(bytes.NewReader(body)).Decode(&response)
	if err != nil {
		logger.Errorf(ctx, "failed to decode Easy Donate response: %s", string(body))
		return "", utils.NewInternalServerError("failed to decode Easy Donate response", err)
	}

	// save easy donate payment id to order
	err = s.orderService.UpdateOrderExternalIDByID(ctx, queries, params.OrderID, strconv.FormatInt(response.Response.Payment.ID, 10))
	if err != nil {
		return "", err
	}

	return response.Response.URL, nil
}
