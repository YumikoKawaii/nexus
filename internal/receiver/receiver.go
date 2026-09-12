package receiver

import (
	"context"
	"encoding/json"
	"log/slog"

	logspb "go.opentelemetry.io/proto/otlp/logs/v1"
	metricspb "go.opentelemetry.io/proto/otlp/metrics/v1"
	tracepb "go.opentelemetry.io/proto/otlp/trace/v1"

	"github.com/yumikokawaii/nexus/internal/config"
	"github.com/yumikokawaii/nexus/internal/constants"
	"github.com/yumikokawaii/nexus/internal/transform"
)

type Producer interface {
	Produce(ctx context.Context, topic, key string, value []byte) error
}
type Service struct {
	cfg      config.Config
	producer Producer
	logger   *slog.Logger
	enabled  map[string]bool
}

func NewService(cfg config.Config, p Producer, logger *slog.Logger) *Service {
	return &Service{
		cfg:      cfg,
		producer: p,
		logger:   logger,
		enabled:  resolveEnabled(cfg.Topics.Enabled),
	}
}

func resolveEnabled(allow []string) map[string]bool {
	all := []string{
		constants.FlatSuffixTraces,
		constants.FlatSuffixLogs,
		constants.FlatSuffixMetricsGauge,
		constants.FlatSuffixMetricsSum,
		constants.FlatSuffixMetricsSummary,
		constants.FlatSuffixMetricsHistogram,
		constants.FlatSuffixMetricsExpHistogram,
	}
	if len(allow) == 0 {
		enabled := make(map[string]bool, len(all))
		for _, s := range all {
			enabled[s] = true
		}
		return enabled
	}
	allowSet := make(map[string]bool, len(allow))
	for _, s := range allow {
		allowSet[s] = true
	}
	enabled := map[string]bool{}
	for _, s := range all {
		if allowSet[s] {
			enabled[s] = true
		}
	}
	return enabled
}

func (s *Service) topic(suffix string) string {
	return s.cfg.OutputTopicPrefix + "." + suffix
}

func (s *Service) isEnabled(suffix string) bool {
	return s.enabled[suffix]
}

func (s *Service) HandleTraces(ctx context.Context, rss []*tracepb.ResourceSpans) {
	if !s.isEnabled(constants.FlatSuffixTraces) {
		return
	}
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
	if !s.isEnabled(constants.FlatSuffixLogs) {
		return
	}
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
	gauge := s.isEnabled(constants.FlatSuffixMetricsGauge)
	sum := s.isEnabled(constants.FlatSuffixMetricsSum)
	summary := s.isEnabled(constants.FlatSuffixMetricsSummary)
	histogram := s.isEnabled(constants.FlatSuffixMetricsHistogram)
	expHistogram := s.isEnabled(constants.FlatSuffixMetricsExpHistogram)
	if !gauge && !sum && !summary && !histogram && !expHistogram {
		return
	}

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

	if gauge {
		for _, row := range batch.Gauges {
			produce(constants.FlatSuffixMetricsGauge, row.ServiceName+row.MetricName, row)
		}
	}
	if sum {
		for _, row := range batch.Sums {
			produce(constants.FlatSuffixMetricsSum, row.ServiceName+row.MetricName, row)
		}
	}
	if summary {
		for _, row := range batch.Summaries {
			produce(constants.FlatSuffixMetricsSummary, row.ServiceName+row.MetricName, row)
		}
	}
	if histogram {
		for _, row := range batch.Histograms {
			produce(constants.FlatSuffixMetricsHistogram, row.ServiceName+row.MetricName, row)
		}
	}
	if expHistogram {
		for _, row := range batch.ExponentialHistograms {
			produce(constants.FlatSuffixMetricsExpHistogram, row.ServiceName+row.MetricName, row)
		}
	}
}
