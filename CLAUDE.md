# CLAUDE.md

## What Nexus is

A Kafka transformer service that consumes raw OTLP JSON batches from `otel.traces`, `otel.metrics`, and `otel.logs` topics, explodes the nested OTLP protobuf-JSON structure into flat rows, remaps field names and types to match the StarRocks `otel` schema, and produces to flat topics consumed by StarRocks Routine Load jobs.


Input topics: `otel.traces`, `otel.metrics`, `otel.logs`
Output topics: `otel.flat.traces`, `otel.flat.metrics.*`, `otel.flat.logs`

## Repo layout

```
cmd/nexus/        — main binary
internal/
  consumer/       — Sarama Kafka consumer group
  producer/       — Sarama Kafka producer
  transform/      — OTLP JSON → flat row mappers (traces, metrics, logs)
  config/         — env-var loading
```

## Build

```bash
go build ./...
go run ./cmd/nexus
```

## Stack

- Go 1.25
- Sarama — Kafka consumer/producer
- No frameworks; plain stdlib + Sarama

## Conventions

- **Commit messages:** `feat`, `fix`, `chore`, `refactor`, `test`, `docs` prefixes. Keep the required `Co-Authored-By` trailer.
- **Git workflow:** only commit when explicitly asked. After committing, always `git push` immediately.
- **Comments:** only comment when something is non-obvious or tricky. Never comment trivial code.
- **Go — OOP best practices:**
  - Small, focused interfaces — 1–3 methods. Define in the consumer package, not the producer.
  - Composition over embedding chains.
  - Constructor functions — every exported type gets a `New*` function.
  - Encapsulation — unexported fields by default.
  - Pointer receivers for mutating or large structs; value receivers for small read-only types. Consistent per type.
  - Dependency injection through constructors — no package-level globals or `init()` side effects. Wire at `main()`.
- **Go idioms:**
  - Handle errors explicitly — no `_` discards on error returns unless truly irrelevant.
  - Prefer `errors.Is` / `errors.As` over string matching.
  - Use `context.Context` as the first parameter on any function that does I/O or can block.
  - Return early on errors — avoid deeply nested happy-path code.
  - Keep goroutines owned — every goroutine spawned must have a clear owner responsible for its lifetime and cancellation.
  - Use `defer` for cleanup at the call site, not inside helpers.
  - Prefer table-driven tests.
