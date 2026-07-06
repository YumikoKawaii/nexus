package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/yumikokawaii/nexus/internal/config"
	"github.com/yumikokawaii/nexus/internal/constants"
	"github.com/yumikokawaii/nexus/internal/consumer"
	"github.com/yumikokawaii/nexus/internal/producer"
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

	h := consumer.NewHandler(cfg, p, logger)
	h.Start(ctx)

	group, err := consumer.NewGroup(cfg, h)
	if err != nil {
		logger.Error("consumer group init failed", "err", err)
		os.Exit(1)
	}
	defer group.Close()

	logger.Info("nexus started",
		"brokers", cfg.KafkaBrokers,
		"group", cfg.ConsumerGroupID,
		"workers", cfg.WorkerCount,
		"channel_buffer", cfg.ChannelBufferSize,
		"producer_mode", cfg.ProducerMode,
	)

	if err := group.Run(ctx); err != nil && err != context.Canceled {
		logger.Error("consumer group exited", "err", err)
		os.Exit(1)
	}
}
