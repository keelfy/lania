package services

import "github.com/google/wire"

var ProviderSet = wire.NewSet(
	NewIdentityService,
	NewKratosService,
	NewMinecraftService,
	NewAccessService,
	NewProfileService,
	NewProfileCosmeticsService,
	NewSeasonService,
	NewMojangService,
	NewProductService,
	NewOrderService,
	NewFreekassaService,
	NewBasketService,
	NewPurchaseService,
	NewIntegrationService,
	NewEasyDonateService,
)
