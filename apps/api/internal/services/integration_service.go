package services

import (
	"bytes"
	"context"
	stdsql "database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/centrifugal/centrifuge-go"
	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/config"
	"github.com/lania-smp/backend/internal/domain"
	"github.com/lania-smp/backend/internal/logger"
	"github.com/lania-smp/backend/internal/storage"
	sql "github.com/lania-smp/backend/internal/storage/main"
	"github.com/lania-smp/backend/internal/utils"
)

type IntegrationService interface {
	GetOAuth2IntegrationByServiceName(ctx context.Context, serviceName domain.IntegrationService) (*domain.OAuth2Integration, error)
	RefreshOAuth2Integration(ctx context.Context, serviceName domain.IntegrationService) error
	ConnectToCentrifugo(ctx context.Context, accessToken string) (string, error)
}

type integrationService struct {
	storage      storage.MainStorage
	orderService OrderService
}

func NewIntegrationService(
	storage storage.MainStorage,
	orderService OrderService,
) IntegrationService {
	return &integrationService{
		storage:      storage,
		orderService: orderService,
	}
}

func (s *integrationService) GetOAuth2IntegrationByServiceName(ctx context.Context, serviceName domain.IntegrationService) (*domain.OAuth2Integration, error) {
	res, err := s.storage.Queries().FindOAuth2IntegrationByServiceName(ctx, serviceName)
	if err == stdsql.ErrNoRows {
		return nil, utils.NewNotFoundError("oauth2 integration not found", err)
	} else if err != nil {
		return nil, utils.NewInternalServerError("failed to get oauth2 integration by service name", err)
	}
	return res, nil
}

func (s *integrationService) RefreshOAuth2Integration(ctx context.Context, serviceName domain.IntegrationService) error {
	integration, err := s.GetOAuth2IntegrationByServiceName(ctx, serviceName)
	if err != nil {
		return err
	}

	if integration.RefreshToken == "" {
		return utils.NewBadRequestError("refresh token not found", err)
	}

	_, err = s.RefreshDonationAlertIntegration(ctx, integration)
	if err != nil {
		return err
	}

	return nil
}

func (s *integrationService) RefreshDonationAlertIntegration(ctx context.Context, integration *domain.OAuth2Integration) (accessToken string, err error) {
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	body := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {integration.RefreshToken},
		"client_id":     {config.GetDonationAlertsClientID()},
		"client_secret": {config.GetDonationAlertsClientSecret()},
		"scope":         {config.GetDonationAlertsScope()},
	}
	bodyReader := strings.NewReader(body.Encode())
	req, err := http.NewRequest("POST", "https://www.donationalerts.com/oauth/token", bodyReader)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to refresh donation alert integration: %s", resp.Status)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var response struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}
	err = json.Unmarshal(bodyBytes, &response)
	if err != nil {
		return "", err
	}

	err = s.storage.Queries().UpdateOAuth2Integration(ctx, sql.UpdateOAuth2IntegrationParams{
		ServiceName:  domain.IntegrationServiceDonationAlert,
		AccessToken:  response.AccessToken,
		RefreshToken: response.RefreshToken,
	})
	if err != nil {
		return "", err
	}

	return response.AccessToken, nil
}

type DAAuthorizeMessage struct {
	Params struct {
		Token string `json:"token"`
	} `json:"params"`
	ID int `json:"id"`
}

type DAAuthorizeMessageResponse struct {
	Result struct {
		Client  uuid.UUID `json:"client"`
		Version string    `json:"version"`
	} `json:"result"`
	ID int `json:"id"`
}

type eventHandler struct {
	integrationService IntegrationService
}

func (h *eventHandler) OnConnect(c *centrifuge.Client, e centrifuge.ConnectEvent) {
	logger.Debugf(context.Background(), "Connected: %s", e.ClientID)
}

func (h *eventHandler) OnError(_ *centrifuge.Client, e centrifuge.ErrorEvent) {
	logger.Debugf(context.Background(), "Error: %s", e.Message)
}

func (h *eventHandler) OnDisconnect(_ *centrifuge.Client, e centrifuge.DisconnectEvent) {
	logger.Debugf(context.Background(), "Disconnected: %s", e.Reason)
}

func (h *eventHandler) OnPrivateSub(c *centrifuge.Client, e centrifuge.PrivateSubEvent) (string, error) {
	/*
			curl \
		    -X POST https://www.donationalerts.com/api/v1/centrifuge/subscribe \
		    -H "Authorization: Bearer <token>" \
		    -H "Content-Type: application/json" \
		    -d '{"channels":["$alerts:donation_<user_id>"], "client":"<uuidv4_client_id>"}'
	*/
	integration, err := h.integrationService.GetOAuth2IntegrationByServiceName(context.Background(), domain.IntegrationServiceDonationAlert)
	if err != nil {
		logger.Errorf(context.Background(), "DA centrifugo failed to get integration: %v", err)
		return "", err
	}

	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	body := struct {
		Channels []string `json:"channels"`
		Client   string   `json:"client"`
	}{
		Channels: []string{fmt.Sprintf("$alerts:donation_%d", config.GetDonationAlertsUserID())},
		Client:   e.ClientID,
	}
	reqBody, err := json.Marshal(body)
	if err != nil {
		logger.Errorf(context.Background(), "DA centrifugo failed to marshal body: %v", err)
		return "", err
	}
	req, err := http.NewRequest("POST", "https://www.donationalerts.com/api/v1/centrifuge/subscribe", bytes.NewReader(reqBody))
	if err != nil {
		logger.Errorf(context.Background(), "DA centrifugo failed to create subscribe request: %v", err)
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", integration.AccessToken))

	resp, err := httpClient.Do(req)
	if err != nil {
		logger.Errorf(context.Background(), "DA centrifugo failed to subscribe: %v", err)
		return "", err
	}
	defer resp.Body.Close()

	var response struct {
		Channels []struct {
			Channel string `json:"channel"`
			Token   string `json:"token"`
		} `json:"channels"`
	}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		logger.Errorf(context.Background(), "DA centrifugo failed to decode response: %v", err)
		return "", err
	}

	if resp.StatusCode >= 300 || resp.StatusCode < 200 {
		logger.Errorf(context.Background(), "DA centrifugo failed to subscribe: %s", resp.Status)
		return "", fmt.Errorf("failed to subscribe: %s", resp.Status)
	}

	if len(response.Channels) == 0 {
		logger.Errorf(context.Background(), "DA centrifugo failed to subscribe: no channels in response")
		return "", fmt.Errorf("no channels in response")
	}

	return response.Channels[0].Token, nil
}

type subEventHandler struct {
	orderService OrderService
}

func (h *subEventHandler) OnSubscribeSuccess(sub *centrifuge.Subscription, _ centrifuge.SubscribeSuccessEvent) {
	logger.Debugf(context.Background(), "Successfully subscribed to private channel %s", sub.Channel())
}

func (h *subEventHandler) OnSubscribeError(sub *centrifuge.Subscription, e centrifuge.SubscribeErrorEvent) {
	logger.Debugf(context.Background(), "Error subscribing to private channel %s: %v", sub.Channel(), e.Error)
}

func (h *subEventHandler) OnUnsubscribe(sub *centrifuge.Subscription, _ centrifuge.UnsubscribeEvent) {
	logger.Debugf(context.Background(), "Unsubscribed from private channel %s", sub.Channel())
}

type DonationMessage struct {
	ID                   int64   `json:"id"`
	Name                 string  `json:"name"`
	Username             string  `json:"username"`
	Message              string  `json:"message"`
	MessageType          string  `json:"message_type"`
	PayingSystem         string  `json:"paying_system"`
	Amount               float64 `json:"amount"`
	Currency             string  `json:"currency"`
	AmountInUserCurrency float64 `json:"amount_in_user_currency"`
	CreatedAt            string  `json:"created_at"`
}

func (h *subEventHandler) OnPublish(sub *centrifuge.Subscription, e centrifuge.PublishEvent) {
	var donationMessage DonationMessage
	err := json.Unmarshal(e.Data, &donationMessage)
	if err != nil {
		logger.Errorf(context.Background(), "[DA Notifications] failed to unmarshal donation message: %v", err)
		return
	}
	logger.Debugf(context.Background(), "[DA Notifications] Message received: %+v", donationMessage)

	orderID := strings.ToLower(strings.TrimSpace(donationMessage.Message))

	go func() {
		ctx := context.Background()
		orderID, err := uuid.Parse(orderID)
		if err != nil {
			logger.Errorf(ctx, "[DA Notifications] failed to parse order id: %v", err)
			return
		}

		order, err := h.orderService.GetOrderByID(ctx, orderID)
		if err != nil {
			logger.Errorf(ctx, "[DA Notifications] failed to get order by id: %v", err)
			return
		}

		expectedAmount := -1.0
		for _, amount := range order.Amounts {
			if strings.EqualFold(string(amount.Currency), donationMessage.Currency) {
				expectedAmount = amount.Amount
				break
			}
		}

		if expectedAmount == -1.0 {
			for _, amount := range order.Amounts {
				if strings.EqualFold(string(amount.Currency), string(domain.DefaultCurrency)) {
					expectedAmount = amount.Amount
					break
				}
			}
		}

		if expectedAmount < donationMessage.Amount {
			logger.Errorf(ctx, "[DA Notifications] amount mismatch: %v < %v", expectedAmount, donationMessage.Amount)
			return
		}

		err = h.orderService.CompleteOrder(ctx, order)
		if err != nil {
			logger.Errorf(ctx, "[DA Notifications] failed to complete order: %v", err)
		}
		logger.Infof(ctx, "[DA Notifications] order completed: %s", orderID.String())
	}()
}

func (s *integrationService) ConnectToCentrifugo(ctx context.Context, accessToken string) (string, error) {
	url := "wss://centrifugo.donationalerts.com/connection/websocket"
	c := centrifuge.NewJsonClient(url, centrifuge.DefaultConfig())

	c.SetToken(config.GetDonationAlertSocketConnectionToken())

	handler := &eventHandler{
		integrationService: s,
	}
	c.OnDisconnect(handler)
	c.OnConnect(handler)
	c.OnError(handler)
	c.OnPrivateSub(handler)

	if err := c.Connect(); err != nil {
		logger.Errorf(ctx, "DA centrifugo failed to connect: %v", err)
		return "", err
	}

	sub, err := c.NewSubscription(fmt.Sprintf("$alerts:donation_%d", config.GetDonationAlertsUserID()))
	if err != nil {
		logger.Errorf(context.Background(), "DA centrifugo failed to create subscription: %v", err)
	}

	subEventHandler := &subEventHandler{
		orderService: s.orderService,
	}
	sub.OnSubscribeSuccess(subEventHandler)
	sub.OnSubscribeError(subEventHandler)
	sub.OnUnsubscribe(subEventHandler)
	sub.OnPublish(subEventHandler)

	// Subscribe on private channel.
	err = sub.Subscribe()
	if err != nil {
		logger.Errorf(context.Background(), "DA centrifugo failed to subscribe: %v", err)
	}

	// Простая блокировка
	select {}
}
