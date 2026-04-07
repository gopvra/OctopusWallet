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
	"github.com/octopuswallet/octopuswallet/internal/coldwallet"
	"github.com/octopuswallet/octopuswallet/internal/config"
	"github.com/octopuswallet/octopuswallet/internal/gasstation"
	"github.com/octopuswallet/octopuswallet/internal/monitor"
	"github.com/octopuswallet/octopuswallet/internal/payout"
	"github.com/octopuswallet/octopuswallet/internal/store/postgres"
	"github.com/octopuswallet/octopuswallet/internal/sweep"
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

	webhookSvc := webhook.NewService(cfg.Webhook.Timeout, cfg.Webhook.MaxRetries, cfg.Webhook.RetryBackoff)

	// Gas Station service
	gasSvc := gasstation.NewService(store, registry, webhookSvc, cfg.GasStation)

	// Monitor service
	monitorSvc := monitor.NewService(store, registry, webhookSvc, cfg.Chains)

	// Sweep service (depends on gas station)
	sweepSvc := sweep.NewService(store, registry, webhookSvc, gasSvc, seed, cfg.Chains)

	// Wire sweep into monitor: trigger auto-sweep when payment completes
	monitorSvc.SetOnPaymentCompleted(sweepSvc.OnPaymentCompleted)

	// Payout service (with approval — only processes approved/auto-released payouts)
	payoutSvc := payout.NewService(store, registry, webhookSvc, seed, cfg.Chains)

	// Cold wallet service
	coldSvc := coldwallet.NewService(store, registry, webhookSvc, seed, cfg.Chains)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go monitorSvc.Start(ctx)
	go payoutSvc.Start(ctx)
	go sweepSvc.Start(ctx)
	go gasSvc.Start(ctx)
	go coldSvc.Start(ctx)

	slog.Info("worker started", "chains", registry.Names())

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down worker...")
	cancel()
	slog.Info("worker stopped")
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
	}
	if solCfg, ok := cfg.Chains["solana"]; ok && solCfg.Enabled {
		if client, err := solana.NewClient(solCfg.RPCURL); err == nil {
			registry.Register(client)
		}
	}
	if tronCfg, ok := cfg.Chains["tron"]; ok && tronCfg.Enabled {
		if client, err := tron.NewClient(tronCfg.RPCURL, tronCfg.APIKey); err == nil {
			registry.Register(client)
		}
	}
	if btcCfg, ok := cfg.Chains["bitcoin"]; ok && btcCfg.Enabled {
		if client, err := bitcoin.NewClient(btcCfg.RPCURL, btcCfg.RPCUser, btcCfg.RPCPass, btcCfg.Network); err == nil {
			registry.Register(client)
		}
	}
}
