package handlers

import (
	"fmt"
	"net/http"
	"slices"
	"sync"

	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/clients"
	"github.com/lania-smp/backend/internal/config"
	"github.com/lania-smp/backend/internal/domain"
	"github.com/lania-smp/backend/internal/logger"
	"github.com/lania-smp/backend/internal/presenter"
	"github.com/lania-smp/backend/internal/services"
	"github.com/lania-smp/backend/internal/storage"
	sql "github.com/lania-smp/backend/internal/storage/main"
	"github.com/lania-smp/backend/internal/transport/http/binders"
	"github.com/lania-smp/backend/internal/transport/http/responses"
	"github.com/lania-smp/backend/internal/utils"
)

type OrderHandler interface {
	CreateOrder(w http.ResponseWriter, r *http.Request)
	GetOrderByID(w http.ResponseWriter, r *http.Request)
	GetOrdersByUserID(w http.ResponseWriter, r *http.Request)
}

type orderHandler struct {
	orderService      services.OrderService
	freekassaService  services.FreekassaService
	easyDonateService services.EasyDonateService
	productService    services.ProductService
	basketService     services.BasketService
	purchaseService   services.PurchaseService
	storage           storage.MainStorage
	oryAPI            clients.OryAPI
}

func NewOrderHandler(
	orderService services.OrderService,
	freekassaService services.FreekassaService,
	easyDonateService services.EasyDonateService,
	productService services.ProductService,
	basketService services.BasketService,
	purchaseService services.PurchaseService,
	storage storage.MainStorage,
	oryAPI clients.OryAPI,
) OrderHandler {
	return &orderHandler{
		orderService:      orderService,
		freekassaService:  freekassaService,
		easyDonateService: easyDonateService,
		productService:    productService,
		basketService:     basketService,
		purchaseService:   purchaseService,
		storage:           storage,
		oryAPI:            oryAPI,
	}
}

type itemToOrder struct {
	Product   *domain.Product
	ProfileID uuid.UUID
	SeasonID  uuid.UUID
	Quantity  int
	Amounts   []*domain.OrderAmounts
}

func (h *orderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	cmd, err := binders.BindCreateOrder(r)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	if err := cmd.Validate(); err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	session, err := h.oryAPI.GetSession(ctx, r.Header.Get("Cookie"))
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	traits := session.Identity.Traits
	if traits == nil {
		utils.HttpError(ctx, w, utils.NewForbiddenError("incorrect identity", nil))
		return
	}

	email, ok := traits.(map[string]interface{})["email"].(string)
	if !ok {
		utils.HttpError(ctx, w, utils.NewForbiddenError("incorrect identity", nil))
		return
	}

	productIDs := make(uuid.UUIDs, len(cmd.Items))
	for i, product := range cmd.Items {
		productIDs[i] = product.ProductID
	}

	products, err := h.productService.GetProductsByIDs(ctx, productIDs)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	seasonID := config.GetActiveSeasonID()
	purchases, err := h.purchaseService.GetPurchasesByProducts(ctx, cmd.UserID, seasonID, products)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	// filter products that are not purchased
	itemsToOrder := make([]*itemToOrder, 0)

	for _, item := range cmd.Items {
		alreadyPurchased := false
		for _, purchase := range purchases {
			if purchase.ProductID == item.ProductID && purchase.ProfileID == item.ProfileID && purchase.SeasonID == item.SeasonID {
				alreadyPurchased = true
				break
			}
		}

		if alreadyPurchased {
			continue
		}

		product := &domain.Product{}
		for _, p := range products {
			if p.ID == item.ProductID {
				product = p
				break
			}
		}

		if product == nil {
			utils.HttpError(ctx, w, utils.NewBadRequestError("product not found", nil))
			return
		}

		itemsToOrder = append(itemsToOrder, &itemToOrder{
			Product:   product,
			ProfileID: item.ProfileID,
			SeasonID:  seasonID,
			Quantity:  item.Quantity,
			Amounts:   []*domain.OrderAmounts{},
		})
	}

	if len(itemsToOrder) == 0 {
		utils.HttpError(ctx, w, utils.NewBadRequestError("no products to order", nil))
		return
	}

	priceNamesToOrder := []domain.ProductPriceName{}
	for _, item := range itemsToOrder {
		if !slices.Contains(priceNamesToOrder, item.Product.PriceName) {
			priceNamesToOrder = append(priceNamesToOrder, item.Product.PriceName)
		}
	}

	uniquePrices, err := h.productService.GetPricesByNames(ctx, priceNamesToOrder)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	overallAmounts := []*domain.OrderAmounts{}
	overallAmountMap := make(map[domain.Currency]float64)

	for _, item := range itemsToOrder {
		productQuantity := 1
		for _, p := range cmd.Items {
			if p.ProductID == item.Product.ID {
				productQuantity = p.Quantity
				break
			}
		}

		amountMap := make(map[domain.Currency]float64)
		for _, price := range uniquePrices {
			if price.Name == item.Product.PriceName {
				amountMap[price.Currency] += price.Amount * float64(productQuantity)
			}
		}

		item.Amounts = []*domain.OrderAmounts{}
		for currency, amount := range amountMap {
			item.Amounts = append(item.Amounts, &domain.OrderAmounts{
				Currency: currency,
				Amount:   amount,
			})
		}

		for currency, amount := range amountMap {
			overallAmountMap[currency] += amount
		}
	}

	for currency, amount := range overallAmountMap {
		overallAmounts = append(overallAmounts, &domain.OrderAmounts{
			Currency: currency,
			Amount:   amount,
		})
	}

	logger.Infof(ctx, "overall amounts: %+v", overallAmounts)

	var paymentURL string

	txErr := h.storage.BeginTx(ctx, func(queries sql.Queries) error {
		orderID, err := h.orderService.CreateOrder(ctx, queries, overallAmounts, cmd)
		if err != nil {
			return err
		}

		for _, product := range itemsToOrder {
			err = h.orderService.CreateOrderItem(ctx, queries, orderID, product.Product, product.ProfileID, product.SeasonID, product.Quantity, product.Amounts)
			if err != nil {
				return err
			}
		}

		currency := utils.GetCurrencyFromCtx(ctx)
		amount := overallAmountMap[currency]

		switch cmd.PaymentMethod {
		case domain.PaymentMethodFreekassa:
			paymentURL = h.freekassaService.ConstructPaymentURL(ctx, email, orderID, amount, currency)
		case domain.PaymentMethodDonationAlerts:
			locale := utils.GetLocaleFromCtx(ctx)
			paymentURL = fmt.Sprintf("/%s/orders/%s/donate", locale, orderID.String())
		case domain.PaymentMethodEasyDonate:
			edProductIDs := make(uuid.UUIDs, 0)
			profileID := uuid.Nil
			for _, item := range itemsToOrder {
				if profileID == uuid.Nil {
					profileID = item.ProfileID
				} else if profileID != item.ProfileID {
					return utils.NewBadRequestError("all products must have the same profile id when using easy donate payment method", nil)
				}
				edProductIDs = append(edProductIDs, item.Product.ID)
			}

			url, err := h.easyDonateService.ConstructPaymentURL(ctx, queries, services.EDConstructPaymentURLParams{
				Email:      email,
				OrderID:    orderID,
				ProductIDs: edProductIDs,
				ProfileID:  profileID,
				Amount:     amount,
				Currency:   currency,
			})
			if err != nil {
				return err
			}
			paymentURL = url
		default:
			return utils.NewBadRequestError("payment method not supported", nil)
		}

		logger.Infof(ctx, "order created successfully: %d", orderID)
		return nil
	})
	if txErr != nil {
		utils.HttpError(ctx, w, txErr)
		return
	}

	basketItems, err := h.basketService.GetBasketItemsByUserID(ctx, cmd.UserID)
	if err != nil {
		logger.Errorf(ctx, "failed to get basket items by user id: %v", err)
	} else {
		basketItemIDs := make([]uuid.UUID, 0)
		for _, item := range itemsToOrder {
			for _, basketItem := range basketItems {
				if basketItem.ProductID == item.Product.ID && basketItem.ProfileID == item.ProfileID {
					basketItemIDs = append(basketItemIDs, basketItem.ID)
				}
			}
		}

		err = h.basketService.DeleteBasketItemByIDs(ctx, h.storage.Queries(), basketItemIDs)
		if err != nil {
			logger.Errorf(ctx, "failed to delete basket items by ids: %v", err)
		}
	}

	res := &responses.CreateOrder{
		PaymentURL: paymentURL,
	}
	utils.WriteHttpJsonResponse(ctx, w, res)
}

func (h *orderHandler) GetOrderByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	orderID, err := binders.BindPathVariableAsUUID(r, "orderId")
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	order, err := h.orderService.GetOrderByID(ctx, orderID)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	authUserID, err := utils.GetUserIDFromCtx(ctx)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	if order.UserID != authUserID {
		utils.HttpError(ctx, w, utils.NewForbiddenError("incorrect order", nil))
		return
	}

	items, err := h.orderService.GetOrderItemsByOrderID(ctx, orderID)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	itemsResponse := make([]*responses.OrderItem, len(items))
	for i, item := range items {
		itemsResponse[i] = presenter.PresentOrderItem(item)
	}
	res := presenter.PresentOrder(order, itemsResponse)
	utils.WriteHttpJsonResponse(ctx, w, res)
}

func (h *orderHandler) GetOrdersByUserID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	authUserID, err := utils.GetUserIDFromCtx(ctx)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	orders, err := h.orderService.GetOrdersByUserID(ctx, authUserID)
	if err != nil {
		utils.HttpError(ctx, w, err)
		return
	}

	orderItemsMap := sync.Map{}
	wg := sync.WaitGroup{}

	for _, order := range orders {
		wg.Add(1)
		go func() {
			defer wg.Done()
			items, err := h.orderService.GetOrderItemsByOrderID(ctx, order.ID)
			if err != nil {
				logger.Errorf(ctx, "failed to get order items by order id: %v", err)
			}
			orderItemsMap.Store(order.ID, items)
		}()
	}
	wg.Wait()

	res := make([]*responses.Order, len(orders))
	for i, order := range orders {
		orderItemsRaw, ok := orderItemsMap.Load(order.ID)
		if !ok {
			logger.Errorf(ctx, "failed to get order items by order id: %v", order.ID)
			res[i] = presenter.PresentOrder(order, []*responses.OrderItem{})
			continue
		}
		orderItems := orderItemsRaw.([]*domain.OrderItem)
		itemsResponse := make([]*responses.OrderItem, len(orderItems))
		for j, item := range orderItems {
			itemsResponse[j] = presenter.PresentOrderItem(item)
		}
		res[i] = presenter.PresentOrder(order, itemsResponse)
	}

	utils.WriteHttpJsonResponse(ctx, w, res)
}
