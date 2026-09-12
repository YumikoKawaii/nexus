CREATE TABLE IF NOT EXISTS otel_traces
(
    `ServiceName`        VARCHAR(256),
    `SpanName`           VARCHAR(256),
    `Timestamp`          DATETIME NOT NULL,
    `TraceId`            VARCHAR(32),
    `SpanId`             VARCHAR(16),
    `ParentSpanId`       VARCHAR(16),
    `TraceState`         VARCHAR(255),
    `SpanKind`           VARCHAR(32),
    `ResourceAttributes` JSON,
    `ScopeName`          VARCHAR(256),
    `ScopeVersion`       VARCHAR(64),
    `SpanAttributes`     JSON,
    `Duration`           BIGINT,
    `StatusCode`         VARCHAR(32),
    `StatusMessage`      VARCHAR(255),
    `Events`             JSON,
    `Links`              JSON
) DUPLICATE KEY(`ServiceName`, `SpanName`, `Timestamp`, `TraceId`, `SpanId`)
PARTITION BY date_trunc('DAY', Timestamp)
DISTRIBUTED BY HASH (TraceId)
PROPERTIES (
    "replication_num" = "1",
    "enable_persistent_index" = "true"
);

CREATE ROUTINE LOAD otel_traces_load ON otel_traces
COLUMNS (`ServiceName`, `SpanName`, `Timestamp`, `TraceId`, `SpanId`, `ParentSpanId`, `TraceState`, `SpanKind`, `ResourceAttributes`, `ScopeName`, `ScopeVersion`, `SpanAttributes`, `Duration`, `StatusCode`, `StatusMessage`, `Events`, `Links`)
PROPERTIES (
    "desired_concurrent_number" = "3",
    "max_batch_interval" = "10",
    "max_error_number" = "1000",
    "strict_mode" = "false",
    "format" = "json",
    "jsonpaths" = "[\"$.ServiceName\",\"$.SpanName\",\"$.Timestamp\",\"$.TraceId\",\"$.SpanId\",\"$.ParentSpanId\",\"$.TraceState\",\"$.SpanKind\",\"$.ResourceAttributes\",\"$.ScopeName\",\"$.ScopeVersion\",\"$.SpanAttributes\",\"$.Duration\",\"$.StatusCode\",\"$.StatusMessage\",\"$.Events\",\"$.Links\"]"
)
FROM KAFKA (
    "kafka_broker_list" = "${KAFKA_BROKERS}",
    "kafka_topic" = "otel.flat.traces",
    "property.kafka_default_offsets" = "OFFSET_END"
);
