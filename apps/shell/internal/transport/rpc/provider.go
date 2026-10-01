package rpc

import "github.com/google/wire"

var ProviderSet = wire.NewSet(
	NewPlayerHandler,
	NewPermissionHandler,
	NewWhitelistHandler,
	NewAuthHandler,
	NewSkinHandler,
	NewPunishmentHandler,
	NewServer,
)
