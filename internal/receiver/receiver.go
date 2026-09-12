package receiver

import (
	"context"
	"encoding/json"
	"log/slog"

	tracepb "go.opentelemetry.io/proto/otlp/trace/v1"
	logspb "go.opentelemetry.io/proto/otlp/logs/v1"
	metricspb "go.opentelemetry.io/proto/otlp/metrics/v1"

	"github.com/yumikokawaii/nexus/internal/config"
	"github.com/yumikokawaii/nexus/internal/constants"
	"github.com/yumikokawaii/nexus/internal/producer"
	"github.com/yumikokawaii/nexus/internal/transform"
)

// Service holds the shared transform+produce pipeline used by both the gRPC
// and HTTP OTLP frontends.
type Service struct {
	cfg      config.Config
	producer producer.Producer
	logger   *slog.Logger
}

func NewService(cfg config.Config, p producer.Producer, logger *slog.Logger) *Service {
	return &Service{cfg: cfg, producer: p, logger: logger}
}

func (s *Service) topic(suffix string) string {
	return s.cfg.OutputTopicPrefix + "." + suffix
}

func (s *Service) HandleTraces(ctx context.Context, rss []*tracepb.ResourceSpans) {
	rows, err := transform.Traces(&tracepb.TracesData{ResourceSpans: rss})
	if err != nil {
		s.logger.Error("transform traces failed", "err", err)
		return
	}
	for _, row := range rows {
		b, _ := json.Marshal(row)
		if err := s.producer.Produce(ctx, s.topic(constants.FlatSuffixTraces), row.TraceId, b); err != nil {
			s.logger.Error("produce trace failed", "traceId", row.TraceId, "err", err)
		}
	}
}

func (s *Service) HandleLogs(ctx context.Context, rls []*logspb.ResourceLogs) {
	rows, err := transform.Logs(&logspb.LogsData{ResourceLogs: rls})
	if err != nil {
		s.logger.Error("transform logs failed", "err", err)
		return
	}
	for _, row := range rows {
		b, _ := json.Marshal(row)
		key := row.TraceId
		if key == "" {
			key = row.ServiceName
		}
		if err := s.producer.Produce(ctx, s.topic(constants.FlatSuffixLogs), key, b); err != nil {
			s.logger.Error("produce log failed", "err", err)
		}
	}
}

func (s *Service) HandleMetrics(ctx context.Context, rms []*metricspb.ResourceMetrics) {
	batch, err := transform.Metrics(&metricspb.MetricsData{ResourceMetrics: rms})
	if err != nil {
		s.logger.Error("transform metrics failed", "err", err)
		return
	}

	produce := func(suffix, key string, v any) {
		b, _ := json.Marshal(v)
		topic := s.topic(suffix)
		if err := s.producer.Produce(ctx, topic, key, b); err != nil {
			s.logger.Error("produce metric failed", "topic", topic, "err", err)
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
}
