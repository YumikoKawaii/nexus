package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/yumikokawaii/nexus/internal/config"
	"github.com/yumikokawaii/nexus/internal/consumer"
	"github.com/yumikokawaii/nexus/internal/producer"
)

func main() {
	cfg := config.Load()

	level := slog.LevelInfo
	switch cfg.LogLevel {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}))

	p, err := producer.New(cfg.KafkaBrokers, producer.AcksFromString(cfg.ProducerAcks))
	if err != nil {
		logger.Error("producer init failed", "err", err)
		os.Exit(1)
	}
	defer p.Close()

	handler := consumer.NewHandler(cfg, p, logger)

	group, err := consumer.NewGroup(cfg, handler)
	if err != nil {
		logger.Error("consumer group init failed", "err", err)
		os.Exit(1)
	}
	defer group.Close()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	logger.Info("nexus started",
		"brokers", cfg.KafkaBrokers,
		"group", cfg.ConsumerGroupID,
		"mode", cfg.ConsumeMode,
		"batch_enabled", cfg.BatchEnabled,
		"batch_size", cfg.BatchSize,
		"batch_timeout", cfg.BatchTimeout,
	)

	if err := group.Run(ctx); err != nil && err != context.Canceled {
		logger.Error("consumer group exited", "err", err)
		os.Exit(1)
	}
}
