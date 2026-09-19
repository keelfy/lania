package rpc

import "github.com/google/wire"

var ProviderSet = wire.NewSet(
	NewPlayerHandler,
	NewPermissionHandler,
	NewWhitelistHandler,
	NewServer,
)
