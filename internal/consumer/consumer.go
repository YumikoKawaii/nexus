package consumer

import (
	"context"
	"fmt"

	"github.com/IBM/sarama"

	"github.com/yumikokawaii/nexus/internal/config"
)

type Group struct {
	cg      sarama.ConsumerGroup
	handler *Handler
	topics  []string
}

func NewGroup(cfg config.Config, handler *Handler) (*Group, error) {
	scfg, err := buildSaramaConfig(cfg)
	if err != nil {
		return nil, err
	}

	cg, err := sarama.NewConsumerGroup(cfg.KafkaBrokers, cfg.ConsumerGroupID, scfg)
	if err != nil {
		return nil, fmt.Errorf("sarama consumer group: %w", err)
	}
	return &Group{cg: cg, handler: handler, topics: cfg.InputTopics}, nil
}

func buildSaramaConfig(cfg config.Config) (*sarama.Config, error) {
	scfg := sarama.NewConfig()

	ver, err := sarama.ParseKafkaVersion(cfg.KafkaVersion)
	if err != nil {
		return nil, fmt.Errorf("invalid KAFKA_VERSION %q: %w", cfg.KafkaVersion, err)
	}
	scfg.Version = ver

	// balance strategy
	switch cfg.ConsumerBalanceStrategy {
	case "range":
		scfg.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRange()}
	case "sticky":
		scfg.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategySticky()}
	default:
		scfg.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRoundRobin()}
	}

	// offset reset
	if cfg.ConsumerOffsetReset == "oldest" {
		scfg.Consumer.Offsets.Initial = sarama.OffsetOldest
	} else {
		scfg.Consumer.Offsets.Initial = sarama.OffsetNewest
	}

	scfg.Consumer.Offsets.AutoCommit.Enable = cfg.ConsumerAutoCommit
	scfg.Consumer.Offsets.AutoCommit.Interval = cfg.ConsumerAutoCommitInterval

	scfg.Consumer.Group.Session.Timeout = cfg.ConsumerSessionTimeout
	scfg.Consumer.Group.Heartbeat.Interval = cfg.ConsumerHeartbeatInterval
	scfg.Consumer.Group.Rebalance.Timeout = cfg.ConsumerRebalanceTimeout

	scfg.Consumer.Fetch.Min = cfg.ConsumerFetchMin
	scfg.Consumer.Fetch.Default = cfg.ConsumerFetchDefault
	scfg.Consumer.Fetch.Max = cfg.ConsumerFetchMax

	return scfg, nil
}

func (g *Group) Run(ctx context.Context) error {
	for {
		if err := g.cg.Consume(ctx, g.topics, g.handler); err != nil {
			return fmt.Errorf("consume: %w", err)
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
	}
}

func (g *Group) Close() error {
	return g.cg.Close()
}
