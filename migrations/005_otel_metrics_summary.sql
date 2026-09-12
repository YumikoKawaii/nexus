CREATE TABLE IF NOT EXISTS otel_metrics_summary
(
    `ServiceName`           VARCHAR(256),
    `MetricName`            VARCHAR(256),
    `TimeUnix`              DATETIME NOT NULL,
    `ResourceAttributes`    JSON,
    `ResourceSchemaUrl`     VARCHAR(255),
    `ScopeName`             VARCHAR(256),
    `ScopeVersion`          VARCHAR(64),
    `ScopeAttributes`       JSON,
    `ScopeDroppedAttrCount` BIGINT,
    `ScopeSchemaUrl`        VARCHAR(255),
    `MetricDescription`     VARCHAR(255),
    `MetricUnit`            VARCHAR(64),
    `Attributes`            JSON,
    `StartTimeUnix`         DATETIME,
    `Count`                 BIGINT,
    `Sum`                   DOUBLE,
    `ValueAtQuantiles`      JSON,
    `Flags`                 INT
) DUPLICATE KEY(`ServiceName`, `MetricName`, `TimeUnix`)
PARTITION BY date_trunc('DAY', TimeUnix)
DISTRIBUTED BY HASH (ServiceName, MetricName)
PROPERTIES (
    "replication_num" = "1",
    "enable_persistent_index" = "true"
);

CREATE ROUTINE LOAD otel_metrics_summary_load ON otel_metrics_summary
COLUMNS (`ServiceName`, `MetricName`, `TimeUnix`, `ResourceAttributes`, `ResourceSchemaUrl`, `ScopeName`, `ScopeVersion`, `ScopeAttributes`, `ScopeDroppedAttrCount`, `ScopeSchemaUrl`, `MetricDescription`, `MetricUnit`, `Attributes`, `StartTimeUnix`, `Count`, `Sum`, `ValueAtQuantiles`, `Flags`)
PROPERTIES (
    "desired_concurrent_number" = "3",
    "max_batch_interval" = "10",
    "max_error_number" = "1000",
    "strict_mode" = "false",
    "format" = "json",
    "jsonpaths" = "[\"$.ServiceName\",\"$.MetricName\",\"$.TimeUnix\",\"$.ResourceAttributes\",\"$.ResourceSchemaUrl\",\"$.ScopeName\",\"$.ScopeVersion\",\"$.ScopeAttributes\",\"$.ScopeDroppedAttrCount\",\"$.ScopeSchemaUrl\",\"$.MetricDescription\",\"$.MetricUnit\",\"$.Attributes\",\"$.StartTimeUnix\",\"$.Count\",\"$.Sum\",\"$.ValueAtQuantiles\",\"$.Flags\"]"
)
FROM KAFKA (
    "kafka_broker_list" = "${KAFKA_BROKERS}",
    "kafka_topic" = "otel.flat.metrics.summary",
    "property.kafka_default_offsets" = "OFFSET_END"
);
