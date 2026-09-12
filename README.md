<div align="center">

```
    ███╗   ██╗███████╗██╗  ██╗██╗   ██╗███████╗
    ████╗  ██║██╔════╝╚██╗██╔╝██║   ██║██╔════╝
    ██╔██╗ ██║█████╗   ╚███╔╝ ██║   ██║███████╗
    ██║╚██╗██║██╔══╝   ██╔██╗ ██║   ██║╚════██║
    ██║ ╚████║███████╗██╔╝ ██╗╚██████╔╝███████║
    ╚═╝  ╚═══╝╚══════╝╚═╝  ╚═╝ ╚═════╝ ╚══════╝
```

### *Where All Telemetry Converges*

---

**An OTLP Endpoint That Flattens the World for StarRocks**

[![Author](https://img.shields.io/badge/Author-Yumiko%20Sturluson-ff69b4?style=for-the-badge)](https://github.com/yumikokawaii)
[![License](https://img.shields.io/badge/License-Private-9370DB?style=for-the-badge)]()
[![Go](https://img.shields.io/badge/Go-1.27-00ADD8?style=for-the-badge&logo=go&logoColor=white)]()
[![Protocol](https://img.shields.io/badge/OTLP-425CC7?style=for-the-badge&logo=opentelemetry&logoColor=white)]()
[![Backend](https://img.shields.io/badge/StarRocks-1E88E5?style=for-the-badge)]()

</div>

---

## About

> *"A nexus is not a place, but a moment — where many roads become one."*

Welcome to **nexus**!

Telemetry arrives from everywhere at once — traces, metrics, logs, from every service, in the deeply nested shape OTLP
was born in. StarRocks wants none of that nesting; it wants flat rows. `nexus` stands at the meeting point: it speaks
the standard **OTLP** protocol on the wire, explodes every nested batch into flat records, and produces them to Kafka
for StarRocks Routine Load.

```
  otel sdk / collector ──OTLP (gRPC :4317 / HTTP :4318)──▶ nexus ──otel.flat.* (Kafka)──▶ StarRocks Routine Load
```

Transformation is a prefilter, not a gamble. A malformed or unexpected record is **logged and dropped at the door** —
never propagated — so nothing downstream can choke on it. OTLP clients always get a success response; every failure
lives in the nexus logs alone.

## Architecture

nexus is a receiver in front of a producer, with a pure transform in between:

- **Receive** — dual OTLP frontends (`internal/receiver`) accept gRPC on `:4317` and HTTP on `:4318`
  (`/v1/{traces,logs,metrics}`, protobuf or JSON). Both hand off to one shared `Service`.
- **Transform** — `internal/transform` explodes each OTLP `ResourceSpans` / `ResourceLogs` / `ResourceMetrics`
  into flat rows matching the StarRocks schema, fanning metrics out by type.
- **Produce** — `internal/producer` (franz-go) writes one JSON record per row to the matching `otel.flat.*` topic.

```
  receiver ──proto──▶ transform.Traces/Logs/Metrics ──rows──▶ producer ──▶ otel.flat.*
```

Each signal lands on its own topic, ready for a StarRocks Routine Load:

| Signal                   | Output topic                              |
|--------------------------|-------------------------------------------|
| Traces                   | `otel.flat.traces`                        |
| Logs                     | `otel.flat.logs`                          |
| Metrics — gauge          | `otel.flat.metrics.gauge`                 |
| Metrics — sum            | `otel.flat.metrics.sum`                   |
| Metrics — summary        | `otel.flat.metrics.summary`               |
| Metrics — histogram      | `otel.flat.metrics.histogram`             |
| Metrics — exp. histogram | `otel.flat.metrics.exponential_histogram` |

## Configuration

All configuration is via environment variables:

| Variable                   | Default     | Purpose                             |
|----------------------------|-------------|-------------------------------------|
| `KAFKA_BROKERS`            | —           | Comma-separated broker list         |
| `OUTPUT_TOPIC_PREFIX`      | `otel.flat` | Flat output topic prefix            |
| `OTLP_GRPC_ADDR`           | `:4317`     | OTLP gRPC listen address            |
| `OTLP_HTTP_ADDR`           | `:4318`     | OTLP HTTP listen address            |
| `LOG_LEVEL`                | `info`      | `debug` / `info` / `warn` / `error` |
| `PRODUCER_MODE`            | `async`     | `sync` or `async`                   |
| `PRODUCER_ACKS`            | `local`     | `none` / `local` / `all`            |
| `PRODUCER_RETRY_MAX`       | `3`         | Max produce retries                 |
| `PRODUCER_RETRY_BACKOFF`   | `100ms`     | Backoff between retries             |
| `PRODUCER_FLUSH_MESSAGES`  | `1000`      | Max buffered records                |
| `PRODUCER_FLUSH_BYTES`     | `1048576`   | Max batch bytes                     |
| `PRODUCER_FLUSH_FREQUENCY` | `2s`        | Producer linger                     |

## Author

<div align="center">

**~ Yumiko Sturluson ~**

*Software Engineer*

*コードよ、わがまほうとなれ*

</div>

---

<div align="center">

*~ Built with cuteness ~*

</div>
