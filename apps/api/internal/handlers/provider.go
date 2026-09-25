package handlers

import (
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(
	NewStatusHandler,
	NewProfileHandler,
	NewAccessHandler,
	NewProductHandler,
	NewOrderHandler,
	NewAcquiringHandler,
	NewProfileCosmeticsHandler,
	NewProfileResyncHandler,
	NewProfileVerificationHandler,
	NewPurchaseHandler,
	NewBasketHandler,
	NewAdminUserHandler,
	NewAdminProfileHandler,
	NewAdminGrantHandler,
	NewAdminCatalogHandler,
	NewSeasonHandler,
	NewSeasonWorldHandler,
	NewChunkClaimHandler,
	NewNotificationHandler,
	NewAccountHandler,
	NewUploadHandler,
)
