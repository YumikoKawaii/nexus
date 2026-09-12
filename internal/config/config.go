package config

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/yumikokawaii/nexus/internal/constants"
)

type Config struct {
	KafkaBrokers      []string
	LogLevel          string
	OutputTopicPrefix string

	OTLPGRPCAddr string
	OTLPHTTPAddr string

	ProducerMode           string
	ProducerAcks           string
	ProducerRetryMax       int
	ProducerRetryBackoff   time.Duration
	ProducerFlushMessages  int
	ProducerFlushBytes     int
	ProducerFlushFrequency time.Duration
}

func Load() Config {
	return Config{
		KafkaBrokers:      splitCSV(env("KAFKA_BROKERS", "")),
		LogLevel:          env("LOG_LEVEL", constants.LogLevelInfo),
		OutputTopicPrefix: env("OUTPUT_TOPIC_PREFIX", "otel.flat"),

		OTLPGRPCAddr: env("OTLP_GRPC_ADDR", constants.DefaultOTLPGRPCAddr),
		OTLPHTTPAddr: env("OTLP_HTTP_ADDR", constants.DefaultOTLPHTTPAddr),

		ProducerMode:           env("PRODUCER_MODE", constants.ProducerModeAsync),
		ProducerAcks:           env("PRODUCER_ACKS", constants.ProducerAcksLocal),
		ProducerRetryMax:       envInt("PRODUCER_RETRY_MAX", constants.DefaultProducerRetryMax),
		ProducerRetryBackoff:   envDuration("PRODUCER_RETRY_BACKOFF", 100*time.Millisecond),
		ProducerFlushMessages:  envInt("PRODUCER_FLUSH_MESSAGES", constants.DefaultProducerFlushMessages),
		ProducerFlushBytes:     envInt("PRODUCER_FLUSH_BYTES", constants.DefaultProducerFlushBytes),
		ProducerFlushFrequency: envDuration("PRODUCER_FLUSH_FREQUENCY", 2*time.Second),
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
