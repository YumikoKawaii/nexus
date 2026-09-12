package producer

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"

	"github.com/yumikokawaii/nexus/internal/config"
	"github.com/yumikokawaii/nexus/internal/constants"
)

type Producer interface {
	Produce(ctx context.Context, topic, key string, value []byte) error
	Close() error
}

func New(cfg config.Config, logger *slog.Logger) (Producer, error) {
	opts := buildClientOptions(cfg)

	cl, err := kgo.NewClient(opts...)
	if err != nil {
		return nil, fmt.Errorf("franz-go client: %w", err)
	}

	if cfg.ProducerMode == constants.ProducerModeAsync {
		return &asyncProducer{cl: cl, logger: logger}, nil
	}
	return &syncProducer{cl: cl}, nil
}

func buildClientOptions(cfg config.Config) []kgo.Opt {
	opts := []kgo.Opt{
		kgo.SeedBrokers(cfg.KafkaBrokers...),
		kgo.RequiredAcks(acksFromString(cfg.ProducerAcks)),
		kgo.ProducerBatchCompression(kgo.SnappyCompression()),
		kgo.RecordRetries(cfg.ProducerRetryMax),
		kgo.RetryBackoffFn(func(int) time.Duration { return cfg.ProducerRetryBackoff }),
	}

	if cfg.ProducerAcks != constants.ProducerAcksAll {
		opts = append(opts, kgo.DisableIdempotentWrite())
	}
	if cfg.ProducerFlushMessages > 0 {
		opts = append(opts, kgo.MaxBufferedRecords(cfg.ProducerFlushMessages))
	}
	if cfg.ProducerFlushBytes > 0 {
		opts = append(opts, kgo.ProducerBatchMaxBytes(int32(cfg.ProducerFlushBytes)))
	}
	if cfg.ProducerFlushFrequency > 0 {
		opts = append(opts, kgo.ProducerLinger(cfg.ProducerFlushFrequency))
	}

	return opts
}

// syncProducer blocks until the broker acks each message.
type syncProducer struct {
	cl *kgo.Client
}

func (s *syncProducer) Produce(ctx context.Context, topic, key string, value []byte) error {
	rec := &kgo.Record{Topic: topic, Key: []byte(key), Value: value}
	return s.cl.ProduceSync(ctx, rec).FirstErr()
}

func (s *syncProducer) Close() error {
	s.cl.Close()
	return nil
}

// asyncProducer enqueues messages and flushes in the background.
// Errors are logged and dropped (log+skip policy).
type asyncProducer struct {
	cl     *kgo.Client
	logger *slog.Logger
}

func (a *asyncProducer) Produce(ctx context.Context, topic, key string, value []byte) error {
	rec := &kgo.Record{Topic: topic, Key: []byte(key), Value: value}
	a.cl.Produce(ctx, rec, func(r *kgo.Record, err error) {
		if err != nil {
			a.logger.Error("async producer error", "topic", r.Topic, "err", err)
		}
	})
	return nil
}

func (a *asyncProducer) Close() error {
	a.cl.Close()
	return nil
}

func acksFromString(s string) kgo.Acks {
	switch s {
	case constants.ProducerAcksNone:
		return kgo.NoAck()
	case constants.ProducerAcksAll:
		return kgo.AllISRAcks()
	default:
		return kgo.LeaderAck()
	}
}
