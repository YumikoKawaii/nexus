package constants

// Output topic suffixes
const (
	FlatSuffixTraces              = "traces"
	FlatSuffixLogs                = "logs"
	FlatSuffixMetricsGauge        = "metrics.gauge"
	FlatSuffixMetricsSum          = "metrics.sum"
	FlatSuffixMetricsSummary      = "metrics.summary"
	FlatSuffixMetricsHistogram    = "metrics.histogram"
	FlatSuffixMetricsExpHistogram = "metrics.exponential_histogram"
)

// OTLP receiver defaults
const (
	DefaultOTLPGRPCAddr = ":4317"
	DefaultOTLPHTTPAddr = ":4318"
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
