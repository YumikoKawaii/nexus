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

	// Consume mode: "sync" or "async"
	ConsumeMode string

	// Batch config
	BatchEnabled  bool
	BatchSize     int
	BatchTimeout  time.Duration

	// Producer acks: "none", "local", "all"
	ProducerAcks string
}

func Load() Config {
	return Config{
		KafkaBrokers:      splitCSV(env("KAFKA_BROKERS", "")),
		ConsumerGroupID:   env("KAFKA_CONSUMER_GROUP", "nexus"),
		LogLevel:          env("LOG_LEVEL", "info"),
		InputTopics:       splitCSV(env("INPUT_TOPICS", "otel.traces,otel.metrics,otel.logs")),
		OutputTopicPrefix: env("OUTPUT_TOPIC_PREFIX", "otel.flat"),

		ConsumeMode: env("CONSUME_MODE", "async"), // sync | async

		BatchEnabled: envBool("BATCH_ENABLED", true),
		BatchSize:    envInt("BATCH_SIZE", 100),
		BatchTimeout: envDuration("BATCH_TIMEOUT", 500*time.Millisecond),

		ProducerAcks: env("PRODUCER_ACKS", "local"), // none | local | all
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
