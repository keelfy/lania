package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/lania-smp/backend/internal/admin"
	"github.com/lania-smp/backend/internal/clients"
)

func main() {
	if err := run(); err != nil {
		slog.Error("admin stopped", "error", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	origin, err := endpoint("ADMIN_ORIGIN")
	if err != nil {
		return err
	}
	parsed, _ := url.Parse(origin)
	if parsed.Path != "" {
		return fmt.Errorf("ADMIN_ORIGIN must not contain a path")
	}
	publicURL, err := endpoint("ORY_URL")
	if err != nil {
		return err
	}
	adminURL, err := endpoint("ORY_ADMIN_URL")
	if err != nil {
		return err
	}
	admins := map[string]bool{}
	for _, raw := range strings.Split(os.Getenv("ADMIN_IDENTITY_IDS"), ",") {
		if strings.TrimSpace(raw) == "" {
			continue
		}
		id, err := uuid.Parse(strings.TrimSpace(raw))
		if err != nil || id == uuid.Nil {
			return fmt.Errorf("ADMIN_IDENTITY_IDS must contain nonzero UUIDs")
		}
		admins[id.String()] = true
	}
	// An empty allowlist intentionally denies all users, allowing safe initial deployment.
	for _, name := range []string{"DATABASE_HOST", "DATABASE_USER", "DATABASE_NAME", "SHELL_ADDRESS", "SHELL_TOKEN"} {
		if os.Getenv(name) == "" {
			return fmt.Errorf("%s is required", name)
		}
	}
	season, err := uuid.Parse(os.Getenv("ACTIVE_SEASON_ID"))
	if err != nil || season == uuid.Nil {
		return fmt.Errorf("ACTIVE_SEASON_ID must be a nonzero UUID")
	}
	color, err := uuid.Parse(os.Getenv("DEFAULT_NAME_COLOR_ID"))
	if err != nil || color == uuid.Nil {
		return fmt.Errorf("DEFAULT_NAME_COLOR_ID must be a nonzero UUID")
	}
	cfg := mysql.NewConfig()
	cfg.Net, cfg.Addr = "tcp", net.JoinHostPort(os.Getenv("DATABASE_HOST"), env("DATABASE_PORT", "3306"))
	cfg.User, cfg.Passwd, cfg.DBName = os.Getenv("DATABASE_USER"), os.Getenv("DATABASE_PASSWORD"), os.Getenv("DATABASE_NAME")
	cfg.ParseTime, cfg.Timeout, cfg.ReadTimeout, cfg.WriteTimeout = true, 5*time.Second, 10*time.Second, 10*time.Second
	db, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		return err
	}
	defer db.Close()
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(2)
	db.SetConnMaxLifetime(3 * time.Minute)
	pingCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	err = db.PingContext(pingCtx)
	cancel()
	if err != nil {
		return err
	}
	shell, cleanup, err := clients.NewShellAPI(ctx)
	if err != nil {
		return err
	}
	defer cleanup()
	panel := &admin.Server{
		Store:    &admin.Store{DB: db, ActiveSeason: season, DefaultColor: color},
		Identity: admin.NewIdentityClient(publicURL, adminURL), Shell: shell, AdminIDs: admins, Origin: origin,
	}
	server := &http.Server{Addr: ":" + env("PORT", "8081"), Handler: panel.Handler(), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 20 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second, MaxHeaderBytes: 32 << 10}
	done := make(chan error, 1)
	go func() { done <- server.ListenAndServe() }()
	slog.Info("admin listening", "address", server.Addr, "administrators", len(admins))
	select {
	case err := <-done:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		return server.Shutdown(shutdownCtx)
	}
}

func endpoint(name string) (string, error) {
	value := strings.TrimRight(os.Getenv(name), "/")
	u, err := url.Parse(value)
	if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") || u.RawQuery != "" || u.Fragment != "" || u.User != nil {
		return "", fmt.Errorf("%s must be an HTTP(S) URL without credentials, query or fragment", name)
	}
	return value, nil
}
func env(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
