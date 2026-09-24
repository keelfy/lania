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
	NewPurchaseHandler,
	NewBasketHandler,
	NewAdminUserHandler,
	NewAdminProfileHandler,
	NewAdminGrantHandler,
	NewAdminCatalogHandler,
	NewSeasonHandler,
	NewChunkClaimHandler,
	NewNotificationHandler,
	NewAccountHandler,
	NewUploadHandler,
)
