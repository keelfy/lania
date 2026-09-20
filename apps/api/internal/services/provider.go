package services

import "github.com/google/wire"

var ProviderSet = wire.NewSet(
	NewIdentityService,
	NewAdminUserService,
	NewAdminProfileService,
	NewAdminGrantService,
	NewKratosService,
	NewMinecraftService,
	NewAccessService,
	NewProfileService,
	NewProfileCosmeticsService,
	NewSeasonService,
	NewMojangService,
	NewPlayerSyncService,
	NewRoleSyncService,
	NewProfileResyncService,
	NewProductService,
	NewFulfillmentService,
	NewOrderService,
	NewFreekassaService,
	NewBasketService,
	NewPurchaseService,
	NewIntegrationService,
	NewNotificationService,
	NewEasyDonateService,
)
