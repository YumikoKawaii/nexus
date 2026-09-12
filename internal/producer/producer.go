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

// Producer publishes flat records to Kafka. In sync mode Produce blocks until
// the broker acks; in async mode it enqueues and errors are logged and dropped
// (log+skip policy).
type Producer struct {
	cl     *kgo.Client
	async  bool
	logger *slog.Logger
}

// New builds a Producer from config.
func New(cfg config.Config, logger *slog.Logger) (*Producer, error) {
	cl, err := kgo.NewClient(clientOptions(cfg)...)
	if err != nil {
		return nil, fmt.Errorf("franz-go client: %w", err)
	}
	return &Producer{
		cl:     cl,
		async:  cfg.ProducerMode == constants.ProducerModeAsync,
		logger: logger,
	}, nil
}

func clientOptions(cfg config.Config) []kgo.Opt {
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

// Produce publishes one record. In async mode it never returns an error;
// delivery failures are logged by the completion callback.
func (p *Producer) Produce(ctx context.Context, topic, key string, value []byte) error {
	rec := &kgo.Record{Topic: topic, Key: []byte(key), Value: value}
	if p.async {
		p.cl.Produce(ctx, rec, func(r *kgo.Record, err error) {
			if err != nil {
				p.logger.Error("async producer error", "topic", r.Topic, "err", err)
			}
		})
		return nil
	}
	return p.cl.ProduceSync(ctx, rec).FirstErr()
}

// Close flushes buffered records and shuts the client down.
func (p *Producer) Close() error {
	p.cl.Close()
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
