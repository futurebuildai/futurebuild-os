package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"

	"github.com/futurebuild/futurebuild-brain/internal/config"
	fboidc "github.com/futurebuild/futurebuild-brain/internal/oidc"
	"github.com/futurebuild/futurebuild-brain/internal/store"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	if err := run(logger); err != nil {
		logger.Error("server exited with error", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	pool, err := store.NewPool(ctx, store.PoolConfig{
		DatabaseURL:    cfg.DatabaseURL,
		MaxConns:       cfg.DBPoolMax,
		MinConns:       cfg.DBPoolMin,
		ConnectTimeout: cfg.DBTimeout,
	})
	if err != nil {
		return fmt.Errorf("connecting to database: %w", err)
	}
	defer pool.Close()

	logger.Info("database connected", "max_conns", cfg.DBPoolMax)

	// Initialize OIDC storage (PostgreSQL-backed)
	oidcStorage := fboidc.NewStorage(pool, logger)

	// Ensure a signing key exists (generates RSA key pair for dev if needed)
	if err := oidcStorage.EnsureSigningKey(ctx); err != nil {
		return fmt.Errorf("ensuring signing key: %w", err)
	}

	// Derive crypto key for OIDC token encryption
	cryptoKey := fboidc.CryptoKeyFromHex(cfg.CryptoKeyHex)

	// Create OIDC provider router
	oidcRouter, err := fboidc.SetupOIDCProvider(cfg.Issuer, oidcStorage, cryptoKey, logger, cfg.DevMode)
	if err != nil {
		return fmt.Errorf("setting up OIDC provider: %w", err)
	}

	// Build the main Chi router
	r := chi.NewRouter()

	// Global middleware
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(chimw.Recoverer)
	r.Use(chimw.Logger)

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		if err := pool.Ping(r.Context()); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"status":"unhealthy","error":"database unreachable"}`))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","version":"0.1.0","system":"fb-brain"}`))
	})

	// Mount OIDC provider at root (serves /.well-known, /authorize, /oauth/token, /keys, /userinfo, /revoke, /login, /consent)
	r.Mount("/", oidcRouter)

	// TODO: Hub Admin API routes (Sprint 1)
	// TODO: MCP Registry routes (Sprint 2)
	// TODO: Maestro routes (Sprint 3)

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		logger.Info("server starting", "port", cfg.Port, "issuer", cfg.Issuer)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		logger.Info("shutdown signal received")
	case err := <-errCh:
		return fmt.Errorf("server error: %w", err)
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server shutdown: %w", err)
	}

	logger.Info("server stopped gracefully")
	return nil
}
