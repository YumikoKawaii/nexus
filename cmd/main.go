package main

import (
	"context"
	"flag"
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
	configPath := flag.String("config", "config.yaml", "path to config file")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		slog.Error("config load failed", "err", err)
		os.Exit(1)
	}

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

	grpcSrv := receiver.NewGRPCServer(cfg.OTLP.GRPCAddr, svc)
	if err := grpcSrv.Start(logger); err != nil {
		logger.Error("otlp grpc start failed", "err", err)
		os.Exit(1)
	}

	httpSrv := receiver.NewHTTPServer(cfg.OTLP.HTTPAddr, svc)
	if err := httpSrv.Start(logger); err != nil {
		logger.Error("otlp http start failed", "err", err)
		os.Exit(1)
	}

	logger.Info("nexus started",
		"brokers", cfg.KafkaBrokers,
		"otlp_grpc", cfg.OTLP.GRPCAddr,
		"otlp_http", cfg.OTLP.HTTPAddr,
		"producer_mode", cfg.Producer.Mode,
	)

	<-ctx.Done()
	logger.Info("shutting down")

	grpcSrv.Stop()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	httpSrv.Stop(shutdownCtx)
}
