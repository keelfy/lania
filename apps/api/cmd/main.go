package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os/signal"
	"syscall"

	// autoload .env file
	_ "github.com/joho/godotenv/autoload"
	_ "github.com/lania-smp/backend/docs"
	"github.com/lania-smp/backend/internal/config"
	"github.com/lania-smp/backend/internal/logger"
	"github.com/ory/graceful"
)

// @title Lania API
// @version 1.0
// @description This is a Lania backend.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host api.lania.gg
// @BasePath /v1
func main() {
	// Create a context with cancellation
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM)
	defer stop()

	logger.PrepareLogger()

	// Initialize dependencies with Wire
	api, cleanup, err := InitializeAPI(ctx)
	if err != nil {
		logger.Fatalf(ctx, "Failed to initialize app: %v", err)
	}
	defer cleanup() // Ensure resources are cleaned up

	r, err := api.BuildAPI(ctx)
	if err != nil {
		logger.Fatalf(ctx, "Failed to build API: %v", err)
	}

	port := config.GetPort()

	server := graceful.WithDefaults(&http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: r,
		// Enable HTTP/2
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
			// Optimize HTTP/2 performance
			NextProtos: []string{"h2", "http/1.1"},
		},
	})

	logger.Infof(ctx, "Starting the server on port %v", port)
	if err := graceful.Graceful(server.ListenAndServe, server.Shutdown); err != nil {
		logger.Fatal(ctx, "Failed to gracefully shutdown")
	}
	logger.Infof(ctx, "Server was shutdown gracefully")
}
