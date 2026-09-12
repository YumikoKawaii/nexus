package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yumikokawaii/nexus/internal/config"
	"github.com/yumikokawaii/nexus/internal/constants"
	"github.com/yumikokawaii/nexus/internal/producer"
	"github.com/yumikokawaii/nexus/internal/receiver"
)

func main() {
	cfg := config.Load()

	level := slog.LevelInfo
	switch cfg.LogLevel {
	case constants.LogLevelDebug:
		level = slog.LevelDebug
	case constants.LogLevelWarn:
		level = slog.LevelWarn
	case constants.LogLevelError:
		level = slog.LevelError
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}))

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	p, err := producer.New(cfg, logger)
	if err != nil {
		logger.Error("producer init failed", "err", err)
		os.Exit(1)
	}
	defer p.Close()

	svc := receiver.NewService(cfg, p, logger)

	grpcSrv := receiver.NewGRPCServer(cfg.OTLPGRPCAddr, svc)
	if err := grpcSrv.Start(logger); err != nil {
		logger.Error("otlp grpc start failed", "err", err)
		os.Exit(1)
	}

	httpSrv := receiver.NewHTTPServer(cfg.OTLPHTTPAddr, svc)
	if err := httpSrv.Start(logger); err != nil {
		logger.Error("otlp http start failed", "err", err)
		os.Exit(1)
	}

	logger.Info("nexus started",
		"brokers", cfg.KafkaBrokers,
		"otlp_grpc", cfg.OTLPGRPCAddr,
		"otlp_http", cfg.OTLPHTTPAddr,
		"producer_mode", cfg.ProducerMode,
	)

	<-ctx.Done()
	logger.Info("shutting down")

	grpcSrv.Stop()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	httpSrv.Stop(shutdownCtx)
}
