package consumer

import (
	"context"
	"errors"
	"fmt"

	"github.com/twmb/franz-go/pkg/kgo"

	"github.com/yumikokawaii/nexus/internal/config"
	"github.com/yumikokawaii/nexus/internal/constants"
)

type Group struct {
	cl      *kgo.Client
	handler *Handler
}

func NewGroup(cfg config.Config, handler *Handler) (*Group, error) {
	opts, err := buildClientOptions(cfg)
	if err != nil {
		return nil, err
	}

	cl, err := kgo.NewClient(opts...)
	if err != nil {
		return nil, fmt.Errorf("franz-go consumer client: %w", err)
	}
	return &Group{cl: cl, handler: handler}, nil
}

func buildClientOptions(cfg config.Config) ([]kgo.Opt, error) {
	var balancer kgo.GroupBalancer
	switch cfg.ConsumerBalanceStrategy {
	case constants.BalanceStrategyRange:
		balancer = kgo.RangeBalancer()
	case constants.BalanceStrategySticky:
		balancer = kgo.StickyBalancer()
	default:
		balancer = kgo.RoundRobinBalancer()
	}

	offset := kgo.NewOffset().AtEnd()
	if cfg.ConsumerOffsetReset == constants.OffsetResetOldest {
		offset = kgo.NewOffset().AtStart()
	}

	opts := []kgo.Opt{
		kgo.SeedBrokers(cfg.KafkaBrokers...),
		kgo.ConsumerGroup(cfg.ConsumerGroupID),
		kgo.ConsumeTopics(cfg.InputTopics...),
		kgo.Balancers(balancer),
		kgo.ConsumeResetOffset(offset),
		kgo.SessionTimeout(cfg.ConsumerSessionTimeout),
		kgo.HeartbeatInterval(cfg.ConsumerHeartbeatInterval),
		kgo.RebalanceTimeout(cfg.ConsumerRebalanceTimeout),
		kgo.FetchMinBytes(cfg.ConsumerFetchMin),
		kgo.FetchMaxBytes(cfg.ConsumerFetchMax),
		kgo.DisableAutoCommit(),
	}

	if cfg.ConsumerAutoCommit {
		opts = append(opts,
			kgo.AutoCommitMarks(),
			kgo.AutoCommitInterval(cfg.ConsumerAutoCommitInterval),
		)
	}

	return opts, nil
}

func (g *Group) Run(ctx context.Context) error {
	for {
		fetches := g.cl.PollFetches(ctx)
		if fetches.IsClientClosed() {
			return nil
		}
		if err := ctx.Err(); err != nil {
			return err
		}

		var fetchErr error
		fetches.EachError(func(topic string, partition int32, err error) {
			if !errors.Is(err, context.Canceled) {
				fetchErr = fmt.Errorf("fetch %s[%d]: %w", topic, partition, err)
			}
		})
		if fetchErr != nil {
			return fetchErr
		}

		g.handler.Dispatch(ctx, g.cl, fetches)
	}
}

func (g *Group) Close() error {
	g.cl.Close()
	return nil
}
