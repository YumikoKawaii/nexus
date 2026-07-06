package consumer

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/IBM/sarama"

	"github.com/yumikokawaii/nexus/internal/config"
	"github.com/yumikokawaii/nexus/internal/producer"
	"github.com/yumikokawaii/nexus/internal/transform"
)

type incomingMsg struct {
	session sarama.ConsumerGroupSession
	msg     *sarama.ConsumerMessage
}

type Handler struct {
	cfg      config.Config
	producer producer.Producer
	logger   *slog.Logger
	ch       chan incomingMsg
}

func NewHandler(cfg config.Config, p producer.Producer, logger *slog.Logger) *Handler {
	return &Handler{
		cfg:      cfg,
		producer: p,
		logger:   logger,
		ch:       make(chan incomingMsg, cfg.ChannelBufferSize),
	}
}

// Start spawns worker goroutines that drain the channel. Call before consuming.
func (h *Handler) Start(ctx context.Context) {
	for i := 0; i < h.cfg.WorkerCount; i++ {
		go h.work(ctx)
	}
}

func (h *Handler) work(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case m, ok := <-h.ch:
			if !ok {
				return
			}
			h.process(ctx, m.session, m.msg)
		}
	}
}

func (h *Handler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (h *Handler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }

func (h *Handler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		h.ch <- incomingMsg{session: session, msg: msg}
	}
	return nil
}

func (h *Handler) process(ctx context.Context, session sarama.ConsumerGroupSession, msg *sarama.ConsumerMessage) {
	var err error
	switch msg.Topic {
	case "otel.traces":
		err = h.handleTraces(ctx, msg)
	case "otel.logs":
		err = h.handleLogs(ctx, msg)
	case "otel.metrics":
		err = h.handleMetrics(ctx, msg)
	}
	if err != nil {
		h.logger.Error("process failed, skipping", "topic", msg.Topic, "offset", msg.Offset, "err", err)
	}
	session.MarkMessage(msg, "")
}

func (h *Handler) handleTraces(ctx context.Context, msg *sarama.ConsumerMessage) error {
	rows, err := transform.Traces(msg.Value)
	if err != nil {
		return err
	}
	for _, row := range rows {
		b, _ := json.Marshal(row)
		if err := h.producer.Produce(ctx, h.cfg.OutputTopicPrefix+".traces", row.TraceId, b); err != nil {
			h.logger.Error("produce trace failed", "traceId", row.TraceId, "err", err)
		}
	}
	return nil
}

func (h *Handler) handleLogs(ctx context.Context, msg *sarama.ConsumerMessage) error {
	rows, err := transform.Logs(msg.Value)
	if err != nil {
		return err
	}
	for _, row := range rows {
		b, _ := json.Marshal(row)
		key := row.TraceId
		if key == "" {
			key = row.ServiceName
		}
		if err := h.producer.Produce(ctx, h.cfg.OutputTopicPrefix+".logs", key, b); err != nil {
			h.logger.Error("produce log failed", "err", err)
		}
	}
	return nil
}

func (h *Handler) handleMetrics(ctx context.Context, msg *sarama.ConsumerMessage) error {
	batch, err := transform.Metrics(msg.Value)
	if err != nil {
		return err
	}

	produce := func(suffix, key string, v any) {
		b, _ := json.Marshal(v)
		topic := h.cfg.OutputTopicPrefix + ".metrics." + suffix
		if err := h.producer.Produce(ctx, topic, key, b); err != nil {
			h.logger.Error("produce metric failed", "topic", topic, "err", err)
		}
	}

	for _, row := range batch.Gauges {
		produce("gauge", row.ServiceName+row.MetricName, row)
	}
	for _, row := range batch.Sums {
		produce("sum", row.ServiceName+row.MetricName, row)
	}
	for _, row := range batch.Summaries {
		produce("summary", row.ServiceName+row.MetricName, row)
	}
	for _, row := range batch.Histograms {
		produce("histogram", row.ServiceName+row.MetricName, row)
	}
	for _, row := range batch.ExponentialHistograms {
		produce("exponential_histogram", row.ServiceName+row.MetricName, row)
	}
	return nil
}
