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
	NewProductService,
	NewFulfillmentService,
	NewOrderService,
	NewFreekassaService,
	NewBasketService,
	NewPurchaseService,
	NewIntegrationService,
	NewEasyDonateService,
)
