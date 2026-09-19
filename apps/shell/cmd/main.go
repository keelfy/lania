package main

import (
	"context"
	"fmt"
	"net"
	"os/signal"
	"syscall"

	// autoload .env file
	_ "github.com/joho/godotenv/autoload"
	"github.com/lania-smp/shell/internal/config"
	"github.com/lania-smp/shell/internal/logger"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	logger.PrepareLogger()

	server, cleanup, err := InitializeServer(ctx)
	if err != nil {
		logger.Fatalf(ctx, "Failed to initialize server: %v", err)
	}
	defer cleanup()

	port := config.GetPort()
	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		logger.Fatalf(ctx, "Failed to listen on port %s: %v", port, err)
	}

	go func() {
		<-ctx.Done()
		server.GracefulStop()
	}()

	logger.Infof(ctx, "Starting the gRPC server on port %v", port)
	if err := server.Serve(listener); err != nil {
		logger.Fatalf(ctx, "Failed to serve: %v", err)
	}
	logger.Infof(ctx, "Server was shutdown gracefully")
}
