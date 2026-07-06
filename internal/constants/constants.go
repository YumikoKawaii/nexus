package constants

// Input topics
const (
	TopicTraces  = "otel.traces"
	TopicMetrics = "otel.metrics"
	TopicLogs    = "otel.logs"
)

// Output topic suffixes
const (
	FlatSuffixTraces                = "traces"
	FlatSuffixLogs                  = "logs"
	FlatSuffixMetricsGauge          = "metrics.gauge"
	FlatSuffixMetricsSum            = "metrics.sum"
	FlatSuffixMetricsSummary        = "metrics.summary"
	FlatSuffixMetricsHistogram      = "metrics.histogram"
	FlatSuffixMetricsExpHistogram   = "metrics.exponential_histogram"
)

// Producer mode
const (
	ProducerModeSync  = "sync"
	ProducerModeAsync = "async"
)

// Producer acks
const (
	ProducerAcksNone  = "none"
	ProducerAcksLocal = "local"
	ProducerAcksAll   = "all"
)

// Consumer offset reset
const (
	OffsetResetNewest = "newest"
	OffsetResetOldest = "oldest"
)

// Consumer balance strategy
const (
	BalanceStrategyRoundRobin = "roundrobin"
	BalanceStrategyRange      = "range"
	BalanceStrategySticky     = "sticky"
)

// Kafka version
const DefaultKafkaVersion = "3.6.0"

// Default consumer config
const (
	DefaultConsumerFetchMinBytes     = 1
	DefaultConsumerFetchDefaultBytes = 1048576  // 1 MiB
	DefaultConsumerFetchMaxBytes     = 10485760 // 10 MiB
)

// Default pipeline config
const (
	DefaultChannelBufferSize = 1000
	DefaultWorkerCount       = 4
	DefaultBatchSize         = 100
)

// Default producer config
const (
	DefaultProducerRetryMax      = 3
	DefaultProducerFlushMessages = 1000
	DefaultProducerFlushBytes    = 1048576 // 1 MiB
)

// Log levels
const (
	LogLevelDebug = "debug"
	LogLevelInfo  = "info"
	LogLevelWarn  = "warn"
	LogLevelError = "error"
)
