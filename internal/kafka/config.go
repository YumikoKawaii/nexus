// Package kafka holds Sarama config shared by the consumer and producer,
// so the common base (Kafka protocol version) is parsed in exactly one place.
package kafka

import (
	"fmt"

	"github.com/IBM/sarama"

	"github.com/yumikokawaii/nexus/internal/config"
)

// BaseConfig returns a Sarama config with the Kafka protocol version applied.
// Consumer- and producer-specific settings are layered on top by their packages.
func BaseConfig(cfg config.Config) (*sarama.Config, error) {
	scfg := sarama.NewConfig()

	ver, err := sarama.ParseKafkaVersion(cfg.KafkaVersion)
	if err != nil {
		return nil, fmt.Errorf("invalid KAFKA_VERSION %q: %w", cfg.KafkaVersion, err)
	}
	scfg.Version = ver

	return scfg, nil
}
