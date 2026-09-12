package consumer

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"

	"github.com/yumikokawaii/nexus/internal/config"
	"github.com/yumikokawaii/nexus/internal/constants"
	"github.com/yumikokawaii/nexus/internal/producer"
	"github.com/yumikokawaii/nexus/internal/transform"
)

type incomingMsg struct {
	rec *kgo.Record
}

type Handler struct {
	cfg      config.Config
	producer producer.Producer
	logger   *slog.Logger
	ch       chan []incomingMsg
	cl       *kgo.Client
}

func NewHandler(cfg config.Config, p producer.Producer, logger *slog.Logger) *Handler {
	return &Handler{
		cfg:      cfg,
		producer: p,
		logger:   logger,
		ch:       make(chan []incomingMsg, cfg.ChannelBufferSize),
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
		case batch, ok := <-h.ch:
			if !ok {
				return
			}
			for _, m := range batch {
				h.process(ctx, m.rec)
			}
		}
	}
}

// Dispatch fans records from a poll into the worker channel, batching per the
// configured batch size / timeout. It blocks until every record is enqueued.
func (h *Handler) Dispatch(ctx context.Context, cl *kgo.Client, fetches kgo.Fetches) {
	h.cl = cl

	if !h.cfg.BatchEnabled {
		fetches.EachRecord(func(rec *kgo.Record) {
			h.enqueue(ctx, []incomingMsg{{rec: rec}})
		})
		return
	}

	batch := make([]incomingMsg, 0, h.cfg.BatchSize)
	deadline := time.Now().Add(h.cfg.BatchTimeout)

	flush := func() {
		if len(batch) == 0 {
			return
		}
		h.enqueue(ctx, batch)
		batch = make([]incomingMsg, 0, h.cfg.BatchSize)
		deadline = time.Now().Add(h.cfg.BatchTimeout)
	}

	fetches.EachRecord(func(rec *kgo.Record) {
		batch = append(batch, incomingMsg{rec: rec})
		if len(batch) >= h.cfg.BatchSize || time.Now().After(deadline) {
			flush()
		}
	})
	flush()
}

func (h *Handler) enqueue(ctx context.Context, batch []incomingMsg) {
	select {
	case <-ctx.Done():
	case h.ch <- batch:
	}
}

func (h *Handler) process(ctx context.Context, rec *kgo.Record) {
	var err error
	switch rec.Topic {
	case constants.TopicTraces:
		err = h.handleTraces(ctx, rec)
	case constants.TopicLogs:
		err = h.handleLogs(ctx, rec)
	case constants.TopicMetrics:
		err = h.handleMetrics(ctx, rec)
	}
	if err != nil {
		h.logger.Error("process failed, skipping", "topic", rec.Topic, "offset", rec.Offset, "err", err)
	}
	if h.cl != nil {
		h.cl.MarkCommitRecords(rec)
	}
}

func (h *Handler) handleTraces(ctx context.Context, rec *kgo.Record) error {
	rows, err := transform.Traces(rec.Value)
	if err != nil {
		return err
	}
	for _, row := range rows {
		b, _ := json.Marshal(row)
		if err := h.producer.Produce(ctx, h.cfg.OutputTopicPrefix+"."+constants.FlatSuffixTraces, row.TraceId, b); err != nil {
			h.logger.Error("produce trace failed", "traceId", row.TraceId, "err", err)
		}
	}
	return nil
}

func (h *Handler) handleLogs(ctx context.Context, rec *kgo.Record) error {
	rows, err := transform.Logs(rec.Value)
	if err != nil {
		return err
	}
	for _, row := range rows {
		b, _ := json.Marshal(row)
		key := row.TraceId
		if key == "" {
			key = row.ServiceName
		}
		if err := h.producer.Produce(ctx, h.cfg.OutputTopicPrefix+"."+constants.FlatSuffixLogs, key, b); err != nil {
			h.logger.Error("produce log failed", "err", err)
		}
	}
	return nil
}

func (h *Handler) handleMetrics(ctx context.Context, rec *kgo.Record) error {
	batch, err := transform.Metrics(rec.Value)
	if err != nil {
		return err
	}

	produce := func(suffix, key string, v any) {
		b, _ := json.Marshal(v)
		topic := h.cfg.OutputTopicPrefix + "." + suffix
		if err := h.producer.Produce(ctx, topic, key, b); err != nil {
			h.logger.Error("produce metric failed", "topic", topic, "err", err)
		}
	}

	for _, row := range batch.Gauges {
		produce(constants.FlatSuffixMetricsGauge, row.ServiceName+row.MetricName, row)
	}
	for _, row := range batch.Sums {
		produce(constants.FlatSuffixMetricsSum, row.ServiceName+row.MetricName, row)
	}
	for _, row := range batch.Summaries {
		produce(constants.FlatSuffixMetricsSummary, row.ServiceName+row.MetricName, row)
	}
	for _, row := range batch.Histograms {
		produce(constants.FlatSuffixMetricsHistogram, row.ServiceName+row.MetricName, row)
	}
	for _, row := range batch.ExponentialHistograms {
		produce(constants.FlatSuffixMetricsExpHistogram, row.ServiceName+row.MetricName, row)
	}
	return nil
}
