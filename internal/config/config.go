package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	KafkaBrokers      []string
	ConsumerGroupID   string
	LogLevel          string
	InputTopics       []string
	OutputTopicPrefix string

	// Consumer
	ConsumerOffsetReset        string // newest | oldest
	ConsumerBalanceStrategy    string // roundrobin | range | sticky
	ConsumerAutoCommit         bool
	ConsumerAutoCommitInterval time.Duration
	ConsumerSessionTimeout     time.Duration
	ConsumerHeartbeatInterval  time.Duration
	ConsumerRebalanceTimeout   time.Duration
	ConsumerFetchMin           int32
	ConsumerFetchDefault       int32
	ConsumerFetchMax           int32

	// Pipeline
	ChannelBufferSize int
	WorkerCount       int
	BatchEnabled      bool
	BatchSize         int
	BatchTimeout      time.Duration

	// Producer
	ProducerMode           string // sync | async
	ProducerAcks           string // none | local | all
	ProducerRetryMax       int
	ProducerRetryBackoff   time.Duration
	ProducerFlushMessages  int
	ProducerFlushBytes     int
	ProducerFlushFrequency time.Duration

	// Kafka version (e.g. "3.6.0")
	KafkaVersion string
}

func Load() Config {
	return Config{
		KafkaBrokers:      splitCSV(env("KAFKA_BROKERS", "")),
		ConsumerGroupID:   env("KAFKA_CONSUMER_GROUP", "nexus"),
		LogLevel:          env("LOG_LEVEL", "info"),
		InputTopics:       splitCSV(env("INPUT_TOPICS", "otel.traces,otel.metrics,otel.logs")),
		OutputTopicPrefix: env("OUTPUT_TOPIC_PREFIX", "otel.flat"),

		ConsumerOffsetReset:        env("CONSUMER_OFFSET_RESET", "newest"),
		ConsumerBalanceStrategy:    env("CONSUMER_BALANCE_STRATEGY", "roundrobin"),
		ConsumerAutoCommit:         envBool("CONSUMER_AUTO_COMMIT", true),
		ConsumerAutoCommitInterval: envDuration("CONSUMER_AUTO_COMMIT_INTERVAL", 1*time.Second),
		ConsumerSessionTimeout:     envDuration("CONSUMER_SESSION_TIMEOUT", 30*time.Second),
		ConsumerHeartbeatInterval:  envDuration("CONSUMER_HEARTBEAT_INTERVAL", 3*time.Second),
		ConsumerRebalanceTimeout:   envDuration("CONSUMER_REBALANCE_TIMEOUT", 60*time.Second),
		ConsumerFetchMin:           int32(envInt("CONSUMER_FETCH_MIN_BYTES", 1)),
		ConsumerFetchDefault:       int32(envInt("CONSUMER_FETCH_DEFAULT_BYTES", 1048576)), // 1 MiB
		ConsumerFetchMax:           int32(envInt("CONSUMER_FETCH_MAX_BYTES", 10485760)),    // 10 MiB

		ChannelBufferSize: envInt("CHANNEL_BUFFER_SIZE", 1000),
		WorkerCount:       envInt("WORKER_COUNT", 4),
		BatchEnabled:      envBool("BATCH_ENABLED", true),
		BatchSize:         envInt("BATCH_SIZE", 100),
		BatchTimeout:      envDuration("BATCH_TIMEOUT", 500*time.Millisecond),

		ProducerMode:           env("PRODUCER_MODE", "async"),
		ProducerAcks:           env("PRODUCER_ACKS", "local"),
		ProducerRetryMax:       envInt("PRODUCER_RETRY_MAX", 3),
		ProducerRetryBackoff:   envDuration("PRODUCER_RETRY_BACKOFF", 100*time.Millisecond),
		ProducerFlushMessages:  envInt("PRODUCER_FLUSH_MESSAGES", 1000),
		ProducerFlushBytes:     envInt("PRODUCER_FLUSH_BYTES", 1048576), // 1 MiB
		ProducerFlushFrequency: envDuration("PRODUCER_FLUSH_FREQUENCY", 2*time.Second),

		KafkaVersion: env("KAFKA_VERSION", "3.6.0"),
	}
}

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}

func envBool(key string, def bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return def
	}
	return b
}

func envInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

func envDuration(key string, def time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return def
	}
	return d
}
