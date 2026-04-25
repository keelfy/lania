package services

import "github.com/google/wire"

var ProviderSet = wire.NewSet(
	NewIdentityService,
	NewKratosService,
	NewPlanService,
	NewFlectoneService,
	NewAccessService,
	NewProfileService,
	NewProfileCosmeticsService,
	NewSeasonService,
	NewMojangService,
	NewProductService,
	NewOrderService,
	NewFreekassaService,
	NewLuckpermsService,
	NewBasketService,
	NewPurchaseService,
	NewIntegrationService,
	NewEasyDonateService,
)
