package consumer

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"
	"time"

	"github.com/IBM/sarama"

	"github.com/yumikokawaii/nexus/internal/config"
	"github.com/yumikokawaii/nexus/internal/producer"
	"github.com/yumikokawaii/nexus/internal/transform"
)

type Handler struct {
	cfg      config.Config
	producer producer.Producer
	logger   *slog.Logger

	// batch state — guarded by mu, used only when cfg.BatchEnabled
	mu      sync.Mutex
	batch   []*sarama.ConsumerMessage
	timer   *time.Timer
	session sarama.ConsumerGroupSession
}

func NewHandler(cfg config.Config, p producer.Producer, logger *slog.Logger) *Handler {
	return &Handler{
		cfg:      cfg,
		producer: p,
		logger:   logger,
	}
}

func (h *Handler) Setup(session sarama.ConsumerGroupSession) error {
	if h.cfg.BatchEnabled {
		h.mu.Lock()
		h.batch = make([]*sarama.ConsumerMessage, 0, h.cfg.BatchSize)
		h.session = session
		h.timer = time.AfterFunc(h.cfg.BatchTimeout, h.timerFlush)
		h.mu.Unlock()
	}
	return nil
}

func (h *Handler) Cleanup(_ sarama.ConsumerGroupSession) error {
	if h.cfg.BatchEnabled {
		h.mu.Lock()
		if h.timer != nil {
			h.timer.Stop()
		}
		h.flush(h.session)
		h.mu.Unlock()
	}
	return nil
}

func (h *Handler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		if h.cfg.BatchEnabled {
			h.mu.Lock()
			h.batch = append(h.batch, msg)
			if len(h.batch) >= h.cfg.BatchSize {
				h.flush(session)
			}
			h.mu.Unlock()
		} else {
			h.process(context.Background(), session, msg)
		}
	}
	return nil
}

func (h *Handler) timerFlush() {
	h.mu.Lock()
	h.flush(h.session)
	if h.timer != nil {
		h.timer.Reset(h.cfg.BatchTimeout)
	}
	h.mu.Unlock()
}

// flush drains the current batch. Caller must hold h.mu.
func (h *Handler) flush(session sarama.ConsumerGroupSession) {
	if len(h.batch) == 0 {
		return
	}
	msgs := h.batch
	h.batch = make([]*sarama.ConsumerMessage, 0, h.cfg.BatchSize)

	if h.cfg.ConsumeMode == "async" {
		go func() {
			for _, m := range msgs {
				h.process(context.Background(), session, m)
			}
		}()
	} else {
		for _, m := range msgs {
			h.process(context.Background(), session, m)
		}
	}
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
