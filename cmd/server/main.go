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

	"github.com/octopuswallet/octopuswallet/internal/api"
	"github.com/octopuswallet/octopuswallet/internal/chain"
	"github.com/octopuswallet/octopuswallet/internal/chain/bitcoin"
	"github.com/octopuswallet/octopuswallet/internal/chain/evm"
	"github.com/octopuswallet/octopuswallet/internal/chain/solana"
	"github.com/octopuswallet/octopuswallet/internal/chain/tron"
	"github.com/octopuswallet/octopuswallet/internal/config"
	"github.com/octopuswallet/octopuswallet/internal/models"
	"github.com/octopuswallet/octopuswallet/internal/store/postgres"
	"github.com/octopuswallet/octopuswallet/internal/wallet"
	"github.com/octopuswallet/octopuswallet/internal/webhook"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg, err := config.Load("")
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	seed, err := wallet.SeedFromMnemonic(cfg.Wallet.MasterSeed)
	if err != nil {
		slog.Error("failed to parse master seed", "error", err)
		os.Exit(1)
	}

	store, err := postgres.New(cfg.Database.URL, cfg.Database.MaxOpenConns)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer store.Close()

	registry := chain.NewRegistry()
	initChains(cfg, registry)

	// Validate admin config
	if cfg.Admin.JWTSecret == "" {
		slog.Error("FATAL: admin.jwt_secret is not configured. Set OCTOPUS_ADMIN_JWT_SECRET or admin.jwt_secret in config.")
		os.Exit(1)
	}
	if len(cfg.Admin.JWTSecret) < 32 {
		slog.Warn("admin.jwt_secret is shorter than 32 characters — consider using a stronger secret")
	}

	// Initialize admin: seed default admin user if none exists
	initAdminUser(store, cfg)

	webhookSvc := webhook.NewService(cfg.Webhook.Timeout, cfg.Webhook.MaxRetries, cfg.Webhook.RetryBackoff)
	hub := api.NewHub()
	router := api.NewRouter(store, registry, seed, webhookSvc, cfg, hub, store)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	go func() {
		slog.Info("starting server", "port", cfg.Server.Port, "chains", registry.Names())
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("server forced shutdown", "error", err)
	}
	slog.Info("server stopped")
}

func initChains(cfg *config.Config, registry *chain.Registry) {
	evmChains := map[string]string{"ethereum": "ETH", "bsc": "BNB", "polygon": "MATIC"}
	for name, symbol := range evmChains {
		chainCfg, ok := cfg.Chains[name]
		if !ok || !chainCfg.Enabled {
			continue
		}
		client, err := evm.NewClient(name, chainCfg.RPCURL, chainCfg.ChainID, symbol)
		if err != nil {
			slog.Warn("failed to initialize chain", "chain", name, "error", err)
			continue
		}
		registry.Register(client)
		slog.Info("chain initialized", "chain", name)
	}
	if solCfg, ok := cfg.Chains["solana"]; ok && solCfg.Enabled {
		if client, err := solana.NewClient(solCfg.RPCURL); err == nil {
			registry.Register(client)
			slog.Info("chain initialized", "chain", "solana")
		}
	}
	if tronCfg, ok := cfg.Chains["tron"]; ok && tronCfg.Enabled {
		if client, err := tron.NewClient(tronCfg.RPCURL, tronCfg.APIKey); err == nil {
			registry.Register(client)
			slog.Info("chain initialized", "chain", "tron")
		}
	}
	if btcCfg, ok := cfg.Chains["bitcoin"]; ok && btcCfg.Enabled {
		if client, err := bitcoin.NewClient(btcCfg.RPCURL, btcCfg.RPCUser, btcCfg.RPCPass, btcCfg.Network); err == nil {
			registry.Register(client)
			slog.Info("chain initialized", "chain", "bitcoin")
		}
	}
}

func initAdminUser(s *postgres.Store, cfg *config.Config) {
	count, err := s.CountAdminUsers(context.Background())
	if err != nil {
		slog.Warn("failed to count admin users", "error", err)
		return
	}
	if count > 0 {
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(cfg.Admin.DefaultPass), bcrypt.DefaultCost)
	if err != nil {
		slog.Error("failed to hash default admin password", "error", err)
		return
	}

	user := &models.AdminUser{
		Username: cfg.Admin.DefaultUser,
		Email:    cfg.Admin.DefaultUser + "@octopus.local",
		Password: string(hash),
		Role:     models.RoleSuperAdmin,
	}
	if err := s.CreateAdminUser(context.Background(), user); err != nil {
		slog.Error("failed to create default admin user", "error", err)
		return
	}
	slog.Info("default admin user created", "username", cfg.Admin.DefaultUser)
}
