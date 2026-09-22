package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

type edPaymentCreateProduct struct {
	ID       int64 `json:"id"`
	Quantity int   `json:"quantity"`
}

type edPaymentCreateRequest struct {
	Username  string                   `json:"username"`
	ServerID  int64                    `json:"server_id"`
	Products  []edPaymentCreateProduct `json:"products"`
	Email     string                   `json:"email"`
	ReturnURL string                   `json:"return_url,omitempty"`
}

type edPaymentCreateResponse struct {
	Success bool `json:"success"`
	Data    struct {
		PaymentID int64  `json:"payment_id"`
		URL       string `json:"url"`
	} `json:"data"`
	Error *struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
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
	uniqueProducts := make(map[uuid.UUID]struct{}, len(params.ProductIDs))
	for _, productID := range params.ProductIDs {
		uniqueProducts[productID] = struct{}{}
	}
	if len(edProductIDs) != len(uniqueProducts) {
		return "", utils.NewConflictError("some products have no EasyDonate ID", nil)
	}

	products := make([]edPaymentCreateProduct, 0, len(edProductIDs))
	for _, edProductID := range edProductIDs {
		products = append(products, edPaymentCreateProduct{ID: edProductID, Quantity: 1})
	}

	payload := edPaymentCreateRequest{
		Username:  profile.MinecraftUsername,
		ServerID:  config.GetEasyDonateProxyServerID(),
		Products:  products,
		Email:     params.Email,
		ReturnURL: fmt.Sprintf(config.GetEasyDonateSuccessURL(), params.OrderID.String()),
	}

	requestBody, err := json.Marshal(payload)
	if err != nil {
		return "", utils.NewInternalServerError("failed to marshal Easy Donate payment create request", err)
	}

	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest(http.MethodPost, config.GetEasyDonateCreatePaymentEndpoint(), bytes.NewReader(requestBody))
	if err != nil {
		return "", utils.NewInternalServerError("failed to create request", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Shop-Key", config.GetEasyDonateKey())
	logger.Debugf(ctx, "easy donate create payment request body: %s", string(requestBody))

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", utils.NewInternalServerError("failed to do Easy Donate request", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", utils.NewInternalServerError("failed to read Easy Donate response body", err)
	}

	var response edPaymentCreateResponse
	if err := json.Unmarshal(body, &response); err != nil {
		logger.Errorf(ctx, "failed to decode Easy Donate response: %s", string(body))
		return "", utils.NewInternalServerError("failed to decode Easy Donate response", err)
	}

	if !response.Success {
		errMsg := string(body)
		if response.Error != nil {
			errMsg = fmt.Sprintf("%s: %s", response.Error.Code, response.Error.Message)
		}
		return "", utils.NewInternalServerError("failed to do Easy Donate request", fmt.Errorf("status code: %d, %s", resp.StatusCode, errMsg))
	}

	// save easy donate payment id to order
	err = s.orderService.UpdateOrderExternalIDByID(ctx, queries, params.OrderID, strconv.FormatInt(response.Data.PaymentID, 10))
	if err != nil {
		return "", err
	}

	return response.Data.URL, nil
}
