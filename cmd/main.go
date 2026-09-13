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

	srv, err := receiver.NewServer(cfg.OTLP.Addr, cfg.OTLP.MaxRecvMsgSizeMiB, svc)
	if err != nil {
		logger.Error("otlp server init failed", "err", err)
		os.Exit(1)
	}
	if err := srv.Start(logger); err != nil {
		logger.Error("otlp server start failed", "err", err)
		os.Exit(1)
	}

	logger.Info("nexus started",
		"brokers", cfg.KafkaBrokers,
		"otlp_addr", cfg.OTLP.Addr,
		"producer_mode", cfg.Producer.Mode,
	)

	<-ctx.Done()
	logger.Info("shutting down")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	srv.Stop(shutdownCtx)
}
