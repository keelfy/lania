package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/commands"
	"github.com/lania-smp/backend/internal/config"
	"github.com/lania-smp/backend/internal/domain"
	"github.com/lania-smp/backend/internal/logger"
	"github.com/lania-smp/backend/internal/storage"
	sql "github.com/lania-smp/backend/internal/storage/main"
	"github.com/lania-smp/backend/internal/utils"
)

type EasyDonateService interface {
	ConstructPaymentURL(ctx context.Context, queries sql.Queries, params EDConstructPaymentURLParams) (string, error)
	// CreateShopProduct creates a position in the EasyDonate control panel using the admin's
	// session cookie and CSRF token, and returns its numeric ID. EasyDonate has no public API for
	// this, so it reproduces what a browser does against the October CMS control panel.
	CreateShopProduct(ctx context.Context, cmd *commands.CreateEDProductCommand) (int64, error)
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

// edProductIDPattern matches the data-product-id markers October renders for every row of the
// shop's product table, in both the products page and the products_row AJAX partial.
var edProductIDPattern = regexp.MustCompile(`data-product-id="(\d+)"`)

type edProductsRowResponse struct {
	ProductsRow string `json:"products_row"`
}

func (s *easyDonateService) CreateShopProduct(ctx context.Context, cmd *commands.CreateEDProductCommand) (int64, error) {
	prices, err := s.storage.Queries().FindPricesByNames(ctx, []domain.ProductPriceName{cmd.PriceName})
	if err != nil {
		return 0, utils.NewInternalServerError("failed to get product tariff", err)
	}
	var amount float64
	found := false
	for _, price := range prices {
		if price.Currency == domain.CurrencyRUB {
			amount = price.Amount
			found = true
			break
		}
	}
	if !found {
		return 0, utils.NewBadRequestError("product tariff has no RUB price", nil)
	}

	shopID := config.GetEasyDonateShopID()
	if shopID == "" {
		return 0, utils.NewInternalServerError("ED_SHOP_ID is not configured", nil)
	}
	productsURL := fmt.Sprintf("%s/shop/%s/products", config.GetEasyDonateControlPanelURL(), shopID)

	httpClient := &http.Client{Timeout: 30 * time.Second}

	existingIDs, err := s.fetchExistingEDProductIDs(ctx, httpClient, productsURL, cmd.SessionKey)
	if err != nil {
		return 0, err
	}

	newID, err := s.postEDProductCreate(ctx, httpClient, productsURL, cmd, amount, existingIDs)
	if err != nil {
		return 0, err
	}

	return newID, nil
}

func (s *easyDonateService) fetchExistingEDProductIDs(ctx context.Context, httpClient *http.Client, productsURL, sessionKey string) (map[string]struct{}, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, productsURL, nil)
	if err != nil {
		return nil, utils.NewInternalServerError("failed to create request", err)
	}
	setEDCommonHeaders(req, sessionKey)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, utils.NewInternalServerError("failed to reach EasyDonate control panel", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, utils.NewInternalServerError("failed to read EasyDonate response body", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, edStatusError(resp.StatusCode, body)
	}

	matches := edProductIDPattern.FindAllStringSubmatch(string(body), -1)
	existingIDs := make(map[string]struct{}, len(matches))
	for _, match := range matches {
		existingIDs[match[1]] = struct{}{}
	}
	if len(matches) == 0 {
		return nil, utils.NewUnauthorizedError("EasyDonate session is invalid or expired", nil)
	}
	return existingIDs, nil
}

func (s *easyDonateService) postEDProductCreate(ctx context.Context, httpClient *http.Client, productsURL string, cmd *commands.CreateEDProductCommand, amount float64, existingIDs map[string]struct{}) (int64, error) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	fields := map[string]string{
		"_token":      cmd.CSRFToken,
		"name":        cmd.Name,
		"price":       strconv.FormatFloat(amount, 'f', -1, 64),
		"number":      "1",
		"description": cmd.Description,
		"type":        "group",
		"commands[0]": "lpv user {user} add permission",
	}
	for field, value := range fields {
		if err := writer.WriteField(field, value); err != nil {
			return 0, utils.NewInternalServerError("failed to build EasyDonate request", err)
		}
	}
	if len(cmd.Image) > 0 {
		part, err := writer.CreateFormFile("image", cmd.ImageFilename)
		if err != nil {
			return 0, utils.NewInternalServerError("failed to build EasyDonate request", err)
		}
		if _, err := part.Write(cmd.Image); err != nil {
			return 0, utils.NewInternalServerError("failed to build EasyDonate request", err)
		}
	}
	if err := writer.Close(); err != nil {
		return 0, utils.NewInternalServerError("failed to build EasyDonate request", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, productsURL, &body)
	if err != nil {
		return 0, utils.NewInternalServerError("failed to create request", err)
	}
	setEDCommonHeaders(req, cmd.SessionKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("X-CSRF-TOKEN", cmd.CSRFToken)
	req.Header.Set("X-October-Request-Handler", "onProductCreate")
	req.Header.Set("X-October-Request-Partials", "products_row")
	req.Header.Set("X-October-Request-Flash", "1")

	resp, err := httpClient.Do(req)
	if err != nil {
		return 0, utils.NewInternalServerError("failed to reach EasyDonate control panel", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, utils.NewInternalServerError("failed to read EasyDonate response body", err)
	}

	// October responds 419 ("Page Expired") when the CSRF token is invalid or stale. There is no
	// net/http constant for it because it isn't in the standard registry.
	const statusPageExpired = 419
	if resp.StatusCode == statusPageExpired {
		return 0, utils.NewUnauthorizedError("EasyDonate CSRF token is invalid or expired", nil)
	}
	if resp.StatusCode != http.StatusOK {
		return 0, edStatusError(resp.StatusCode, respBody)
	}

	var response edProductsRowResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return 0, utils.NewInternalServerError("EasyDonate returned an unexpected response", fmt.Errorf("status %d: %s", resp.StatusCode, truncate(respBody, 300)))
	}

	matches := edProductIDPattern.FindAllStringSubmatch(response.ProductsRow, -1)
	var newIDs []int64
	for _, match := range matches {
		if _, known := existingIDs[match[1]]; known {
			continue
		}
		id, err := strconv.ParseInt(match[1], 10, 64)
		if err != nil {
			continue
		}
		newIDs = append(newIDs, id)
	}

	if len(newIDs) == 0 {
		return 0, utils.NewConflictError("EasyDonate did not report a new product", nil)
	}
	largest := newIDs[0]
	for _, id := range newIDs[1:] {
		if id > largest {
			largest = id
		}
	}
	if len(newIDs) > 1 {
		logger.Warnf(ctx, "easydonate reported %d new product ids, picking the largest", len(newIDs))
	}
	return largest, nil
}

// setEDCommonHeaders sets the headers October's control panel expects on every AJAX request,
// including the session cookie built from the admin's session key. It never logs cmd, the cookie
// or the token.
func setEDCommonHeaders(req *http.Request, sessionKey string) {
	req.Header.Set("Cookie", "easydonate_session="+sessionKey)
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("User-Agent", "Mozilla/5.0")
}

// edStatusError maps an unexpected EasyDonate control panel response to an error, without ever
// including the request that produced it.
func edStatusError(statusCode int, body []byte) error {
	switch statusCode {
	case http.StatusUnauthorized, http.StatusForbidden:
		return utils.NewUnauthorizedError("EasyDonate session is invalid or expired", nil)
	default:
		return utils.NewInternalServerError("EasyDonate control panel returned an unexpected response", fmt.Errorf("status %d: %s", statusCode, truncate(body, 300)))
	}
}

func truncate(body []byte, limit int) string {
	if len(body) <= limit {
		return string(body)
	}
	return string(body[:limit])
}
