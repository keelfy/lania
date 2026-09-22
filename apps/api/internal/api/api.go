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
	profileResyncHandler    handlers.ProfileResyncHandler
	productHandler          handlers.ProductHandler
	orderHandler            handlers.OrderHandler
	freekassaHandler        handlers.AcquiringHandler
	purchaseHandler         handlers.PurchaseHandler
	basketHandler           handlers.BasketHandler
	adminUserHandler        handlers.AdminUserHandler
	adminProfileHandler     handlers.AdminProfileHandler
	adminGrantHandler       handlers.AdminGrantHandler
	adminCatalogHandler     handlers.AdminCatalogHandler
	seasonHandler           handlers.SeasonHandler
	notificationHandler     handlers.NotificationHandler
	accountHandler          handlers.AccountHandler
	integrationService      services.IntegrationService
	mojangService           services.MojangService
	playerSyncService       services.PlayerSyncService
	roleSyncService         services.RoleSyncService
	tokenAuth               *jwtAuth.JWTAuth
	oryAPI                  clients.OryAPI
}

func NewLaniaAPI(
	statusHandler handlers.StatusHandler,
	accessHandler handlers.AccessHandler,
	profileHandler handlers.ProfileHandler,
	profileCosmeticsHandler handlers.ProfileCosmeticsHandler,
	profileResyncHandler handlers.ProfileResyncHandler,
	productHandler handlers.ProductHandler,
	orderHandler handlers.OrderHandler,
	freekassaHandler handlers.AcquiringHandler,
	purchaseHandler handlers.PurchaseHandler,
	basketHandler handlers.BasketHandler,
	adminUserHandler handlers.AdminUserHandler,
	adminProfileHandler handlers.AdminProfileHandler,
	adminGrantHandler handlers.AdminGrantHandler,
	adminCatalogHandler handlers.AdminCatalogHandler,
	seasonHandler handlers.SeasonHandler,
	notificationHandler handlers.NotificationHandler,
	accountHandler handlers.AccountHandler,
	integrationService services.IntegrationService,
	mojangService services.MojangService,
	playerSyncService services.PlayerSyncService,
	roleSyncService services.RoleSyncService,
	oryAPI clients.OryAPI,
) LaniaAPI {
	return &laniaAPI{
		statusHandler:           statusHandler,
		accessHandler:           accessHandler,
		profileHandler:          profileHandler,
		profileCosmeticsHandler: profileCosmeticsHandler,
		profileResyncHandler:    profileResyncHandler,
		productHandler:          productHandler,
		orderHandler:            orderHandler,
		freekassaHandler:        freekassaHandler,
		purchaseHandler:         purchaseHandler,
		basketHandler:           basketHandler,
		adminUserHandler:        adminUserHandler,
		adminProfileHandler:     adminProfileHandler,
		adminGrantHandler:       adminGrantHandler,
		adminCatalogHandler:     adminCatalogHandler,
		seasonHandler:           seasonHandler,
		notificationHandler:     notificationHandler,
		accountHandler:          accountHandler,
		integrationService:      integrationService,
		mojangService:           mojangService,
		playerSyncService:       playerSyncService,
		roleSyncService:         roleSyncService,
		oryAPI:                  oryAPI,
		tokenAuth:               jwtAuth.New("HS256", config.GetJWTSecret(), nil),
	}
}

func (api *laniaAPI) BuildAPI(ctx context.Context) (*chi.Mux, error) {
	if err := api.seasonHandler.InitializePrimarySeason(ctx); err != nil {
		return nil, err
	}
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

	// look up mojang uuids of profiles in background to stay within the mojang rate limit
	go api.mojangService.RunProfileSync(ctx)

	// copy playtime and seen dates from the minecraft server so profile pages never wait for shell
	go api.playerSyncService.RunPlayerSync(ctx)

	// push roles changed in profiles to every season server, profiles are the source of truth for roles
	go api.roleSyncService.RunRoleSync(ctx)

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
			r.Get("/by-username/{username}", api.profileHandler.GetProfileDetailsByUsername)
		})

		r.Group(func(r chi.Router) {
			api.useUnprotectedRoutes(r)

			r.Get("/", api.profileHandler.GetPublicProfiles)
			r.Get("/top-playtime", api.profileHandler.GetTopPlaytimeProfiles)
			r.Get("/stats", api.profileHandler.GetProfilesStats)
		})

		r.Route("/{profileId}", func(r chi.Router) {
			r.Group(func(r chi.Router) {
				api.useUnprotectedRoutes(r)

				r.Get("/", api.profileHandler.GetUserProfileDetails)
				r.Get("/stats", api.profileHandler.GetProfileStats)
			})

			r.Group(func(r chi.Router) {
				api.useProtectedRoutes(r)

				r.Get("/cosmetics/options", api.profileCosmeticsHandler.GetProfileCosmeticOptions)
				r.Post("/cosmetics/name-color", api.profileCosmeticsHandler.SelectProfileNameColor)
				r.Post("/cosmetics/name-prefix/{type}", api.profileCosmeticsHandler.SelectProfileNamePrefix)
				r.Post("/resync", api.profileResyncHandler.ResyncProfile)
			})
		})
	})

	r.Route("/seasons", func(r chi.Router) {
		r.Get("/", api.seasonHandler.GetSeasons)
		r.Get("/{seasonId}/screenshots", api.seasonHandler.GetSeasonScreenshots)
	})

	r.Route("/seasons/{seasonId}", func(r chi.Router) {
		api.useProtectedRoutes(r)

		r.Post("/access/pre-register", api.accessHandler.ObtainFreeAccessForProfiles)
		r.Post("/access/register", api.accessHandler.RegisterProfilesForSeason)
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

	r.Route("/admin", func(r chi.Router) {
		api.useProtectedRoutes(r)
		r.Use(middleware.AdminOnly())

		r.Get("/users", api.adminUserHandler.GetUsers)
		r.Get("/users/{userId}", api.adminUserHandler.GetUserDetails)

		r.Route("/cosmetics", func(r chi.Router) {
			r.Get("/", api.adminCatalogHandler.GetCosmetics)
			r.Post("/name-colors", api.adminCatalogHandler.CreateNameColor)
			r.Put("/name-colors/{cosmeticId}", api.adminCatalogHandler.UpdateNameColor)
			r.Post("/name-prefixes", api.adminCatalogHandler.CreateNamePrefix)
			r.Put("/name-prefixes/{cosmeticId}", api.adminCatalogHandler.UpdateNamePrefix)
		})

		r.Route("/products", func(r chi.Router) {
			r.Get("/", api.adminCatalogHandler.GetProducts)
			r.Post("/", api.adminCatalogHandler.CreateProduct)
			r.Put("/{productId}", api.adminCatalogHandler.UpdateProduct)
		})

		r.Route("/seasons", func(r chi.Router) {
			r.Get("/", api.seasonHandler.GetAdminSeasons)
			r.Post("/", api.seasonHandler.CreateSeason)
			r.Put("/{seasonId}", api.seasonHandler.UpdateSeason)
			r.Delete("/{seasonId}", api.seasonHandler.DeleteSeason)

			r.Route("/{seasonId}/screenshots", func(r chi.Router) {
				r.Post("/", api.seasonHandler.CreateSeasonScreenshot)
				r.Put("/{screenshotId}", api.seasonHandler.UpdateSeasonScreenshot)
				r.Delete("/{screenshotId}", api.seasonHandler.DeleteSeasonScreenshot)
			})
		})

		r.Get("/profiles", api.adminProfileHandler.GetProfiles)
		r.Route("/profiles/{profileId}", func(r chi.Router) {
			r.Get("/", api.adminProfileHandler.GetProfileDetails)
			r.Put("/owner", api.adminProfileHandler.TransferProfile)
			r.Put("/role", api.adminProfileHandler.SetProfileRole)
			r.Delete("/owner", api.adminProfileHandler.ReleaseProfile)
			r.Post("/resync", api.profileResyncHandler.AdminResyncProfile)
			r.Get("/merge", api.adminProfileHandler.PreviewMergeProfiles)
			r.Post("/merge", api.adminProfileHandler.MergeProfiles)
			r.Get("/merges", api.adminProfileHandler.GetProfileMerges)

			r.Get("/grants", api.adminGrantHandler.GetGrants)
			r.Post("/grants", api.adminGrantHandler.GrantProduct)
			r.Post("/grants/cosmetic", api.adminGrantHandler.GrantCosmetic)
			r.Delete("/grants/{grantType}/{grantId}", api.adminGrantHandler.RevokeGrant)
		})
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

	r.Route("/notifications", func(r chi.Router) {
		api.useProtectedRoutes(r)

		r.Get("/", api.notificationHandler.GetNotifications)
		r.Post("/read", api.notificationHandler.MarkNotificationsRead)
	})

	r.Route("/account", func(r chi.Router) {
		api.useProtectedRoutes(r)

		r.Delete("/", api.accountHandler.DeleteAccount)
	})

	r.Route("/purchases", func(r chi.Router) {
		api.useProtectedRoutes(r)

		r.Get("/", api.purchaseHandler.GetPurchasedProducts)
	})

	return r
}
