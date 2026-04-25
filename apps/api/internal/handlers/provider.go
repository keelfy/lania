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
	NewPurchaseHandler,
	NewBasketHandler,
)
