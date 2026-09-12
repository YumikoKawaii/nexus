package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"sigs.k8s.io/yaml"

	"github.com/yumikokawaii/nexus/internal/constants"
)

type Duration time.Duration

func (d Duration) Unwrap() time.Duration { return time.Duration(d) }

func (d *Duration) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	v, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	*d = Duration(v)
	return nil
}

type Config struct {
	KafkaBrokers      []string `json:"kafkaBrokers"`
	LogLevel          string   `json:"logLevel"`
	OutputTopicPrefix string   `json:"outputTopicPrefix"`

	OTLP     OTLP     `json:"otlp"`
	Topics   Topics   `json:"topics"`
	Producer Producer `json:"producer"`
}

type OTLP struct {
	GRPCAddr          string `json:"grpcAddr"`
	HTTPAddr          string `json:"httpAddr"`
	MaxRecvMsgSizeMiB int    `json:"maxRecvMsgSizeMiB"`
}

type Topics struct {
	Enabled []string `json:"enabled"`
}

type Producer struct {
	Mode           string   `json:"mode"`
	Acks           string   `json:"acks"`
	RetryMax       int      `json:"retryMax"`
	RetryBackoff   Duration `json:"retryBackoff"`
	FlushMessages  int      `json:"flushMessages"`
	FlushBytes     int      `json:"flushBytes"`
	FlushFrequency Duration `json:"flushFrequency"`
}

func Load(path string) (Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config %s: %w", path, err)
	}
	cfg := defaults()
	if err := yaml.UnmarshalStrict(b, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse config %s: %w", path, err)
	}
	return cfg, nil
}

func defaults() Config {
	return Config{
		LogLevel:          constants.LogLevelInfo,
		OutputTopicPrefix: "otel.flat",
		OTLP: OTLP{
			GRPCAddr:          constants.DefaultOTLPGRPCAddr,
			HTTPAddr:          constants.DefaultOTLPHTTPAddr,
			MaxRecvMsgSizeMiB: constants.DefaultOTLPMaxRecvMsgSizeMiB,
		},
		Producer: Producer{
			Mode:           constants.ProducerModeAsync,
			Acks:           constants.ProducerAcksLocal,
			RetryMax:       constants.DefaultProducerRetryMax,
			RetryBackoff:   Duration(100 * time.Millisecond),
			FlushMessages:  constants.DefaultProducerFlushMessages,
			FlushBytes:     constants.DefaultProducerFlushBytes,
			FlushFrequency: Duration(2 * time.Second),
		},
	}
}
