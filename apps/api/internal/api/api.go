package api

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	jwtAuth "github.com/go-chi/jwtauth/v5"
	"github.com/lania-smp/backend/internal/clients"
	"github.com/lania-smp/backend/internal/config"
	"github.com/lania-smp/backend/internal/domain"
	"github.com/lania-smp/backend/internal/handlers"
	"github.com/lania-smp/backend/internal/logger"
	"github.com/lania-smp/backend/internal/middleware"
	"github.com/lania-smp/backend/internal/services"
	httpSwagger "github.com/swaggo/http-swagger"
)

type LaniaAPI interface {
	BuildAPI(ctx context.Context) (*chi.Mux, error)
}

type laniaAPI struct {
	statusHandler           handlers.StatusHandler
	accessHandler           handlers.AccessHandler
	profileHandler          handlers.ProfileHandler
	profileCosmeticsHandler handlers.ProfileCosmeticsHandler
	productHandler          handlers.ProductHandler
	orderHandler            handlers.OrderHandler
	freekassaHandler        handlers.AcquiringHandler
	purchaseHandler         handlers.PurchaseHandler
	basketHandler           handlers.BasketHandler
	integrationService      services.IntegrationService
	tokenAuth               *jwtAuth.JWTAuth
	oryAPI                  clients.OryAPI
}

func NewLaniaAPI(
	statusHandler handlers.StatusHandler,
	accessHandler handlers.AccessHandler,
	profileHandler handlers.ProfileHandler,
	profileCosmeticsHandler handlers.ProfileCosmeticsHandler,
	productHandler handlers.ProductHandler,
	orderHandler handlers.OrderHandler,
	freekassaHandler handlers.AcquiringHandler,
	purchaseHandler handlers.PurchaseHandler,
	basketHandler handlers.BasketHandler,
	integrationService services.IntegrationService,
	oryAPI clients.OryAPI,
) LaniaAPI {
	return &laniaAPI{
		statusHandler:           statusHandler,
		accessHandler:           accessHandler,
		profileHandler:          profileHandler,
		profileCosmeticsHandler: profileCosmeticsHandler,
		productHandler:          productHandler,
		orderHandler:            orderHandler,
		freekassaHandler:        freekassaHandler,
		purchaseHandler:         purchaseHandler,
		basketHandler:           basketHandler,
		integrationService:      integrationService,
		oryAPI:                  oryAPI,
		tokenAuth:               jwtAuth.New("HS256", config.GetJWTSecret(), nil),
	}
}

func (api *laniaAPI) BuildAPI(ctx context.Context) (*chi.Mux, error) {
	r := chi.NewRouter()

	// connect to donation alerts centrifugo
	integration, err := api.integrationService.GetOAuth2IntegrationByServiceName(ctx, domain.IntegrationServiceDonationAlert)
	if err != nil {
		logger.Errorf(ctx, "failed to get oauth2 integration by service name: %v", err)
		return nil, err
	}

	go func() {
		api.integrationService.ConnectToCentrifugo(ctx, integration.AccessToken)
	}()

	// middlewares
	r.Use(chiMiddleware.RequestID)
	r.Use(chiMiddleware.RealIP)
	r.Use(chiMiddleware.Recoverer)
	r.Use(middleware.CORS)
	r.Use(middleware.LocaleMiddleware())
	r.Use(middleware.CurrencyMiddleware())
	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Timeout(config.GetContextTimeoutMs() * time.Millisecond))

	// /v1 routes
	r.Mount("/v1", api.v1RouteHandler())

	// swagger endpoint
	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	logger.Info(ctx, "API is ready")
	return r, nil
}

func (api *laniaAPI) useProtectedRoutes(r chi.Router) {
	r.Use(middleware.SessionMiddleware(api.oryAPI, false))
}

func (api *laniaAPI) useUnprotectedRoutes(r chi.Router) {
	r.Use(middleware.SessionMiddleware(api.oryAPI, true))
}

func (api *laniaAPI) useApiKey(r chi.Router) {
	r.Use(middleware.ApiKey())
}

func (api *laniaAPI) v1RouteHandler() http.Handler {
	r := chi.NewRouter()

	r.Get("/health", api.statusHandler.Health)

	r.Route("/profiles", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			api.useUnprotectedRoutes(r)

			r.Get("/check-username/{username}", api.accessHandler.CheckUsernames)
		})

		r.Group(func(r chi.Router) {
			api.useUnprotectedRoutes(r)

			r.Get("/", api.profileHandler.GetPublicProfiles)
		})

		r.Route("/{profileId}", func(r chi.Router) {
			r.Group(func(r chi.Router) {
				api.useUnprotectedRoutes(r)

				r.Get("/", api.profileHandler.GetUserProfileDetails)
			})

			r.Group(func(r chi.Router) {
				api.useProtectedRoutes(r)

				r.Get("/cosmetics/options", api.profileCosmeticsHandler.GetProfileCosmeticOptions)
				r.Post("/cosmetics/name-color", api.profileCosmeticsHandler.SelectProfileNameColor)
				r.Post("/cosmetics/name-prefix/{type}", api.profileCosmeticsHandler.SelectProfileNamePrefix)
			})
		})
	})

	r.Route("/seasons/{seasonId}", func(r chi.Router) {
		api.useProtectedRoutes(r)

		r.Post("/access/pre-register", api.accessHandler.ObtainFreeAccessForProfiles)
		r.Post("/get-access", api.accessHandler.ObtainAccessForProfiles)
	})

	r.Route("/products", func(r chi.Router) {
		r.Get("/", api.productHandler.GetProducts)
		r.Get("/{productId}", api.productHandler.GetProductByID)
	})

	r.Route("/orders", func(r chi.Router) {
		api.useProtectedRoutes(r)

		r.Post("/", api.orderHandler.CreateOrder)
		r.Get("/{orderId}", api.orderHandler.GetOrderByID)
	})

	r.Route("/users/{userId}", func(r chi.Router) {
		api.useProtectedRoutes(r)

		r.Get("/profiles", api.profileHandler.GetUserProfiles)
		r.Get("/orders", api.orderHandler.GetOrdersByUserID)
	})

	r.Route("/callbacks", func(r chi.Router) {
		r.Post("/freekassa", api.freekassaHandler.HandleFreekassaResult)
		r.Post("/easydonate", api.freekassaHandler.HandleEasyDonateResult)
	})

	r.Route("/basket", func(r chi.Router) {
		api.useProtectedRoutes(r)

		r.Get("/", api.basketHandler.GetBasketItems)
		r.Post("/", api.basketHandler.AddBasketItem)
		r.Delete("/", api.basketHandler.DeleteBasketItem)
	})

	r.Route("/purchases", func(r chi.Router) {
		api.useProtectedRoutes(r)

		r.Get("/", api.purchaseHandler.GetPurchasedProducts)
	})

	return r
}
