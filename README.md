# nexus

Kafka transformer that bridges raw OTLP JSON batches to flat StarRocks-ingestible records.

Consumes `otel.traces`, `otel.metrics`, `otel.logs` → explodes nested OTLP protobuf-JSON → produces flat rows to
`otel.flat.*` topics for StarRocks Routine Load.

## Stack

- Go 1.27
- [franz-go](https://github.com/twmb/franz-go) — Kafka consumer/producer

## Run

```bash
cp .env.example .env
go run ./cmd
```

## Environment

| Variable               | Description                                           |
|------------------------|-------------------------------------------------------|
| `KAFKA_BROKERS`        | Comma-separated broker list                           |
| `KAFKA_CONSUMER_GROUP` | Consumer group ID (default: `nexus`)                  |
| `LOG_LEVEL`            | `debug` / `info` / `warn` / `error` (default: `info`) |
