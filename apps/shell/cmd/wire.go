//go:build wireinject
// +build wireinject

package main

import (
	"context"

	"github.com/google/wire"
	"github.com/lania-smp/shell/internal/clients"
	"github.com/lania-smp/shell/internal/services"
	"github.com/lania-smp/shell/internal/storage"
	"github.com/lania-smp/shell/internal/transport/rpc"
	"google.golang.org/grpc"
)

func InitializeServer(ctx context.Context) (*grpc.Server, func(), error) {
	wire.Build(
		clients.ProviderSet,
		storage.ProviderSet,
		services.ProviderSet,
		rpc.ProviderSet,
	)
	return nil, nil, nil
}
