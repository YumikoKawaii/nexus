CREATE TABLE IF NOT EXISTS otel_metrics_exponential_histogram
(
    `ServiceName`            VARCHAR(256),
    `MetricName`             VARCHAR(256),
    `TimeUnix`               DATETIME NOT NULL,
    `ResourceAttributes`     JSON,
    `ResourceSchemaUrl`      VARCHAR(255),
    `ScopeName`              VARCHAR(256),
    `ScopeVersion`           VARCHAR(64),
    `ScopeAttributes`        JSON,
    `ScopeDroppedAttrCount`  BIGINT,
    `ScopeSchemaUrl`         VARCHAR(255),
    `MetricDescription`      VARCHAR(255),
    `MetricUnit`             VARCHAR(64),
    `Attributes`             JSON,
    `StartTimeUnix`          DATETIME,
    `Count`                  BIGINT,
    `Sum`                    DOUBLE,
    `Scale`                  INT,
    `ZeroCount`              BIGINT,
    `PositiveOffset`         INT,
    `PositiveBucketCounts`   JSON,
    `NegativeOffset`         INT,
    `NegativeBucketCounts`   JSON,
    `Exemplars`              JSON,
    `Flags`                  INT,
    `Min`                    DOUBLE,
    `Max`                    DOUBLE,
    `AggregationTemporality` INT
) DUPLICATE KEY(`ServiceName`, `MetricName`, `TimeUnix`)
PARTITION BY date_trunc('DAY', TimeUnix)
DISTRIBUTED BY HASH (ServiceName, MetricName)
PROPERTIES (
    "replication_num" = "1",
    "enable_persistent_index" = "true"
);

CREATE ROUTINE LOAD otel_metrics_exponential_histogram_load ON otel_metrics_exponential_histogram
COLUMNS (`ServiceName`, `MetricName`, `TimeUnix`, `ResourceAttributes`, `ResourceSchemaUrl`, `ScopeName`, `ScopeVersion`, `ScopeAttributes`, `ScopeDroppedAttrCount`, `ScopeSchemaUrl`, `MetricDescription`, `MetricUnit`, `Attributes`, `StartTimeUnix`, `Count`, `Sum`, `Scale`, `ZeroCount`, `PositiveOffset`, `PositiveBucketCounts`, `NegativeOffset`, `NegativeBucketCounts`, `Exemplars`, `Flags`, `Min`, `Max`, `AggregationTemporality`)
PROPERTIES (
    "desired_concurrent_number" = "3",
    "max_batch_interval" = "10",
    "max_error_number" = "1000",
    "strict_mode" = "false",
    "format" = "json",
    "jsonpaths" = "[\"$.ServiceName\",\"$.MetricName\",\"$.TimeUnix\",\"$.ResourceAttributes\",\"$.ResourceSchemaUrl\",\"$.ScopeName\",\"$.ScopeVersion\",\"$.ScopeAttributes\",\"$.ScopeDroppedAttrCount\",\"$.ScopeSchemaUrl\",\"$.MetricDescription\",\"$.MetricUnit\",\"$.Attributes\",\"$.StartTimeUnix\",\"$.Count\",\"$.Sum\",\"$.Scale\",\"$.ZeroCount\",\"$.PositiveOffset\",\"$.PositiveBucketCounts\",\"$.NegativeOffset\",\"$.NegativeBucketCounts\",\"$.Exemplars\",\"$.Flags\",\"$.Min\",\"$.Max\",\"$.AggregationTemporality\"]"
)
FROM KAFKA (
    "kafka_broker_list" = "${KAFKA_BROKERS}",
    "kafka_topic" = "otel.flat.metrics.exponential_histogram",
    "property.kafka_default_offsets" = "OFFSET_END"
);
