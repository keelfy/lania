package services

import "github.com/google/wire"

var ProviderSet = wire.NewSet(
	NewPlayerService,
	NewPermissionService,
	NewWhitelistService,
	NewAuthService,
	NewSkinService,
	NewPunishmentService,
)
