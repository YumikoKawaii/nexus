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

type Producer struct {
	cl     *kgo.Client
	async  bool
	logger *slog.Logger
}

func New(cfg config.Config, logger *slog.Logger) (*Producer, error) {
	cl, err := kgo.NewClient(clientOptions(cfg)...)
	if err != nil {
		return nil, fmt.Errorf("franz-go client: %w", err)
	}
	return &Producer{
		cl:     cl,
		async:  cfg.Producer.Mode == constants.ProducerModeAsync,
		logger: logger,
	}, nil
}

func clientOptions(cfg config.Config) []kgo.Opt {
	p := cfg.Producer
	opts := []kgo.Opt{
		kgo.SeedBrokers(cfg.KafkaBrokers...),
		kgo.RequiredAcks(acksFromString(p.Acks)),
		kgo.ProducerBatchCompression(kgo.SnappyCompression()),
		kgo.RecordRetries(p.RetryMax),
		kgo.RetryBackoffFn(func(int) time.Duration { return p.RetryBackoff.Unwrap() }),
	}

	if p.Acks != constants.ProducerAcksAll {
		opts = append(opts, kgo.DisableIdempotentWrite())
	}
	if p.FlushMessages > 0 {
		opts = append(opts, kgo.MaxBufferedRecords(p.FlushMessages))
	}
	if p.FlushBytes > 0 {
		opts = append(opts, kgo.ProducerBatchMaxBytes(int32(p.FlushBytes)))
	}
	if p.FlushFrequency > 0 {
		opts = append(opts, kgo.ProducerLinger(p.FlushFrequency.Unwrap()))
	}

	return opts
}

func (p *Producer) Produce(ctx context.Context, topic, key string, value []byte) error {
	rec := &kgo.Record{Topic: topic, Key: []byte(key), Value: value}
	if p.async {
		p.cl.Produce(context.WithoutCancel(ctx), rec, func(r *kgo.Record, err error) {
			if err != nil {
				p.logger.Error("async producer error", "topic", r.Topic, "err", err)
			}
		})
		return nil
	}
	return p.cl.ProduceSync(ctx, rec).FirstErr()
}

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
