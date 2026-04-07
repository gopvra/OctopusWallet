package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/octopuswallet/octopuswallet/internal/chain"
	"github.com/octopuswallet/octopuswallet/internal/chain/bitcoin"
	"github.com/octopuswallet/octopuswallet/internal/chain/evm"
	"github.com/octopuswallet/octopuswallet/internal/chain/solana"
	"github.com/octopuswallet/octopuswallet/internal/chain/tron"
	"github.com/octopuswallet/octopuswallet/internal/config"
	"github.com/octopuswallet/octopuswallet/internal/monitor"
	"github.com/octopuswallet/octopuswallet/internal/payout"
	"github.com/octopuswallet/octopuswallet/internal/store/postgres"
	"github.com/octopuswallet/octopuswallet/internal/wallet"
	"github.com/octopuswallet/octopuswallet/internal/webhook"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg, err := config.Load("")
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	store, err := postgres.New(cfg.Database.URL, cfg.Database.MaxOpenConns)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer store.Close()

	// Initialize chain registry
	registry := chain.NewRegistry()
	initChains(cfg, registry)

	// Initialize webhook service
	webhookSvc := webhook.NewService(cfg.Webhook.Timeout, cfg.Webhook.MaxRetries, cfg.Webhook.RetryBackoff)

	// Initialize master seed
	seed, err := wallet.SeedFromMnemonic(cfg.Wallet.MasterSeed)
	if err != nil {
		slog.Error("failed to parse master seed", "error", err)
		os.Exit(1)
	}

	// Initialize monitor service
	monitorSvc := monitor.NewService(store, registry, webhookSvc, cfg.Chains)

	// Initialize payout service
	payoutSvc := payout.NewService(store, registry, webhookSvc, seed, cfg.Chains)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		slog.Info("monitor started", "chains", registry.Names())
		monitorSvc.Start(ctx)
	}()

	go func() {
		slog.Info("payout service started")
		payoutSvc.Start(ctx)
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down worker...")
	cancel()
	slog.Info("worker stopped")
}

func initChains(cfg *config.Config, registry *chain.Registry) {
	evmChains := map[string]string{
		"ethereum": "ETH",
		"bsc":      "BNB",
		"polygon":  "MATIC",
	}
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
		client, err := solana.NewClient(solCfg.RPCURL)
		if err != nil {
			slog.Warn("failed to initialize solana", "error", err)
		} else {
			registry.Register(client)
			slog.Info("chain initialized", "chain", "solana")
		}
	}

	if tronCfg, ok := cfg.Chains["tron"]; ok && tronCfg.Enabled {
		client, err := tron.NewClient(tronCfg.RPCURL, tronCfg.APIKey)
		if err != nil {
			slog.Warn("failed to initialize tron", "error", err)
		} else {
			registry.Register(client)
			slog.Info("chain initialized", "chain", "tron")
		}
	}

	if btcCfg, ok := cfg.Chains["bitcoin"]; ok && btcCfg.Enabled {
		client, err := bitcoin.NewClient(btcCfg.RPCURL, btcCfg.RPCUser, btcCfg.RPCPass, btcCfg.Network)
		if err != nil {
			slog.Warn("failed to initialize bitcoin", "error", err)
		} else {
			registry.Register(client)
			slog.Info("chain initialized", "chain", "bitcoin")
		}
	}
}
