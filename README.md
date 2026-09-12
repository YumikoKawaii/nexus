# nexus

OTLP endpoint that transforms OTLP telemetry into flat StarRocks-ingestible records.

Accepts OTLP over gRPC (`:4317`) and HTTP (`:4318`) → explodes nested OTLP into flat rows →
produces to `otel.flat.*` Kafka topics for StarRocks Routine Load.

## Stack

- Go 1.27
- [OTLP proto](https://github.com/open-telemetry/opentelemetry-proto) + [gRPC](https://google.golang.org/grpc) — OTLP receiver
- [franz-go](https://github.com/twmb/franz-go) — Kafka producer

## Endpoints

| Transport | Address  | Paths                                   |
|-----------|----------|-----------------------------------------|
| gRPC      | `:4317`  | OTLP `Export` (traces / logs / metrics) |
| HTTP      | `:4318`  | `/v1/traces`, `/v1/logs`, `/v1/metrics` |

HTTP accepts both `application/x-protobuf` and `application/json` bodies.

## Delivery policy

nexus acts as a prefilter in front of StarRocks. Transform and produce errors are
logged and skipped rather than propagated — a malformed or unexpected record is
dropped here so it can never break downstream Routine Load ingestion. OTLP clients
always receive a success response; failures surface only in nexus logs.

## Run

```bash
cp .env.example .env
go run ./cmd
```

## Environment

| Variable              | Description                                           |
|-----------------------|-------------------------------------------------------|
| `KAFKA_BROKERS`       | Comma-separated broker list                           |
| `OUTPUT_TOPIC_PREFIX` | Flat output topic prefix (default: `otel.flat`)       |
| `OTLP_GRPC_ADDR`      | gRPC listen address (default: `:4317`)                |
| `OTLP_HTTP_ADDR`      | HTTP listen address (default: `:4318`)                |
| `LOG_LEVEL`           | `debug` / `info` / `warn` / `error` (default: `info`) |
