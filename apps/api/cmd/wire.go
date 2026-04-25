//go:build wireinject
// +build wireinject

package main

import (
	"context"

	"github.com/google/wire"
	"github.com/lania-smp/backend/internal/api"
	"github.com/lania-smp/backend/internal/clients"
	"github.com/lania-smp/backend/internal/handlers"
	"github.com/lania-smp/backend/internal/services"
	"github.com/lania-smp/backend/internal/storage"
)

func InitializeAPI(ctx context.Context) (api.LaniaAPI, func(), error) {
	wire.Build(
		storage.NewMainStorage,
		storage.NewPlanStorage,
		storage.NewFlectoneStorage,
		storage.NewCacheStorage,
		clients.ProviderSet,
		services.ProviderSet,
		handlers.ProviderSet,
		api.NewLaniaAPI,
	)
	return nil, nil, nil
}
