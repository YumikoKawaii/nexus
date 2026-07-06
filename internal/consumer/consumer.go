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
	scfg := sarama.NewConfig()
	scfg.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRoundRobin()}
	scfg.Consumer.Offsets.Initial = sarama.OffsetNewest

	cg, err := sarama.NewConsumerGroup(cfg.KafkaBrokers, cfg.ConsumerGroupID, scfg)
	if err != nil {
		return nil, fmt.Errorf("sarama consumer group: %w", err)
	}
	return &Group{cg: cg, handler: handler, topics: cfg.InputTopics}, nil
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
