CREATE TABLE IF NOT EXISTS otel_logs
(
    `ServiceName`           VARCHAR(256),
    `Timestamp`             DATETIME NOT NULL,
    `TraceId`               VARCHAR(32),
    `SpanId`                VARCHAR(16),
    `SeverityText`          VARCHAR(32),
    `SeverityNumber`        INT,
    `Body`                  VARCHAR(65533),
    `ScopeName`             VARCHAR(256),
    `ServiceVersion`        VARCHAR(64),
    `DeploymentEnvironment` VARCHAR(64),
    `ResourceAttributes`    JSON,
    `LogAttributes`         JSON,
    `EventName`             VARCHAR(256)
)
DUPLICATE KEY(`ServiceName`, `Timestamp`, `TraceId`, `SpanId`)
PARTITION BY date_trunc('DAY', Timestamp)
DISTRIBUTED BY HASH (ServiceName)
PROPERTIES (
    "replication_num" = "1",
    "enable_persistent_index" = "true"
);

CREATE ROUTINE LOAD otel_logs_load ON otel_logs
COLUMNS (`ServiceName`, `Timestamp`, `TraceId`, `SpanId`, `SeverityText`, `SeverityNumber`, `Body`, `ScopeName`, `ServiceVersion`, `DeploymentEnvironment`, `ResourceAttributes`, `LogAttributes`, `EventName`)
PROPERTIES (
    "desired_concurrent_number" = "3",
    "max_batch_interval" = "10",
    "max_error_number" = "1000",
    "strict_mode" = "false",
    "format" = "json",
    "jsonpaths" = "[\"$.ServiceName\",\"$.Timestamp\",\"$.TraceId\",\"$.SpanId\",\"$.SeverityText\",\"$.SeverityNumber\",\"$.Body\",\"$.ScopeName\",\"$.ServiceVersion\",\"$.DeploymentEnvironment\",\"$.ResourceAttributes\",\"$.LogAttributes\",\"$.EventName\"]"
)
FROM KAFKA (
    "kafka_broker_list" = "${KAFKA_BROKERS}",
    "kafka_topic" = "otel.flat.logs",
    "property.kafka_default_offsets" = "OFFSET_END"
);
