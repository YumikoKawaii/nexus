package transform

// OTLP JSON wire format structs for unmarshalling raw Kafka messages.
// Field names match the protobuf-JSON encoding from otelcol-contrib kafka exporter.

type OTLPTracePayload struct {
	ResourceSpans []ResourceSpan `json:"resourceSpans"`
}

type ResourceSpan struct {
	Resource   OTLPResource `json:"resource"`
	SchemaUrl  string       `json:"schemaUrl"`
	ScopeSpans []ScopeSpan  `json:"scopeSpans"`
}

type ScopeSpan struct {
	Scope     OTLPScope `json:"scope"`
	SchemaUrl string    `json:"schemaUrl"`
	Spans     []Span    `json:"spans"`
}

type Span struct {
	TraceId           string       `json:"traceId"`
	SpanId            string       `json:"spanId"`
	ParentSpanId      string       `json:"parentSpanId"`
	Name              string       `json:"name"`
	Kind              int32        `json:"kind"`
	StartTimeUnixNano string       `json:"startTimeUnixNano"`
	EndTimeUnixNano   string       `json:"endTimeUnixNano"`
	Attributes        []OTLPKV     `json:"attributes"`
	Status            SpanStatus   `json:"status"`
	Events            []SpanEvent  `json:"events"`
	Links             []SpanLink   `json:"links"`
}

type SpanStatus struct {
	Code    int32  `json:"code"`
	Message string `json:"message"`
}

type SpanEvent struct {
	TimeUnixNano string   `json:"timeUnixNano"`
	Name         string   `json:"name"`
	Attributes   []OTLPKV `json:"attributes"`
}

type SpanLink struct {
	TraceId    string   `json:"traceId"`
	SpanId     string   `json:"spanId"`
	Attributes []OTLPKV `json:"attributes"`
}

type OTLPLogsPayload struct {
	ResourceLogs []ResourceLog `json:"resourceLogs"`
}

type ResourceLog struct {
	Resource  OTLPResource `json:"resource"`
	SchemaUrl string       `json:"schemaUrl"`
	ScopeLogs []ScopeLog   `json:"scopeLogs"`
}

type ScopeLog struct {
	Scope      OTLPScope   `json:"scope"`
	SchemaUrl  string      `json:"schemaUrl"`
	LogRecords []LogRecord `json:"logRecords"`
}

type LogRecord struct {
	TimeUnixNano         string   `json:"timeUnixNano"`
	SeverityNumber       int32    `json:"severityNumber"`
	SeverityText         string   `json:"severityText"`
	Body                 OTLPAny  `json:"body"`
	Attributes           []OTLPKv `json:"attributes"`
	TraceId              string   `json:"traceId"`
	SpanId               string   `json:"spanId"`
}

type OTLPMetricsPayload struct {
	ResourceMetrics []ResourceMetric `json:"resourceMetrics"`
}

type ResourceMetric struct {
	Resource     OTLPResource  `json:"resource"`
	SchemaUrl    string        `json:"schemaUrl"`
	ScopeMetrics []ScopeMetric `json:"scopeMetrics"`
}

type ScopeMetric struct {
	Scope     OTLPScope `json:"scope"`
	SchemaUrl string    `json:"schemaUrl"`
	Metrics   []Metric  `json:"metrics"`
}

type Metric struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Unit        string `json:"unit"`
	// only one of these will be set
	Gauge                *GaugeData                `json:"gauge,omitempty"`
	Sum                  *SumData                  `json:"sum,omitempty"`
	Summary              *SummaryData              `json:"summary,omitempty"`
	Histogram            *HistogramData            `json:"histogram,omitempty"`
	ExponentialHistogram *ExponentialHistogramData `json:"exponentialHistogram,omitempty"`
}

type GaugeData struct {
	DataPoints []NumberDataPoint `json:"dataPoints"`
}

type SumData struct {
	DataPoints             []NumberDataPoint `json:"dataPoints"`
	AggregationTemporality int32             `json:"aggregationTemporality"`
	IsMonotonic            bool              `json:"isMonotonic"`
}

type SummaryData struct {
	DataPoints []SummaryDataPoint `json:"dataPoints"`
}

type HistogramData struct {
	DataPoints             []HistogramDataPoint `json:"dataPoints"`
	AggregationTemporality int32                `json:"aggregationTemporality"`
}

type ExponentialHistogramData struct {
	DataPoints             []ExponentialHistogramDataPoint `json:"dataPoints"`
	AggregationTemporality int32                           `json:"aggregationTemporality"`
}

type NumberDataPoint struct {
	Attributes        []OTLPKv `json:"attributes"`
	StartTimeUnixNano string   `json:"startTimeUnixNano"`
	TimeUnixNano      string   `json:"timeUnixNano"`
	AsDouble          *float64 `json:"asDouble,omitempty"`
	AsInt             *int64   `json:"asInt,omitempty"`
	Flags             int32    `json:"flags"`
}

type SummaryDataPoint struct {
	Attributes        []OTLPKv         `json:"attributes"`
	StartTimeUnixNano string           `json:"startTimeUnixNano"`
	TimeUnixNano      string           `json:"timeUnixNano"`
	Count             uint64           `json:"count"`
	Sum               float64          `json:"sum"`
	QuantileValues    []QuantileValue  `json:"quantileValues"`
	Flags             int32            `json:"flags"`
}

type QuantileValue struct {
	Quantile float64 `json:"quantile"`
	Value    float64 `json:"value"`
}

type HistogramDataPoint struct {
	Attributes        []OTLPKv  `json:"attributes"`
	StartTimeUnixNano string    `json:"startTimeUnixNano"`
	TimeUnixNano      string    `json:"timeUnixNano"`
	Count             uint64    `json:"count"`
	Sum               *float64  `json:"sum,omitempty"`
	BucketCounts      []uint64  `json:"bucketCounts"`
	ExplicitBounds    []float64 `json:"explicitBounds"`
	Exemplars         []Exemplar `json:"exemplars"`
	Flags             int32     `json:"flags"`
	Min               *float64  `json:"min,omitempty"`
	Max               *float64  `json:"max,omitempty"`
}

type ExponentialHistogramDataPoint struct {
	Attributes        []OTLPKv    `json:"attributes"`
	StartTimeUnixNano string      `json:"startTimeUnixNano"`
	TimeUnixNano      string      `json:"timeUnixNano"`
	Count             uint64      `json:"count"`
	Sum               *float64    `json:"sum,omitempty"`
	Scale             int32       `json:"scale"`
	ZeroCount         uint64      `json:"zeroCount"`
	Positive          BucketBands `json:"positive"`
	Negative          BucketBands `json:"negative"`
	Exemplars         []Exemplar  `json:"exemplars"`
	Flags             int32       `json:"flags"`
	Min               *float64    `json:"min,omitempty"`
	Max               *float64    `json:"max,omitempty"`
}

type BucketBands struct {
	Offset       int32    `json:"offset"`
	BucketCounts []uint64 `json:"bucketCounts"`
}

type Exemplar struct {
	TimeUnixNano string   `json:"timeUnixNano"`
	AsDouble     *float64 `json:"asDouble,omitempty"`
	AsInt        *int64   `json:"asInt,omitempty"`
	TraceId      string   `json:"traceId"`
	SpanId       string   `json:"spanId"`
}

type OTLPResource struct {
	Attributes []OTLPKv `json:"attributes"`
	SchemaUrl  string   `json:"schemaUrl"`
}

type OTLPScope struct {
	Name                   string   `json:"name"`
	Version                string   `json:"version"`
	Attributes             []OTLPKv `json:"attributes"`
	DroppedAttributesCount int64    `json:"droppedAttributesCount"`
}

type OTLPKv struct {
	Key   string  `json:"key"`
	Value OTLPAny `json:"value"`
}

// OTLPKV is an alias kept for compatibility within this package.
type OTLPKV = OTLPKv

type OTLPAny struct {
	StringValue *string  `json:"stringValue,omitempty"`
	IntValue    *int64   `json:"intValue,omitempty"`
	DoubleValue *float64 `json:"doubleValue,omitempty"`
	BoolValue   *bool    `json:"boolValue,omitempty"`
}
