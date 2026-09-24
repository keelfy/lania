package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strconv"
	"strings"
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
	// user_auth cookie, and returns its numeric ID. EasyDonate has no public API for
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

// edCSRFTagPatterns find the tags a page can carry its CSRF token in: the meta tag October's AJAX
// framework reads, and the hidden _token input of its forms. Attribute order and quotes vary, so
// the tag is matched first and its value attribute second.
var edCSRFTagPatterns = []struct {
	tag   *regexp.Regexp
	value *regexp.Regexp
}{
	{
		tag:   regexp.MustCompile(`(?i)<meta\b[^>]*\bname\s*=\s*["']csrf-token["'][^>]*>`),
		value: regexp.MustCompile(`(?i)\bcontent\s*=\s*["']([^"']+)["']`),
	},
	{
		tag:   regexp.MustCompile(`(?i)<input\b[^>]*\bname\s*=\s*["']_token["'][^>]*>`),
		value: regexp.MustCompile(`(?i)\bvalue\s*=\s*["']([^"']+)["']`),
	},
}

var edTitlePattern = regexp.MustCompile(`(?i)<title>([^<]*)`)

const (
	// edRememberCookie is the RainLab.User "remember me" cookie. It logs the admin in again and
	// makes October start a fresh session, so the session cookie is never asked for.
	edRememberCookie = "user_auth"
	// edXSRFCookie is Laravel's copy of the session's CSRF token, used when a page has no meta tag.
	edXSRFCookie = "XSRF-TOKEN"
)

type edProductsRowResponse struct {
	ProductsRow string `json:"products_row"`
}

// edProductsPage is what the products page tells us before a position is created: the rows that
// already exist and the CSRF token of the session October just started.
type edProductsPage struct {
	existingIDs map[string]struct{}
	csrfToken   string
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
	productsURL, err := url.Parse(fmt.Sprintf("%s/shop/%s/products", config.GetEasyDonateControlPanelURL(), shopID))
	if err != nil {
		return 0, utils.NewInternalServerError("EasyDonate control panel URL is invalid", err)
	}

	// The jar carries user_auth into the first request and the session October starts in
	// response into the second one, like a browser tab would.
	jar, err := cookiejar.New(nil)
	if err != nil {
		return 0, utils.NewInternalServerError("failed to create cookie jar", err)
	}
	jar.SetCookies(productsURL, []*http.Cookie{{Name: edRememberCookie, Value: cmd.UserAuth, Path: "/"}})
	httpClient := &http.Client{Timeout: 30 * time.Second, Jar: jar}

	page, err := s.fetchEDProductsPage(ctx, httpClient, productsURL)
	if err != nil {
		return 0, err
	}

	newID, err := s.postEDProductCreate(ctx, httpClient, productsURL, cmd, amount, page)
	if err != nil {
		return 0, err
	}

	return newID, nil
}

func (s *easyDonateService) fetchEDProductsPage(ctx context.Context, httpClient *http.Client, productsURL *url.URL) (*edProductsPage, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, productsURL.String(), nil)
	if err != nil {
		return nil, utils.NewInternalServerError("failed to create request", err)
	}
	setEDCommonHeaders(req)

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
	if len(matches) == 0 {
		return nil, utils.NewUnauthorizedError("EasyDonate user_auth is invalid or expired", edPageDiagnostics(resp, body))
	}
	existingIDs := make(map[string]struct{}, len(matches))
	for _, match := range matches {
		existingIDs[match[1]] = struct{}{}
	}

	page := &edProductsPage{existingIDs: existingIDs}
	page.csrfToken = findEDCSRFToken(body)
	return page, nil
}

func (s *easyDonateService) postEDProductCreate(ctx context.Context, httpClient *http.Client, productsURL *url.URL, cmd *commands.CreateEDProductCommand, amount float64, page *edProductsPage) (int64, error) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	fields := map[string]string{
		"name":        cmd.Name,
		"price":       strconv.FormatFloat(amount, 'f', -1, 64),
		"number":      "1",
		"description": cmd.Description,
		"type":        "group",
		"commands[0]": "lpv user {user} add permission",
	}
	if page.csrfToken != "" {
		fields["_token"] = page.csrfToken
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

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, productsURL.String(), &body)
	if err != nil {
		return 0, utils.NewInternalServerError("failed to create request", err)
	}
	setEDCommonHeaders(req)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("X-October-Request-Handler", "onProductCreate")
	req.Header.Set("X-October-Request-Partials", "products_row")
	req.Header.Set("X-October-Request-Flash", "1")
	if page.csrfToken != "" {
		req.Header.Set("X-CSRF-TOKEN", page.csrfToken)
	} else if xsrf := edJarCookie(httpClient.Jar, productsURL, edXSRFCookie); xsrf != "" {
		// Laravel accepts the XSRF-TOKEN cookie echoed back decoded, which is what axios does.
		decoded, err := url.PathUnescape(xsrf)
		if err != nil {
			decoded = xsrf
		}
		req.Header.Set("X-XSRF-TOKEN", decoded)
	} else {
		// October may run with CSRF protection off. If it doesn't, the POST comes back 419 below.
		logger.Warnf(ctx, "easydonate products page has no CSRF token, posting without one")
	}

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
		return 0, utils.NewUnauthorizedError("EasyDonate rejected the CSRF token", fmt.Errorf("sent token: %t", page.csrfToken != "" || req.Header.Get("X-XSRF-TOKEN") != ""))
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
		if _, known := page.existingIDs[match[1]]; known {
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

// setEDCommonHeaders sets the headers October's control panel expects on every AJAX request.
// Cookies come from the client's jar. Nothing here logs cmd, the cookies or the token.
func setEDCommonHeaders(req *http.Request) {
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("User-Agent", "Mozilla/5.0")
}

func findEDCSRFToken(body []byte) string {
	for _, pattern := range edCSRFTagPatterns {
		for _, tag := range pattern.tag.FindAll(body, -1) {
			if match := pattern.value.FindSubmatch(tag); match != nil {
				return html.UnescapeString(string(match[1]))
			}
		}
	}
	return ""
}

func edJarCookie(jar http.CookieJar, u *url.URL, name string) string {
	for _, cookie := range jar.Cookies(u) {
		if cookie.Name == name {
			return cookie.Value
		}
	}
	return ""
}

// edPageDiagnostics describes a products page that has no product rows, so the log shows whether
// it was a login page, a redirect or an anti-bot stub. It carries no cookies and no body.
func edPageDiagnostics(resp *http.Response, body []byte) error {
	title := ""
	if match := edTitlePattern.FindSubmatch(body); match != nil {
		title = strings.TrimSpace(html.UnescapeString(string(match[1])))
	}
	return fmt.Errorf("status %d, final path %q, title %q", resp.StatusCode, resp.Request.URL.Path, title)
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
