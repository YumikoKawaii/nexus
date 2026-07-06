package producer

import (
	"fmt"

	"github.com/IBM/sarama"
)

type Producer interface {
	Produce(ctx interface{ Done() <-chan struct{} }, topic, key string, value []byte) error
	Close() error
}

type syncProducer struct {
	p sarama.SyncProducer
}

func New(brokers []string, acks sarama.RequiredAcks) (Producer, error) {
	cfg := sarama.NewConfig()
	cfg.Producer.RequiredAcks = acks
	cfg.Producer.Return.Successes = true
	cfg.Producer.Compression = sarama.CompressionSnappy

	p, err := sarama.NewSyncProducer(brokers, cfg)
	if err != nil {
		return nil, fmt.Errorf("sarama sync producer: %w", err)
	}
	return &syncProducer{p: p}, nil
}

func (s *syncProducer) Produce(_ interface{ Done() <-chan struct{} }, topic, key string, value []byte) error {
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.ByteEncoder(value),
	}
	_, _, err := s.p.SendMessage(msg)
	return err
}

func (s *syncProducer) Close() error {
	return s.p.Close()
}

func AcksFromString(s string) sarama.RequiredAcks {
	switch s {
	case "none":
		return sarama.NoResponse
	case "all":
		return sarama.WaitForAll
	default:
		return sarama.WaitForLocal
	}
}
