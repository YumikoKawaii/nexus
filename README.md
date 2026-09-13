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

**An OTLP Endpoint That Flattens Telemetry into Kafka**

[![Author](https://img.shields.io/badge/Author-Yumiko%20Sturluson-ff69b4?style=for-the-badge)](https://github.com/yumikokawaii)
[![License](https://img.shields.io/badge/License-Private-9370DB?style=for-the-badge)]()
[![Go](https://img.shields.io/badge/Go-1.27-00ADD8?style=for-the-badge&logo=go&logoColor=white)]()
[![Protocol](https://img.shields.io/badge/OTLP-425CC7?style=for-the-badge&logo=opentelemetry&logoColor=white)]()
[![Transport](https://img.shields.io/badge/Kafka-231F20?style=for-the-badge&logo=apachekafka&logoColor=white)]()

</div>

---

## About

> *"A nexus is not a place, but a moment — where many roads become one."*

Welcome to **nexus**!

Telemetry arrives from everywhere at once — traces, metrics, logs, from every service, in the deeply nested shape OTLP
was born in. Downstreams want flat rows, not nesting. `nexus` stands at the meeting point: it speaks the standard
**OTLP** protocol on the wire, explodes every nested batch into flat records, and produces one JSON message per row to
Kafka — ready for whatever consumes it next.

```
  otel sdk / collector ──OTLP (gRPC + HTTP on :4317)──▶ nexus ──otel.flat.* (Kafka)──▶ downstream
```

## Architecture

nexus is a receiver in front of a producer, with a pure transform in between:

- **Receive** — one OTLP frontend (`internal/receiver`) on `:4317`, powered by
  [Vanguard](https://github.com/connectrpc/vanguard-go): a single Connect-RPC service handler
  transcodes gRPC, gRPC-Web, Connect, and REST (`/v1/{traces,logs,metrics}`, protobuf or JSON,
  with standard compression) onto the same `Export` implementations — no hand-written parsing
- **Transform** — explodes each OTLP `ResourceSpans` / `ResourceLogs` / `ResourceMetrics` into flat rows, fanning
  metrics out by type.
- **Produce** — writes one JSON record per row to the matching `otel.flat.*` topic.

```
  receiver ──proto──▶ transform.Traces/Logs/Metrics ──rows──▶ producer ──▶ otel.flat.*
```

Each signal lands on its own topic, addressed by suffix:

| Signal                   | Topic suffix                    | Output topic                              |
|--------------------------|---------------------------------|-------------------------------------------|
| Traces                   | `traces`                        | `otel.flat.traces`                        |
| Logs                     | `logs`                          | `otel.flat.logs`                          |
| Metrics — gauge          | `metrics.gauge`                 | `otel.flat.metrics.gauge`                 |
| Metrics — sum            | `metrics.sum`                   | `otel.flat.metrics.sum`                   |
| Metrics — summary        | `metrics.summary`               | `otel.flat.metrics.summary`               |
| Metrics — histogram      | `metrics.histogram`             | `otel.flat.metrics.histogram`             |
| Metrics — exp. histogram | `metrics.exponential_histogram` | `otel.flat.metrics.exponential_histogram` |

## Configuration

nexus reads a single YAML file, given by `--config` (default `config.yaml`):

```yaml
kafkaBrokers:
  - localhost:9092
outputTopicPrefix: otel.flat   # topics are <prefix>.<suffix>
logLevel: info                 # debug | info | warn | error

otlp:
  addr: ":4317"               # single port: gRPC, gRPC-Web, Connect, REST
  maxRecvMsgSizeMiB: 4         # max decoded OTLP request size

topics:
  # Allowlist of topic suffixes to publish. Empty (or omitted) = all.
  # e.g. [traces, metrics.gauge] publishes only those two.
  enabled: [ ]

producer:
  mode: async                  # sync | async
  acks: local                  # none | local | all
  retryMax: 3
  retryBackoff: 100ms
  flushMessages: 1000
  flushBytes: 1048576
  flushFrequency: 2s
```

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
