package constants

const (
	FlatSuffixTraces              = "traces"
	FlatSuffixLogs                = "logs"
	FlatSuffixMetricsGauge        = "metrics.gauge"
	FlatSuffixMetricsSum          = "metrics.sum"
	FlatSuffixMetricsSummary      = "metrics.summary"
	FlatSuffixMetricsHistogram    = "metrics.histogram"
	FlatSuffixMetricsExpHistogram = "metrics.exponential_histogram"
)

const (
	DefaultOTLPGRPCAddr          = ":4317"
	DefaultOTLPHTTPAddr          = ":4318"
	DefaultOTLPMaxRecvMsgSizeMiB = 4
)

const (
	ProducerModeSync  = "sync"
	ProducerModeAsync = "async"
)

const (
	ProducerAcksNone  = "none"
	ProducerAcksLocal = "local"
	ProducerAcksAll   = "all"
)

const (
	DefaultProducerRetryMax      = 3
	DefaultProducerFlushMessages = 1000
	DefaultProducerFlushBytes    = 1048576
)

const (
	LogLevelDebug = "debug"
	LogLevelInfo  = "info"
	LogLevelWarn  = "warn"
	LogLevelError = "error"
)
