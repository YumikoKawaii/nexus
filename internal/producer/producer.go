package producer

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/IBM/sarama"

	"github.com/yumikokawaii/nexus/internal/config"
	"github.com/yumikokawaii/nexus/internal/constants"
	"github.com/yumikokawaii/nexus/internal/kafka"
)

type Producer interface {
	Produce(ctx context.Context, topic, key string, value []byte) error
	Close() error
}

func New(cfg config.Config, logger *slog.Logger) (Producer, error) {
	scfg, err := buildProducerConfig(cfg)
	if err != nil {
		return nil, err
	}

	if cfg.ProducerMode == constants.ProducerModeAsync {
		scfg.Producer.Return.Successes = false
		scfg.Producer.Return.Errors = true
		p, err := sarama.NewAsyncProducer(cfg.KafkaBrokers, scfg)
		if err != nil {
			return nil, fmt.Errorf("sarama async producer: %w", err)
		}
		ap := &asyncProducer{p: p}
		go ap.drainErrors(logger)
		return ap, nil
	}

	scfg.Producer.Return.Successes = true
	p, err := sarama.NewSyncProducer(cfg.KafkaBrokers, scfg)
	if err != nil {
		return nil, fmt.Errorf("sarama sync producer: %w", err)
	}
	return &syncProducer{p: p}, nil
}

func buildProducerConfig(cfg config.Config) (*sarama.Config, error) {
	scfg, err := kafka.BaseConfig(cfg)
	if err != nil {
		return nil, err
	}

	scfg.Producer.RequiredAcks = acksFromString(cfg.ProducerAcks)
	scfg.Producer.Compression = sarama.CompressionSnappy

	scfg.Producer.Retry.Max = cfg.ProducerRetryMax
	scfg.Producer.Retry.Backoff = cfg.ProducerRetryBackoff

	if cfg.ProducerFlushMessages > 0 {
		scfg.Producer.Flush.Messages = cfg.ProducerFlushMessages
	}
	if cfg.ProducerFlushBytes > 0 {
		scfg.Producer.Flush.Bytes = cfg.ProducerFlushBytes
	}
	if cfg.ProducerFlushFrequency > 0 {
		scfg.Producer.Flush.Frequency = cfg.ProducerFlushFrequency
	}

	return scfg, nil
}

// syncProducer blocks until the broker acks each message.
type syncProducer struct {
	p sarama.SyncProducer
}

func (s *syncProducer) Produce(ctx context.Context, topic, key string, value []byte) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	_, _, err := s.p.SendMessage(&sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.ByteEncoder(value),
	})
	return err
}

func (s *syncProducer) Close() error { return s.p.Close() }

// asyncProducer enqueues messages and flushes in the background.
// Errors are logged and dropped (log+skip policy).
type asyncProducer struct {
	p sarama.AsyncProducer
}

func (a *asyncProducer) Produce(ctx context.Context, topic, key string, value []byte) error {
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.ByteEncoder(value),
	}
	select {
	case a.p.Input() <- msg:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (a *asyncProducer) Close() error {
	a.p.AsyncClose()
	return nil
}

func (a *asyncProducer) drainErrors(logger *slog.Logger) {
	for err := range a.p.Errors() {
		logger.Error("async producer error", "topic", err.Msg.Topic, "err", err.Err)
	}
}

func acksFromString(s string) sarama.RequiredAcks {
	switch s {
	case constants.ProducerAcksNone:
		return sarama.NoResponse
	case constants.ProducerAcksAll:
		return sarama.WaitForAll
	default:
		return sarama.WaitForLocal
	}
}
