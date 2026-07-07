package transform

type FlatTrace struct {
	ServiceName        string `json:"ServiceName"`
	SpanName           string `json:"SpanName"`
	Timestamp          string `json:"Timestamp"`
	TraceId            string `json:"TraceId"`
	SpanId             string `json:"SpanId"`
	ParentSpanId       string `json:"ParentSpanId"`
	TraceState         string `json:"TraceState"`
	SpanKind           string `json:"SpanKind"`
	ResourceAttributes string `json:"ResourceAttributes"`
	ScopeName          string `json:"ScopeName"`
	ScopeVersion       string `json:"ScopeVersion"`
	SpanAttributes     string `json:"SpanAttributes"`
	Duration           int64  `json:"Duration"`
	StatusCode         string `json:"StatusCode"`
	StatusMessage      string `json:"StatusMessage"`
	Events             string `json:"Events"`
	Links              string `json:"Links"`
}

type FlatLog struct {
	ServiceName           string `json:"ServiceName"`
	Timestamp             string `json:"Timestamp"`
	TraceId               string `json:"TraceId"`
	SpanId                string `json:"SpanId"`
	SeverityText          string `json:"SeverityText"`
	SeverityNumber        int32  `json:"SeverityNumber"`
	Body                  string `json:"Body"`
	ScopeName             string `json:"ScopeName"`
	ServiceVersion        string `json:"ServiceVersion"`
	DeploymentEnvironment string `json:"DeploymentEnvironment"`
	ResourceAttributes    string `json:"ResourceAttributes"`
	LogAttributes         string `json:"LogAttributes"`
	EventName             string `json:"EventName"`
}

type FlatGauge struct {
	ServiceName           string `json:"ServiceName"`
	MetricName            string `json:"MetricName"`
	TimeUnix              string `json:"TimeUnix"`
	ResourceAttributes    string `json:"ResourceAttributes"`
	ResourceSchemaUrl     string `json:"ResourceSchemaUrl"`
	ScopeName             string `json:"ScopeName"`
	ScopeVersion          string `json:"ScopeVersion"`
	ScopeAttributes       string `json:"ScopeAttributes"`
	ScopeDroppedAttrCount int64  `json:"ScopeDroppedAttrCount"`
	ScopeSchemaUrl        string `json:"ScopeSchemaUrl"`
	MetricDescription     string `json:"MetricDescription"`
	MetricUnit            string `json:"MetricUnit"`
	Attributes            string `json:"Attributes"`
	StartTimeUnix         string `json:"StartTimeUnix"`
	Value                 string `json:"Value"`
	Flags                 int32  `json:"Flags"`
	Exemplars             string `json:"Exemplars"`
}

type FlatSum struct {
	FlatGauge
	AggregationTemporality int32 `json:"AggregationTemporality"`
	IsMonotonic            bool  `json:"IsMonotonic"`
}

type FlatSummary struct {
	ServiceName           string  `json:"ServiceName"`
	MetricName            string  `json:"MetricName"`
	TimeUnix              string  `json:"TimeUnix"`
	ResourceAttributes    string  `json:"ResourceAttributes"`
	ResourceSchemaUrl     string  `json:"ResourceSchemaUrl"`
	ScopeName             string  `json:"ScopeName"`
	ScopeVersion          string  `json:"ScopeVersion"`
	ScopeAttributes       string  `json:"ScopeAttributes"`
	ScopeDroppedAttrCount int64   `json:"ScopeDroppedAttrCount"`
	ScopeSchemaUrl        string  `json:"ScopeSchemaUrl"`
	MetricDescription     string  `json:"MetricDescription"`
	MetricUnit            string  `json:"MetricUnit"`
	Attributes            string  `json:"Attributes"`
	StartTimeUnix         string  `json:"StartTimeUnix"`
	Count                 uint64  `json:"Count"`
	Sum                   float64 `json:"Sum"`
	ValueAtQuantiles      string  `json:"ValueAtQuantiles"`
	Flags                 int32   `json:"Flags"`
}

type FlatHistogram struct {
	ServiceName            string  `json:"ServiceName"`
	MetricName             string  `json:"MetricName"`
	TimeUnix               string  `json:"TimeUnix"`
	ResourceAttributes     string  `json:"ResourceAttributes"`
	ResourceSchemaUrl      string  `json:"ResourceSchemaUrl"`
	ScopeName              string  `json:"ScopeName"`
	ScopeVersion           string  `json:"ScopeVersion"`
	ScopeAttributes        string  `json:"ScopeAttributes"`
	ScopeDroppedAttrCount  int64   `json:"ScopeDroppedAttrCount"`
	ScopeSchemaUrl         string  `json:"ScopeSchemaUrl"`
	MetricDescription      string  `json:"MetricDescription"`
	MetricUnit             string  `json:"MetricUnit"`
	Attributes             string  `json:"Attributes"`
	StartTimeUnix          string  `json:"StartTimeUnix"`
	Count                  uint64  `json:"Count"`
	Sum                    float64 `json:"Sum"`
	BucketCounts           string  `json:"BucketCounts"`
	ExplicitBounds         string  `json:"ExplicitBounds"`
	Exemplars              string  `json:"Exemplars"`
	Flags                  int32   `json:"Flags"`
	Min                    float64 `json:"Min"`
	Max                    float64 `json:"Max"`
	AggregationTemporality int32   `json:"AggregationTemporality"`
}

type FlatExponentialHistogram struct {
	ServiceName            string  `json:"ServiceName"`
	MetricName             string  `json:"MetricName"`
	TimeUnix               string  `json:"TimeUnix"`
	ResourceAttributes     string  `json:"ResourceAttributes"`
	ResourceSchemaUrl      string  `json:"ResourceSchemaUrl"`
	ScopeName              string  `json:"ScopeName"`
	ScopeVersion           string  `json:"ScopeVersion"`
	ScopeAttributes        string  `json:"ScopeAttributes"`
	ScopeDroppedAttrCount  int64   `json:"ScopeDroppedAttrCount"`
	ScopeSchemaUrl         string  `json:"ScopeSchemaUrl"`
	MetricDescription      string  `json:"MetricDescription"`
	MetricUnit             string  `json:"MetricUnit"`
	Attributes             string  `json:"Attributes"`
	StartTimeUnix          string  `json:"StartTimeUnix"`
	Count                  uint64  `json:"Count"`
	Sum                    float64 `json:"Sum"`
	Scale                  int32   `json:"Scale"`
	ZeroCount              uint64  `json:"ZeroCount"`
	PositiveOffset         int32   `json:"PositiveOffset"`
	PositiveBucketCounts   string  `json:"PositiveBucketCounts"`
	NegativeOffset         int32   `json:"NegativeOffset"`
	NegativeBucketCounts   string  `json:"NegativeBucketCounts"`
	Exemplars              string  `json:"Exemplars"`
	Flags                  int32   `json:"Flags"`
	Min                    float64 `json:"Min"`
	Max                    float64 `json:"Max"`
	AggregationTemporality int32   `json:"AggregationTemporality"`
}
